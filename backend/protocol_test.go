package backend

import "testing"

func TestProtocolURL(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"m3u8://https://example.com/video.m3u8", "https://example.com/video.m3u8"},
		{"m3u8://http://example.com/video.m3u8", "http://example.com/video.m3u8"},
		{"m3u8://example.com/video.m3u8", "https://example.com/video.m3u8"},
		{"m3u8:https://example.com/video.m3u8", "https://example.com/video.m3u8"},
		{"M3U8://example.com/video.m3u8", "https://example.com/video.m3u8"},
		{"  m3u8://example.com/video.m3u8  ", "https://example.com/video.m3u8"},
		{"https://example.com/plain.m3u8", ""}, // not a protocol arg
		{"m3u8://", ""},                         // nothing after the scheme
	}
	for _, c := range cases {
		if got := protocolURL(c.in); got != c.want {
			t.Errorf("protocolURL(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestProtocolURLFromArgs(t *testing.T) {
	if got := protocolURLFromArgs([]string{"m3u8-downloader.exe", "m3u8://example.com/v.m3u8"}); got != "https://example.com/v.m3u8" {
		t.Fatalf("unexpected: %q", got)
	}
	if got := protocolURLFromArgs([]string{"m3u8-downloader.exe"}); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
