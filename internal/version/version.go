// Package version carries what a build knows about itself: which release it
// is, and which commit it was cut from. A release stamps both in at link
// time; a build nobody stamped falls back to what the Go toolchain recorded
// in the binary, so that a `go install ...@v0.1.0` names its release and a
// build from a checkout names its commit without the source tree they came
// from.
package version

import (
	"runtime/debug"
	"strings"
)

// The unstamped spellings. The fallback keys on these exactly: a value that
// differs from them was stamped by someone, and what they stamped wins.
const (
	defaultVersion = "0.0.0-dev"
	defaultDigest  = "unknown"
)

// Version is the release this build claims. A build nobody stamped says so,
// rather than claiming a release it is not.
var Version = defaultVersion

// BuildDigest is the short commit this build was cut from, or "unknown" when
// nobody stamped it in.
var BuildDigest = defaultDigest

// String is the one spelling of a build's identity: the release, then the
// commit it came from, so that a bug report quoting it names both. Build
// info is read on every call; it is cheap, and the answer never changes.
func String() string {
	info, ok := debug.ReadBuildInfo()
	version, digest := resolve(Version, BuildDigest, info, ok)
	return version + " (" + digest + ")"
}

// resolve decides what a build calls itself. A stamped value stands as it
// is; a default is filled in from build info where build info knows better:
// the main module's version for the release, the VCS revision for the digest.
// Missing build info leaves both as they are.
func resolve(version, digest string, info *debug.BuildInfo, ok bool) (string, string) {
	if !ok || info == nil {
		return version, digest
	}
	if version == defaultVersion {
		version = moduleVersion(info.Main.Version, version)
	}
	if digest == defaultDigest {
		digest = vcsDigest(info.Settings, digest)
	}
	return version, digest
}

// moduleVersion is the release the toolchain recorded for the main module,
// without its leading "v", or fallback when it recorded nothing worth
// repeating: "(devel)" is what a plain build says, and says nothing.
func moduleVersion(recorded, fallback string) string {
	if recorded == "" || recorded == "(devel)" {
		return fallback
	}
	return strings.TrimPrefix(recorded, "v")
}

// shortDigestLen is how much of a commit hash names the commit here: the
// seven characters git itself abbreviates to, and what the release stamps.
const shortDigestLen = 7

// vcsDigest is the short commit the toolchain recorded, "-dirty" appended
// when the tree it built from had uncommitted changes, or fallback when no
// revision was recorded -- a build from the module cache has no tree at all.
func vcsDigest(settings []debug.BuildSetting, fallback string) string {
	revision, modified := "", false
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	if revision == "" {
		return fallback
	}
	if len(revision) > shortDigestLen {
		revision = revision[:shortDigestLen]
	}
	if modified {
		revision += "-dirty"
	}
	return revision
}
