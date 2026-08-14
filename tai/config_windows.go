package main

const CLI_NAME = "tai-cli.exe"
const PLAY_NAME = "play.exe"
const SHORTCUT_NAME = "Tribes Ascend.lnk"
const UNINSTALL_SHORTCUT_NAME = "Uninstall.lnk"
const SHORTCUTS_FAILURE_HINT = "You may need to re-run the installer with admin permissions to create shortcuts."
const UNINSTALLER_REGISTRY_PATH = `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Tribes Ascend`

// The location of the utilities' checksums, relative to the repository. Fetched at verify
// time and hashed against the installed utilities.
const UTILITIES_CHECKSUMS_SOURCE = "/installer/Installer.json"

// The installation utilities package. Checksums are deliberately omitted: they're fetched
// from UTILITIES_CHECKSUMS_SOURCE. The file names are still needed so
// installationUtilitiesAreMissing can tell whether the utilities are already installed.
var INSTALLATION_UTILITIES = Asset{
	Name:        "Installation Utilities",
	Type:        "prerequisite",
	Source:      "/installer/Installer.zip",
	Destination: "/Installer",
	Files: []File{
		{File: CLI_NAME},
		{File: PLAY_NAME},
		{File: "uninstall.exe"},
		{File: "update.exe"},
		{File: "verify.exe"},
	},
}
