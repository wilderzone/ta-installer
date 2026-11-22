---
order: 3
---

# Installation Utilities

Alongside the game, TAI installs a small set of helper executables into the `Installer` directory.

| Utility     | Purpose                                                                                                                               |
| :---------- | :------------------------------------------------------------------------------------------------------------------------------------ |
| `tai-cli`   | A local copy of TAI CLI.                                                                                                              |
| `play`      | Launches the game, after checking for updates. All launch shortcuts point at this utility instead of the game's executable, so any changes to the game's launch behaviour or compatibility layer can be easily managed between updates. |
| `uninstall` | Uninstalls the game. On Windows, it's run from the Control Panel or the Settings menu. On Linux, it's run from the applications menu or the game's right-click actions. |
| `update`    | Updates the game. On Linux, this is included in the game's right-click actions.                                                       |
| `verify`    | Verifies the game's files, and repairs any that are damaged or missing. On Linux, this is included in the game's right-click actions. |

`uninstall`, `update`, and `verify` each run one subcommand of the local CLI, and show their progress and result in native dialogs (via [zenity](https://github.com/ncruces/zenity)) instead of terminal windows:
- A progress dialog follows the CLI's machine-readable output. Pressing "Cancel" asks the CLI to stop at its next safe point.
- A result dialog reports how the run went. "Show Details" opens the run's log: `Logs/update.txt` or `Logs/verify.txt` inside the installation. `uninstall` logs to `tai-uninstall.txt` in the temp directory, since it removes the installation.
- The CLI is always run with `-y`. Anything worth asking is asked by the utility first: `uninstall` runs `check`, so its confirmation can say exactly what will be deleted. The exception is a lack of free space, which `update` and `verify` only discover once they've checked the files: they're run with `--stop-if-short`, and if they stop, the utility asks whether to continue anyway, then runs them again.
- Passing `-s` runs a utility without any dialogs or questions. A utility exits with a non-zero status if the run failed, only partly succeeded, or was interrupted or declined.
- On Linux, the dialogs need `zenity`, `qarma`, or `matedialog`. Without one, a utility reopens itself in a terminal.


## Launching

Before it starts the game, `play` checks for an update, behind a "Checking for updates..." dialog that can skip the check. If there is one, it asks whether to install it first:
- **Update** installs it behind the same progress dialog as `update`, then starts the game. If the update doesn't complete, `play` explains what went wrong, and offers **Play Anyway** or **Close**. If there may not be enough free space, it asks whether to continue anyway first; **Cancel** starts the game on its current version, since nothing has been downloaded.
- **Play Now** starts the game straight away.

Nothing about the check should stand in the way of launching the game. Without a connection, when the check fails or takes longer than 10 seconds, or on Linux without a way to show dialogs, the game starts as it would have without it, and the reason is written to `Logs/launcher.txt`. Like `update`, the check compares the installation with the default repository. `play -prepare`, which `install` runs on Linux to set up the compatibility layer, never checks for updates.

On Windows, each of the game's shortcuts carries an AppUserModelID (`Wilderzone.TribesAscend` or `Wilderzone.TribesAscend.PUGs`), and `play` starts the game with the ID of the shortcut it was started from. The taskbar uses the ID to tie the game's window back to the shortcut, so the game shares the shortcut's taskbar button. Pinning the game then pins the shortcut rather than `TribesAscend.exe`, which can't reach a login server by itself. If `play` is started any other way (eg, directly, or from a shortcut made by an older installer), pinning is disabled to prevent a broken pin from being created.


## Uninstalling

`uninstall` runs the CLI's `check` subcommand first, so its confirmation can say exactly what will happen: the installation folder, and everything in it, will be permanently deleted. If the installation's manifest can't be found/read, the prompt includes an additional warning (since there's no telling which of the files belong to the game). Cancel is the default answer.

`uninstall` only removes the installation folder after a complete run, and only when nothing but the utilities directory is left in it. Otherwise, it reports a partial uninstall.
