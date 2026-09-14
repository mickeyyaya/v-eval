// Package version carries what a build knows about itself: which release it
// is, and which commit it was cut from. Both are stamped in at link time, so
// a binary can answer for itself without the source tree it came from.
package version

// Version is the release this build claims. A build nobody stamped says so,
// rather than claiming a release it is not.
var Version = "0.0.0-dev"

// BuildDigest is the short commit this build was cut from, or "unknown" when
// nobody stamped it in.
var BuildDigest = "unknown"

// String is the one spelling of a build's identity: the release, then the
// commit it came from, so that a bug report quoting it names both.
func String() string {
	return Version + " (" + BuildDigest + ")"
}
