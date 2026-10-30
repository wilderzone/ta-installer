---
order: 3
---

# Installation Utilities

Alongside the game, TAI installs a small set of helper executables into the `Installer` directory.

| Utility     | Windows | Linux | Purpose                                                                                                                           |
| :---------- | :------ | :---- | :-------------------------------------------------------------------------------------------------------------------------------- |
| `tai-cli`   | ✓       | ✓     | A local copy of TAI CLI.                                                                                                          |
| `play`      | ✓       | ✓     | Launches the game, after [checking for updates](#checking-for-updates), with [TAMods](#tamods) if it's installed. All shortcuts point at this utility instead of the game's executable, so any changes to the game's launch behaviour or compatibility layer can be easily managed between updates. |
| `uninstall` | ✓       | ✓     | Uninstalls the game. On Windows it's run from the Control Panel or the Settings menu; on Linux, from the applications menu or the game's right-click actions. |
| `update`    | ✓       | ✓     | Updates the game. On Windows, double-click it; on Linux, use the game's right-click actions or the entry in the installation folder. |
| `verify`    | ✓       | ✓     | Verifies the game's files, and repairs any that are damaged. It's run the same way as `update`.                                   |

## Dialogs

`uninstall`, `update`, and `verify` each run one TAI CLI [subcommand](/cli/subcommands) against the installation they live in, and show their progress and result in native dialogs instead of a terminal window. They're built from a shared Go module in `/utilities`, and use [zenity](https://github.com/ncruces/zenity) for the dialogs.

- **Progress:** A progress dialog follows the CLI's [machine-readable output](/cli/output), showing each step and the progress of any download. Pressing **Cancel** asks the CLI to stop at its next safe point (see [Interrupting a Run](/cli/output#interrupting-a-run)), and the result explains how to pick up where it left off.
- **Result:** Once the CLI finishes, a dialog reports how it went: a success, a success with warnings (eg, the shortcuts couldn't be created), a partial result, or an error. **Show Details** opens the run's full log, including the CLI's verbose output.
- **Logs:** `update` and `verify` log to `Logs/update.txt` and `Logs/verify.txt` inside the installation, replacing the log on each run. `uninstall` logs to `tai-uninstall.txt` in the temp directory instead, since it removes the installation.
- **Questions:** The CLI is always run with `-y`, so it never prompts. Anything worth asking is asked by the utility, before the CLI starts. The exception is a lack of free space, which `update` and `verify` only discover once they've checked the files. They're run with `--stop-if-short`, and if they stop, the utility asks whether to continue anyway, then runs them again (checking the files over again first).
- **Silent mode:** Passing `-s` runs a utility without any dialogs or questions. Windows uses this for a quiet uninstall.
- **Exit code:** A utility exits with a non-zero status if the run failed, only partly succeeded, or was interrupted or declined. The CLI exits cleanly from some of these, so its own exit code isn't enough to go on.

### Uninstalling

`uninstall` runs the CLI's `check` subcommand first, so its confirmation can say exactly what will happen: the installation folder, and everything in it, will be permanently deleted. If the installation's manifest can't be read, the question also warns that there's no telling which of the files belong to the game. Cancel is the default answer.

The CLI removes everything except the utilities directory it runs from. `uninstall` then removes that directory, and the installation folder around it - but only after a complete run, and only when nothing else is left in the folder. Anything the CLI couldn't remove is left where it is, along with the folder.

### Checking for Updates

Before it starts the game, `play` checks for an update, behind a "Checking for updates..." dialog that can skip the check. If there is one, it asks whether to install it first:
- **Update** installs it behind the same progress dialog as `update`, then starts the game. If the update doesn't complete, `play` explains what went wrong, and offers **Play Anyway** or **Close**. If there may not be enough free space, it asks whether to continue anyway first; **Cancel** starts the game on its current version, since nothing has been downloaded.
- **Play Now** starts the game straight away.

Nothing about the check stands in the way of playing. Without a connection, when the check fails or takes longer than 10 seconds, or on Linux without a way to show dialogs, the game starts as it would have without it, and the reason is written to `Logs/launcher.txt`. Like `update`, the check compares the installation with the default repository. `play -prepare`, which `install` runs on Linux to set up the compatibility layer, never checks.

### TAMods

If the TA Mods asset is installed, `play` loads TAMods into the game the same way T:A Launcher does. Before the game starts, a copy of `Mods/TAMods/TAMods.dll` takes the place of each of the game's `FonixTtsDtSimple*.dll` text-to-speech libraries in `Binaries/Win32`, and the game's own copies are set aside beside them as `.tai-original` files. Once the game exits, they're put back, so `verify` finds the installation as it was installed. If `play` doesn't get to put them back (eg, the computer lost power), it does so before the next launch. If TAMods can't be applied, the game starts without it, and the reason is written to `Logs/launcher.txt`.

`play` waits for the game to exit so it can put the DLLs back, and only `play` applies TAMods: the game started any other way (eg, by running `TribesAscend.exe` directly) runs without it.

### On Linux

The game's desktop entry offers each utility as an action: right-click the game in the applications menu or the dock for **Verify Files**, **Check for Updates**, and **Uninstall**. The installation folder also holds a desktop entry for each utility, and the "Uninstall Tribes: Ascend" entry in the applications menu stands in for Add/Remove Programs.

The dialogs need one of `zenity`, `qarma`, or `matedialog` to be installed, which most desktops include. Without one, a utility reopens itself in a terminal, where it shows the CLI's output and the CLI asks its own questions.

## Building

Build these utilities automatically by running the `build.sh` script in the `/utilities` directory of this repo. This script must be run on Linux.
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

Publish both `Installer.zip` and the generated `Installer.json` under `/installer/` in the repository, renamed to `Installer-windows-amd64.zip` and `Installer-windows-amd64.json`. Repeat for Linux, publishing `Installer-linux-amd64.zip` and its `.json`. The zip must be built **on Linux** so the executables keep their execute bits. A [release](/development/building#releasing) builds and publishes both packages in CI.

Build `play.exe`, `uninstall.exe`, `update.exe`, and `verify.exe` for Windows as GUI programs, so that starting the game, or showing a utility's dialogs, doesn't open a console window as well. They're all built from the `/utilities` module:

```sh
cd utilities
GOOS=windows go build -ldflags="-H=windowsgui" -o play.exe ./play
GOOS=windows go build -ldflags="-H=windowsgui" -o verify.exe ./verify   # And likewise for uninstall and update.
```

The `play` utility writes to `Logs/launcher.txt` inside the installation, replacing it on each launch. The log includes output from the compatibility layer as well.
The other three utilities keep the logs described in [Dialogs](#dialogs).

</details>
