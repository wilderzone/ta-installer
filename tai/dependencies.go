package main

import (
	"fmt"
	"slices"
	"strings"
)

// Traversal states used to detect cycles while resolving dependencies.
const (
	unvisited = iota
	visiting
	visited
)

// foundational reports whether an asset underpins the whole installation. These are
// never pruned automatically - an installation without them isn't a working game - so
// they only ever leave by explicit request.
func foundational(asset Asset) bool {
	return asset.Type == CORE || asset.Type == PREREQUISITE
}

// resolveDependencies returns the transitive closure of `roots` over the `available`
// catalogue, ordered so that every asset appears after the assets it depends on. It
// fails if an asset names a dependency that isn't in the catalogue, or if the
// dependencies form a cycle.
func resolveDependencies(available []Asset, roots []Asset) ([]Asset, error) {
	catalogue := make(map[string]Asset, len(available))
	for _, asset := range available {
		catalogue[asset.Name] = asset
	}

	state := make(map[string]int, len(available))
	trail := []string{}
	closure := []Asset{}

	var visit func(name string, requiredBy string) error
	visit = func(name string, requiredBy string) error {
		switch state[name] {
		case visited:
			return nil
		case visiting:
			// The asset is already on the trail, so everything from there on is the loop.
			loop := trail[slices.Index(trail, name):]
			return fmt.Errorf("dependency cycle: %s → %s", strings.Join(loop, " → "), name)
		}

		asset, known := catalogue[name]
		if !known {
			if len(requiredBy) == 0 {
				return fmt.Errorf("unknown asset %q", name)
			}
			return fmt.Errorf("%q requires %q, which doesn't exist", requiredBy, name)
		}

		state[name] = visiting
		trail = append(trail, name)
		for _, dependency := range asset.Dependencies {
			err := visit(dependency, name)
			if err != nil {
				return err
			}
		}
		trail = trail[:len(trail)-1]
		state[name] = visited

		// Appending only once the dependencies have been visited puts them first.
		closure = append(closure, asset)
		return nil
	}

	for _, root := range roots {
		err := visit(root.Name, "")
		if err != nil {
			return nil, err
		}
	}

	return closure, nil
}

// validateDependencies checks that every dependency in a collection of assets resolves
// and that the graph is acyclic. Using every asset as a root visits the whole graph, so
// both problems surface at once.
func validateDependencies(assets []Asset) error {
	_, err := resolveDependencies(assets, assets)
	return err
}

// dependentsOf returns the installed assets that transitively depend on one of the
// `targets`, excluding the targets themselves. The fixpoint loop avoids recursion, so a
// manifest that somehow contains a cycle can't hang an uninstall.
func dependentsOf(installed []Asset, targets []Asset) []Asset {
	targeted := make(map[string]bool, len(targets))
	doomed := make(map[string]bool, len(targets))
	for _, asset := range targets {
		targeted[asset.Name] = true
		doomed[asset.Name] = true
	}

	for changed := true; changed; {
		changed = false
		for _, asset := range installed {
			if doomed[asset.Name] {
				continue
			}
			for _, dependency := range asset.Dependencies {
				if doomed[dependency] {
					doomed[asset.Name] = true
					changed = true
					break
				}
			}
		}
	}

	dependents := []Asset{}
	for _, asset := range installed {
		if doomed[asset.Name] && !targeted[asset.Name] {
			dependents = append(dependents, asset)
		}
	}
	return dependents
}

// pruneOrphans repeatedly drops the assets that were only installed to satisfy another
// asset's dependency and are no longer needed by anything that remains. It returns the
// assets that stay installed, and those that were pruned. The loop matters: pruning one
// asset can orphan its own dependencies.
func pruneOrphans(survivors []Asset) ([]Asset, []Asset) {
	kept := slices.Clone(survivors)
	pruned := []Asset{}

	for {
		needed := map[string]bool{}
		for _, asset := range kept {
			for _, dependency := range asset.Dependencies {
				needed[dependency] = true
			}
		}

		remaining := []Asset{}
		for _, asset := range kept {
			if asset.Automatic && !foundational(asset) && !needed[asset.Name] {
				pruned = append(pruned, asset)
			} else {
				remaining = append(remaining, asset)
			}
		}
		if len(remaining) == len(kept) {
			return kept, pruned
		}
		kept = remaining
	}
}

// dependentOf returns the name of the first asset that lists `name` as a dependency, or
// an empty string if none do. Used to explain dependency conflicts to the user.
func dependentOf(assets []Asset, name string) string {
	for _, asset := range assets {
		if slices.Contains(asset.Dependencies, name) {
			return asset.Name
		}
	}
	return ""
}

// explicitAssets returns the assets the user asked for directly, as opposed to those
// that were installed to satisfy a dependency.
func explicitAssets(assets []Asset) []Asset {
	explicit := []Asset{}
	for _, asset := range assets {
		if !asset.Automatic {
			explicit = append(explicit, asset)
		}
	}
	return explicit
}

// assetNames renders a comma-separated list of asset names for log messages.
func assetNames(assets []Asset) string {
	names := make([]string, 0, len(assets))
	for _, asset := range assets {
		names = append(names, asset.Name)
	}
	return strings.Join(names, ", ")
}
