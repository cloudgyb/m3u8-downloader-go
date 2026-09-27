//go:build windows

package backend

import "golang.org/x/sys/windows"

var procClipboardSequence = windows.NewLazySystemDLL("user32.dll").NewProc("GetClipboardSequenceNumber")

// clipboardSequence returns the Windows clipboard sequence number, which is
// incremented on every clipboard write — even when the new text is identical —
// so re-copying the same link is still detected as a change. It returns 0 when
// the value can't be read; the watcher then falls back to text comparison.
func clipboardSequence() uint32 {
	r, _, _ := procClipboardSequence.Call()
	return uint32(r)
}
