package main

import (
	"flag"
	"fmt"

	"pkt.systems/version"
)

func main() {
	semver := flag.Bool("semver", false, "strip leading v from version output")
	dirty := flag.Bool("dirty", false, "include +dirty when available")
	flag.Parse()

	var value string
	switch {
	case *semver && *dirty:
		value = version.CurrentSemverWithDirty()
	case *semver:
		value = version.CurrentSemver()
	case *dirty:
		value = version.CurrentWithDirty()
	default:
		value = version.Current()
	}

	fmt.Println(value)
}
