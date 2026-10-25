# Output

By default, TAI will print human-readable output to the console. For example:
```sh{1}
> tai-cli.exe verify --path="G:/tribes"

Reading local manifest from: G:\tribes\manifest.json
[1/34] Prerequisites: Verified
[2/34] Core: Verified
[3/34] Engine: Verified
⠋ [3/34] Verifying...
```
While a slow step is running, an animated spinner is shown on the last line, so it's clear that TAI is still working during long pauses. It is only drawn when the output is an interactive terminal; when the output is piped or redirected, or when `-m` or `-s` is passed, it is omitted.

Passing the `-m` [flag](/cli/options#flags) will instruct TAI to output machine-readable JSON instead of human-readable messages. For example:
```sh{1}
> tai-cli.exe verify -m --path="G:/tribes"

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
- `DOWNLOAD_COMPLETE`
- `DOWNLOAD_PROGRESS`
- `DOWNLOAD_STARTED`
- `DOWNLOADING_MANIFEST`
- `ERROR`
- `INSTALL_FAILURE`
- `INSTALL_FOUND`
- `INSTALL_INTERRUPTED`
- `INSTALL_NOT_FOUND`
- `INSTALL_PARTIAL`
- `INSTALL_SUCCESS`
- `INSTALL_UP_TO_DATE`
- `INSTALL_VERIFIED`
- `INSTALLING`
- `INSTALLING_DEPENDENCIES`
- `INTERRUPTED`
- `MANIFEST_CONTENTS`
- `MANIFEST_CREATED`
- `MANIFEST_UNREADABLE`
- `PATCHING_ASSETS`
- `PREREQUISITES_DOWNLOADING`
- `PREREQUISITES_FAILED`
- `PREREQUISITES_INSTALLING`
- `PREREQUISITES_REQUIRED`
- `PROGRESS`
- `PROMPT_PROCEED`
- `PROMPT_REMOVE_DEPENDENTS`
- `PROMPT_UNINSTALL_ANYWAY`
- `READING_LOCAL_MANIFEST`
- `READING_MANIFEST_FAILED`
- `REDOWNLOADING`
- `REPO_FOUND`
- `REPO_NOT_FOUND`
- `RUNTIME_CHECK_FAILED`
- `RUNTIME_CHECK_SKIPPED`
- `RUNTIME_DAMAGED`
- `RUNTIME_DOWNLOADING`
- `RUNTIME_MISSING`
- `RUNTIME_PREPARE_FAILED`
- `RUNTIME_PREPARING`
- `RUNTIME_READY`
- `RUNTIME_REPAIRED`
- `RUNTIME_VERIFIED`
- `SHORTCUTS_FAILURE`
- `SHORTCUTS_REMOVE_FAILURE`
- `SHORTCUTS_SUCCESS`
- `UNINSTALL_DEPENDENTS`
- `UNINSTALL_FAILURE`
- `UNINSTALL_INTERRUPTED`
- `UNINSTALL_MATCHED_ALL`
- `UNINSTALL_ORPHANS`
- `UNINSTALL_PARTIAL`
- `UNINSTALL_SUCCESS`
- `UNINSTALLING`
- `UPDATE_FAILURE`
- `UPDATE_FOUND`
- `UPDATE_INTERRUPTED`
- `UPDATE_PARTIAL`
- `UPDATE_SUCCESS`
- `UTILITIES_CHECK_SKIPPED`
- `UTILITIES_DAMAGED`
- `UTILITIES_FAILURE`
- `UTILITIES_REPAIRED`
- `UTILITIES_REPAIR_DEFERRED`
- `UTILITIES_SUCCESS`
- `UTILITIES_VERIFIED`
- `VERIFY_INTERRUPTED`
</details>


### Interrupting a Run

Pressing `Ctrl+C` stops a running [subcommand](/cli/subcommands) at its next safe point, rather than killing it mid-write. TAI reports what was interrupted and which subcommand to run to pick up where it left off, and exits with a non-zero status. Pressing `Ctrl+C` a second time ends the process immediately.

A parent process that has no console of its own (for example, TAI GUI) can't send an interrupt signal on Windows, so it asks the same way by writing the line `cancel` to TAI's standard input. This is only listened for when standard input is a pipe (not a terminal), so that it can't accidentally swallow the answers to confirmation prompts.

Passing the `-m` flag will instruct TAI to output machine-readable JSON instead. For example:
```
> tai-cli.exe verify -m --path="G:/tribes"

{"type":"READING_LOCAL_MANIFEST","message":"Reading local manifest from: G:\\tribes\\manifest.json"}
{"type":"PROGRESS","index":1,"total":34,"message":"Prerequisites: Verified"}
{"type":"PROGRESS","index":2,"total":34,"message":"Core: Verified"}
{"type":"PROGRESS","index":3,"total":34,"message":"Engine: Verified"}
...
```


### Download Progress

Assets are downloaded in batches and only logged as each one finishes, so a large download can sit on the same line for a long time. In a terminal, the spinner fills that gap with the bytes as they arrive:
```
> tai-cli.exe install --path="G:/tribes"

Installing...
[1/34] Prerequisites: Downloaded
⠋ Downloading... 42% of 2.8 GB - 4 minutes left
```

An interface driving TAI has no spinner to read, so passing the `-p` flag instructs TAI to produce JSON-formatted progress updates instead. A download declares the total size, reports the number of bytes that have arrived several times per second, and declares when it is complete:
```
> tai-cli.exe install -p --path="G:/tribes"
...
{"size":2782785536,"message":"","type":"DOWNLOAD_STARTED"}
{"bytes":15728640,"size":2782785536,"rate":6553600,"message":"","type":"DOWNLOAD_PROGRESS"}
{"bytes":32505856,"size":2782785536,"rate":6684672,"message":"","type":"DOWNLOAD_PROGRESS"}
...
{"bytes":2782785536,"size":2782785536,"message":"","type":"DOWNLOAD_COMPLETE"}
```

`bytes` is what has arrived so far, `size` what is expected in total, and `rate` a smoothed average of the bytes per second. These lines are always JSON, with or without `-m`. A manifest that doesn't record the size of every asset it is downloading can't be measured, and reports nothing rather than an inaccurate total.
