# TA Installer
A command line utility for managing Tribes: Ascend installations.


```sh
tai.exe <subcommand> --repo=<url> --path=<path>
```
`--repo` is optional. TAI will retrieve files from the Wilderzone CDN by default.


## Subcommands

### Install
Download and install TA into the given directory.
```sh
tai.exe install --repo=<url> --path=<path>
```


### Verify
Verify an existing TA installation at the given path.
```sh
tai.exe verify --repo=<url> --path=<path>
```


### Patch/Update
Update an existing TA installation.
```sh
tai.exe patch --repo=<url> --path=<path>
tai.exe update --repo=<url> --path=<path>
```


### Uninstall
Remove a TA installation from the system.
```sh
tai.exe uninstall --repo=<url> --path=<path>
```


### Asset
Generate an asset file with the given name from the given path. The `<name>.json` file will be exported at the root of `<path>`.
```sh
tai.exe asset --name=<name> --path=<path>
```


### Manifest
Generate a manifest file from a collection of asset files at the given path. The `manifest.json` file will be exported at the root of `<path>`.
```sh
tai.exe manifest --path=<path>
```
