package version

import (
	"runtime/debug"
	"testing"
)

// buildInfo constructs the build info a binary reads back about itself: the
// main module's version and whichever VCS settings the linker recorded.
func buildInfo(mainVersion string, settings ...debug.BuildSetting) *debug.BuildInfo {
	return &debug.BuildInfo{
		Main:     debug.Module{Path: "github.com/mickeyyaya/v-eval", Version: mainVersion},
		Settings: settings,
	}
}

func setting(key, value string) debug.BuildSetting {
	return debug.BuildSetting{Key: key, Value: value}
}

// TestResolvePrefersStampedValuesAndFallsBackToBuildInfo pins where a build's
// identity comes from. A value the linker stamped wins outright; only the
// defaults are filled in from build info, and only from what the build info
// actually knows -- "(devel)" and an empty version say nothing.
func TestResolvePrefersStampedValuesAndFallsBackToBuildInfo(t *testing.T) {
	const fullRevision = "2959506c7f1e4a9b8d0e6f5a4b3c2d1e0f9a8b7c"
	cases := []struct {
		name        string
		version     string
		digest      string
		info        *debug.BuildInfo
		ok          bool
		wantVersion string
		wantDigest  string
	}{
		{"stamped values win", "0.1.0", "c591242",
			buildInfo("v9.9.9", setting("vcs.revision", fullRevision), setting("vcs.modified", "true")), true,
			"0.1.0", "c591242"},
		{"missing build info leaves the defaults", defaultVersion, defaultDigest, nil, false,
			defaultVersion, defaultDigest},
		{"nil build info leaves the defaults", defaultVersion, defaultDigest, nil, true,
			defaultVersion, defaultDigest},
		{"module version fills the default, v stripped", defaultVersion, defaultDigest,
			buildInfo("v0.1.0"), true, "0.1.0", defaultDigest},
		{"devel is ignored", defaultVersion, defaultDigest, buildInfo("(devel)"), true,
			defaultVersion, defaultDigest},
		{"empty version is ignored", defaultVersion, defaultDigest, buildInfo(""), true,
			defaultVersion, defaultDigest},
		{"revision is shortened to seven characters", defaultVersion, defaultDigest,
			buildInfo("(devel)", setting("vcs.revision", fullRevision)), true,
			defaultVersion, "2959506"},
		{"a modified tree is marked dirty", defaultVersion, defaultDigest,
			buildInfo("(devel)", setting("vcs.revision", fullRevision), setting("vcs.modified", "true")), true,
			defaultVersion, "2959506-dirty"},
		{"an unmodified tree is not marked", defaultVersion, defaultDigest,
			buildInfo("(devel)", setting("vcs.revision", fullRevision), setting("vcs.modified", "false")), true,
			defaultVersion, "2959506"},
		{"a short revision is kept whole", defaultVersion, defaultDigest,
			buildInfo("(devel)", setting("vcs.revision", "abc")), true, defaultVersion, "abc"},
		{"modified without a revision names nothing", defaultVersion, defaultDigest,
			buildInfo("(devel)", setting("vcs.modified", "true")), true, defaultVersion, defaultDigest},
		{"both fall back together", defaultVersion, defaultDigest,
			buildInfo("v0.2.0", setting("vcs.revision", fullRevision)), true, "0.2.0", "2959506"},
		{"a stamped version keeps a build-info digest", "0.1.0", defaultDigest,
			buildInfo("v0.1.0", setting("vcs.revision", fullRevision)), true, "0.1.0", "2959506"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotVersion, gotDigest := resolve(tc.version, tc.digest, tc.info, tc.ok)
			if gotVersion != tc.wantVersion || gotDigest != tc.wantDigest {
				t.Errorf("resolve(%q, %q, ...) = %q, %q, want %q, %q",
					tc.version, tc.digest, gotVersion, gotDigest, tc.wantVersion, tc.wantDigest)
			}
		})
	}
}

// TestStringSpellsStampedValuesAsBefore holds the stamped path to the one
// spelling a release note quotes: the release, then the commit in parens.
func TestStringSpellsStampedValuesAsBefore(t *testing.T) {
	version, digest := Version, BuildDigest
	t.Cleanup(func() { Version, BuildDigest = version, digest })
	Version, BuildDigest = "0.1.0", "c591242"
	if got, want := String(), "0.1.0 (c591242)"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

// TestDefaultsAreTheUnstampedSpelling pins the two sentinels the fallback
// keys on: a build nobody stamped says exactly this, and nothing else does.
func TestDefaultsAreTheUnstampedSpelling(t *testing.T) {
	if defaultVersion != "0.0.0-dev" || defaultDigest != "unknown" {
		t.Errorf("defaults = %q, %q; want 0.0.0-dev, unknown", defaultVersion, defaultDigest)
	}
}
