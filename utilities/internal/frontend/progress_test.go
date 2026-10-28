package frontend

import (
	"strings"
	"testing"
)

func TestProgressCountsItemsBetweenDownloads(t *testing.T) {
	view := progress{}
	view.apply(line{Type: "PROGRESS", Index: 1, Total: 4, Message: "Core: Verified"})

	if view.value != PROGRESS_SCALE / 4 {
		t.Errorf("got value %d, want %d", view.value, PROGRESS_SCALE / 4)
	}
	if view.text != "[1/4] Core: Verified" {
		t.Errorf("got text %q, want the line as the CLI would print it", view.text)
	}
}

func TestProgressFollowsTheBytesOfADownload(t *testing.T) {
	view := progress{}
	view.apply(line{Type: "DOWNLOAD_STARTED", Size: 200, Message: "0% of 200 B"})
	view.apply(line{Type: "DOWNLOAD_PROGRESS", Bytes: 50, Size: 200, Message: "25% of 200 B"})
	// An asset finishing mid-download would otherwise yank the bar to a different scale.
	view.apply(line{Type: "PROGRESS", Index: 2, Total: 3, Message: "Core: Downloaded"})

	if view.value != PROGRESS_SCALE / 4 {
		t.Errorf("got value %d, want %d", view.value, PROGRESS_SCALE / 4)
	}
	if view.text != "Downloading... 25% of 200 B" {
		t.Errorf("got text %q, want the download's summary", view.text)
	}

	view.apply(line{Type: "DOWNLOAD_COMPLETE", Bytes: 200, Size: 200})
	view.apply(line{Type: "PROGRESS", Index: 3, Total: 3, Message: "Engine: Downloaded"})
	if view.text != "[3/3] Engine: Downloaded" {
		t.Errorf("got text %q, want the count to take over once the download is complete", view.text)
	}
}

func TestProgressNeverFillsTheBar(t *testing.T) {
	view := progress{}
	view.apply(line{Type: "PROGRESS", Index: 4, Total: 4, Message: "Engine: Verified"})
	if view.value >= PROGRESS_SCALE {
		t.Errorf("got value %d, want less than %d, so the dialog's OK button stays disabled", view.value, PROGRESS_SCALE)
	}

	view.apply(line{Type: "DOWNLOAD_PROGRESS", Bytes: 300, Size: 200})
	if view.value >= PROGRESS_SCALE {
		t.Errorf("got value %d for an overshooting download, want less than %d", view.value, PROGRESS_SCALE)
	}
}

func TestProgressKeepsTheTextOnOneLine(t *testing.T) {
	view := progress{}
	view.apply(line{Type: "UPDATE_FOUND", Message: "Update found: v1 → v2\n"})
	if strings.ContainsAny(view.text, "\r\n") {
		t.Errorf("got text %q, which would be read as more than one command by a Linux dialog tool", view.text)
	}
}

func TestProgressIgnoresEmptyMessages(t *testing.T) {
	view := progress{text: "Verifying..."}
	if view.apply(line{Type: "INSTALL_UP_TO_DATE"}) {
		t.Error("an empty message was reported as a change")
	}
	if view.text != "Verifying..." {
		t.Errorf("got text %q, want it left alone", view.text)
	}
}
