package frontend

import (
	"errors"
	"strings"
	"testing"
)

var TEST_UTILITY = Utility{Success: "Verified.", Failure: "Couldn't verify."}

func TestDescribeASuccessfulRun(t *testing.T) {
	outcome := outcome{}
	outcome.observe(line{Type: "INSTALL_VERIFIED", Message: "Installation verified."})

	got := outcome.describe(TEST_UTILITY, nil)
	if got.kind != "info" || got.text != TEST_UTILITY.Success {
		t.Errorf("got %+v, want the utility's success message", got)
	}
}

func TestDescribeASuccessfulRunWithWarnings(t *testing.T) {
	outcome := outcome{}
	outcome.observe(line{Type: "SHORTCUTS_FAILURE", Message: "Failed to create shortcuts."})

	got := outcome.describe(TEST_UTILITY, nil)
	if got.kind != "warning" {
		t.Errorf("got kind %q, want a warning", got.kind)
	}
	if !strings.Contains(got.text, TEST_UTILITY.Success) || !strings.Contains(got.text, "Failed to create shortcuts.") {
		t.Errorf("got text %q, want both the success and the warning", got.text)
	}
}

func TestDescribeAFailedRun(t *testing.T) {
	outcome := outcome{}
	outcome.fail("Some assets couldn't be re-downloaded.")

	got := outcome.describe(TEST_UTILITY, errors.New("exit status 1"))
	if got.kind != "error" || got.text != "Some assets couldn't be re-downloaded." {
		t.Errorf("got %+v, want the CLI's own error", got)
	}
}

func TestDescribeAFailedRunWithoutAnExplanation(t *testing.T) {
	outcome := outcome{}

	got := outcome.describe(TEST_UTILITY, errors.New("exit status 2"))
	if got.kind != "error" || got.text != TEST_UTILITY.Failure {
		t.Errorf("got %+v, want the utility's failure message", got)
	}
}

func TestDescribeAFailureTheCLIExitedCleanlyFrom(t *testing.T) {
	outcome := outcome{}
	outcome.observe(line{Type: "UPDATE_FAILURE", Message: "Update failed."})

	// Every asset failed to patch, but the CLI reports that as a finding and exits 0.
	got := outcome.describe(TEST_UTILITY, nil)
	if got.kind != "error" || got.text != TEST_UTILITY.Failure {
		t.Errorf("got %+v, want the utility's failure message", got)
	}
}

func TestDescribeAPartialRun(t *testing.T) {
	outcome := outcome{}
	message := "Partially updated. You may need to run the \"verify\" command."
	outcome.observe(line{Type: "UPDATE_PARTIAL", Message: message})

	got := outcome.describe(TEST_UTILITY, nil)
	if got.kind != "warning" {
		t.Errorf("got kind %q, want a warning", got.kind)
	}
	if got.text != message {
		t.Errorf("got text %q, want only what was left undone, without the success message it contradicts", got.text)
	}
}

func TestExitCodeOfARunThatDidntFinishItsWork(t *testing.T) {
	cases := map[string]line{
		"failed":      {Type: "UNINSTALL_FAILURE", Message: "Failed to uninstall."},
		"partial":     {Type: "UNINSTALL_PARTIAL", Message: "Partially uninstalled."},
		"interrupted": {Type: "UNINSTALL_INTERRUPTED", Message: "Uninstall interrupted."},
	}
	for name, l := range cases {
		outcome := outcome{}
		outcome.observe(l)
		// The CLI exits cleanly from each of these, so its exit code alone would say otherwise.
		if outcome.exitCode(nil) == 0 {
			t.Errorf("%s: got exit code 0, want non-zero", name)
		}
	}

	warned := outcome{}
	warned.observe(line{Type: "SHORTCUTS_REMOVE_FAILURE", Message: "Failed to remove some shortcuts."})
	if warned.exitCode(nil) != 0 {
		t.Error("a warning alone made the exit code non-zero")
	}
}

func TestDescribeAConclusion(t *testing.T) {
	utility := TEST_UTILITY
	utility.Conclusions = map[string]string{"INSTALL_UP_TO_DATE": "Already up to date."}
	reached := outcome{}
	reached.observe(line{Type: "INSTALL_UP_TO_DATE", Message: "Installation is up to date: v12"})

	got := reached.describe(utility, nil)
	if got.kind != "info" || got.text != "Already up to date." {
		t.Errorf("got %+v, want the conclusion in place of the success message", got)
	}

	// A conclusion the run didn't reach leaves the success message alone.
	unreached := outcome{}
	got = unreached.describe(utility, nil)
	if got.text != utility.Success {
		t.Errorf("got text %q, want the success message", got.text)
	}
}

func TestDescribeAnInterruptedRun(t *testing.T) {
	outcome := outcome{}
	message := "Verification interrupted. Run \"verify\" again to finish checking the installation."
	outcome.observe(line{Type: "VERIFY_INTERRUPTED", Message: message})

	// An interrupted run exits with an error, but stopped safely - it isn't a failure.
	got := outcome.describe(TEST_UTILITY, errors.New("exit status 1"))
	if got.kind != "warning" || got.text != message {
		t.Errorf("got %+v, want the CLI's explanation of how to resume", got)
	}
}
