---
order: 5
---

# Releasing


## Main Bundle

CI automatically publishes releases to the CDN. To create a release:
1. Bump `version` in `gui/src-tauri/tauri.conf.json` (and anywhere else), commit to `main`, and push `main` before the tag.
2. Tag that commit with the same version, and push the tag:
   ```sh
   git tag v1.2.3
   git push origin v1.2.3
   ```
3. Once every build has passed, the "Publish to the CDN" job waits for approval in the `CDN` environment. Download the run's artifacts and test them on Windows and Linux before approving it.

The job fails if the tag isn't on `origin/main`, or doesn't match the version in `tauri.conf.json`. Otherwise it uploads the GUI executables and both installation utilities packages to `/ta/game/installer/` on the CDN, then purges those URLs from Cloudflare's cache. Nothing else under that directory is touched, so the GUI's models are left alone.

The `CDN` environment holds the job's secrets:

| Secret                                     | Purpose                                                         |
| :----------------------------------------- | :-------------------------------------------------------------- |
| `R2_ACCOUNT_ID`, `R2_BUCKET`               | Where to upload the files for distribution.                     |
| `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY` | An R2 API token with Object Read & Write on that bucket only.   |
| `CLOUDFLARE_ZONE_ID`                       | The zone that serves `cdn.wilderzone.org`.                      |
| `CLOUDFLARE_PURGE_TOKEN`                   | An API token with only the Cache Purge permission on that zone. |


## Compatibility Layer

The compatibility layer is published separately from `manifest.json`, alongside the installation utilities and for the same reason the DirectX 9 redistributable is: the manifest describes the *game*, and is served by whatever repository `--repo` names. A custom repository shouldn't have to mirror half a gigabyte of Proton for a Linux install to work, so the runtime always comes from the official CDN regardless of `--repo`.

Publish the compatibility layer as a `.tar.gz` rather than a `.zip` (it carries symlinks and execute bits that a zip either can't represent or loses when packaged on Windows) and generate its checksum file the same way as the utilities':
```sh
tai-cli asset --path=<extracted runtime> --name=Runtime   # Produces Runtime.json
```

Publish both `Runtime.tar.gz` and `Runtime.json` under `/ta/game/runtime/` on the CDN. `verify` fetches the checksums, and replaces the whole runtime if any file fails - which is also how a newly pinned Proton build reaches existing installations.
