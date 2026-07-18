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

| Subcommand         | Description                                                                                                                                     | Example                                                     |
| :----------------- | :---------------------------------------------------------------------------------------------------------------------------------------------- | :---------------------------------------------------------- |
| `install`          | Download and install TA into the given directory.                                                                                               | `tai.exe install --repo=<url> --path=<path>`                |
| `check`            | Check for an existing installation in the given directory.                                                                                      | `tai.exe check --path=<path>`                               |
| `check-repo`       | Check that the provided URL points to a valid repository.                                                                                       | `tai.exe check-repo --repo=<url>`                           |
| `verify`           | Verify an existing TA installation in the given directory.                                                                                      | `tai.exe verify --path=<path>`                              |
| `update`           | Update an existing TA installation.                                                                                                             | `tai.exe update --repo=<url> --path=<path>`                 |
| `uninstall`        | Remove a TA installation from the system.                                                                                                       | `tai.exe uninstall --path=<path>`                           |
| `asset`            | Generate an asset file with the given name from the given path. The `<name>.json` file will be exported at the root of `<path>`.                | `tai.exe asset --name=<name> --path=<path>`                 |
| `manifest`         | Generate a manifest file from a collection of asset files at the given path. The `manifest.json` file will be exported at the root of `<path>`. | `tai.exe manifest --path=<path>`                            |
| `list`             | List the contents of a manifest file in the given path or repo.                                                                                 | `tai.exe list --path=<path>` or `tai.exe list --repo=<url>` |


### Output

By default, TAI will print human-readable output to the console. For example:
```
> tai.exe verify --path="G:/tribes"

Reading local manifest from: G:\tribes\manifest.json
[1/34] Prerequisites: Verified
[2/34] Core: Verified
[3/34] Engine: Verified
...
```

Passing the `-m` flag will instruct TAI to output machine-readable JSON instead. For example:
```
> tai.exe verify -m --path="G:/tribes"

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
- `INSTALL_FAILURE`
- `INSTALL_PARTIAL`
- `INSTALL_SUCCESS`
- `INSTALL_UP_TO_DATE`
- `INSTALL_VERIFIED`
- `INSTALLING_PREREQUISITES`
- `MANIFEST_CONTENTS`
- `MANIFEST_CREATED`
- `PATCHING_ASSETS`
- `PROMPT_PROCEED`
- `READING_LOCAL_MANIFEST`
- `READING_MANIFEST_FAILED`
- `REDOWNLOADING`
- `REPO_FOUND`
- `REPO_NOT_FOUND`
- `RETRIEVING_MANIFEST`
- `UNINSTALL_FAILURE`
- `UNINSTALL_PARTIAL`
- `UNINSTALL_SUCCESS`
- `UNINSTALLING`
- `UPDATE_FAILURE`
- `UPDATE_FOUND`
- `UPDATE_PARTIAL`
- `UPDATE_SUCCESS`
- `PROGRESS`
</details>
