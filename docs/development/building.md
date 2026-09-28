# Building

The GUI embeds the CLI, and the embedded binary is gitignored. A fresh repository clone can't compile the Rust crate until the CLI has been built:

```sh
cd gui
npm ci
npm run build-sidecar        # cross-compile with GOOS=linux / GOOS=windows
npm run tauri build          # `pretauri` runs build-sidecar for you
```

Building the GUI on Linux needs the WebKitGTK stack: `libwebkit2gtk-4.1-dev`, `libgtk-3-dev`, `librsvg2-dev`, `libdbus-1-dev`, `patchelf`, and `rpm` for the `.rpm` bundle. CI builds both platforms on every push.
