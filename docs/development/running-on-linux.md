---
order: 2
---

# Running on Linux

Tribes: Ascend is a 32-bit Windows game, so a Linux installation includes a compatibility layer alongside the game files. TAI manages all of it automatically, including validating the compatibility layer when you run `verify`.

A few things differ from Windows by necessity:

- Shortcuts are desktop entries that are placed in the applications menu, on the desktop, and in the installation directory, and they run the `play` utility rather than the game directly - so a shortcut written by an older install keeps working when the way the game is launched changes. The uninstall entry stands in for the Add/Remove Programs listing.
- The default install path is under your home directory.
- DirectX 9 is served by the compatibility layer's own Direct3D 9 implementation, rather than installed system-wide.
- The compatibility layer's container image is fetched by umu**, into `~/.local/share/umu`, the first time a Wine prefix is created. That download is outside TAI's own progress reporting, so it appears as "Preparing the compatibility layer".
- **Install the GUI from the `.deb` or `.rpm` rather than running the binary on its own.** Its taskbar icon comes from the desktop entry those packages install, and under Wayland that is the only mechanism there is - a loose binary has nothing for the shell to match it against, and falls back to a generic icon. The binary is named `tribes-ascend-installer` to match the entry's filename, which is what the shell matches on.
- The GUI draws its own window corners and shadow. An undecorated window gets neither from the desktop - the compositor won't decorate it, and GTK's own decorations are what turning decorations off removes - so it dresses itself against a transparent window, and gives up `--window-shadow` pixels per side to do it.
- TA and TAMods settings live inside the Wine prefix, under `Prefix/pfx/drive_c/users/`, because that's where the game looks for its Documents folder. Uninstalling removes them along with everything else, whereas on Windows they sit in your real Documents folder and survive an uninstall. Copy that directory out first if you want to keep your game settings.

`install` reports anything about the machine that would get in the way (eg, a missing Vulkan driver, no `python3`, or unprivileged user namespaces being disabled) before committing to a multi-gigabyte download. None of them stops an install. `verify` runs the same checks as well.

The runtime is published separately from `manifest.json`, alongside the installation utilities and for the same reason the DirectX 9 redistributable is: the manifest describes the *game*, and is served by whatever repository `--repo` names. A custom repository shouldn't have to mirror half a gigabyte of Proton for a Linux install to work, so the runtime always comes from the official CDN regardless of `--repo`.

Publish it as a `.tar.gz` rather than a `.zip` (the runtime carries symlinks and execute bits that a zip either can't represent or loses when packaged on Windows) and generate its checksum file the same way as the utilities':

```sh
tai-cli asset --path=<extracted runtime> --name=Runtime   # produces Runtime.json
```

Publish both `Runtime.tar.gz` and `Runtime.json` under `/runtime/` on the CDN. `verify` fetches the checksums, and replaces the whole runtime if any file fails - which is also how a newly pinned Proton build reaches existing installations.

## Graphics Drivers

The game is a 32-bit program, so it needs the 32-bit build of your graphics card's Vulkan driver. The compatibility layer's container supplies its own 32-bit system libraries, but the graphics drivers can't come from the container - they have to match the running kernel, so they're taken from the host. Without them the game exits silently during startup. TAI checks for this case and prints the exact commands for your distribution, but can't actually run the commands itself. On Debian and Ubuntu the commands are:

```sh
sudo dpkg --add-architecture i386
sudo apt update
sudo apt install libvulkan1:i386 mesa-vulkan-drivers:i386 libgl1-mesa-dri:i386
```
