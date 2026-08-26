package main

const DEFAULT_REPO = "https://cdn.wilderzone.org/ta/game"
const GAME_BINARY_NAME = "TribesAscend.exe"
const MANIFEST_NAME = "manifest.json"
const PRODUCT_NAME = "Tribes: Ascend"
const SHORTCUT_DESCRIPTION = "Play Tribes: Ascend"
const SHORTCUT_DIRECTORY = "Tribes Ascend"
const INSTALLATION_UTILITIES_DIRECTORY = "Installer"
const LOGS_DIRECTORY = "Logs"
const TEMP_DIRECTORY = "Temp"
const UNINSTALLER_DISPLAY_VERSION = "0.1.0"
const UNINSTALLER_ABOUT_LINK = "https://github.com/wilderzone/ta-installer"
const UNINSTALLER_HELP_LINK = "https://github.com/wilderzone/ta-installer"
const UNINSTALLER_UPDATES_LINK = "https://github.com/wilderzone/ta-installer/releases"
const UNINSTALLER_PUBLISHER = "Wilderzone"

// Maximum number of assets to process concurrently. Downloads are network-bound
// (kept modest to avoid hammering the CDN); verification is disk/CPU-bound.
const DOWNLOAD_CONCURRENCY = 4
const VERIFY_CONCURRENCY = 4

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
