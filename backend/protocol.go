package backend

import "strings"

// m3u8Scheme is the custom URL scheme registered with the OS by the installer.
// A browser launching the app through an <a href="m3u8://…"> link passes the
// whole link as the first command-line argument.
const m3u8Scheme = "m3u8:"

// protocolURL converts a single m3u8:// launch argument into the http(s) URL
// it points at. Accepted forms (the browser may or may not keep the http(s)
// scheme after the protocol prefix):
//
//	m3u8://example.com/video.m3u8
//	m3u8://https://example.com/video.m3u8
//	m3u8:https://example.com/video.m3u8
func protocolURL(arg string) string {
	arg = strings.TrimSpace(arg)
	if !strings.HasPrefix(strings.ToLower(arg), m3u8Scheme) {
		return ""
	}
	rest := strings.TrimLeft(arg[len(m3u8Scheme):], "/")
	if rest == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(rest), "http://") || strings.HasPrefix(strings.ToLower(rest), "https://") {
		return rest
	}
	return "https://" + rest
}

// protocolURLFromArgs returns the first m3u8:// target found in launch
// arguments, or "" when the app wasn't launched through the protocol.
func protocolURLFromArgs(args []string) string {
	for _, arg := range args {
		if u := protocolURL(arg); u != "" {
			return u
		}
	}
	return ""
}
