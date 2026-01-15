package version

import (
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// buildVersion is set via -ldflags "-X pkt.systems/version.buildVersion=...".
var buildVersion = ""

var (
	defaultModuleMu sync.RWMutex
	defaultModule   string
)

// SetDefaultModule sets the fallback module path used when build info is unavailable.
// Passing an empty string clears the fallback.
func SetDefaultModule(path string) {
	defaultModuleMu.Lock()
	defaultModule = strings.TrimSpace(path)
	defaultModuleMu.Unlock()
}

// Current returns the best available version string (without dirty suffix).
func Current() string {
	return currentFromBuildInfo(false)
}

// CurrentWithDirty returns the best available version string (including dirty suffix when available).
func CurrentWithDirty() string {
	return currentFromBuildInfo(true)
}

// CurrentSemver returns the best available version string without a leading "v".
func CurrentSemver() string {
	return stripLeadingV(Current())
}

// CurrentSemverWithDirty returns the best available version string without a
// leading "v", including dirty suffix when available.
func CurrentSemverWithDirty() string {
	return stripLeadingV(CurrentWithDirty())
}

// Module returns the module path from build info when available.
// If unavailable, it returns the configured fallback or "unknown".
func Module() string {
	info, ok := debug.ReadBuildInfo()
	if ok {
		return moduleFromBuildInfo(info)
	}
	return moduleFromBuildInfo(nil)
}

// ModuleVersion returns the best available version string for the given module path
// (without dirty suffix). If the path matches the main module, it falls back to
// Current().
func ModuleVersion(path string) string {
	return moduleVersion(path, false)
}

// ModuleVersionWithDirty returns the best available version string for the given
// module path (including dirty suffix when available). If the path matches the
// main module, it falls back to CurrentWithDirty().
func ModuleVersionWithDirty(path string) string {
	return moduleVersion(path, true)
}

// ModuleVersionSemver returns the best available version string for the given
// module path without a leading "v".
func ModuleVersionSemver(path string) string {
	return stripLeadingV(ModuleVersion(path))
}

// ModuleVersionSemverWithDirty returns the best available version string for the
// given module path without a leading "v", including dirty suffix when available.
func ModuleVersionSemverWithDirty(path string) string {
	return stripLeadingV(ModuleVersionWithDirty(path))
}

func moduleFromBuildInfo(info *debug.BuildInfo) string {
	if info != nil {
		if path := strings.TrimSpace(info.Main.Path); path != "" {
			return path
		}
	}
	fallback := defaultModuleFallback()
	if fallback != "" {
		return fallback
	}
	return "unknown"
}

func moduleVersion(path string, includeDirty bool) string {
	target := strings.TrimSpace(path)
	if target == "" {
		return "v0.0.0-unknown"
	}
	info, ok := debug.ReadBuildInfo()
	if ok && info != nil {
		if strings.TrimSpace(info.Main.Path) == target {
			return currentFromBuildInfo(includeDirty)
		}
		return moduleVersionFromBuildInfo(info, target, includeDirty)
	}
	return "v0.0.0-unknown"
}

func moduleVersionFromBuildInfo(info *debug.BuildInfo, path string, includeDirty bool) string {
	if info == nil || strings.TrimSpace(path) == "" {
		return "v0.0.0-unknown"
	}
	for _, dep := range info.Deps {
		if dep == nil {
			continue
		}
		if strings.TrimSpace(dep.Path) == path {
			if v := versionFromModule(dep, includeDirty); v != "" {
				return v
			}
			return "v0.0.0-unknown"
		}
		if dep.Replace != nil && strings.TrimSpace(dep.Replace.Path) == path {
			if v := versionFromModule(dep.Replace, includeDirty); v != "" {
				return v
			}
			return "v0.0.0-unknown"
		}
	}
	return "v0.0.0-unknown"
}

func versionFromModule(module *debug.Module, includeDirty bool) string {
	if module == nil {
		return ""
	}
	if v := strings.TrimSpace(module.Version); v != "" && v != "(devel)" {
		return normalizeVersion(v, includeDirty)
	}
	return ""
}

func defaultModuleFallback() string {
	defaultModuleMu.RLock()
	fallback := defaultModule
	defaultModuleMu.RUnlock()
	return fallback
}

func currentFromBuildInfo(includeDirty bool) string {
	if strings.TrimSpace(buildVersion) != "" {
		return normalizeVersion(buildVersion, includeDirty)
	}
	info, ok := debug.ReadBuildInfo()
	if ok {
		if v := strings.TrimSpace(info.Main.Version); v != "" && v != "(devel)" {
			return normalizeVersion(v, includeDirty)
		}
		if v := pseudoFromBuildInfo(info, includeDirty); v != "" {
			return normalizeVersion(v, includeDirty)
		}
	}
	return "v0.0.0-unknown"
}

func normalizeVersion(v string, includeDirty bool) string {
	value := strings.TrimSpace(v)
	if includeDirty {
		return value
	}
	return strings.TrimSuffix(value, "+dirty")
}

func stripLeadingV(v string) string {
	value := strings.TrimSpace(v)
	if strings.HasPrefix(value, "v") {
		return strings.TrimPrefix(value, "v")
	}
	return value
}

func pseudoFromBuildInfo(info *debug.BuildInfo, includeDirty bool) string {
	if info == nil {
		return ""
	}
	var revision string
	var vcsTime string
	var modified bool
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.time":
			vcsTime = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	if revision == "" || vcsTime == "" {
		return ""
	}
	parsed, err := time.Parse(time.RFC3339, vcsTime)
	if err != nil {
		return ""
	}
	rev := revision
	if len(rev) > 12 {
		rev = rev[:12]
	}
	ver := "v0.0.0-" + parsed.UTC().Format("20060102150405") + "-" + rev
	if modified && includeDirty {
		ver += "+dirty"
	}
	return ver
}
