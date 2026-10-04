package main

import (
	"runtime/debug"
	"strings"

	skillembed "github.com/mpyw/go-skill-embed"

	"github.com/mpyw/molint"
)

// skills is the command that puts the skills where an agent reads them. It
// carries the release of this binary, so an installed copy names the version
// of the rules it describes.
//
//declscope:package // main.go hands it the arguments before the driver sees them
var skills = skillembed.NewInstaller(
	molint.Skills,
	skillembed.WithToolName("molint"),
	skillembed.WithVersion(skillsVersion(skillsRelease, debug.ReadBuildInfo)),
)

// skillsRelease is the release a release build stamps, with -X
// main.skillsRelease=<version>. .goreleaser.yaml passes it. It is empty in
// every other build.
//
// The module version the go command records is not enough for a release.
// goreleaser builds every platform into dist/ in one checkout, so each build
// after the first sees untracked files and records the version as +dirty.
var skillsRelease string

// skillsVersion is the release this binary was built from, without its
// leading v, or "devel". The stamp comes first. Then the module version the
// go command records, which is what `go install <pkg>@<v>` gives. A build of
// a checkout has neither.
func skillsVersion(stamp string, buildInfo func() (*debug.BuildInfo, bool)) string {
	if stamp != "" {
		return strings.TrimPrefix(stamp, "v")
	}
	if info, ok := buildInfo(); ok {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return strings.TrimPrefix(v, "v")
		}
	}
	return "devel"
}
