package main

const DEFAULT_REPO = "https://cdn.wilderzone.live/ta"
const MANIFEST_NAME = "manifest.json"
const PRODUCT_NAME = "Tribes: Ascend"
const SHORTCUT_ARGUMENTS = "-hostx=18.197.240.229 -abslog=%q"
const SHORTCUT_DESCRIPTION = "Tribes: Ascend"
const SHORTCUT_DIRECTORY = "Tribes Ascend"
const SHORTCUT_NAME = "Tribes Ascend.lnk"
const LOG_NAME = "log.txt"
const INSTALLATION_UTILITIES_DIRECTORY = "Installer"
const LOGS_DIRECTORY = "Logs"
const TEMP_DIRECTORY = "Temp"
const UNINSTALLER_REGISTRY_PATH = `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Tribes Ascend`
const UNINSTALLER_DISPLAY_VERSION = "0.1.0"
const UNINSTALLER_ABOUT_LINK = "https://github.com/wilderzone/ta-installer"
const UNINSTALLER_HELP_LINK = "https://github.com/wilderzone/ta-installer"
const UNINSTALLER_UPDATES_LINK = "https://github.com/wilderzone/ta-installer/releases"
const UNINSTALLER_PUBLISHER = "Wilderzone"

// Maximum number of assets to process concurrently. Downloads are network-bound
// (kept modest to avoid hammering the CDN); verification is disk/CPU-bound.
const DOWNLOAD_CONCURRENCY = 4
const VERIFY_CONCURRENCY = 4

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
		{File: "tai-cli.exe"},
		{File: "uninstall.exe"},
		{File: "update.exe"},
		{File: "verify.exe"},
	},
}

var PREREQUISITES = Asset{
	Name:        "Prerequisites",
	Type:        "prerequisite",
	Source:      "/Redist.zip",
	Destination: "/Binaries/Redist",
	Files: []File{
		{
			Checksum: "8c35296f330c0c4e431a7cf155013f3e4abec4a74efd68bb6bfcd76f08c46193",
			File:     "/directx_Jun2010_redist.exe",
		},
	},
}
