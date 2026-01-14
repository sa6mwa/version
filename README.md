# pkt.systems/version

Small Go helpers for surfacing build-time version metadata.

This package prefers Go build info (`debug.ReadBuildInfo`) and VCS metadata when
available. When build info is missing (for example, `go run` or builds without
VCS data), it returns safe defaults and allows a configurable module fallback.

## Install

```bash
go get pkt.systems/version@latest
```

## Usage

```go
package main

import (
	"fmt"

	"pkt.systems/version"
)

func main() {
	version.SetDefaultModule("example.com/service")

	fmt.Println("module:", version.Module())
	fmt.Println("version:", version.Current())
	fmt.Println("version (dirty):", version.CurrentWithDirty())
	fmt.Println("dep version:", version.ModuleVersion("example.com/service"))
}
```

## Build-time version override

You can override the version string at build time by setting the
`buildVersion` variable via linker flags:

```bash
go build -ldflags "-X pkt.systems/version.buildVersion=v1.2.3"
```

## API notes

- `version.Current()` returns the best available version string without a dirty
  suffix.
- `version.CurrentWithDirty()` includes a `+dirty` suffix when VCS data indicates
  a modified working tree.
- `version.Module()` returns the module path from build info, or the configured
  fallback, or `"unknown"` if neither is available.
- `version.ModuleVersion(path)` returns the version for a specific dependency
  module path (or the main module if it matches), falling back to
  `"v0.0.0-unknown"` when build info is unavailable.
- `version.SetDefaultModule(path)` sets the fallback module path; pass an empty
  string to clear it.
