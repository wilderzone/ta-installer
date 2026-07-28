package assets

import "slices"

// MatchesTerms reports whether any of the given filter terms matches the asset, by
// its type (e.g. "dlc") or its exact name (e.g. "Foo Bar"). Matching is case-sensitive.
func MatchesTerms(asset Asset, terms []string) bool {
	return slices.Contains(terms, string(asset.Type)) || slices.Contains(terms, asset.Name)
}

// FilterAssets returns the assets that satisfy the include/exclude terms. Include is
// applied first (an empty include list means "include all"), then exclude.
func FilterAssets(assets []Asset, include []string, exclude []string) []Asset {
	if len(include) == 0 && len(exclude) == 0 {
		return assets
	}
	filtered := []Asset{}
	for _, asset := range assets {
		included := len(include) == 0 || MatchesTerms(asset, include)
		if included && !MatchesTerms(asset, exclude) {
			filtered = append(filtered, asset)
		}
	}
	return filtered
}

// IntersectByName returns the assets from `available` whose name matches one of
// the `installed` assets, preserving the order of `available`.
func IntersectByName(available []Asset, installed []Asset) []Asset {
	names := make(map[string]bool, len(installed))
	for _, asset := range installed {
		names[asset.Name] = true
	}
	result := []Asset{}
	for _, asset := range available {
		if names[asset.Name] {
			result = append(result, asset)
		}
	}
	return result
}

// DifferenceByName returns the assets from `all` whose name is not among the
// `removed` assets, preserving the order of `all`.
func DifferenceByName(all []Asset, removed []Asset) []Asset {
	names := make(map[string]bool, len(removed))
	for _, asset := range removed {
		names[asset.Name] = true
	}
	result := []Asset{}
	for _, asset := range all {
		if !names[asset.Name] {
			result = append(result, asset)
		}
	}
	return result
}

// MergeAssets returns the union of the already-installed `existing` assets and the
// newly-installed `added` assets, keyed by name, so that repeated installs are
// additive. For a name present in both, the `added` definition wins (it was just
// downloaded). Existing assets keep their original order; genuinely new assets are
// appended.
func MergeAssets(existing []Asset, added []Asset) []Asset {
	addedByName := make(map[string]Asset, len(added))
	for _, asset := range added {
		addedByName[asset.Name] = asset
	}
	merged := []Asset{}
	seen := make(map[string]bool, len(existing)+len(added))
	for _, asset := range existing {
		if replacement, ok := addedByName[asset.Name]; ok {
			merged = append(merged, replacement)
		} else {
			merged = append(merged, asset)
		}
		seen[asset.Name] = true
	}
	for _, asset := range added {
		if !seen[asset.Name] {
			merged = append(merged, asset)
			seen[asset.Name] = true
		}
	}
	return merged
}

// ApplyAutomatic stamps the `Automatic` flag onto each asset, returning a new slice.
// An asset named in `explicit` was asked for directly, so it is never automatic (which
// also promotes an asset that was previously pulled in as a dependency); an asset
// already recorded in `previous` keeps the flag it had; anything else is new and only
// present to satisfy a dependency.
//
// Repository manifests never carry the flag, and MergeAssets lets the freshly
// downloaded definition win, so it has to be re-applied whenever the local manifest is
// rewritten.
func ApplyAutomatic(assets []Asset, explicit []Asset, previous []Asset) []Asset {
	requested := make(map[string]bool, len(explicit))
	for _, asset := range explicit {
		requested[asset.Name] = true
	}
	installed := make(map[string]bool, len(previous))
	for _, asset := range previous {
		installed[asset.Name] = asset.Automatic
	}

	stamped := make([]Asset, 0, len(assets))
	for _, asset := range assets {
		if requested[asset.Name] {
			asset.Automatic = false
		} else {
			automatic, known := installed[asset.Name]
			asset.Automatic = !known || automatic
		}
		stamped = append(stamped, asset)
	}
	return stamped
}
