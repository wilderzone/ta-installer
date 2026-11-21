package main

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// The Windows Shell's shortcut object, and the two of its interfaces used here: the shortcut
// file, and the properties stored in it.
//
// The play utility reads these properties back, with its own copy of this plumbing in
// utilities/play/appid_windows.go.
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
const STGM_READWRITE = 0x2

// The positions of the methods used here in their interface's method table. Every table
// starts with IUnknown's three methods.
const (
	METHOD_QUERY_INTERFACE   = 0
	METHOD_RELEASE           = 2
	PERSIST_FILE_LOAD        = 5
	PERSIST_FILE_SAVE        = 6
	PROPERTY_STORE_GET_VALUE = 5
	PROPERTY_STORE_SET_VALUE = 6
	PROPERTY_STORE_COMMIT    = 7
)

var procCoCreateInstance = windows.NewLazySystemDLL("ole32.dll").NewProc("CoCreateInstance")

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

// hresultError returns the error an HRESULT stands for, or nil if it reports success.
// An HRESULT is 32 bits, so anything above 32 bits is ignored.
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
	// It was already initialised in the other threading model. That's still usable, but it's
	// someone else's to undo.
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

// setShortcutAppID gives a shortcut an AppUserModelID (System.AppUserModel.ID). Windows uses
// it to tie a running program's windows back to the shortcut, so that they share a taskbar
// button, and pinning the program pins the shortcut.
//
// COM is per-thread, so the caller has to keep its goroutine locked to its thread.
func setShortcutAppID(path string, appID string) error {
	uninitialise, err := initialiseCOM()
	if err != nil {
		return fmt.Errorf("initialise COM: %w", err)
	}
	defer uninitialise()

	file, err := openShortcut(path, STGM_READWRITE)
	if err != nil {
		return err
	}
	defer file.release()
	store, err := file.queryInterface(&IID_PROPERTY_STORE)
	if err != nil {
		return fmt.Errorf("open the properties of %s: %w", path, err)
	}
	defer store.release()

	value, err := windows.UTF16PtrFromString(appID)
	if err != nil {
		return err
	}
	variant := propVariant{kind: VT_LPWSTR, value: value}
	hr, _, _ := syscall.SyscallN(store.methods[PROPERTY_STORE_SET_VALUE], uintptr(unsafe.Pointer(store)), uintptr(unsafe.Pointer(&PKEY_APP_USER_MODEL_ID)), uintptr(unsafe.Pointer(&variant)))
	if err := hresultError(hr); err != nil {
		return fmt.Errorf("set the AppUserModelID of %s: %w", path, err)
	}
	hr, _, _ = syscall.SyscallN(store.methods[PROPERTY_STORE_COMMIT], uintptr(unsafe.Pointer(store)))
	if err := hresultError(hr); err != nil {
		return fmt.Errorf("set the AppUserModelID of %s: %w", path, err)
	}

	// Save it back over the file it was loaded from.
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	hr, _, _ = syscall.SyscallN(file.methods[PERSIST_FILE_SAVE], uintptr(unsafe.Pointer(file)), uintptr(unsafe.Pointer(name)), 1)
	if err := hresultError(hr); err != nil {
		return fmt.Errorf("save %s: %w", path, err)
	}
	return nil
}
