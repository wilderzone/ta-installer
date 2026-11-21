---
order: 5
---

# Releasing

CI also publishes releases to the CDN. To cut a release:

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
