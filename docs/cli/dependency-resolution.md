---
order: 4
---

# Dependency Resolution

Some assets don't function until other assets are present. An asset file declares this in
its `dependencies` array, which holds the exact names of the assets it requires:

```json
{
	"name": "Arx Novena",
	"type": "dlc",
	"dependencies": ["Core", "Map Textures"],
	...
}
```

The CLI's `manifest` [subcommand](/cli/subcommands) validates these relationships when a repository is built, so a
misspelled name or a circular dependency is caught before the manifest is published.

TAI resolves the graph automatically:
- `install` also installs everything the selected assets require. A required asset that is already installed is left alone rather than downloaded again.
- Because `--exclude` is an explicit instruction, excluding an asset that the selection requires raises an error, instead of silently overriding.
- `uninstall` also removes anything that requires the assets being removed, after asking for confirmation.
- `uninstall` then prunes assets that TAI installed only to satisfy a dependency, if they are no longer required. Assets that were asked for by name, and foundational `core` and `prerequisite` assets, are never pruned this way.

The local `manifest.json` records this by marking dependency-installed assets with `"automatic": true` (repository manifests don't carry this field).
