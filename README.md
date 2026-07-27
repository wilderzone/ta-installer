# <img src="./gui/public/logo.svg" style="width: 3em;"> TA Installer
A [CLI](https://en.wikipedia.org/wiki/Command-line_interface) and [GUI](https://en.wikipedia.org/wiki/Graphical_user_interface) utility for managing Tribes: Ascend installations.


## GUI

To use TAI GUI, download and run the portable `tai-gui.exe` executable from the [releases page](https://github.com/wilderzone/ta-installer/releases/latest).
You'll be prompted to select the folder in which to install TA, and the repository to download the game files from. Leaving the repository option as default is a good idea if you just want to download the vanilla game.

_Note: `tai-gui.exe` requires [VC Redist](https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist) and [WebView2](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) to be installed on your computer. Whilst it's very likely you already have them, it's possible you will need to install them before running TAI._

<p align="center">
  <img width="582" height="682" alt="image" src="https://github.com/user-attachments/assets/70cdef8a-983a-46ed-bbe8-b1598c200346" />
</p>

## CLI

TAI CLI is provided for use in applications that require programmatic control of the TA installation process, such as launchers.  
To use TAI CLI, download the portable `tai-cli.exe` executable from the [releases page](https://github.com/wilderzone/ta-installer/releases/latest).
This executable can be bundled with your application. See the [GUI](./gui) package for an example of this kind of integration.

Then run TAI, passing a subcommand and options:
```sh
tai-cli.exe <subcommand> --repo=<url> --path=<path>
```

### Subcommands
_For all subcommands, `--repo` is optional. TAI will retrieve files from the [Wilderzone CDN](https://cdn.wilderzone.live) by default._  
_Commands such as `verify` may prompt for confirmation. Pass the `-y` flag to skip these prompts._  
_`install` and `list` accept `--include` and `--exclude` lists of comma-separated values to filter assets by type (e.g. `dlc`) or exact name (e.g. `"Map Pack"`). Include is applied first, then exclude. Names containing spaces must be quoted, and names cannot contain commas._  
_Assets can depend on other assets, and TAI manages those relationships for you: `install` also installs whatever the selected assets require, and `uninstall` also removes assets that are no longer required. See [Dependencies](#dependencies) for more information._

| Subcommand         | Description                                                                                                                                                          | Example                                                     |
| :----------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :---------------------------------------------------------- |
| `install`          | Download and install TA into the given directory. Install a subset of assets with `--include`/`--exclude` (see above). Any assets the selection requires are installed too. | `tai-cli.exe install --path=<path> --exclude=dlc`               |
| `check`            | Check for an existing installation in the given directory.                                                                                                           | `tai-cli.exe check --path=<path>`                               |
| `check-repo`       | Check that the provided URL points to a valid repository.                                                                                                            | `tai-cli.exe check-repo --repo=<url>`                           |
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
| `-s`               | _NA_                                                | Run silently.                                                                                                   |
| `-v`               | _NA_                                                | Enable verbose logging of otherwise-hidden errors (such as failed best-effort steps) to stderr, for debugging.  |
| `-y`               | _NA_                                                | Automatically approve confirmation prompts.                                                                     |


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

The `manifest` subcommand validates these relationships when a repository is built, so a
misspelled name or a circular dependency is caught before the manifest is published.

TAI resolves the graph automatically:
- `install` also installs everything the selected assets require. A required asset that is already installed is left alone rather than downloaded again.
- Because `--exclude` is an explicit instruction, excluding an asset that the selection requires raises an error, instead of silently overriding.
- `uninstall` also removes anything that requires the assets being removed, after asking for confirmation.
- `uninstall` then prunes assets that TAI installed only to satisfy a dependency, if they are no longer required. Assets that were asked for by name, and foundational `core` and `prerequisite` assets, are never pruned this way.

The local `manifest.json` records this by marking dependency-installed assets with `"automatic": true` (repository manifests don't carry this field).


### Installation utilities

Alongside the game, TAI installs a small set of helper executables (`tai-cli.exe`, `uninstall.exe`, `update.exe`, `verify.exe`) into the `Installer` directory, so Windows can uninstall the game through the normal Settings/Control Panel channel. `verify` checks these files too, and offers to repair any that are damaged (re-downloading the package and, for an executable that's currently running, replacing it after the window is closed).

Their checksums are **not** compiled into `tai-cli.exe` — that would be circular, since the checksum of `tai-cli.exe` can't be known until after it's built. Instead they're published on the CDN and fetched at verify time. When cutting a release, after building the four executables into a directory, generate the checksum file and publish it next to `Installer.zip`:

```sh
tai-cli.exe asset --path=<installer build dir> --name=Installer   # produces Installer.json
```

Publish both `Installer.zip` and the generated `Installer.json` to the repository root.


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
- `DOWNLOADING_MANIFEST`
- `ERROR`
- `INSTALL_FAILURE`
- `INSTALL_FOUND`
- `INSTALL_NOT_FOUND`
- `INSTALL_PARTIAL`
- `INSTALL_SUCCESS`
- `INSTALL_UP_TO_DATE`
- `INSTALL_VERIFIED`
- `INSTALLING`
- `INSTALLING_DEPENDENCIES`
- `MANIFEST_CONTENTS`
- `MANIFEST_CREATED`
- `MANIFEST_UNREADABLE`
- `PATCHING_ASSETS`
- `PREREQUISITES_DOWNLOADING`
- `PREREQUISITES_FAILED`
- `PREREQUISITES_INSTALLING`
- `PREREQUISITES_REQUIRED`
- `PROMPT_PROCEED`
- `PROMPT_REMOVE_DEPENDENTS`
- `PROMPT_UNINSTALL_ANYWAY`
- `READING_LOCAL_MANIFEST`
- `READING_MANIFEST_FAILED`
- `REDOWNLOADING`
- `REPO_FOUND`
- `REPO_NOT_FOUND`
- `SHORTCUTS_CREATING`
- `SHORTCUTS_FAILURE`
- `SHORTCUTS_REMOVE_FAILURE`
- `SHORTCUTS_SUCCESS`
- `UNINSTALL_DEPENDENTS`
- `UNINSTALL_FAILURE`
- `UNINSTALL_MATCHED_ALL`
- `UNINSTALL_ORPHANS`
- `UNINSTALL_PARTIAL`
- `UNINSTALL_SUCCESS`
- `UNINSTALLING`
- `UPDATE_FAILURE`
- `UPDATE_FOUND`
- `UPDATE_PARTIAL`
- `UPDATE_SUCCESS`
- `UTILITIES_CHECK_SKIPPED`
- `UTILITIES_DAMAGED`
- `UTILITIES_FAILURE`
- `UTILITIES_REPAIRED`
- `UTILITIES_REPAIR_DEFERRED`
- `UTILITIES_SUCCESS`
- `UTILITIES_VERIFIED`
- `UTILITIES_VERIFYING`
- `PROGRESS`
</details>
