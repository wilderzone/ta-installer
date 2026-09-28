# Installation Utilities

Alongside the game, TAI installs a small set of helper executables into the `Installer` directory.

| Utility     | Windows | Linux | Purpose                                                                                                                           |
| :---------- | :------ | :---- | :-------------------------------------------------------------------------------------------------------------------------------- |
| `tai-cli`   | ✓       | ✓     | A local copy of TAI CLI.                                                                                                          |
| `play`      | ✓       | ✓     | Launches the game. All shortcuts point at this utility instead of the game's executable, so any changes to the game's launch behaviour or compatibility layer can be easily managed between updates. |
| `uninstall` | ✓       |       | Allows the game to be uninstalled via the Control Panel or the Settings menu. On Linux this action is handled by desktop entries. |
| `update`    | ✓       |       | Allows the game to be updated via a simple double-click. On Linux this action is handled by desktop entries.                      |
| `verify`    | ✓       |       | Allows the game files to be verified via a simple double-click. On Linux this action is handled by desktop entries.               |

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

Publish both `Installer.zip` and the generated `Installer.json` to the repository root. Repeat for Linux, publishing `Installer-linux-amd64.zip` and its `.json`. The zip must be built **on Linux** so `tai-cli` and `play` keep their execute bits.

Build `play.exe` for Windows, configured so starting the game doesn't flash up a console window:

```sh
GOOS=windows go build -ldflags="-H=windowsgui" -o play.exe ./utilities/play
```

The `play` utility writes to `Logs/launcher.txt` inside the installation, replacing it on each launch. The log includes output from the compatibility layer as well.
The other three utilities print their output to the console window they were launched from.

</details>
