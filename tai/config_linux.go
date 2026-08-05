package main

import (
	"os"
	"path/filepath"
)

const CLI_NAME = "tai-cli"
const SHORTCUT_NAME = "tribes-ascend.desktop"
const UNINSTALL_SHORTCUT_NAME = "tribes-ascend-uninstall.desktop"

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
	},
}

// defaultInstallPath is where the game goes when the user doesn't choose somewhere else.
// It's reported by the CLI so the GUI doesn't have to hardcode a second copy of it.
//
// ~/Games is the convention other Linux game managers use, and being under the user's own
// home means an install never needs elevating - which is why there's no Linux counterpart
// to runElevated.
func defaultInstallPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Games", SHORTCUT_DIRECTORY)
}
