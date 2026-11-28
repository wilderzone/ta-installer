package main

import (
	"encoding/binary"
	"fmt"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Values from the Windows SDK's CommCtrl.h, WinUser.h and WinBase.h.
const (
	TDF_USE_HICON_MAIN                   = 0x0002
	TDF_ALLOW_DIALOG_CANCELLATION        = 0x0008
	TDCBF_CANCEL_BUTTON                  = 0x0008
	// MAKEINTRESOURCE(-1).
	TD_WARNING_ICON                      = 0xFFFF
	IDCANCEL                             = 2
	IDI_QUESTION                         = 32514
	ACTCTX_FLAG_ASSEMBLY_DIRECTORY_VALID = 0x004
	ACTCTX_FLAG_RESOURCE_NAME_VALID      = 0x008
)

// ACCEPT_BUTTON identifies the dialog's own button, which answers yes. Its ID only has to
// differ from the standard buttons' (IDOK, IDCANCEL, ...).
const ACCEPT_BUTTON = 100

// The resource in shell32.dll that holds its manifest. The manifest asks for version 6 of the
// common controls, which is the only version that has task dialogs.
const SHELL32_MANIFEST = 124

// The DPI awareness contexts to ask for, best first: DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2,
// _PER_MONITOR_AWARE, and _SYSTEM_AWARE (-4, -3, and -2). Older versions of Windows don't know
// the newer ones.
var DPI_AWARENESS_CONTEXTS = []uintptr{^uintptr(3), ^uintptr(2), ^uintptr(1)}

var (
	comctl32 = windows.NewLazySystemDLL("comctl32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	user32   = windows.NewLazySystemDLL("user32.dll")

	procTaskDialogIndirect           = comctl32.NewProc("TaskDialogIndirect")
	procCreateActCtx                 = kernel32.NewProc("CreateActCtxW")
	procActivateActCtx               = kernel32.NewProc("ActivateActCtx")
	procDeactivateActCtx             = kernel32.NewProc("DeactivateActCtx")
	procReleaseActCtx                = kernel32.NewProc("ReleaseActCtx")
	procLoadIcon                     = user32.NewProc("LoadIconW")
	procSetThreadDpiAwarenessContext = user32.NewProc("SetThreadDpiAwarenessContext")
)

// taskDialog is a question asked in a task dialog: the message box's successor, which can carry a
// checkbox beneath its text. zenity only shows message boxes.
type taskDialog struct {
	title    string
	// Shown above the text, in larger type.
	heading  string
	text     string
	// The label of the button that answers yes. The other button is Cancel, which is the default.
	accept   string
	// The label of a checkbox, which starts unticked. If it's empty, there's no checkbox.
	checkbox string
	// Shows a warning sign, instead of a question mark.
	warning  bool
}

// show asks the question, and reports whether the user accepted it and whether they ticked the
// checkbox. An error means the dialog couldn't be shown at all.
func (d taskDialog) show() (accepted bool, checked bool, err error) {
	// Activation contexts and DPI awareness both belong to a thread, so this has to stay on one.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	deactivate, err := activateCommonControls()
	if err != nil {
		return false, false, err
	}
	defer deactivate()
	defer useHighDPI()()

	// Version 6 of the common controls has to be active before this loads them for the first time.
	err = procTaskDialogIndirect.Find()
	if err != nil {
		return false, false, fmt.Errorf("find TaskDialogIndirect: %w", err)
	}

	// The structure holds pointers that Go can't see, so what they point at is pinned until
	// the dialog has closed.
	var pinner runtime.Pinner
	defer pinner.Unpin()
	icon, flags := d.mainIcon()
	config, err := d.config(&pinner, icon, flags | TDF_ALLOW_DIALOG_CANCELLATION)
	if err != nil {
		return false, false, err
	}

	var button int32
	var ticked int32
	result, _, _ := procTaskDialogIndirect.Call(
		uintptr(unsafe.Pointer(&config[0])),
		uintptr(unsafe.Pointer(&button)),
		0,
		uintptr(unsafe.Pointer(&ticked)),
	)
	// A negative HRESULT is a failure.
	if int32(result) < 0 {
		return false, false, fmt.Errorf("show a task dialog: HRESULT %#x", uint32(result))
	}
	return button == ACCEPT_BUTTON, ticked != 0, nil
}

// mainIcon returns the icon to show beside the heading, and the flag that says what kind of
// icon it is.
func (d taskDialog) mainIcon() (uintptr, uint32) {
	if d.warning {
		return TD_WARNING_ICON, 0
	}
	// Task dialogs don't have a question mark of their own, so the system's icon is handed over instead.
	icon, _, _ := procLoadIcon.Call(0, IDI_QUESTION)
	if icon == 0 {
		return 0, 0
	}
	return icon, TDF_USE_HICON_MAIN
}

// config lays the dialog out as a TASKDIALOGCONFIG structure. Its text is pinned with `pinner`,
// which must stay pinned until the dialog has closed.
func (d taskDialog) config(pinner *runtime.Pinner, icon uintptr, flags uint32) ([]byte, error) {
	var failed error
	text := func(s string) uintptr {
		if s == "" {
			return 0
		}
		pointer, err := windows.UTF16PtrFromString(s)
		if err != nil {
			failed = err
			return 0
		}
		pinner.Pin(pointer)
		return uintptr(unsafe.Pointer(pointer))
	}

	// A TASKDIALOG_BUTTON for the button that answers yes.
	buttons := packed{}
	buttons.dword(ACCEPT_BUTTON)                // nButtonID
	buttons.pointer(text(d.accept))             // pszButtonText
	pinner.Pin(&buttons.bytes[0])
	accept := uintptr(unsafe.Pointer(&buttons.bytes[0]))

	config := packed{}
	config.dword(0)                             // cbSize, filled in once it's known
	config.pointer(0)                           // hwndParent
	config.pointer(0)                           // hInstance
	config.dword(flags)                         // dwFlags
	config.dword(TDCBF_CANCEL_BUTTON)           // dwCommonButtons
	config.pointer(text(d.title))               // pszWindowTitle
	config.pointer(icon)                        // hMainIcon or pszMainIcon
	config.pointer(text(d.heading))             // pszMainInstruction
	config.pointer(text(d.text))                // pszContent
	config.dword(1)                             // cButtons
	config.pointer(accept)                      // pButtons
	config.dword(IDCANCEL)                      // nDefaultButton
	config.dword(0)                             // cRadioButtons
	config.pointer(0)                           // pRadioButtons
	config.dword(0)                             // nDefaultRadioButton
	config.pointer(text(d.checkbox))            // pszVerificationText
	config.pointer(0)                           // pszExpandedInformation
	config.pointer(0)                           // pszExpandedControlText
	config.pointer(0)                           // pszCollapsedControlText
	config.pointer(0)                           // hFooterIcon or pszFooterIcon
	config.pointer(0)                           // pszFooter
	config.pointer(0)                           // pfCallback
	config.pointer(0)                           // lpCallbackData
	config.dword(0)                             // cxWidth
	binary.LittleEndian.PutUint32(config.bytes, uint32(len(config.bytes)))

	if failed != nil {
		return nil, fmt.Errorf("prepare the task dialog's text: %w", failed)
	}
	return config.bytes, nil
}

// packed builds a structure that's laid out the way CommCtrl.h declares the task dialog's:
// with no padding between its fields. Go can't declare a structure like that, so its fields
// are written out one by one. The utilities are only built for 64-bit Windows, so a pointer
// takes 8 bytes.
type packed struct {
	bytes []byte
}

func (p *packed) dword(value uint32) {
	p.bytes = binary.LittleEndian.AppendUint32(p.bytes, value)
}

func (p *packed) pointer(value uintptr) {
	p.bytes = binary.LittleEndian.AppendUint64(p.bytes, uint64(value))
}

// actCtx is the ACTCTXW structure, which describes an activation context.
type actCtx struct {
	size                  uint32
	flags                 uint32
	source                *uint16
	processorArchitecture uint16
	langID                uint16
	assemblyDirectory     *uint16
	resourceName          uintptr
	applicationName       *uint16
	module                windows.Handle
}

// activateCommonControls switches this thread to version 6 of the common controls, which is the
// only version with task dialogs. It returns a function that switches the thread back.
//
// Programs normally ask for version 6 in a manifest of their own. The utilities don't have one,
// so this borrows the manifest that shell32.dll carries, as zenity does for its own dialogs.
func activateCommonControls() (func(), error) {
	system, err := windows.GetSystemDirectory()
	if err != nil {
		return nil, fmt.Errorf("find the system directory: %w", err)
	}
	source, err := windows.UTF16PtrFromString(filepath.Join(system, "shell32.dll"))
	if err != nil {
		return nil, err
	}
	directory, err := windows.UTF16PtrFromString(system)
	if err != nil {
		return nil, err
	}

	activation := actCtx{
		flags:             ACTCTX_FLAG_RESOURCE_NAME_VALID | ACTCTX_FLAG_ASSEMBLY_DIRECTORY_VALID,
		source:            source,
		assemblyDirectory: directory,
		resourceName:      SHELL32_MANIFEST,
	}
	activation.size = uint32(unsafe.Sizeof(activation))
	handle, _, err := procCreateActCtx.Call(uintptr(unsafe.Pointer(&activation)))
	if windows.Handle(handle) == windows.InvalidHandle {
		return nil, fmt.Errorf("create an activation context: %w", err)
	}

	var cookie uintptr
	ok, _, err := procActivateActCtx.Call(handle, uintptr(unsafe.Pointer(&cookie)))
	if ok == 0 {
		procReleaseActCtx.Call(handle)
		return nil, fmt.Errorf("activate an activation context: %w", err)
	}
	return func() {
		procDeactivateActCtx.Call(0, cookie)
		procReleaseActCtx.Call(handle)
	}, nil
}

// useHighDPI has Windows draw this thread's windows at the display's own resolution, as zenity
// does for its dialogs, instead of stretching them. It returns a function that puts the thread
// back as it was. Versions of Windows that can't do this stretch the dialog instead.
func useHighDPI() func() {
	if procSetThreadDpiAwarenessContext.Find() != nil {
		return func() {}
	}
	for _, awareness := range DPI_AWARENESS_CONTEXTS {
		previous, _, _ := procSetThreadDpiAwarenessContext.Call(awareness)
		if previous != 0 {
			return func() {
				procSetThreadDpiAwarenessContext.Call(previous)
			}
		}
	}
	return func() {}
}
