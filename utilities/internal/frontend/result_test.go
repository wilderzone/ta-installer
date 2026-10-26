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
