package main

import (
	"bufio"
	"context"
	"os"
	"os/signal"
	"strings"
)

// CANCEL_COMMAND can be written to this process's standard input to ask it to stop.
// The GUI has no console of its own, so on Windows it has no way to send an interrupt
// signal to the CLI it launched; a line on stdin is the channel it does have.
const CANCEL_COMMAND = "cancel"

// cancellation returns the context that long-running commands run under, cancelled
// with Ctrl+C at a terminal, or with CANCEL_COMMAND from a parent process such as
// the GUI. Commands stop at their next safe point and report how to resume,
// rather than being killed mid-write.
func cancellation() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	watchInterrupts(cancel)
	watchInput(cancel)
	return ctx, cancel
}

// watchInterrupts cancels the run on the first Ctrl+C, then steps out of the way so that a
// second one falls through to the default handler and ends the process outright.
func watchInterrupts(cancel context.CancelFunc) {
	interrupts := make(chan os.Signal, 1)
	signal.Notify(interrupts, os.Interrupt)
	go func() {
		<-interrupts
		signal.Stop(interrupts)
		cancel()
	}()
}

// watchInput cancels the run when CANCEL_COMMAND arrives on standard input.
//
// This only listens when stdin is a pipe, which is how a parent process supplies it.
// At an interactive terminal Ctrl+C already does the job, and reading here would
// swallow the answers to confirmation prompts.
func watchInput(cancel context.CancelFunc) {
	info, err := os.Stdin.Stat()
	if err != nil || info.Mode() & os.ModeCharDevice != 0 {
		return
	}
	go func() {
		lines := bufio.NewScanner(os.Stdin)
		for lines.Scan() {
			if strings.TrimSpace(lines.Text()) == CANCEL_COMMAND {
				cancel()
				return
			}
		}
	}()
}
