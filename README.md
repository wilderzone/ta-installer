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

### Check
Check for an existing installation in the given directory.
```sh
tai.exe check --path=<path>
```

### Check Repository
Check that the provided URL points to a valid repository.
```sh
tai.exe check-repo --repo=<url>
```

### Verify
Verify an existing TA installation in the given directory.
```sh
tai.exe verify --path=<path>
```

### Update
Update an existing TA installation.
```sh
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


## Output

By default, TAI will print human-readable output to the console. For example:
```
> tai.exe verify --path="G:/tribes"

Reading local manifest from: G:\tribes\manifest.json
[1/34] Prerequisites: Verified
[2/34] Core: Verified
[3/34] Engine: Verified
...
```

Passing the `-m` flag will prompt TAI to output machine-readable JSON instead. For example:
```
> tai.exe verify -m --path="G:/tribes"

{"type":"READING_LOCAL_MANIFEST","message":"Reading local manifest from: G:\\tribes\\manifest.json"}
{"type":"PROGRESS","index":1,"total":34,"message":"Prerequisites: Verified"}
{"type":"PROGRESS","index":2,"total":34,"message":"Core: Verified"}
{"type":"PROGRESS","index":3,"total":34,"message":"Engine: Verified"}
...
```

<details>
<summary>Machine-readable message types</summary>

- `ASSET_CREATED`
- `ASSETS_ARE_DAMAGED`
- `CHECKING_ASSETS`
- `CHECKING_FOR_UPDATES`
- `DOWNLOADING_MANIFEST`
- `INSTALL_FAILURE`
- `INSTALL_PARTIAL`
- `INSTALL_SUCCESS`
- `INSTALL_UP_TO_DATE`
- `INSTALL_VERIFIED`
- `INSTALLING_PREREQUISITES`
- `MANIFEST_CREATED`
- `PATCHING_ASSETS`
- `PROMPT_PROCEED`
- `READING_LOCAL_MANIFEST`
- `REDOWNLOADING`
- `REPO_FOUND`
- `REPO_NOT_FOUND`
- `RETRIEVING_MANIFEST`
- `UPDATE_FAILURE`
- `UPDATE_FOUND`
- `UPDATE_PARTIAL`
- `UPDATE_SUCCESS`
- `PROGRESS`
</details>
