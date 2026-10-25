# <img src="./gui/src/assets/logo.svg" style="width: 3em;"> TA Installer
A [CLI](https://en.wikipedia.org/wiki/Command-line_interface) and [GUI](https://en.wikipedia.org/wiki/Graphical_user_interface) utility for managing Tribes: Ascend installations.

> [!TIP]
> For detailed instructions on how to use TAI, and how it works, please see the [documentation website](https://wilderzone.org/docs/ta-installer).

## GUI

To use TAI GUI, download and run the appropriate portable executable for your operating system from the [releases page](https://github.com/wilderzone/ta-installer/releases/latest).
You'll be prompted to select the game type (Vanilla, Modded, or Custom) and the folder in which to install TA. Modded is the recommended option, and includes community maps and CTF routes.

_Note: `tai-gui.exe` requires [VC Redist](https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist) and [WebView2](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) to be installed on your computer. Whilst it's very likely you already have them, it's possible you will need to install them before running TAI._

<p align="center">
  <img width="1262" height="902" alt="TAI GUI example" src="https://github.com/user-attachments/assets/c4da7aa5-9c3a-40e2-a74f-635f9fb22405" />
</p>

## CLI

TAI CLI is provided as a means to install custom variants of TA, and for use in applications that require programmatic control of the TA installation process, such as launchers.  
To use TAI CLI, download the portable `tai-cli.exe` executable from the [releases page](https://github.com/wilderzone/ta-installer/releases/latest).
_This executable can be bundled with your application. See the [GUI package](./gui) for an example of this kind of integration._

Then run TAI, passing a subcommand and options:
```sh
tai-cli.exe <subcommand> --repo=<url> --path=<path>
```
