---
order: 1
---

# Project Structure

TAI is built from a [single Git repository](https://github.com/wilderzone/ta-installer), with subdirectories for the CLI, GUI, shared Go libraries, installation utilities, and documentation.

| Component              | Directory    | Languages    | Description                  |
| :--------------------- | :----------- | :----------- | :--------------------------- |
| Documentation          | `/docs`      | TS, Markdown | This documentation website.  |
| CLI                    | `/tai`       | Go           | TAI CLI, as is built to `tai-cli.exe`. |
| GUI                    | `/gui`       | Rust, TS     | TAI GUI, as is built to `tai-gui.exe`. |
| Shared Go libraries    | `/internal`  | Go           | Pure TAI CLI functions, separated into dedicated libraries for improved organisation and easier testing. |
| Installation utilities | `/utilities` | Go           | Runtime installation utilities, as built to `verify.exe`, `uninstall.exe`, etc. |
