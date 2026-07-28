package assets

import (
	"slices"
	"strings"
	"testing"
)

// dep builds an asset with the given name and dependencies.
func dep(name string, deps ...string) Asset {
	return Asset{Name: name, Dependencies: deps}
}

// names extracts the names of a slice of assets, in order.
func names(assets []Asset) []string {
	out := make([]string, len(assets))
	for i, a := range assets {
		out[i] = a.Name
	}
	return out
}

// assertDependenciesFirst fails if any asset appears before one of its dependencies.
func assertDependenciesFirst(t *testing.T, closure []Asset) {
	t.Helper()
	position := map[string]int{}
	for i, a := range closure {
		position[a.Name] = i
	}
	for i, a := range closure {
		for _, d := range a.Dependencies {
			if at, ok := position[d]; ok && at > i {
				t.Errorf("%q (index %d) appears before its dependency %q (index %d)", a.Name, i, d, at)
			}
		}
	}
}

func TestResolveDependenciesChain(t *testing.T) {
	available := []Asset{dep("A", "B"), dep("B", "C"), dep("C")}
	closure, err := ResolveDependencies(available, []Asset{dep("A", "B")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := names(closure); !slices.Equal(got, []string{"C", "B", "A"}) {
		t.Errorf("got order %v, want [C B A]", got)
	}
}

func TestResolveDependenciesDiamondVisitsOnce(t *testing.T) {
	// A depends on B and C; both depend on D. D must appear once, before everything.
	available := []Asset{dep("A", "B", "C"), dep("B", "D"), dep("C", "D"), dep("D")}
	closure, err := ResolveDependencies(available, []Asset{available[0]})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(closure) != 4 {
		t.Fatalf("got %d assets, want 4 (D must not be duplicated): %v", len(closure), names(closure))
	}
	assertDependenciesFirst(t, closure)
	if closure[0].Name != "D" || closure[len(closure)-1].Name != "A" {
		t.Errorf("got order %v, want D first and A last", names(closure))
	}
}

func TestResolveDependenciesCycle(t *testing.T) {
	available := []Asset{dep("A", "B"), dep("B", "A")}
	_, err := ResolveDependencies(available, []Asset{available[0]})
	if err == nil {
		t.Fatal("expected a cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "dependency cycle") || !strings.Contains(err.Error(), "A → B → A") {
		t.Errorf("error %q does not name the cycle A → B → A", err)
	}
}

func TestResolveDependenciesUnknownRoot(t *testing.T) {
	_, err := ResolveDependencies([]Asset{dep("A")}, []Asset{dep("X")})
	if err == nil || !strings.Contains(err.Error(), `unknown asset "X"`) {
		t.Errorf("got %v, want an 'unknown asset \"X\"' error", err)
	}
}

func TestResolveDependenciesUnknownDependency(t *testing.T) {
	_, err := ResolveDependencies([]Asset{dep("A", "Ghost")}, []Asset{dep("A", "Ghost")})
	if err == nil || !strings.Contains(err.Error(), `"A" requires "Ghost"`) {
		t.Errorf("got %v, want an error naming that A requires Ghost", err)
	}
}

func TestDependentsOfIsTransitive(t *testing.T) {
	core := Asset{Name: "Core", Type: CORE}
	maps := dep("Maps", "Core")
	skins := dep("Skins", "Maps")
	installed := []Asset{core, maps, skins}

	dependents := DependentsOf(installed, []Asset{core})
	got := names(dependents)
	slices.Sort(got)
	if !slices.Equal(got, []string{"Maps", "Skins"}) {
		t.Errorf("got %v, want [Maps Skins] (transitive, excluding Core)", got)
	}
}

func TestPruneOrphansKeepsFoundationalAndNeeded(t *testing.T) {
	core := Asset{Name: "Core", Type: CORE, Automatic: true} // automatic but foundational -> immune
	needed := Asset{Name: "Needed", Automatic: true}
	user := Asset{Name: "User", Dependencies: []string{"Needed"}} // explicit, keeps Needed alive
	orphan := Asset{Name: "Orphan", Automatic: true}

	kept, pruned := PruneOrphans([]Asset{core, needed, user, orphan})
	keptNames := names(kept)
	slices.Sort(keptNames)
	if !slices.Equal(keptNames, []string{"Core", "Needed", "User"}) {
		t.Errorf("kept %v, want [Core Needed User]", keptNames)
	}
	if got := names(pruned); !slices.Equal(got, []string{"Orphan"}) {
		t.Errorf("pruned %v, want [Orphan]", got)
	}
}

func TestPruneOrphansFixpointCascade(t *testing.T) {
	// A is automatic and depends on B (also automatic). Nothing external needs A, so A is
	// pruned; that orphans B, which the loop must then prune too.
	a := Asset{Name: "A", Automatic: true, Dependencies: []string{"B"}}
	b := Asset{Name: "B", Automatic: true}

	kept, pruned := PruneOrphans([]Asset{a, b})
	if len(kept) != 0 {
		t.Errorf("kept %v, want nothing (both should cascade away)", names(kept))
	}
	got := names(pruned)
	slices.Sort(got)
	if !slices.Equal(got, []string{"A", "B"}) {
		t.Errorf("pruned %v, want [A B]", got)
	}
}
