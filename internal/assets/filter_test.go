package assets

import (
	"slices"
	"testing"
)

func TestFilterAssetsIncludeThenExclude(t *testing.T) {
	catalogue := []Asset{
		{Name: "Core", Type: CORE},
		{Name: "Arx Novena", Type: DLC},
		{Name: "Drydock", Type: DLC},
	}
	// Include everything of type dlc, then exclude one by exact name.
	got := names(FilterAssets(catalogue, []string{"dlc"}, []string{"Drydock"}))
	if !slices.Equal(got, []string{"Arx Novena"}) {
		t.Errorf("got %v, want [Arx Novena]", got)
	}
}

func TestFilterAssetsEmptyReturnsAll(t *testing.T) {
	catalogue := []Asset{{Name: "Core"}, {Name: "Maps"}}
	if got := names(FilterAssets(catalogue, nil, nil)); !slices.Equal(got, []string{"Core", "Maps"}) {
		t.Errorf("got %v, want all assets unchanged", got)
	}
}

func TestFilterByPlatformKeepsUnrestrictedAssets(t *testing.T) {
	catalogue := []Asset{
		{Name: "Core", Type: CORE},
		{Name: "Compatibility Runtime", Type: PREREQUISITE, Platform: "linux"},
		{Name: "Installation Utilities", Type: PREREQUISITE, Platform: "windows"},
	}
	// An asset without a platform belongs to every platform, so the game's own files
	// survive the filter on both.
	if got := names(FilterByPlatform(catalogue, "linux")); !slices.Equal(got, []string{"Core", "Compatibility Runtime"}) {
		t.Errorf("got %v, want [Core Compatibility Runtime]", got)
	}
	if got := names(FilterByPlatform(catalogue, "windows")); !slices.Equal(got, []string{"Core", "Installation Utilities"}) {
		t.Errorf("got %v, want [Core Installation Utilities]", got)
	}
}

func TestFilterByPlatformExcludesForeignDependencies(t *testing.T) {
	// A dependency on an asset for another platform must not survive the filter, or
	// resolution would go looking for something that was never downloadable here.
	catalogue := []Asset{
		{Name: "Core", Type: CORE},
		{Name: "Runtime", Type: PREREQUISITE, Platform: "linux", Dependencies: []string{"Core"}},
	}
	if got := names(FilterByPlatform(catalogue, "windows")); !slices.Equal(got, []string{"Core"}) {
		t.Errorf("got %v, want [Core]", got)
	}
}

func TestApplyAutomaticThreeCases(t *testing.T) {
	assets := []Asset{{Name: "Explicit"}, {Name: "Kept"}, {Name: "New"}}
	explicit := []Asset{{Name: "Explicit"}}
	previous := []Asset{{Name: "Kept", Automatic: true}}

	stamped := ApplyAutomatic(assets, explicit, previous)
	flags := map[string]bool{}
	for _, a := range stamped {
		flags[a.Name] = a.Automatic
	}

	if flags["Explicit"] {
		t.Error("an explicitly requested asset must not be automatic")
	}
	if !flags["Kept"] {
		t.Error("an asset already recorded as automatic must keep that flag")
	}
	if !flags["New"] {
		t.Error("a newly pulled-in asset must be marked automatic")
	}
}

func TestApplyAutomaticPromotesToExplicit(t *testing.T) {
	// An asset that was previously automatic but is now asked for by name is promoted.
	assets := []Asset{{Name: "Baz"}}
	stamped := ApplyAutomatic(assets, []Asset{{Name: "Baz"}}, []Asset{{Name: "Baz", Automatic: true}})
	if stamped[0].Automatic {
		t.Error("requesting an asset by name should clear its automatic flag")
	}
}

func TestMergeAssetsIsAdditiveAndAddedWins(t *testing.T) {
	existing := []Asset{{Name: "A", Source: "old"}, {Name: "B"}}
	added := []Asset{{Name: "A", Source: "new"}, {Name: "C"}}

	merged := MergeAssets(existing, added)
	if got := names(merged); !slices.Equal(got, []string{"A", "B", "C"}) {
		t.Fatalf("got order %v, want [A B C]", got)
	}
	if merged[0].Source != "new" {
		t.Errorf("A.Source = %q, want the freshly added %q", merged[0].Source, "new")
	}
}
