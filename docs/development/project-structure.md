---
order: 1
---

# Project Structure

TAI is built from a [single Git repository](https://github.com/wilderzone/ta-installer), with one subdirectory for each component.

| Component              | Directory    | Languages     | Description                                                                                              |
| :--------------------- | :----------- | :------------ | :------------------------------------------------------------------------------------------------------- |
| Documentation          | `/docs`      | Markdown      | The content for this documentation website.                                                              |
| CLI                    | `/tai`       | Go            | TAI CLI, as is built to `tai-cli.exe` (Windows) & `tai-cli` (Linux).                                     |
| GUI                    | `/gui`       | Rust, TS, Vue | TAI GUI, as is built to `ta-installer-windows.exe` (Windows) & `ta-installer-linux.run` (Linux).         |
| Shared Go libraries    | `/internal`  | Go            | Pure TAI CLI functions, separated into dedicated libraries for improved organisation and easier testing. |
| Installation utilities | `/utilities` | Go            | Runtime installation utilities, as built to `verify.exe`, `uninstall.exe`, etc.                          |
