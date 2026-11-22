---
order: 4
---

# Dependency Resolution

Some assets can't function until other assets are present. An asset file declares this in its `dependencies` array, which contains the exact names of the assets it requires:
```json
{
	"name": "Arx Novena",
	"type": "dlc",
	"dependencies": ["Core", "Map Textures"],
	...
}
```

The CLI's `manifest` [subcommand](/cli/subcommands) validates these relationships when a repository is built, so a misspelled name or a circular dependency is caught before the manifest is published.

TAI resolves the graph automatically:
- `install` also installs everything the selected assets require. A dependency that is already installed is ignored rather than downloaded again.
- Because `--exclude` is an explicit instruction, excluding a dependency of a selected asset raises an error, instead of silently overriding the selection.
- `uninstall` also removes any dependents of the assets being removed, after asking for confirmation.
- `uninstall` then prunes assets that TAI installed only to satisfy a dependency, if they are no longer required. Assets that were installed by name (using `--include`), and foundational `core` and `prerequisite` assets, are never pruned. The local `manifest.json` records this by marking dependency-installed assets with `"automatic": true` (repository manifests don't carry this field).
