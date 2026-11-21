package main

import (
	"fmt"
	"io"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// UNPINNABLE_APP_ID is the AppUserModelID the game is started with when there's no shortcut
// to take one from. There's no shortcut to pin in the game's place either, so its windows
// can't be pinned at all.
const UNPINNABLE_APP_ID = "Wilderzone.TribesAscend.Unpinnable"

// The flags of a process's startup information that say what its title holds.
const STARTF_TITLEISLINKNAME = 0x00000800
const STARTF_TITLEISAPPID = 0x00001000
const STARTF_PREVENTPINNING = 0x00002000

// The Windows Shell's shortcut object, and the two of its interfaces used here: the shortcut
// file, and the properties stored in it.
//
// The installer writes these properties, with its own copy of this plumbing in
// tai/appid_windows.go.
var CLSID_SHELL_LINK = windows.GUID{Data1: 0x00021401, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
var IID_PERSIST_FILE = windows.GUID{Data1: 0x0000010B, Data2: 0x0000, Data3: 0x0000, Data4: [8]byte{0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46}}
var IID_PROPERTY_STORE = windows.GUID{Data1: 0x886D8EEB, Data2: 0x8CF2, Data3: 0x4446, Data4: [8]byte{0x8D, 0x02, 0xCD, 0xBA, 0x1D, 0xBD, 0xCF, 0x99}}

// propertyKey identifies a property of a shell object (a PROPERTYKEY).
type propertyKey struct {
	format windows.GUID
	id     uint32
}

// PKEY_APP_USER_MODEL_ID is System.AppUserModel.ID.
var PKEY_APP_USER_MODEL_ID = propertyKey{windows.GUID{Data1: 0x9F4C2855, Data2: 0x9F79, Data3: 0x4B39, Data4: [8]byte{0xA8, 0xD0, 0xE1, 0xD4, 0x2D, 0xE1, 0xD5, 0xF3}}, 5}

// propVariant is a PROPVARIANT holding a string, the only kind of value used here. Its value
// is a union two pointers wide on 64-bit Windows, so the second is padding.
type propVariant struct {
	kind     uint16
	reserved [3]uint16
	value    *uint16
	_        uintptr
}

const VT_LPWSTR = 31

const STGM_READ = 0x0

// The positions of the methods used here in their interface's method table. Every table
// starts with IUnknown's three methods.
const (
	METHOD_QUERY_INTERFACE   = 0
	METHOD_RELEASE           = 2
	PERSIST_FILE_LOAD        = 5
	PROPERTY_STORE_GET_VALUE = 5
)

var ole32 = windows.NewLazySystemDLL("ole32.dll")
var procCoCreateInstance = ole32.NewProc("CoCreateInstance")
var procPropVariantClear = ole32.NewProc("PropVariantClear")

// comObject is a COM interface pointer, whose first word points to the interface's method
// table. The table is declared larger than any interface used here needs, and only the
// methods an interface actually has are ever read from it.
type comObject struct {
	methods *[16]uintptr
}

// release gives up this reference to the object.
func (o *comObject) release() {
	syscall.SyscallN(o.methods[METHOD_RELEASE], uintptr(unsafe.Pointer(o)))
}

// queryInterface returns another of the object's interfaces. The caller has to release it.
func (o *comObject) queryInterface(iid *windows.GUID) (*comObject, error) {
	var result *comObject
	hr, _, _ := syscall.SyscallN(o.methods[METHOD_QUERY_INTERFACE], uintptr(unsafe.Pointer(o)), uintptr(unsafe.Pointer(iid)), uintptr(unsafe.Pointer(&result)))
	if err := hresultError(hr); err != nil {
		return nil, err
	}
	return result, nil
}

// hresultError returns the error an HRESULT stands for, or nil if it reports success. An
// HRESULT is 32 bits, so anything above them is ignored.
func hresultError(hr uintptr) error {
	if int32(hr) < 0 {
		return syscall.Errno(uint32(hr))
	}
	return nil
}

// initialiseCOM initialises COM on the current thread, and returns the function that undoes
// it. COM is per-thread, so the caller has to keep its goroutine locked to the thread until
// it's done.
func initialiseCOM() (func(), error) {
	err := windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED)
	code, isCode := err.(syscall.Errno)
	// S_FALSE means it was already initialised on this thread, which still has to be balanced.
	if err == nil || (isCode && uint32(code) == uint32(windows.S_FALSE)) {
		return windows.CoUninitialize, nil
	}
	// It was already initialised in the other threading model. That's still usable, but should not be uninitialised here.
	if isCode && uint32(code) == uint32(windows.RPC_E_CHANGED_MODE) {
		return func() {}, nil
	}
	return nil, err
}

// openShortcut loads a shortcut file, and returns the shortcut object it was loaded into as
// the file's interface. The caller has to release it.
func openShortcut(path string, mode uint32) (*comObject, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	var file *comObject
	hr, _, _ := syscall.SyscallN(procCoCreateInstance.Addr(), uintptr(unsafe.Pointer(&CLSID_SHELL_LINK)), 0, uintptr(windows.CLSCTX_INPROC_SERVER), uintptr(unsafe.Pointer(&IID_PERSIST_FILE)), uintptr(unsafe.Pointer(&file)))
	if err := hresultError(hr); err != nil {
		return nil, fmt.Errorf("create a shortcut object: %w", err)
	}
	hr, _, _ = syscall.SyscallN(file.methods[PERSIST_FILE_LOAD], uintptr(unsafe.Pointer(file)), uintptr(unsafe.Pointer(name)), uintptr(mode))
	if err := hresultError(hr); err != nil {
		file.release()
		return nil, fmt.Errorf("load %s: %w", path, err)
	}
	return file, nil
}

// clearVariant frees whatever a property store put in a variant.
func clearVariant(variant *propVariant) {
	syscall.SyscallN(procPropVariantClear.Addr(), uintptr(unsafe.Pointer(variant)))
}

// shortcutAppID reads a shortcut's AppUserModelID (System.AppUserModel.ID). A shortcut
// without one reads as empty.
func shortcutAppID(path string) (string, error) {
	// COM is per-thread, so this goroutine has to stay on the thread it's initialised on.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	uninitialise, err := initialiseCOM()
	if err != nil {
		return "", fmt.Errorf("initialise COM: %w", err)
	}
	defer uninitialise()

	file, err := openShortcut(path, STGM_READ)
	if err != nil {
		return "", err
	}
	defer file.release()
	store, err := file.queryInterface(&IID_PROPERTY_STORE)
	if err != nil {
		return "", fmt.Errorf("open the properties of %s: %w", path, err)
	}
	defer store.release()

	var variant propVariant
	hr, _, _ := syscall.SyscallN(store.methods[PROPERTY_STORE_GET_VALUE], uintptr(unsafe.Pointer(store)), uintptr(unsafe.Pointer(&PKEY_APP_USER_MODEL_ID)), uintptr(unsafe.Pointer(&variant)))
	if err := hresultError(hr); err != nil {
		return "", fmt.Errorf("read the AppUserModelID of %s: %w", path, err)
	}
	defer clearVariant(&variant)
	if variant.kind != VT_LPWSTR || variant.value == nil {
		return "", nil
	}
	return windows.UTF16PtrToString(variant.value), nil
}

// startupTitle returns the title in this process's startup information, along with the flags
// that say what it holds: the shortcut it was started from, if it was started by the shell,
// or an AppUserModelID it was given.
func startupTitle() (string, uint32) {
	var startup windows.StartupInfo
	if windows.GetStartupInfo(&startup) != nil || startup.Title == nil {
		return "", 0
	}
	return windows.UTF16PtrToString(startup.Title), startup.Flags
}

// gameAppID returns the AppUserModelID to start the game with, and whether its windows can be
// pinned to the taskbar.
//
// The taskbar groups and pins windows by the process that owns them, and the game isn't the
// process a shortcut starts: this launcher is. On its own, the game would get a taskbar button
// of its own, and pinning it would pin the game's executable, which can't reach a login server
// by itself. Started with the AppUserModelID of the shortcut this launcher was started from,
// the game shares that shortcut's button, and pinning it pins the shortcut instead.
//
// Without an ID to pass on - this launcher wasn't started from a shortcut, or from one written
// before shortcuts had IDs - there's no shortcut for the taskbar to pin in the game's place,
// so the game's windows can't be pinned at all.
func gameAppID(log io.Writer) (string, bool) {
	title, flags := startupTitle()
	// Started with an ID already, which is passed on as it is: pinnable or not.
	if flags & STARTF_TITLEISAPPID != 0 && title != "" {
		pinnable := flags & STARTF_PREVENTPINNING == 0
		fmt.Fprintf(log, "Taskbar: %s, as given (pinnable=%t)\n", title, pinnable)
		return title, pinnable
	}
	if flags & STARTF_TITLEISLINKNAME == 0 || title == "" {
		fmt.Fprintln(log, "Taskbar: not started from a shortcut, so the game can't be pinned.")
		return UNPINNABLE_APP_ID, false
	}
	link := title
	appID, err := shortcutAppID(link)
	if err != nil {
		fmt.Fprintf(log, "Taskbar: couldn't read an AppUserModelID, so the game can't be pinned: %s\n", err.Error())
		return UNPINNABLE_APP_ID, false
	}
	if appID == "" {
		fmt.Fprintf(log, "Taskbar: %s has no AppUserModelID, so the game can't be pinned.\n", link)
		return UNPINNABLE_APP_ID, false
	}
	fmt.Fprintf(log, "Taskbar: %s, from %s\n", appID, link)
	return appID, true
}
