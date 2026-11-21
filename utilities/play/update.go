package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ncruces/zenity"
	"wilderzone.org/tai-utilities/internal/frontend"
)

// UPDATE_CHECK_TIMEOUT is how long the update check can hold up the game. A slow or missing
// connection shouldn't stop anyone playing.
const UPDATE_CHECK_TIMEOUT = 10 * time.Second

// offerUpdate checks for an update before the game starts, and offers to install it. The user
// can always ignore it and play anyway. It returns false if the user chose not to play after
// all, which they can only do once an update has failed.
//
// Nothing about the check stands in the way of playing: when it can't be done, or doesn't
// finish in time, the game starts as it would have without it, and the reason is logged.
func offerUpdate(log io.Writer) bool {
	if !zenity.IsAvailable() {
		fmt.Fprintln(log, "Skipped the update check: there's no way to show dialogs on this system.")
		return true
	}
	found, err := frontend.Locate()
	if err != nil {
		fmt.Fprintf(log, "Skipped the update check: %s\n", err.Error())
		return true
	}

	update, err := checkForUpdates(found)
	if err != nil {
		fmt.Fprintf(log, "Skipped the update check: %s\n", err.Error())
		return true
	}
	if update == "" {
		return true
	}

	err = zenity.Question(update + "\n\nWould you like to update " + frontend.PRODUCT_NAME + " before playing?",
		zenity.Title(frontend.UPDATE.Title),
		zenity.OKLabel("Update"),
		zenity.CancelLabel("Play Now"),
	)
	if err != nil {
		fmt.Fprintln(log, "Skipped the update.")
		return true
	}

	report := frontend.RunWithProgress(found, frontend.UPDATE)
	// Nothing was downloaded, so the game is as it was before the update was offered.
	if report.Declined() {
		fmt.Fprintln(log, "Skipped the update: there may not be enough free space.")
		return true
	}
	if !report.Complete() {
		// A half-applied update may leave the game unable to start, so ask the user before starting the game.
		return report.Ask("Play Anyway", "Close")
	}
	if report.Warned() {
		report.Show()
	}
	return true
}

// checkForUpdates checks for an update behind a dialog that says so, which the user can use
// to skip the check. It gives up after UPDATE_CHECK_TIMEOUT.
func checkForUpdates(found frontend.Installation) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), UPDATE_CHECK_TIMEOUT)
	defer cancel()

	dialog, err := zenity.Progress(zenity.Title(frontend.UPDATE.Title), zenity.Pulsate(), zenity.CancelLabel("Skip"))
	if err != nil {
		return "", err
	}
	_ = dialog.Text("Checking for updates...")
	go func() {
		select {
			case <-dialog.Done():
				cancel()

			case <-ctx.Done():
		}
	}()

	update, err := found.CheckForUpdates(ctx)
	_ = dialog.Close()

	if errors.Is(err, context.DeadlineExceeded) {
		return "", fmt.Errorf("it took longer than %s", UPDATE_CHECK_TIMEOUT)
	}
	if errors.Is(err, context.Canceled) {
		return "", errors.New("the user skipped it")
	}
	return update, err
}
