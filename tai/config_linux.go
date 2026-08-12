package main

import "path/filepath"

const CLI_NAME = "tai-cli"
const SHORTCUT_NAME = "tribes-ascend.desktop"
const UNINSTALL_SHORTCUT_NAME = "tribes-ascend-uninstall.desktop"

// The hint shown when shortcuts couldn't be created. Everything TAI writes on Linux is
// under the user's own home, so a failure is about that directory, not about permissions.
const SHORTCUTS_FAILURE_HINT = "Check that your home directory is writable."

// The location of the utilities' checksums, relative to the repository. Fetched at verify
// time and hashed against the installed utilities.
const UTILITIES_CHECKSUMS_SOURCE = "/installer/Installer-linux-amd64.json"

// The installation utilities package. Checksums are deliberately omitted: they're fetched
// from UTILITIES_CHECKSUMS_SOURCE. The file names are still needed so
// installationUtilitiesAreMissing can tell whether the utilities are already installed.
//
// Linux only needs the CLI itself. The Windows package also carries uninstall/update/verify
// executables, which exist purely to be double-clicked from Explorer; desktop entries do
// that job here, and they point straight at the CLI.
var INSTALLATION_UTILITIES = Asset{
	Name:        "Installation Utilities",
	Type:        "prerequisite",
	Source:      "/installer/Installer-linux-amd64.zip",
	Destination: "/Installer",
	Files: []File{
		{File: CLI_NAME},
		{File: ICON_FILE},
	},
}

// The location of the compatibility runtime's checksums, relative to the repository. As with
// the utilities, they're fetched at verify time rather than compiled in: the runtime is
// thousands of files, and its checksums change every time the pinned build does.
const RUNTIME_CHECKSUMS_SOURCE = "/runtime/Runtime.json"

// The compatibility runtime: a pinned Proton build and the umu launcher that drives it.
//
// This is a bespoke package rather than a manifest asset, for the same reason the DirectX 9
// redistributable is on Windows. The manifest describes the game, and is served by whatever
// repository --repo names; a custom repository shouldn't have to mirror half a gigabyte of
// Proton for a Linux install to work. So, like the utilities, this always comes from
// DEFAULT_REPO regardless of --repo.
//
// PackedSize only scales the progress bar, and is pinned alongside Source - a stale figure
// mis-scales the bar but breaks nothing.
var COMPATIBILITY_RUNTIME = Asset{
	Name:         "Compatibility Runtime",
	Type:         "prerequisite",
	Source:       "/runtime/Runtime.tar.gz",
	Destination:  "/" + RUNTIME_DIRECTORY,
	PackedSize:   532500000,
	Files: []File{
		{File: filepath.Join(UMU_DIRECTORY, UMU_EXECUTABLE)},
	},
}
