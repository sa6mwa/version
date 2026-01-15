// Package version exposes build-time metadata such as the module path and a
// version string derived from Go build info or linker flags.
//
// It prefers values from Go's build info (debug.ReadBuildInfo) when available,
// falling back to a configurable module path and a safe "v0.0.0-unknown"
// version string when no metadata is present (for example, in "go run" or
// development builds without VCS data).
//
// To override the version string at build time, set the buildVersion variable
// via linker flags:
//
//   go build -ldflags "-X pkt.systems/version.buildVersion=v1.2.3"
//
// Example usage:
//
//   package main
//
//   import (
//       "fmt"
//       "pkt.systems/version"
//   )
//
//   func main() {
//       version.SetDefaultModule("example.com/service")
//       fmt.Println("module:", version.Module())
//       fmt.Println("version:", version.Current())
//       fmt.Println("dep version:", version.ModuleVersion("example.com/service"))
//       fmt.Println("version (no v):", version.CurrentSemver())
//   }
package version
