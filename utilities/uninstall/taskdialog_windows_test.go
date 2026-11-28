package main

import (
	"encoding/binary"
	"runtime"
	"testing"
)

// Where TASKDIALOGCONFIG's fields are on 64-bit Windows, as the Windows SDK's CommCtrl.h lays it
// out. The structure is packed, so these differ from where Go would put the same fields.
const (
	CONFIG_SIZE                = 160
	OFFSET_FLAGS               = 20
	OFFSET_MAIN_INSTRUCTION    = 44
	OFFSET_BUTTON_COUNT        = 60
	OFFSET_DEFAULT_BUTTON      = 72
	OFFSET_VERIFICATION_TEXT   = 92
)

func testConfig(t *testing.T, dialog taskDialog) []byte {
	t.Helper()
	var pinner runtime.Pinner
	t.Cleanup(pinner.Unpin)
	config, err := dialog.config(&pinner, 0, TDF_ALLOW_DIALOG_CANCELLATION)
	if err != nil {
		t.Fatalf("lay out the dialog: %v", err)
	}
	return config
}

func TestTaskDialogConfigMatchesTheWindowsLayout(t *testing.T) {
	config := testConfig(t, taskDialog{title: "Title", heading: "Heading", text: "Text", accept: "Accept", checkbox: "Checkbox"})

	if len(config) != CONFIG_SIZE {
		t.Fatalf("got %d bytes, want %d", len(config), CONFIG_SIZE)
	}
	if size := binary.LittleEndian.Uint32(config[0:]); size != CONFIG_SIZE {
		t.Errorf("got cbSize %d, want %d", size, CONFIG_SIZE)
	}
	if flags := binary.LittleEndian.Uint32(config[OFFSET_FLAGS:]); flags != TDF_ALLOW_DIALOG_CANCELLATION {
		t.Errorf("got dwFlags %#x, want %#x", flags, TDF_ALLOW_DIALOG_CANCELLATION)
	}
	if binary.LittleEndian.Uint64(config[OFFSET_MAIN_INSTRUCTION:]) == 0 {
		t.Error("got no pszMainInstruction")
	}
	if count := binary.LittleEndian.Uint32(config[OFFSET_BUTTON_COUNT:]); count != 1 {
		t.Errorf("got cButtons %d, want 1", count)
	}
	if button := binary.LittleEndian.Uint32(config[OFFSET_DEFAULT_BUTTON:]); button != IDCANCEL {
		t.Errorf("got nDefaultButton %d, want Cancel", button)
	}
	if binary.LittleEndian.Uint64(config[OFFSET_VERIFICATION_TEXT:]) == 0 {
		t.Error("got no pszVerificationText")
	}
}

func TestTaskDialogConfigWithoutACheckbox(t *testing.T) {
	config := testConfig(t, taskDialog{title: "Title", heading: "Heading", text: "Text", accept: "Accept"})

	// Without verification text, Windows doesn't show a checkbox.
	if binary.LittleEndian.Uint64(config[OFFSET_VERIFICATION_TEXT:]) != 0 {
		t.Error("got pszVerificationText for a dialog without a checkbox")
	}
}
