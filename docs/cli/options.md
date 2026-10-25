# Options

Options can be used to gain fine-grained control over the behaviour of TAI CLI. Options start with `--` (double hyphen), and typically take an argument in the form of `--option="argument"`. Whether an option is required or not depends on the provided [subcommand](/cli/subcommands).

| Option             | Argument                                            | Description                                                                                                     |
| :----------------- | :-------------------------------------------------- | :-------------------------------------------------------------------------------------------------------------- |
| `--path`           | An absolute directory path.                         | The directory in which to install the game, or where an installation is already located.                        |
| `--repo`           | A URL. _Defaults to the Wilderzone CDN if omitted._ | The repository from which to download the game.                                                                 |
| `--shortcuts`      | `true` or `false`. _Defaults to `true` if omitted._ | Whether to create Start Menu and Desktop shortcuts when installing the game.                                    |
| `--include`        | A comma-separated list of asset names or types.     | Filter the results of the `install` and `list` subcommand by including only the provided assets. Assets that the provided assets require are installed as well. |
| `--exclude`        | A comma-separated list of asset names or types.     | Filter the results of the `install` and `list` subcommand by excluding all of the provided assets. Excluding an asset that another selected asset requires is an error. |
| `--remove`         | A comma-separated list of asset names or types.     | Uninstall only the provided assets. Assets that require them are uninstalled as well.                           |
| `--keep`           | A comma-separated list of asset names or types.     | Uninstall everything except the provided assets.                                                                |


# Flags

Flags can be used for course control of TAI CLI. Flags start with `-` (single hyphen), and take no arguments. All flags are optional.

| Flag               | Description                                                                                                        |
| :----------------- | :----------------------------------------------------------------------------------------------------------------- |
| `-m`               | Output machine-readable JSON instead of plain-text messages.                                                       |
| `-p`               | Report the progress of a download as JSON lines on stdout. See [Download progress](/cli/output#download-progress). |
| `-s`               | Run silently.                                                                                                      |
| `-v`               | Enable verbose logging of otherwise-hidden errors (such as failed best-effort steps) to stderr, for debugging.     |
| `-y`               | Automatically approve confirmation prompts.                                                                        |
