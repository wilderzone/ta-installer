---
order: 4
---

# Building

The application is built for production in CI, so you should only need to build it manually if you are testing some part of the build/distribution process.

See the [build workflow](https://github.com/wilderzone/ta-installer/blob/main/.github/workflows/build.yml) for reference.


## CLI

The CLI is built in two parts:
- The core module under [`/tai`](https://github.com/wilderzone/ta-installer/tree/main/tai).
- The installation utilities under [`/utilities`](https://github.com/wilderzone/ta-installer/tree/main/utilities).

The both parts should be built on Linux, using the following script:
```sh
cd ./utilities && bash ./build.sh
```

This will produce bundles & artifacts for Windows and Linux, into `/utilities/dist`.


## GUI

The GUI embeds the CLI, so the embedded binary needs to be built before the GUI. This is handled automatically through the `pretauri` NPM script, so you only need to run the following commands:
```sh
cd gui
npm ci
npm run tauri build
```

Building the GUI on Linux needs the WebKitGTK stack: `libwebkit2gtk-4.1-dev`, `libgtk-3-dev`, `librsvg2-dev`, `libdbus-1-dev`, `patchelf`, and `rpm` for the `.rpm` bundle. CI builds both platforms on every push.
