# <img src="./gui/src/assets/logo.svg" style="width: 3em;"> TA Installer
A [CLI](https://en.wikipedia.org/wiki/Command-line_interface) and [GUI](https://en.wikipedia.org/wiki/Graphical_user_interface) utility for managing Tribes: Ascend installations.


## GUI

To use TAI GUI, download and run the appropriate portable executable for your operating system from the [releases page](https://github.com/wilderzone/ta-installer/releases/latest).
You'll be prompted to select the game type (Vanilla, Modded, or Custom) and the folder in which to install TA. Modded is the recommended option, and includes community maps and CTF routes.

_Note: `tai-gui.exe` requires [VC Redist](https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist) and [WebView2](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) to be installed on your computer. Whilst it's very likely you already have them, it's possible you will need to install them before running TAI._

<p align="center">
  <img width="1262" height="902" alt="image" src="https://github.com/user-attachments/assets/c4da7aa5-9c3a-40e2-a74f-635f9fb22405" />
</p>

## CLI

TAI CLI is provided as a means to install custom variants of TA, and for use in applications that require programmatic control of the TA installation process, such as launchers.  
To use TAI CLI, download the portable `tai-cli.exe` executable from the [releases page](https://github.com/wilderzone/ta-installer/releases/latest).
_This executable can be bundled with your application. See the [GUI](./gui) package for an example of this kind of integration._

Then run TAI, passing a subcommand and options:
```sh
tai-cli.exe <subcommand> --repo=<url> --path=<path>
```

### Subcommands
- _For all subcommands, `--repo` is optional. TAI will retrieve files from the [Wilderzone CDN](https://cdn.wilderzone.live) by default._  
- _Commands such as `verify` may prompt for confirmation. Pass the `-y` flag to skip these prompts._  
- _`install` and `list` accept `--include` and `--exclude` lists of comma-separated values to filter assets by type (e.g. `dlc`) or exact name (e.g. `"Map Pack"`). Include is applied first, then exclude. Names containing spaces must be quoted, and names cannot contain commas._  
- _Assets can depend on other assets, and TAI manages those relationships for you: `install` also installs whatever the selected assets require, and `uninstall` also removes assets that are no longer required. See [Dependencies](#dependencies) for more information._

| Subcommand         | Description                                                                                                                                                          | Example                                                     |
| :----------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :---------------------------------------------------------- |
| `install`          | Download and install TA into the given directory. Install a subset of assets with `--include`/`--exclude` (see above). Any assets the selection requires are installed too. | `tai-cli.exe install --path=<path> --exclude=dlc`               |
| `check`            | Check for an existing installation in the given directory.                                                                                                           | `tai-cli.exe check --path=<path>`                               |
| `check-repo`       | Check that the provided URL points to a valid repository.                                                                                                            | `tai-cli.exe check-repo --repo=<url>`                           |
| `check-updates`    | Check for updates to an existing TA installation.                                                                                                                    | `tai-cli.exe check-updates --path=<path> --repo=<url>`          |
| `verify`           | Verify an existing TA installation in the given directory. Also checks the bundled installation utilities against checksums fetched from the CDN.                    | `tai-cli.exe verify --path=<path> --repo=<url>`                 |
| `update`           | Update an existing TA installation.                                                                                                                                  | `tai-cli.exe update --path=<path> --repo=<url>`                 |
| `uninstall`        | Remove a TA installation. Remove only a subset with `--remove`/`--keep`. Assets that require a removed asset are removed too. A full uninstall also removes the Start Menu and Desktop shortcuts. | `tai-cli.exe uninstall --path=<path> --remove=dlc`              |
| `asset`            | Generate an asset file with the given name from the given path. The `<name>.json` file will be exported at the root of `<path>`.                                     | `tai-cli.exe asset --path=<path> --name=<name>`                 |
| `manifest`         | Generate a manifest file from a collection of asset files at the given path. The `manifest.json` file will be exported at the root of `<path>`.                      | `tai-cli.exe manifest --path=<path>`                            |
| `list`             | List the contents of a manifest file in the given path or repo. Results can be filtered by type or name with `--include=<a,b,...>` and `--exclude=<a,b,...>`         | `tai-cli.exe list --path=<path>` or `tai-cli.exe list --repo=<url>` |


### Flags
| Flag               | Options                                             | Description                                                                                                     |
| :----------------- | :-------------------------------------------------- | :-------------------------------------------------------------------------------------------------------------- |
| `--path`           | An absolute directory path.                         | The directory in which to install the game, or where an installation is already located.                        |
| `--repo`           | A URL. _Defaults to the Wilderzone CDN if omitted._ | The repository from which to download the game.                                                                 |
| `--shortcuts`      | `true` or `false`. _Defaults to `true` if omitted._ | Whether to create Start Menu and Desktop shortcuts when installing the game.                                    |
| `--include`        | A comma-separated list of asset names or types.     | Filter the results of the `install` and `list` subcommand by including only the provided assets. Assets that the provided assets require are installed as well. |
| `--exclude`        | A comma-separated list of asset names or types.     | Filter the results of the `install` and `list` subcommand by excluding all of the provided assets. Excluding an asset that another selected asset requires is an error. |
| `--remove`         | A comma-separated list of asset names or types.     | Uninstall only the provided assets. Assets that require them are uninstalled as well.                           |
| `--keep`           | A comma-separated list of asset names or types.     | Uninstall everything except the provided assets.                                                                |
| `-m`               | _NA_                                                | Output machine-readable JSON instead of plain-text messages.                                                    |
| `-p`               | _NA_                                                | Report the progress of a download as JSON lines on stdout. See [Download progress](#download-progress).         |
| `-s`               | _NA_                                                | Run silently.                                                                                                   |
| `-v`               | _NA_                                                | Enable verbose logging of otherwise-hidden errors (such as failed best-effort steps) to stderr, for debugging.  |
| `-y`               | _NA_                                                | Automatically approve confirmation prompts.                                                                     |


### Output

By default, TAI will print human-readable output to the console. For example:
```
> tai-cli.exe verify --path="G:/tribes"

Reading local manifest from: G:\tribes\manifest.json
[1/34] Prerequisites: Verified
[2/34] Core: Verified
[3/34] Engine: Verified
⠋ [3/34] Verifying...
```
While a slow step is running, an animated spinner is shown on the last line, so it's clear that TAI is still working during long pauses. It is only drawn when the output is an interactive terminal; when the output is piped or redirected, or when `-m` or `-s` is passed, it is omitted entirely.


### Interrupting a run

Pressing `Ctrl+C` stops a running subcommand at its next safe point, rather than killing it mid-write. TAI reports what was interrupted and which subcommand to run to pick up where it left off, and exits with a non-zero status. Pressing `Ctrl+C` a second time ends the process immediately.

A parent process that has no console of its own (for example, TAI GUI) can't send an interrupt signal on Windows, so it asks the same way by writing the line `cancel` to TAI's standard input. This is only listened for when standard input is a pipe (not a terminal), so that it can't accidentally swallow the answers to confirmation prompts.

Passing the `-m` flag will instruct TAI to output machine-readable JSON instead. For example:
```
> tai-cli.exe verify -m --path="G:/tribes"

{"type":"READING_LOCAL_MANIFEST","message":"Reading local manifest from: G:\\tribes\\manifest.json"}
{"type":"PROGRESS","index":1,"total":34,"message":"Prerequisites: Verified"}
{"type":"PROGRESS","index":2,"total":34,"message":"Core: Verified"}
{"type":"PROGRESS","index":3,"total":34,"message":"Engine: Verified"}
...
```

<details>
<summary>Machine-readable message types</summary>

- `ASSET_CREATED`
- `ASSETS_ARE_DAMAGED`
- `CHECKING_ASSETS`
- `CHECKING_FOR_UPDATES`
- `DOWNLOAD_COMPLETE`
- `DOWNLOAD_PROGRESS`
- `DOWNLOAD_STARTED`
- `DOWNLOADING_MANIFEST`
- `ERROR`
- `INSTALL_FAILURE`
- `INSTALL_FOUND`
- `INSTALL_INTERRUPTED`
- `INSTALL_NOT_FOUND`
- `INSTALL_PARTIAL`
- `INSTALL_SUCCESS`
- `INSTALL_UP_TO_DATE`
- `INSTALL_VERIFIED`
- `INSTALLING`
- `INSTALLING_DEPENDENCIES`
- `INTERRUPTED`
- `MANIFEST_CONTENTS`
- `MANIFEST_CREATED`
- `MANIFEST_UNREADABLE`
- `PATCHING_ASSETS`
- `PREREQUISITES_DOWNLOADING`
- `PREREQUISITES_FAILED`
- `PREREQUISITES_INSTALLING`
- `PREREQUISITES_REQUIRED`
- `PROGRESS`
- `PROMPT_PROCEED`
- `PROMPT_REMOVE_DEPENDENTS`
- `PROMPT_UNINSTALL_ANYWAY`
- `READING_LOCAL_MANIFEST`
- `READING_MANIFEST_FAILED`
- `REDOWNLOADING`
- `REPO_FOUND`
- `REPO_NOT_FOUND`
- `RUNTIME_CHECK_FAILED`
- `RUNTIME_CHECK_SKIPPED`
- `RUNTIME_DAMAGED`
- `RUNTIME_DOWNLOADING`
- `RUNTIME_MISSING`
- `RUNTIME_PREPARE_FAILED`
- `RUNTIME_PREPARING`
- `RUNTIME_READY`
- `RUNTIME_REPAIRED`
- `RUNTIME_VERIFIED`
- `SHORTCUTS_FAILURE`
- `SHORTCUTS_REMOVE_FAILURE`
- `SHORTCUTS_SUCCESS`
- `UNINSTALL_DEPENDENTS`
- `UNINSTALL_FAILURE`
- `UNINSTALL_INTERRUPTED`
- `UNINSTALL_MATCHED_ALL`
- `UNINSTALL_ORPHANS`
- `UNINSTALL_PARTIAL`
- `UNINSTALL_SUCCESS`
- `UNINSTALLING`
- `UPDATE_FAILURE`
- `UPDATE_FOUND`
- `UPDATE_INTERRUPTED`
- `UPDATE_PARTIAL`
- `UPDATE_SUCCESS`
- `UTILITIES_CHECK_SKIPPED`
- `UTILITIES_DAMAGED`
- `UTILITIES_FAILURE`
- `UTILITIES_REPAIRED`
- `UTILITIES_REPAIR_DEFERRED`
- `UTILITIES_SUCCESS`
- `UTILITIES_VERIFIED`
- `VERIFY_INTERRUPTED`
</details>


### Download progress

Assets are downloaded in batches and only logged as each one finishes, so a large download can sit on the same line for a long time. In a terminal, the spinner fills that gap with the bytes as they arrive:
```
> tai-cli.exe install --path="G:/tribes"

Installing...
[1/34] Prerequisites: Downloaded
⠋ Downloading... 42% of 2.8 GB - 4 minutes left
```

An interface driving TAI has no spinner to read, so passing the `-p` flag instructs TAI to produce JSON-formatted progress updates instead. A download declares the total size, reports the number of bytes that have arrived several times per second, and declares when it is complete:
```
> tai-cli.exe install -p --path="G:/tribes"
...
{"size":2782785536,"message":"","type":"DOWNLOAD_STARTED"}
{"bytes":15728640,"size":2782785536,"rate":6553600,"message":"","type":"DOWNLOAD_PROGRESS"}
{"bytes":32505856,"size":2782785536,"rate":6684672,"message":"","type":"DOWNLOAD_PROGRESS"}
...
{"bytes":2782785536,"size":2782785536,"message":"","type":"DOWNLOAD_COMPLETE"}
```

`bytes` is what has arrived so far, `size` what is expected in total, and `rate` a smoothed average of the bytes per second. These lines are always JSON, with or without `-m`. A manifest that doesn't record the size of every asset it is downloading can't be measured, and reports nothing rather than an inaccurate total.


## Development

### Dependencies

Some assets don't function until other assets are present. An asset file declares this in
its `dependencies` array, which holds the exact names of the assets it requires:

```json
{
	"name": "Arx Novena",
	"type": "dlc",
	"dependencies": ["Core", "Map Textures"],
	...
}
```

The CLI's `manifest` subcommand validates these relationships when a repository is built, so a
misspelled name or a circular dependency is caught before the manifest is published.

TAI resolves the graph automatically:
- `install` also installs everything the selected assets require. A required asset that is already installed is left alone rather than downloaded again.
- Because `--exclude` is an explicit instruction, excluding an asset that the selection requires raises an error, instead of silently overriding.
- `uninstall` also removes anything that requires the assets being removed, after asking for confirmation.
- `uninstall` then prunes assets that TAI installed only to satisfy a dependency, if they are no longer required. Assets that were asked for by name, and foundational `core` and `prerequisite` assets, are never pruned this way.

The local `manifest.json` records this by marking dependency-installed assets with `"automatic": true` (repository manifests don't carry this field).

### Running on Linux

Tribes: Ascend is a 32-bit Windows game, so a Linux installation includes a compatibility layer alongside the game files. TAI manages all of it automatically, including validating the compatibility layer when you run `verify`.

A few things differ from Windows by necessity:

- Shortcuts are desktop entries that are placed in the applications menu, on the desktop, and in the installation directory, and they run the `play` utility rather than the game directly - so a shortcut written by an older install keeps working when the way the game is launched changes. The uninstall entry stands in for the Add/Remove Programs listing.
- The default install path is under your home directory.
- DirectX 9 is served by the compatibility layer's own Direct3D 9 implementation, rather than installed system-wide.
- The compatibility layer's container image is fetched by umu**, into `~/.local/share/umu`, the first time a Wine prefix is created. That download is outside TAI's own progress reporting, so it appears as "Preparing the compatibility layer".
- TA and TAMods settings live inside the Wine prefix, under `Prefix/pfx/drive_c/users/`, because that's where the game looks for its Documents folder. Uninstalling removes them along with everything else, whereas on Windows they sit in your real Documents folder and survive an uninstall. Copy that directory out first if you want to keep your game settings.

`install` reports anything about the machine that would get in the way (eg, a missing Vulkan driver, no `python3`, or unprivileged user namespaces being disabled) before committing to a multi-gigabyte download. None of them stops an install.

The runtime is published separately from `manifest.json`, alongside the installation utilities and for the same reason the DirectX 9 redistributable is: the manifest describes the *game*, and is served by whatever repository `--repo` names. A custom repository shouldn't have to mirror half a gigabyte of Proton for a Linux install to work, so the runtime always comes from the official CDN regardless of `--repo`.

Publish it as a `.tar.gz` rather than a `.zip` (the runtime carries symlinks and execute bits that a zip either can't represent or loses when packaged on Windows) and generate its checksum file the same way as the utilities':

```sh
tai-cli asset --path=<extracted runtime> --name=Runtime   # produces Runtime.json
```

Publish both `Runtime.tar.gz` and `Runtime.json` under `/runtime/` on the CDN. `verify` fetches the checksums, and replaces the whole runtime if any file fails - which is also how a newly pinned Proton build reaches existing installations.

### Installation utilities

Alongside the game, TAI installs a small set of helper executables into the `Installer` directory.

| Utility     | Windows | Linux | Purpose                                                                                                                           |
| :---------- | :------ | :---- | :-------------------------------------------------------------------------------------------------------------------------------- |
| `tai-cli`   | ✓       | ✓     | A local copy of TAI CLI.                                                                                                          |
| `play`      | ✓       | ✓     | Launches the game. All shortcuts point at this utility instead of the game's executable, so any changes to the game's launch behaviour or compatibility layer can be easily managed between updates. |
| `uninstall` | ✓       |       | Allows the game to be uninstalled via the Control Panel or the Settings menu. On Linux this action is handled by desktop entries. |
| `update`    | ✓       |       | Allows the game to be updated via a simple double-click. On Linux this action is handled by desktop entries.                      |
| `verify`    | ✓       |       | Allows the game files to be verified via a simple double-click. On Linux this action is handled by desktop entries.               |

Build these utilities automatically by running the `build.sh` script in the `/utilities` directory of this repo.
```sh
./utilities/build.sh
```

<details>
<summary>Manual steps</summary>

On Linux, `icon.png` is also included in this directory. Desktop entries can't read icons out of a Windows executable, whereas a Windows shortcut can. It is extracted from `TribesAscend.exe` at release time:

```sh
wrestool -x --type=group_icon --output=tribes.ico Binaries/Win32/TribesAscend.exe
icotool --list tribes.ico          # Check what icon sizes exist in the executable.
icotool --extract --index=1 --output=icon.png tribes.ico
```

The icon size doesn't matter; `install` reads the PNG's dimensions and files it under the matching `hicolor` size directory, it just needs to be a PNG. `verify` checks these files too, and offers to repair any that are damaged (re-downloading the package and, for an executable that's currently running, replacing it after the `verify` window is closed).

Their checksums are **not** compiled into `tai-cli.exe` - that would be circular, since the checksum of `tai-cli.exe` can't be known until after it's built. Instead they're published on the CDN and fetched at verify time. When cutting a release, after building the executables into a directory, generate the checksum file and publish it next to `Installer.zip`:

```sh
tai-cli.exe asset --path=<installer build dir> --name=Installer
```

Publish both `Installer.zip` and the generated `Installer.json` to the repository root. Repeat for Linux, publishing `Installer-linux-amd64.zip` and its `.json`. The zip must be built **on Linux** so `tai-cli` and `play` keep their execute bits.

Build `play.exe` for Windows, configured so starting the game doesn't flash up a console window:

```sh
GOOS=windows go build -ldflags="-H=windowsgui" -o play.exe ./utilities/play
```

The `play` utility writes to `Logs/launcher.txt` inside the installation, replacing it on each launch. The log includes output from the compatibility layer as well.
The other three utilities print their output to the console window they were launched from.

</details>

### Building

The GUI embeds the CLI, and the embedded binary is gitignored. A fresh repository clone can't compile the Rust crate until it has been built:

```sh
cd gui
npm ci
npm run build-sidecar        # cross-compile with GOOS=linux / GOOS=windows
npm run tauri build          # `pretauri` runs build-sidecar for you
```

Building the GUI on Linux needs the WebKitGTK stack: `libwebkit2gtk-4.1-dev`, `libgtk-3-dev`, `librsvg2-dev`, `libdbus-1-dev`, `patchelf`, and `rpm` for the `.rpm` bundle. CI builds both platforms on every push.
