package version

import (
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

func TestCurrentPrefersBuildVersion(t *testing.T) {
	old := buildVersion
	buildVersion = "v1.2.3"
	t.Cleanup(func() { buildVersion = old })

	if got := Current(); got != "v1.2.3" {
		t.Fatalf("expected build version, got %q", got)
	}
}

func TestModuleFromBuildInfoFallback(t *testing.T) {
	SetDefaultModule("example.com/service")
	t.Cleanup(func() { SetDefaultModule("") })

	if got := moduleFromBuildInfo(nil); got != "example.com/service" {
		t.Fatalf("expected fallback module, got %q", got)
	}

	SetDefaultModule("")
	if got := moduleFromBuildInfo(nil); got != "unknown" {
		t.Fatalf("expected unknown fallback, got %q", got)
	}
}

func TestModuleFromBuildInfoUsesPath(t *testing.T) {
	SetDefaultModule("example.com/service")
	t.Cleanup(func() { SetDefaultModule("") })

	info := &debug.BuildInfo{Main: debug.Module{Path: "example.com/real"}}
	if got := moduleFromBuildInfo(info); got != "example.com/real" {
		t.Fatalf("expected build info path, got %q", got)
	}
}

func TestModuleVersionFromBuildInfoDeps(t *testing.T) {
	info := &debug.BuildInfo{
		Deps: []*debug.Module{
			{Path: "example.com/dep", Version: "v1.2.3+dirty"},
		},
	}
	if got := moduleVersionFromBuildInfo(info, "example.com/dep", false); got != "v1.2.3" {
		t.Fatalf("expected normalized dep version, got %q", got)
	}
	if got := moduleVersionFromBuildInfo(info, "example.com/dep", true); got != "v1.2.3+dirty" {
		t.Fatalf("expected dirty dep version, got %q", got)
	}
	if got := moduleVersionFromBuildInfo(info, "example.com/missing", false); got != "v0.0.0-unknown" {
		t.Fatalf("expected unknown for missing dep, got %q", got)
	}
}

func TestModuleVersionFromBuildInfoReplace(t *testing.T) {
	info := &debug.BuildInfo{
		Deps: []*debug.Module{
			{
				Path:    "example.com/original",
				Version: "v1.0.0",
				Replace: &debug.Module{Path: "example.com/replaced", Version: "v2.0.0"},
			},
		},
	}
	if got := moduleVersionFromBuildInfo(info, "example.com/replaced", false); got != "v2.0.0" {
		t.Fatalf("expected replaced dep version, got %q", got)
	}
}

func TestPseudoFromBuildInfo(t *testing.T) {
	ts := time.Date(2025, time.January, 2, 3, 4, 5, 0, time.UTC)
	info := &debug.BuildInfo{
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "1234567890abcdef"},
			{Key: "vcs.time", Value: ts.Format(time.RFC3339)},
			{Key: "vcs.modified", Value: "true"},
		},
	}
	got := pseudoFromBuildInfo(info, false)
	if got == "" {
		t.Fatalf("expected pseudo version")
	}
	if wantPrefix := "v0.0.0-20250102030405-1234567890ab"; got[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("unexpected version prefix: %q", got)
	}
	if strings.Contains(got, "+dirty") {
		t.Fatalf("expected no dirty suffix, got %q", got)
	}
	gotDirty := pseudoFromBuildInfo(info, true)
	if !strings.Contains(gotDirty, "+dirty") {
		t.Fatalf("expected dirty suffix, got %q", gotDirty)
	}
	if got := normalizeVersion("v1.2.3+dirty", false); got != "v1.2.3" {
		t.Fatalf("expected dirty suffix removed, got %q", got)
	}
	if got := normalizeVersion("v1.2.3+dirty", true); got != "v1.2.3+dirty" {
		t.Fatalf("expected dirty suffix preserved, got %q", got)
	}
	if pseudoFromBuildInfo(nil, false) != "" {
		t.Fatalf("expected empty version for nil build info")
	}
}
