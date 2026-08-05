package main

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
