//go:build windows

package version

// ANSI escape codes are not supported on Windows; use empty strings so that
// version output is printed without colour sequences.
var (
	colorReset = ""
	colorGreen = ""
)
