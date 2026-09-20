package backend

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseMediaPlaylist(t *testing.T) {
	body := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:9.009,
seg0.ts
#EXTINF:9.009,
seg1.ts
#EXT-X-ENDLIST
`
	base, _ := url.Parse("http://example.com/video/index.m3u8")
	pl := parsePlaylist([]byte(body), base)

	if pl.master {
		t.Fatal("expected a media playlist")
	}
	if len(pl.segments) != 2 {
		t.Fatalf("want 2 segments, got %d", len(pl.segments))
	}
	if pl.segments[0].uri != "http://example.com/video/seg0.ts" {
		t.Fatalf("bad segment uri: %s", pl.segments[0].uri)
	}
	if pl.segments[0].seq != 0 || pl.segments[1].seq != 1 {
		t.Fatalf("bad sequence numbers: %d, %d", pl.segments[0].seq, pl.segments[1].seq)
	}
	if pl.totalDur != 18.018 {
		t.Fatalf("bad total duration: %f", pl.totalDur)
	}
}

func TestParseMasterPlaylist(t *testing.T) {
	body := `#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=1280x720,CODECS="avc1.4d401f"
low/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=2400000,RESOLUTION=1920x1080
hi/index.m3u8
`
	base, _ := url.Parse("http://example.com/master.m3u8")
	pl := parsePlaylist([]byte(body), base)

	if !pl.master {
		t.Fatal("expected a master playlist")
	}
	if len(pl.variants) != 2 {
		t.Fatalf("want 2 variants, got %d", len(pl.variants))
	}
	if pl.variants[0].URI != "http://example.com/low/index.m3u8" {
		t.Fatalf("bad variant uri: %s", pl.variants[0].URI)
	}
	if pl.variants[0].Bandwidth != 800000 {
		t.Fatalf("bad bandwidth: %d", pl.variants[0].Bandwidth)
	}
	if pl.variants[0].Resolution != "1280x720" {
		t.Fatalf("bad resolution: %s", pl.variants[0].Resolution)
	}
	if pl.variants[0].Codecs != "avc1.4d401f" {
		t.Fatalf("bad codecs: %s", pl.variants[0].Codecs)
	}
	if pl.variants[1].URI != "http://example.com/hi/index.m3u8" {
		t.Fatalf("bad variant uri: %s", pl.variants[1].URI)
	}
	if pl.variants[1].Bandwidth != 2400000 {
		t.Fatalf("bad bandwidth: %d", pl.variants[1].Bandwidth)
	}
}

func TestParseKey(t *testing.T) {
	base, _ := url.Parse("http://example.com/a/video/index.m3u8")
	k := parseKey(`#EXT-X-KEY:METHOD=AES-128,URI="key.bin",IV=0x000000000000000000000000000000FF`, base)

	if k == nil {
		t.Fatal("expected a non-nil key")
	}
	if k.method != "AES-128" {
		t.Fatalf("bad method: %s", k.method)
	}
	if k.uri != "http://example.com/a/video/key.bin" {
		t.Fatalf("key URI not resolved: %s", k.uri)
	}
	if len(k.iv) != 16 || k.iv[15] != 0xFF {
		t.Fatalf("bad IV: %x", k.iv)
	}
}

func TestAESDecryptPadded(t *testing.T) {
	key := make([]byte, 16)
	for i := range key {
		key[i] = byte(i)
	}
	iv := make([]byte, 16)
	plain := []byte("hello world")

	block, _ := aes.NewCipher(key)
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	padded := append(append([]byte{}, plain...), bytes.Repeat([]byte{byte(pad)}, pad)...)
	enc := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(enc, padded)

	out, err := aesDecrypt(key, iv, enc)
	if err != nil {
		t.Fatalf("decrypt error: %v", err)
	}
	if !bytes.Equal(out, plain) {
		t.Fatalf("got %q, want %q", out, plain)
	}
}

func TestBaseName(t *testing.T) {
	if got := baseName("", "http://example.com/video/my show.m3u8"); got != "my show" {
		t.Fatalf("derive from URL: got %q", got)
	}
	if got := baseName("a/b:c*.mkv", ""); got != "a_b_c_" {
		t.Fatalf("illegal chars + strip ext: got %q", got)
	}
	if got := baseName("plain", ""); got != "plain" {
		t.Fatalf("plain name: got %q", got)
	}
}

func TestFinalOutputPath(t *testing.T) {
	if got := finalOutputPath("C:/dl", "movie", "", "mp4"); got != filepath.Join("C:/dl", "movie.mp4") {
		t.Fatalf("mp4: got %q", got)
	}
	if got := finalOutputPath("C:/dl", "movie.mkv", "", "mp4"); got != filepath.Join("C:/dl", "movie.mp4") {
		t.Fatalf("strip ext: got %q", got)
	}
	if got := finalOutputPath("C:/dl", "", "http://x/a/v.m3u8", "ts"); got != filepath.Join("C:/dl", "v.ts") {
		t.Fatalf("ts from url: got %q", got)
	}
}

func ffmpegName() string {
	if runtime.GOOS == "windows" {
		return "ffmpeg.exe"
	}
	return "ffmpeg"
}

func TestFFmpegPathLocal(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, ffmpegName())
	if err := os.WriteFile(p, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := ffmpegPath(dir); got != p {
		t.Fatalf("want %q, got %q", p, got)
	}
}

func TestFFmpegPathBinSubdir(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(bin, ffmpegName())
	if err := os.WriteFile(p, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := ffmpegPath(dir); got != p {
		t.Fatalf("want %q, got %q", p, got)
	}
}

func TestLoadSessionResume(t *testing.T) {
	dir := t.TempDir()
	req := DownloadRequest{URL: "http://x/v.m3u8", Name: "v"}
	pl := &playlist{segments: []segment{
		{uri: "http://x/s0.ts"},
		{uri: "http://x/s1.ts"},
		{uri: "http://x/s2.ts"},
	}}

	s := loadSession(dir, req, "", pl)
	if len(s.completed) != 0 {
		t.Fatal("fresh session should have no completed segments")
	}
	s.markComplete(0, 100)
	s.markComplete(1, 200)
	if s.man.DoneBytes != 300 {
		t.Fatalf("bad done bytes: %d", s.man.DoneBytes)
	}

	// Reload from disk as if the app restarted.
	s2 := loadSession(dir, req, "", pl)
	if !s2.isComplete(0) || !s2.isComplete(1) {
		t.Fatal("resume lost completed segments")
	}
	if s2.isComplete(2) {
		t.Fatal("segment 2 should not be complete")
	}
	if s2.man.DoneBytes != 300 {
		t.Fatalf("bad resumed done bytes: %d", s2.man.DoneBytes)
	}
}

func TestLoadSessionRejectsMismatch(t *testing.T) {
	dir := t.TempDir()
	req := DownloadRequest{URL: "http://x/v.m3u8", Name: "v"}
	pl := &playlist{segments: []segment{
		{uri: "http://x/s0.ts"},
		{uri: "http://x/s1.ts"},
	}}
	s := loadSession(dir, req, "", pl)
	s.markComplete(0, 100)
	s.markComplete(1, 100)

	// Playlist content changed → partial state must be discarded.
	pl2 := &playlist{segments: []segment{
		{uri: "http://x/NEW.ts"},
		{uri: "http://x/s1.ts"},
	}}
	s2 := loadSession(dir, req, "", pl2)
	if len(s2.completed) != 0 {
		t.Fatalf("expected mismatch to reset completed, got %d", len(s2.completed))
	}
}
