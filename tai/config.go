package main

const DEFAULT_REPO = "https://cdn.wilderzone.live/ta"
const MANIFEST_NAME = "manifest.json"
const SHORTCUT_ARGUMENTS = "-hostx=18.197.240.229 -abslog=%q"
const SHORTCUT_DESCRIPTION = "Tribes: Ascend"
const SHORTCUT_DIRECTORY = "Tribes Ascend"
const SHORTCUT_NAME = "Tribes Ascend.lnk"
const LOG_NAME = "log.txt"
const LOGS_DIRECTORY = "Logs"
const TEMP_DIRECTORY = "Temp"

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
