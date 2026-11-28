---
order: 2
---

# Options

Options can be used to gain fine-grained control over the behaviour of TAI CLI. Options start with `--` (double hyphen), and typically take an argument in the form of `--option="argument"`. Whether an option is required or not depends on the provided [subcommand](/cli/subcommands).

| Option              | Argument                                            | Description                                                                                                     |
| :------------------ | :-------------------------------------------------- | :-------------------------------------------------------------------------------------------------------------- |
| `--path`            | An absolute or relative directory path, in either Windows or Unix style. | The directory in which to install the game, or where an installation is already located.                        |
| `--repo`            | A URL. _Defaults to the [Wilderzone CDN](https://cdn.wilderzone.org/ta/game/manifest.json) if omitted._ | The repository from which to download the game, or check for updates against. |
| `--shortcuts`       | `true` or `false`. _Defaults to `true` if omitted._ | Whether to create Desktop shortcuts when installing the game. Start Menu (or applications menu) shortcuts are always created. |
| `--include`         | A comma-separated list of asset names or types.     | Filter the results of the `install` and `list` subcommand by including only the provided assets. Assets that the provided assets depend on will be included as well. |
| `--exclude`         | A comma-separated list of asset names or types.     | Filter the results of the `install` and `list` subcommand by excluding all of the provided assets. Excluding an asset that another selected asset depends on raises an error. |
| `--remove`          | A comma-separated list of asset names or types.     | Uninstall only the provided assets. Dependents of the provided assets will be uninstalled as well.              |
| `--keep`            | A comma-separated list of asset names or types.     | Uninstall everything except the provided assets.                                                                |
| `--remove-settings` | _None._                                             | When performing a complete uninstall, also remove the game's settings and logs from `Documents\My Games\Tribes Ascend`. This option can't be combined with `--remove` or `--keep`. |
| `--stop-if-short`   | _None._                                             | Stop before downloading anything if there may not be enough free space, rather than prompting for confirmation (even with `-y`). For use with interfaces that have their own prompting logic. |


# Flags

Flags can be used for course control of TAI CLI. Flags start with `-` (single hyphen), and take no arguments. All flags are optional.

| Flag               | Description                                                                                                          |
| :----------------- | :------------------------------------------------------------------------------------------------------------------- |
| `-m`               | Output machine-readable JSON instead of plain-text messages.                                                         |
| `-p`               | Report the progress of a download as JSON lines on `stdout`. See [Download Progress](/cli/output#download-progress). |
| `-s`               | Run silently.                                                                                                        |
| `-v`               | Enable verbose logging of otherwise-hidden errors (such as failed best-effort steps) to `stderr`.                    |
| `-y`               | Automatically approve confirmation prompts.                                                                          |
