package main

const DEFAULT_REPO = "https://cdn.wilderzone.live/ta"
const MANIFEST_NAME = "manifest.json"
const SHORTCUT_ARGUMENTS = "-hostx=18.197.240.229 -abslog=%q"
const SHORTCUT_DESCRIPTION = "Tribes: Ascend"
const SHORTCUT_DIRECTORY = "Tribes Ascend"
const SHORTCUT_NAME = "Tribes Ascend.lnk"
const LOG_NAME = "log.txt"
const INSTALLATION_UTILITIES_DIRECTORY = "Installer"
const LOGS_DIRECTORY = "Logs"
const TEMP_DIRECTORY = "Temp"
const UNINSTALLER_REGISTRY_PATH = `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Tribes Ascend`
const UNINSTALLER_DISPLAY_NAME = "Tribes: Ascend"
const UNINSTALLER_DISPLAY_VERSION = "0.1.0"
const UNINSTALLER_ABOUT_LINK = "https://github.com/wilderzone/ta-installer"
const UNINSTALLER_HELP_LINK = "https://github.com/wilderzone/ta-installer"
const UNINSTALLER_UPDATES_LINK = "https://github.com/wilderzone/ta-installer/releases"
const UNINSTALLER_PUBLISHER = "Wilderzone"

// Maximum number of assets to process concurrently. Downloads are network-bound
// (kept modest to avoid hammering the CDN); verification is disk/CPU-bound.
const DOWNLOAD_CONCURRENCY = 4
const VERIFY_CONCURRENCY = 4

var INSTALLATION_UTILITIES = Asset{
	Name:        "Installation Utilities",
	Type:        "prerequisite",
	Source:      "/Installer.zip",
	Destination: "/Installer",
	Files: []File{
		{
			Checksum: "651dc1d9186cc2ce4a6dd2d36b22a44318e85e52a4825dd2e0b6d592757e2e27",
			File: "tai-cli.exe",
		},
		{
			Checksum: "5fa6d0694c007b848584d622b045ede3b3af459cf6dfd5abdd53c7ab2d680de4",
			File: "uninstall.exe",
		},
		{
			Checksum: "2e8c60056434ccf3f9c477b31ae0c2496e8147580acd23efd92e2df86612d323",
			File: "update.exe",
		},
		{
			Checksum: "4ca96519a88d60cd0d90f2b2cabd5ceec7cafb575654abe0a88e485b856026f3",
			File: "verify.exe",
		},
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
