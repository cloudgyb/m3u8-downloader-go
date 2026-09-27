package backend

import (
	"regexp"
	"strings"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// linkRe matches http(s) URLs inside arbitrary clipboard text. Quotes, angle
// brackets, whitespace (including the full-width space) terminate the match; a
// trailing run of common prose punctuation is stripped separately.
var linkRe = regexp.MustCompile(`https?://[^\s\x{3000}"'<>]+`)

// trailingPunct is punctuation that frequently sticks to a copied URL — ASCII
// plus common CJK full-width marks — trimmed from the end of each match.
const trailingPunct = ".,;:!?)]}，。、；：！？）】》〉」』"

// extractM3u8Links returns the unique m3u8 URLs found in `text`, in order of
// first appearance. Only URLs that contain ".m3u8" are treated as download
// links — everything else (a copied web page URL, etc.) is ignored.
func extractM3u8Links(text string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, u := range linkRe.FindAllString(text, -1) {
		u = strings.TrimRight(u, trailingPunct)
		if !strings.Contains(strings.ToLower(u), ".m3u8") {
			continue
		}
		if seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	return out
}

// startClipboardWatcher polls the OS clipboard once per second and, whenever
// its content changes to something containing m3u8 links, emits a
// "clipboard:links" event carrying the list of links. The frontend listens for
// that event and shows a confirmation dialog. Polling lives in the backend so
// it keeps running even when the WebView throttles background timers.
func (a *App) startClipboardWatcher() {
	ctx := a.ctx
	if ctx == nil {
		return
	}
	go func() {
		lastText := ""
		lastSeq := uint32(0)
		first := true
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			text, err := wruntime.ClipboardGetText(ctx)
			if err != nil {
				continue
			}
			seq := clipboardSequence()
			// The first read only establishes a baseline so a link that was
			// already on the clipboard at launch doesn't pop the dialog.
			if first {
				lastText = text
				lastSeq = seq
				first = false
				continue
			}
			// Prefer the clipboard sequence number (Windows): it changes on every
			// write, so re-copying the same link still counts as a change. Fall
			// back to text comparison when no sequence number is available.
			changed := seq != lastSeq
			if seq == 0 || lastSeq == 0 {
				changed = text != lastText
			}
			lastText = text
			lastSeq = seq
			if !changed {
				continue
			}
			// Only emit while the user has auto-detection enabled.
			if !a.clipboardWatch.Load() {
				continue
			}
			if links := extractM3u8Links(text); len(links) > 0 {
				wruntime.EventsEmit(ctx, "clipboard:links", links)
			}
		}
	}()
}
