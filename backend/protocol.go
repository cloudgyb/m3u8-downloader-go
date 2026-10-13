package backend

import (
	"net/url"
	"strings"
)

// ProtocolLink is a m3u8:// launch payload: the http(s) download target plus an
// optional suggested save name from the "title" query parameter.
type ProtocolLink struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// m3u8Scheme is the custom URL scheme registered with the OS by the installer.
// The protocol uses the form "m3u8://download?title=<name>&url=<http(s) url>".
const m3u8Scheme = "m3u8://"

// parseProtocolLink parses a single m3u8:// launch argument into its download
// target and suggested save name. The "url" parameter is always last and is
// taken verbatim to the end of the argument, because m3u8 targets frequently
// carry their own auth/sign query params ("?"/"&") that would otherwise be
// ambiguous. "title" is percent-decoded so non-ASCII names survive the trip
// through the browser/OS command line.
func parseProtocolLink(arg string) (ProtocolLink, bool) {
	arg = strings.TrimSpace(arg)
	if !strings.HasPrefix(strings.ToLower(arg), m3u8Scheme) {
		return ProtocolLink{}, false
	}
	rest := arg[len(m3u8Scheme):]
	qIdx := strings.Index(rest, "?")
	if qIdx < 0 {
		return ProtocolLink{}, false
	}
	query := rest[qIdx+1:]

	// Locate the "url=" parameter. In the fixed layout it follows "title=" and
	// is the final parameter, so everything after it is the target URL.
	urlIdx := strings.Index(query, "&url=")
	if urlIdx < 0 {
		if !strings.HasPrefix(query, "url=") {
			return ProtocolLink{}, false
		}
		urlIdx = 0
	} else {
		urlIdx++ // skip the leading '&'
	}
	target := strings.TrimSpace(query[urlIdx+len("url="):])
	if target == "" {
		return ProtocolLink{}, false
	}

	// "title=" sits before "url=" in the fixed layout.
	var title string
	if tIdx := strings.Index(query[:urlIdx], "title="); tIdx >= 0 {
		title = strings.TrimRight(query[:urlIdx][tIdx+len("title="):], "&")
		if dec, err := url.QueryUnescape(title); err == nil {
			title = dec
		}
		title = strings.TrimSpace(title)
	}

	return ProtocolLink{URL: target, Title: title}, true
}

// protocolLinkFromArgs returns the first m3u8:// launch payload found in the
// launch arguments, or false when the app wasn't launched through the protocol.
func protocolLinkFromArgs(args []string) (ProtocolLink, bool) {
	for _, arg := range args {
		if link, ok := parseProtocolLink(arg); ok {
			return link, true
		}
	}
	return ProtocolLink{}, false
}
