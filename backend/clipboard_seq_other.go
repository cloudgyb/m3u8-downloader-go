//go:build !windows

package backend

// clipboardSequence returns 0 on platforms without a clipboard sequence
// number; the watcher falls back to text comparison there.
func clipboardSequence() uint32 {
	return 0
}
