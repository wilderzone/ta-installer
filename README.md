# <img src="./gui/src/assets/logo.svg" style="width: 3em;"> TA Installer
A [CLI](https://en.wikipedia.org/wiki/Command-line_interface) and [GUI](https://en.wikipedia.org/wiki/Graphical_user_interface) utility for managing Tribes: Ascend installations.


## GUI

To use TAI GUI, download and run the portable `tai-gui.exe` executable from the [releases page](https://github.com/wilderzone/ta-installer/releases/latest).
You'll be prompted to select the game type (Vanilla, Modded, or Custom) and the folder in which to install TA. Modded is the recommended option, and includes community maps and CTF routes.

_Note: `tai-gui.exe` requires [VC Redist](https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist) and [WebView2](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) to be installed on your computer. Whilst it's very likely you already have them, it's possible you will need to install them before running TAI._

On Linux, download the `.deb` or `.rpm` package for your distribution instead. Everything else works the same way: pick a folder, pick a variant, and install. See [Running on Linux](#running-on-linux) for what TAI sets up behind the scenes.

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
| `play`             | Launch an installed game. On Linux this starts it through the compatibility layer; the shortcuts an install creates point here rather than at the game itself. | `tai-cli.exe play --path=<path>`                                |
| `asset`            | Generate an asset file with the given name from the given path. The `<name>.json` file will be exported at the root of `<path>`.                                     | `tai-cli.exe asset --path=<path> --name=<name>`                 |
| `manifest`         | Generate a manifest file from a collection of asset files at the given path. The `manifest.json` file will be exported at the root of `<path>`.                      | `tai-cli.exe manifest --path=<path>`                            |
| `list`             | List the contents of a manifest file in the given path or repo. Results can be filtered by type or name with `--include=<a,b,...>` and `--exclude=<a,b,...>`         | `tai-cli.exe list --path=<path>` or `tai-cli.exe list --repo=<url>` |


### Flags
| Flag               | Options                                             | Description                                                                                                     |
| :----------------- | :-------------------------------------------------- | :-------------------------------------------------------------------------------------------------------------- |
| `--path`           | An absolute directory path.                         | The directory in which to install the game, or where an installation is already located.                        |
| `--repo`           | A URL. _Defaults to the Wilderzone CDN if omitted._ | The repository from which to download the game.                                                                 |
| `--shortcuts`      | `true` or `false`. _Defaults to `true` if omitted._ | Whether to create Start Menu and Desktop shortcuts when installing the game.                                    |
| `--platform`       | `windows` or `linux`. _Defaults to every platform if omitted._ | Restrict a generated asset to one platform, for the `asset` subcommand. An asset with a platform is ignored everywhere else. |
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
- `LAUNCHING`
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
- `RUNTIME_DOWNLOADING`
- `RUNTIME_FAILED`
- `RUNTIME_MISSING`
- `RUNTIME_PREPARING`
- `RUNTIME_READY`
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

Tribes: Ascend is a 32-bit Windows game, so a Linux installation includes a compatibility layer alongside the game files. TAI manages all of it — there is no Wine to install and no drivers to fix up:

```
~/Games/Tribes Ascend/
  Binaries/ TribesGame/ Engine/   the game, byte-for-byte the same as on Windows
  Runtime/                        a pinned Proton build, plus the umu launcher
  Prefix/                         the Wine prefix this installation runs in
  Installer/tai-cli               the CLI, which the desktop entries launch through
```

Everything lives inside the installation directory, so uninstalling remains "remove one folder", and `verify` checks the runtime the same way it checks the game.

A few things differ from Windows by necessity:

- **Shortcuts are desktop entries.** They go in the applications menu, on the desktop, and in the installation directory, and they run `tai-cli play` rather than the game directly — so a shortcut written by an older install keeps working when the way the game is launched changes. The uninstall entry stands in for the Add/Remove Programs listing.
- **Nothing is elevated.** The default install path is under your home directory, and DirectX is served by the runtime's own Direct3D 9 implementation rather than installed system-wide, so no step needs a password.
- **The container image is fetched by umu**, into `~/.local/share/umu`, the first time a prefix is created. That download is outside TAI's own progress reporting, so it appears as "Preparing the compatibility layer".

`install` reports anything about the machine that would get in the way — a missing Vulkan driver, no `python3`, or unprivileged user namespaces being disabled — before committing you to a multi-gigabyte download. None of them stops an install.

The runtime is published as an ordinary manifest asset tagged with `"platform": "linux"`. Assets that name a platform are filtered out everywhere else, and assets that name none — the game's own files — are installed on both. Generate one with:

```sh
tai-cli asset --path=<extracted runtime> --name=Runtime --platform=linux
```

Publish it as a `.tar.gz` rather than a `.zip`: the runtime carries symlinks and execute bits that a zip either can't represent or loses when packaged on Windows.

### Installation utilities

Alongside the game, TAI installs a small set of helper executables (`tai-cli.exe`, `uninstall.exe`, `update.exe`, `verify.exe`) into the `Installer` directory, so Windows can uninstall the game through the normal Settings/Control Panel channel. The Linux package contains only `tai-cli` — the other three exist to be double-clicked from Explorer, and desktop entries do that job instead. `verify` checks these files too, and offers to repair any that are damaged (re-downloading the package and, for an executable that's currently running, replacing it after the window is closed).

Their checksums are **not** compiled into `tai-cli.exe` — that would be circular, since the checksum of `tai-cli.exe` can't be known until after it's built. Instead they're published on the CDN and fetched at verify time. When cutting a release, after building the four executables into a directory, generate the checksum file and publish it next to `Installer.zip`:

```sh
tai-cli.exe asset --path=<installer build dir> --name=Installer   # produces Installer.json
```

Publish both `Installer.zip` and the generated `Installer.json` to the repository root. Repeat for Linux, publishing `Installer-linux-amd64.zip` and its `.json` — and build that zip **on Linux**, so the CLI keeps its execute bit.

### Building

The GUI embeds the CLI, and the embedded binary is gitignored, so a fresh clone can't compile the Rust crate until it has been built:

```sh
cd gui
npm ci
npm run build-sidecar        # cross-compile with GOOS=linux / GOOS=windows
npm run tauri build          # `pretauri` runs build-sidecar for you
```

Building the GUI on Linux needs the WebKitGTK stack: `libwebkit2gtk-4.1-dev`, `libgtk-3-dev`, `librsvg2-dev`, `libdbus-1-dev`, `patchelf`, and `rpm` for the `.rpm` bundle. CI builds both platforms on every push.
