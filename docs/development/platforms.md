---
order: 2
---

# Platforms


## Windows

Thanks to Window's famous backwards-compatibility, Tribes: Ascend should run fine on any version of Windows from XP SP3 to 11. It only requires the DX9 runtime be installed, and TAI handles this automatically.

However, TAI itself relies on more modern technologies. So when using TAI to install Tribes: Ascend, you will need at least Windows 10 (1803), with [VC Redist](https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist) and [WebView2](https://developer.microsoft.com/en-us/microsoft-edge/webview2/). If you're using Windows 11, you will likely already have both of these components.

To do development on Windows, you will also need [git for Windows](https://git-scm.com/install/windows), and the [Go](https://go.dev/) and [Rust](https://rust-lang.org/) toolchains. These are not necessary to run the distributed TAI binaries, but they are necessary if you wish to develop/compile TAI yourself.


## Linux

Tribes: Ascend is a 32-bit Windows game, so a compatibility layer is required to run it on Linux. When installed with TAI, this compatibility layer ([GE-Proton](https://github.com/GloriousEggroll/proton-ge-custom) & [umu](https://github.com/Open-Wine-Components/umu-launcher)) is included alongside the game files. Maintenance is handled automatically, including updates and validation.

A few things differ from Windows by necessity:
- The default install path is under the user's home directory.
- Shortcuts are desktop entries that are placed in the installation directory, the applications menu, and (unless `--shortcuts=false`) on the desktop. The uninstall entry stands in for the Add/Remove Programs listing, and the game's desktop entry offers the [installation utilities](/development/installation-utilities#on-linux) as right-click actions.
- DirectX 9 is served by the compatibility layer's own Direct3D 9 implementation, rather than installed system-wide.
- The compatibility layer's container image is fetched by [umu](https://github.com/Open-Wine-Components/umu-launcher), into `~/.local/share/umu`, the first time a Wine prefix is created. That download is outside TAI's own progress reporting, so it appears as "Preparing the compatibility layer".
- TA and TAMods settings live inside the Wine prefix, under `Prefix/pfx/drive_c/users/`, because that's where TA looks for its Documents folder. Uninstalling removes them along with the installation, whereas on Windows they remain in the real Documents folder and survive an uninstall.
- Before committing to a multi-gigabyte download, `install` reports anything about the machine that would prevent a successful installation (eg, a missing Vulkan driver, no `python3`, or unprivileged user namespaces being disabled). `verify` runs the same checks as well.

### Graphics Drivers

On Linux, Tribes: Ascend requires 32-bit Vulkan drivers. These must match the running Kernel, so they can't easily be supplied by the compatibility layer. Without the correct drivers the game exits silently during startup. TAI checks for this case and prints the exact commands for your distribution, but can't actually run the commands itself. On Debian and Ubuntu the commands are:

```sh
sudo dpkg --add-architecture i386
sudo apt update
sudo apt install libvulkan1:i386 mesa-vulkan-drivers:i386 libgl1-mesa-dri:i386
```


## MacOS

Neither TAI nor Tribes: Ascend are compatible with MacOS yet.
