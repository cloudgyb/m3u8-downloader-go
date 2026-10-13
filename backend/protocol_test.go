package backend

import "testing"

func TestParseProtocolLink(t *testing.T) {
	cases := []struct {
		in        string
		wantURL   string
		wantTitle string
		ok        bool
	}{
		{"m3u8://download?title=MyVideo&url=https://example.com/v.m3u8", "https://example.com/v.m3u8", "MyVideo", true},
		// m3u8 targets with their own query params survive verbatim.
		{"m3u8://download?title=MyVideo&url=https://example.com/v.m3u8?token=abc&expires=123", "https://example.com/v.m3u8?token=abc&expires=123", "MyVideo", true},
		// Empty title falls through as an empty string.
		{"m3u8://download?title=&url=https://example.com/v.m3u8", "https://example.com/v.m3u8", "", true},
		// Percent-encoded (non-ASCII) title is decoded.
		{"m3u8://download?title=%E6%88%91%E7%9A%84%E8%A7%86%E9%A2%91&url=https://example.com/v.m3u8", "https://example.com/v.m3u8", "我的视频", true},
		// No title parameter at all.
		{"m3u8://download?url=https://example.com/v.m3u8", "https://example.com/v.m3u8", "", true},
		// Not a protocol argument.
		{"https://example.com/plain.m3u8", "", "", false},
		{"m3u8://download?title=x", "", "", false},
		{"m3u8://download?title=x&url=", "", "", false},
	}
	for _, c := range cases {
		link, ok := parseProtocolLink(c.in)
		if ok != c.ok || link.URL != c.wantURL || link.Title != c.wantTitle {
			t.Errorf("parseProtocolLink(%q) = (%+v, %v), want (%q, %q, %v)",
				c.in, link, ok, c.wantURL, c.wantTitle, c.ok)
		}
	}
}

func TestProtocolLinkFromArgs(t *testing.T) {
	link, ok := protocolLinkFromArgs([]string{"m3u8-downloader.exe", "m3u8://download?title=a&url=https://example.com/v.m3u8"})
	if !ok || link.URL != "https://example.com/v.m3u8" || link.Title != "a" {
		t.Fatalf("unexpected: %+v %v", link, ok)
	}
	if _, ok := protocolLinkFromArgs([]string{"m3u8-downloader.exe"}); ok {
		t.Fatal("expected no protocol link")
	}
}
