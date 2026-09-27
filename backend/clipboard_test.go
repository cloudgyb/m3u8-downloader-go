package backend

import (
	"reflect"
	"testing"
)

func TestExtractM3u8Links(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "single url",
			in:   "https://cdn.example.com/video/index.m3u8",
			want: []string{"https://cdn.example.com/video/index.m3u8"},
		},
		{
			name: "multiple urls newline separated",
			in:   "https://a.example.com/1.m3u8\nhttps://b.example.com/2.m3u8",
			want: []string{"https://a.example.com/1.m3u8", "https://b.example.com/2.m3u8"},
		},
		{
			name: "urls embedded in prose",
			in:   "看这个 https://a.example.com/1.m3u8 还有这个 https://b.example.com/2.m3u8！",
			want: []string{"https://a.example.com/1.m3u8", "https://b.example.com/2.m3u8"},
		},
		{
			name: "dedupes repeated url",
			in:   "https://a.example.com/1.m3u8 https://a.example.com/1.m3u8",
			want: []string{"https://a.example.com/1.m3u8"},
		},
		{
			name: "ignores non-m3u8 urls",
			in:   "https://example.com/page https://example.com/x.m3u8",
			want: []string{"https://example.com/x.m3u8"},
		},
		{
			name: "strips trailing punctuation",
			in:   "(https://a.example.com/1.m3u8),",
			want: []string{"https://a.example.com/1.m3u8"},
		},
		{
			name: "uppercase m3u8",
			in:   "https://a.example.com/1.M3U8",
			want: []string{"https://a.example.com/1.M3U8"},
		},
		{
			name: "query string preserved",
			in:   "https://a.example.com/1.m3u8?token=abc&expires=1",
			want: []string{"https://a.example.com/1.m3u8?token=abc&expires=1"},
		},
		{
			name: "no links",
			in:   "hello world, nothing to see here",
			want: []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := extractM3u8Links(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %#v, want %#v", got, c.want)
			}
		})
	}
}
