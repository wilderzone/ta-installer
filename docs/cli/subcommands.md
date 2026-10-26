---
order: 1
---

# Subcommands

The subcommand provided to TAI CLI determines the action to perform. Subcommands must always be passed as the first argument to TAI CLI. The subcommand is then followed by [options](/cli/options) and [flags](/cli/options#flags).

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

Notes:
- _For all subcommands, `--repo` is optional. TAI will retrieve files from the [Wilderzone CDN](https://cdn.wilderzone.org) by default._  
- _Subcommands such as `verify` may prompt for confirmation. Pass the `-y` flag to skip these prompts._  
- _`install` and `list` accept `--include` and `--exclude` lists of comma-separated values to filter assets by type (e.g. `dlc`) or exact name (e.g. `"Arx Novena"`). Include is applied first, then exclude. Names containing spaces must be quoted, and names cannot contain commas._  
- _Assets can depend on other assets. TAI manages these relationships automatically: `install` also installs whatever the selected assets require, and `uninstall` also removes assets that are no longer required. See [Dependency Resolution](/cli/dependency-resolution) for more information._
