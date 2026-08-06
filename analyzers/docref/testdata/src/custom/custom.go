// Package custom — boundary fixture of GID-262: the marker list comes from
// settings (patterns replace the built-in list, extra adds the tracker key).
package custom

// Upload starts the export (UDMP-3762) — the tracker key is in settings.extra. // want `GID-262: comment references development documentation`
func Upload() bool { return true }

// Cleanup drops the segment (see the wiki page "segments"). // want `GID-262: comment references development documentation`
func Cleanup() bool { return true }

// Resend repeats the upload (ARD Р-11, задача 29) — the built-in markers are
// replaced by settings.patterns, so this comment stays clean.
func Resend() bool { return true }
