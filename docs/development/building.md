---
order: 4
---

# Building

The GUI embeds the CLI, and the embedded binary is gitignored. A fresh repository clone can't compile the Rust crate until the CLI has been built:

```sh
cd gui
npm ci
npm run build-sidecar        # cross-compile with GOOS=linux / GOOS=windows
npm run tauri build          # `pretauri` runs build-sidecar for you
```

Building the GUI on Linux needs the WebKitGTK stack: `libwebkit2gtk-4.1-dev`, `libgtk-3-dev`, `librsvg2-dev`, `libdbus-1-dev`, `patchelf`, and `rpm` for the `.rpm` bundle. CI builds both platforms on every push.

## Releasing

CI also publishes releases to the CDN. To cut one:

1. Bump `version` in `gui/src-tauri/tauri.conf.json`, commit it to `main`, and push `main` before the tag.
2. Tag that commit with the same version, and push the tag:
   ```sh
   git tag v1.2.3
   git push origin v1.2.3
   ```
3. Once every build has passed, the _Publish to the CDN_ job waits for approval in the `CDN` environment. Download the run's artifacts and test them on Windows and Linux before approving it.

The job fails if the tag isn't on `main` on GitHub, or doesn't match the version in `tauri.conf.json`. Otherwise it uploads the GUI executables and both installation utilities packages (with their checksum files) to `ta/game/installer/` on the CDN, then purges those URLs from Cloudflare's cache. Nothing else under that directory is touched, so the GUI's models are left alone.

The `CDN` environment holds the job's secrets:

| Secret                                     | Purpose                                                         |
| :----------------------------------------- | :-------------------------------------------------------------- |
| `R2_ACCOUNT_ID`, `R2_BUCKET`               | Where to upload the files.                                      |
| `R2_ACCESS_KEY_ID`, `R2_SECRET_ACCESS_KEY` | An R2 API token with Object Read & Write on that bucket only.   |
| `CLOUDFLARE_ZONE_ID`                       | The zone that serves `cdn.wilderzone.org`.                      |
| `CLOUDFLARE_PURGE_TOKEN`                   | An API token with only the Cache Purge permission on that zone. |
