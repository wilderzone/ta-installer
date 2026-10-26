package main

const CLI_NAME = "tai-cli"
const PLAY_NAME = "play"
const UNINSTALL_NAME = "uninstall"
const UPDATE_NAME = "update"
const VERIFY_NAME = "verify"
const SHORTCUT_NAME = "tribes-ascend.desktop"
const UNINSTALL_SHORTCUT_NAME = "tribes-ascend-uninstall.desktop"
const UPDATE_SHORTCUT_NAME = "tribes-ascend-update.desktop"
const VERIFY_SHORTCUT_NAME = "tribes-ascend-verify.desktop"

// The hint shown when shortcuts couldn't be created. Everything TAI writes on Linux is
// under the user's own home, so a failure is about that directory, not about permissions.
const SHORTCUTS_FAILURE_HINT = "Check that your home directory is writable."
const UTILITIES_FAILURE_HINT = "Uninstalling the game via the applications menu may not be possible."

// The location of the utilities' checksums, relative to the repository.
const UTILITIES_CHECKSUMS_SOURCE = "/installer/Installer-linux-amd64.json"

// The installation utilities package. Checksums are deliberately omitted: they're fetched
// from UTILITIES_CHECKSUMS_SOURCE. The file names are still needed so TAI can tell whether
// the utilities are already installed.
//
// These are the same utilities as on Windows, plus the icons that desktop entries can't read
// out of an executable. The desktop entries run the utilities.
var INSTALLATION_UTILITIES = Asset{
	Name:        "Installation Utilities",
	Type:        "prerequisite",
	Source:      "/installer/Installer-linux-amd64.zip",
	Destination: "/Installer",
	Files: []File{
		{File: CLI_NAME},
		{File: PLAY_NAME},
		{File: UNINSTALL_NAME},
		{File: UPDATE_NAME},
		{File: VERIFY_NAME},
		{File: ICON_FILE},
		{File: UNINSTALL_ICON_FILE},
	},
}

// The location of the compatibility runtime's checksums, relative to the repository.
const RUNTIME_CHECKSUMS_SOURCE = "/runtime/Runtime.json"

// The compatibility runtime: a pinned Proton build and the umu launcher that drives it.
//
// This is a bespoke package rather than a manifest asset, for the same reason the DirectX 9
// redistributable is on Windows.
//
// PackedSize only scales the progress bar, and is pinned alongside Source - a stale figure
// mis-scales the bar but breaks nothing.
var COMPATIBILITY_RUNTIME = Asset{
	Name:         "Compatibility Layer",
	Type:         "prerequisite",
	Source:       "/runtime/Runtime.tar.gz",
	Destination:  "/" + RUNTIME_DIRECTORY,
	PackedSize:   533499712,
	Files: []File{
		{File: "umu/umu-run"},
	},
}
