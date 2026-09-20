package backend

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/proxy"
)

// ── request / event models ────────────────────────────────────────────────

// ProxyConfig describes the optional proxy to use for downloads.
type ProxyConfig struct {
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"` // "http" | "socks5"
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

// DownloadRequest starts a new download. VariantURI is the chosen media
// playlist (from a master playlist); empty means "best available".
type DownloadRequest struct {
	URL        string      `json:"url"`
	Name       string      `json:"name"`
	SaveDir    string      `json:"saveDir"`
	Threads    int         `json:"threads"`
	VariantURI string      `json:"variantUri"`
	Format     string      `json:"format"` // "mp4" | "ts"
	Proxy      ProxyConfig `json:"proxy"`
}

// DownloadEvent is emitted to the frontend as "download:event".
type DownloadEvent struct {
	ID     int64  `json:"id"`
	Status string `json:"status"` // downloading | completed | failed | canceled
	Done   int64  `json:"done"`
	Total  int64  `json:"total"`
	Speed  int64  `json:"speed"`
	File   string `json:"file"`
	Error  string `json:"error"`
}

// Variant is one rendition of a master playlist.
type Variant struct {
	URI        string `json:"uri"`
	Bandwidth  int64  `json:"bandwidth"`
	Resolution string `json:"resolution"`
	Codecs     string `json:"codecs"`
	Name       string `json:"name"`
}

// ── playlist model & parser ───────────────────────────────────────────────

type segKey struct {
	method string
	uri    string
	iv     []byte
}

type segment struct {
	uri       string
	duration  float64
	key       *segKey
	byteRange string
	seq       int64
}

type playlist struct {
	master   bool
	variants []Variant
	segments []segment
	mediaSeq int64
	totalDur float64
}

func parseInt(s string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return n
}

func parseAttrs(s string) map[string]string {
	m := map[string]string{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if !strings.Contains(part, "=") {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		m[strings.ToUpper(strings.TrimSpace(kv[0]))] = strings.Trim(strings.TrimSpace(kv[1]), `"`)
	}
	return m
}

func parseExtinf(line string) float64 {
	s := strings.TrimSpace(strings.TrimPrefix(line, "#EXTINF:"))
	if i := strings.IndexByte(s, ','); i >= 0 {
		s = s[:i]
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseKey(line string, base *url.URL) *segKey {
	s := strings.TrimPrefix(line, "#EXT-X-KEY:")
	k := &segKey{method: "NONE"}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if !strings.Contains(part, "=") {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		key := strings.TrimSpace(kv[0])
		val := strings.Trim(strings.TrimSpace(kv[1]), `"`)
		switch strings.ToUpper(key) {
		case "METHOD":
			k.method = strings.ToUpper(val)
		case "URI":
			k.uri = val
		case "IV":
			if strings.HasPrefix(strings.ToUpper(val), "0X") {
				k.iv, _ = hex.DecodeString(val[2:])
			}
		}
	}
	if k.method == "NONE" {
		return nil
	}
	if k.uri != "" {
		k.uri = resolveURL(base, k.uri)
	}
	return k
}

func resolveURL(base *url.URL, ref string) string {
	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(u).String()
}

func parsePlaylist(body []byte, base *url.URL) *playlist {
	p := &playlist{}
	var curKey *segKey
	var curRange string
	var lastDur float64
	seq := int64(0)
	pendingSeg := false
	var pendingVariant Variant
	pendingVariantSet := false

	sc := bufio.NewScanner(strings.NewReader(string(body)))
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, "#EXT-X-STREAM-INF"):
			p.master = true
			a := parseAttrs(strings.TrimPrefix(line, "#EXT-X-STREAM-INF:"))
			pendingVariant = Variant{
				Bandwidth:  parseInt(a["BANDWIDTH"]),
				Resolution: a["RESOLUTION"],
				Codecs:     a["CODECS"],
				Name:       a["NAME"],
			}
			pendingVariantSet = true
			pendingSeg = false
		case strings.HasPrefix(line, "#EXT-X-MEDIA:"):
			// audio/subtitle renditions: ignore
		case strings.HasPrefix(line, "#EXTINF"):
			lastDur = parseExtinf(line)
			pendingSeg = true
			pendingVariantSet = false
		case strings.HasPrefix(line, "#EXT-X-KEY:"):
			curKey = parseKey(line, base)
		case strings.HasPrefix(line, "#EXT-X-BYTERANGE:"):
			curRange = strings.TrimSpace(strings.TrimPrefix(line, "#EXT-X-BYTERANGE:"))
		case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"):
			p.mediaSeq = parseInt(strings.TrimPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"))
			seq = p.mediaSeq
		case strings.HasPrefix(line, "#"):
			// other tags / comments
		default:
			u := resolveURL(base, line)
			if pendingVariantSet {
				pendingVariant.URI = u
				p.variants = append(p.variants, pendingVariant)
				pendingVariant = Variant{}
				pendingVariantSet = false
			} else {
				if pendingSeg {
					p.segments = append(p.segments, segment{uri: u, duration: lastDur, key: curKey, byteRange: curRange, seq: seq})
					p.totalDur += lastDur
					pendingSeg = false
					curRange = ""
				} else {
					p.segments = append(p.segments, segment{uri: u, key: curKey, seq: seq})
				}
				seq++
			}
		}
	}
	return p
}

// ── HTTP helpers ──────────────────────────────────────────────────────────

var downloadUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"

func buildClient(p ProxyConfig) (*http.Client, error) {
	tr := &http.Transport{}
	if p.Enabled {
		addr := net.JoinHostPort(p.Host, strconv.Itoa(p.Port))
		if p.Type == "socks5" {
			d, err := proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
			if err != nil {
				return nil, err
			}
			tr.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
				return d.Dial(network, address)
			}
		} else {
			tr.Proxy = http.ProxyURL(&url.URL{Scheme: "http", Host: addr})
		}
	}
	return &http.Client{Transport: tr}, nil
}

func fetchPlaylist(client *http.Client, ctx context.Context, rawURL string) (*playlist, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", downloadUA)
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d while fetching playlist", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	base, _ := url.Parse(rawURL)
	return parsePlaylist(body, base), nil
}

// ── AES-128 decryption ────────────────────────────────────────────────────

func seqToIV(seq int64) []byte {
	iv := make([]byte, 16)
	binary.BigEndian.PutUint64(iv[8:], uint64(seq))
	return iv
}

func aesDecrypt(key, iv, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(iv) < block.BlockSize() {
		iv = make([]byte, block.BlockSize())
	}
	iv = iv[:block.BlockSize()]
	if len(data) == 0 {
		return data, nil
	}
	if len(data)%block.BlockSize() != 0 {
		data = data[:len(data)-len(data)%block.BlockSize()]
	}
	out := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, data)
	return pkcs7Unpad(out), nil
}

func pkcs7Unpad(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	n := int(b[len(b)-1])
	if n > 0 && n <= len(b) {
		valid := true
		for _, v := range b[len(b)-n:] {
			if int(v) != n {
				valid = false
				break
			}
		}
		if valid {
			return b[:len(b)-n]
		}
	}
	return b
}

// ── per-task state ────────────────────────────────────────────────────────

type dlTask struct {
	id      int64
	req     DownloadRequest
	ctx     context.Context
	cancel  context.CancelFunc
	paused  atomic.Bool
	done    atomic.Int64
	total   atomic.Int64
	speed   atomic.Int64
	segDone atomic.Int64
	segTot  int64
	file    string // final output path, set on completion
}

func (t *dlTask) waitIfPaused() {
	for t.paused.Load() {
		if t.ctx.Err() != nil {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// ── resume manifest ───────────────────────────────────────────────────────

type manifest struct {
	URL        string   `json:"url"`
	Name       string   `json:"name"`
	VariantURI string   `json:"variantUri"`
	Segments   []string `json:"segments"`
	Completed  []int    `json:"completed"`
	DoneBytes  int64    `json:"doneBytes"`
}

type dlSession struct {
	mu           sync.Mutex
	workDir      string
	manifestPath string
	man          manifest
	completed    map[int]bool
}

func workDirFor(saveDir, rawURL, name, variant string) string {
	h := sha1.Sum([]byte(rawURL + "|" + name + "|" + variant))
	return filepath.Join(saveDir, ".m3u8dl-"+hex.EncodeToString(h[:])[:16])
}

func loadSession(workDir string, req DownloadRequest, variantURI string, pl *playlist) *dlSession {
	s := &dlSession{
		workDir:      workDir,
		manifestPath: filepath.Join(workDir, "manifest.json"),
		completed:    map[int]bool{},
		man: manifest{
			URL:        req.URL,
			Name:       req.Name,
			VariantURI: variantURI,
			Segments:   make([]string, len(pl.segments)),
		},
	}
	for i, seg := range pl.segments {
		s.man.Segments[i] = seg.uri
	}

	if b, err := os.ReadFile(s.manifestPath); err == nil {
		var old manifest
		if json.Unmarshal(b, &old) == nil &&
			old.URL == req.URL && old.Name == req.Name &&
			len(old.Segments) == len(pl.segments) {
			valid := true
			for _, idx := range old.Completed {
				if idx < 0 || idx >= len(pl.segments) || old.Segments[idx] != pl.segments[idx].uri {
					valid = false
					break
				}
			}
			if valid {
				s.man = old
				for _, idx := range old.Completed {
					s.completed[idx] = true
				}
			}
		}
	}
	return s
}

func (s *dlSession) isComplete(idx int) bool {
	return s.completed[idx]
}

func (s *dlSession) markComplete(idx int, bytes int64) {
	s.mu.Lock()
	if !s.completed[idx] {
		s.completed[idx] = true
		s.man.Completed = append(s.man.Completed, idx)
		s.man.DoneBytes += bytes
		s.saveLocked()
	}
	s.mu.Unlock()
}

func (s *dlSession) save() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.saveLocked()
}

func (s *dlSession) saveLocked() {
	b, err := json.Marshal(s.man)
	if err != nil {
		return
	}
	tmp := s.manifestPath + ".tmp"
	_ = os.WriteFile(tmp, b, 0o644)
	_ = os.Rename(tmp, s.manifestPath)
}

// ── download manager ──────────────────────────────────────────────────────

// DownloadManager owns all in-flight downloads and emits progress events.
type DownloadManager struct {
	mu     sync.Mutex
	tasks  map[int64]*dlTask
	nextID int64
	emit   func(eventName string, data interface{})
}

// NewDownloadManager creates a manager. emit is set once the Wails context
// is available (OnStartup) so events can be pushed to the frontend.
func NewDownloadManager() *DownloadManager {
	return &DownloadManager{tasks: make(map[int64]*dlTask)}
}

// SetEmitter wires the Wails EventsEmit function.
func (m *DownloadManager) SetEmitter(fn func(eventName string, data interface{})) {
	m.emit = fn
}

func (m *DownloadManager) emitEvent(t *dlTask, status, errStr string) {
	ev := DownloadEvent{
		ID:     t.id,
		Status: status,
		Done:   t.done.Load(),
		Total:  t.total.Load(),
		Speed:  t.speed.Load(),
		Error:  errStr,
	}
	if status == "completed" {
		ev.Done = ev.Total
		ev.File = t.file
	}
	if m.emit != nil {
		m.emit("download:event", ev)
	}
}

// FetchVariants returns the renditions of a master playlist, or an empty
// slice when the URL is already a media playlist.
func (m *DownloadManager) FetchVariants(rawURL string, p ProxyConfig) ([]Variant, error) {
	client, err := buildClient(p)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pl, err := fetchPlaylist(client, ctx, rawURL)
	if err != nil {
		return nil, err
	}
	return pl.variants, nil
}

func (m *DownloadManager) Start(req DownloadRequest) (int64, error) {
	if strings.TrimSpace(req.URL) == "" {
		return 0, errors.New("empty url")
	}
	if req.Threads <= 0 {
		req.Threads = 8
	}
	if req.Threads > 64 {
		req.Threads = 64
	}
	if strings.TrimSpace(req.SaveDir) == "" {
		req.SaveDir = "C:/Downloads/M3U8"
	}
	if req.Format != "mp4" && req.Format != "ts" {
		req.Format = "mp4"
	}
	if strings.TrimSpace(req.Name) == "" {
		req.Name = baseName("", req.URL)
	}
	if err := os.MkdirAll(req.SaveDir, 0o755); err != nil {
		return 0, err
	}

	m.mu.Lock()
	id := m.nextID
	m.nextID++
	ctx, cancel := context.WithCancel(context.Background())
	t := &dlTask{id: id, req: req, ctx: ctx, cancel: cancel}
	m.tasks[id] = t
	m.mu.Unlock()

	go m.run(t)
	return id, nil
}

func (m *DownloadManager) Pause(id int64) {
	m.mu.Lock()
	t := m.tasks[id]
	m.mu.Unlock()
	if t != nil {
		t.paused.Store(true)
	}
}

func (m *DownloadManager) Resume(id int64) {
	m.mu.Lock()
	t := m.tasks[id]
	m.mu.Unlock()
	if t != nil {
		t.paused.Store(false)
	}
}

func (m *DownloadManager) Cancel(id int64) {
	m.mu.Lock()
	t := m.tasks[id]
	m.mu.Unlock()
	if t != nil {
		t.cancel()
	}
}

func (m *DownloadManager) cleanup(t *dlTask) {
	m.mu.Lock()
	delete(m.tasks, t.id)
	m.mu.Unlock()
}

// ── download pipeline ─────────────────────────────────────────────────────

func (m *DownloadManager) run(t *dlTask) {
	defer m.cleanup(t)

	client, err := buildClient(t.req.Proxy)
	if err != nil {
		m.emitEvent(t, "failed", err.Error())
		return
	}

	pl, err := fetchPlaylist(client, t.ctx, t.req.URL)
	if err != nil {
		m.emitEvent(t, "failed", err.Error())
		return
	}

	variantURI := ""
	if pl.master {
		if len(pl.variants) == 0 {
			m.emitEvent(t, "failed", "master playlist has no variants")
			return
		}
		chosen := pl.variants[0]
		for _, v := range pl.variants {
			if v.Bandwidth > chosen.Bandwidth {
				chosen = v
			}
		}
		if t.req.VariantURI != "" {
			for _, v := range pl.variants {
				if v.URI == t.req.VariantURI {
					chosen = v
					break
				}
			}
		}
		variantURI = chosen.URI
		pl, err = fetchPlaylist(client, t.ctx, chosen.URI)
		if err != nil {
			m.emitEvent(t, "failed", err.Error())
			return
		}
	}
	if len(pl.segments) == 0 {
		m.emitEvent(t, "failed", "playlist has no segments")
		return
	}

	workDir := workDirFor(t.req.SaveDir, t.req.URL, t.req.Name, variantURI)
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		m.emitEvent(t, "failed", err.Error())
		return
	}
	sess := loadSession(workDir, t.req, variantURI, pl)

	t.segTot = int64(len(pl.segments))
	t.done.Store(sess.man.DoneBytes)
	t.segDone.Store(int64(len(sess.man.Completed)))
	m.updateTotal(t)

	go m.reportProgress(t)

	// Only queue segments that are not yet downloaded.
	missing := make([]int, 0, len(pl.segments))
	for idx := range pl.segments {
		if !sess.isComplete(idx) {
			missing = append(missing, idx)
		}
	}

	if len(missing) == 0 {
		// Fully resumed: everything is already on disk, just merge.
		m.finishDownload(t, sess, workDir, len(pl.segments))
		return
	}

	m.downloadMissing(t, client, pl, workDir, sess, missing)
}

func (m *DownloadManager) downloadMissing(t *dlTask, client *http.Client, pl *playlist, workDir string, sess *dlSession, missing []int) {
	type job struct {
		idx int
		seg segment
	}
	jobs := make(chan job)

	workers := t.req.Threads
	if workers > len(missing) {
		workers = len(missing)
	}

	var firstErr error
	var errOnce sync.Once
	abort := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			keyCache := map[string][]byte{}
			for {
				select {
				case <-abort:
					return
				case <-t.ctx.Done():
					return
				case j, ok := <-jobs:
					if !ok {
						return
					}
					t.waitIfPaused()
					if t.ctx.Err() != nil {
						return
					}
					n, err := downloadSegment(t, client, j.seg, j.idx, workDir, keyCache)
					if err != nil {
						errOnce.Do(func() {
							firstErr = err
							close(abort)
						})
						return
					}
					sess.markComplete(j.idx, n)
					t.done.Add(n)
					t.segDone.Add(1)
					m.updateTotal(t)
				}
			}
		}()
	}

sendLoop:
	for _, idx := range missing {
		select {
		case jobs <- job{idx: idx, seg: pl.segments[idx]}:
		case <-abort:
			break sendLoop
		case <-t.ctx.Done():
			break sendLoop
		}
	}
	close(jobs)
	wg.Wait()

	// User cancellation discards partials; a download error keeps them for resume.
	if t.ctx.Err() != nil {
		os.RemoveAll(workDir)
		m.emitEvent(t, "canceled", "")
		return
	}
	if firstErr != nil {
		sess.save()
		m.emitEvent(t, "failed", firstErr.Error())
		return
	}

	m.finishDownload(t, sess, workDir, len(pl.segments))
}

// finishDownload merges the segments into the final file (remuxing to MP4
// when requested and ffmpeg is available) and emits the terminal event.
func (m *DownloadManager) finishDownload(t *dlTask, sess *dlSession, workDir string, n int) {
	finalPath, err := writeFinalFile(t, workDir, n)
	if err != nil {
		sess.save()
		m.emitEvent(t, "failed", err.Error())
		return
	}
	os.RemoveAll(workDir)
	if fi, statErr := os.Stat(finalPath); statErr == nil {
		t.total.Store(fi.Size())
		t.done.Store(fi.Size())
	}
	t.file = finalPath
	m.emitEvent(t, "completed", "")
}

func (m *DownloadManager) reportProgress(t *dlTask) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	lastDone := int64(0)
	last := time.Now()
	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			done := t.done.Load()
			dt := time.Since(last).Seconds()
			speed := int64(0)
			if dt > 0 {
				speed = int64(float64(done-lastDone) / dt)
				if speed < 0 {
					speed = 0
				}
			}
			lastDone = done
			last = time.Now()
			if t.paused.Load() {
				speed = 0
			}
			t.speed.Store(speed)
			m.emitEvent(t, "downloading", "")
		}
	}
}

func (m *DownloadManager) updateTotal(t *dlTask) {
	done := t.done.Load()
	completed := t.segDone.Load()
	remaining := t.segTot - completed
	if completed <= 0 || remaining <= 0 {
		return
	}
	avg := done / completed
	total := done + avg*remaining
	for {
		cur := t.total.Load()
		if total <= cur {
			break
		}
		if t.total.CompareAndSwap(cur, total) {
			break
		}
	}
}

// ── segment download & merge ──────────────────────────────────────────────

func formatByteRange(br string) string {
	if i := strings.IndexByte(br, '@'); i >= 0 {
		length := parseInt(br[:i])
		offset := parseInt(br[i+1:])
		if length > 0 {
			return fmt.Sprintf("%d-%d", offset, offset+length-1)
		}
	}
	length := parseInt(br)
	if length > 0 {
		return fmt.Sprintf("0-%d", length-1)
	}
	return ""
}

func resolveKey(client *http.Client, ctx context.Context, k *segKey, cache map[string][]byte) ([]byte, error) {
	if b, ok := cache[k.uri]; ok {
		return b, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", downloadUA)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d fetching key", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	cache[k.uri] = b
	return b, nil
}

func downloadSegment(t *dlTask, client *http.Client, seg segment, idx int, workDir string, keyCache map[string][]byte) (int64, error) {
	req, err := http.NewRequestWithContext(t.ctx, http.MethodGet, seg.uri, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", downloadUA)
	if br := formatByteRange(seg.byteRange); br != "" {
		req.Header.Set("Range", "bytes="+br)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("HTTP %d for %s", resp.StatusCode, seg.uri)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if seg.key != nil && seg.key.method == "AES-128" {
		key, err := resolveKey(client, t.ctx, seg.key, keyCache)
		if err != nil {
			return 0, err
		}
		iv := seg.key.iv
		if len(iv) == 0 {
			iv = seqToIV(seg.seq)
		}
		data, err = aesDecrypt(key, iv, data)
		if err != nil {
			return 0, err
		}
	}

	dst := filepath.Join(workDir, fmt.Sprintf("seg_%06d.ts", idx))
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return 0, err
	}
	return int64(len(data)), nil
}

func mergeSegments(workDir, outPath string, n int) error {
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	bw := bufio.NewWriterSize(out, 1<<20)
	for i := 0; i < n; i++ {
		src := filepath.Join(workDir, fmt.Sprintf("seg_%06d.ts", i))
		f, err := os.Open(src)
		if err != nil {
			return err
		}
		_, err = io.Copy(bw, f)
		f.Close()
		if err != nil {
			return err
		}
	}
	if err := bw.Flush(); err != nil {
		return err
	}
	return nil
}

// ── output naming & remux ────────────────────────────────────────────────

var illegalNameChars = []string{`/`, `\`, `:`, `*`, `?`, `"`, `<`, `>`, `|`}

var mediaExts = map[string]bool{
	".ts": true, ".mp4": true, ".m4v": true, ".mkv": true, ".mov": true,
	".flv": true, ".avi": true, ".webm": true, ".m4a": true, ".mp3": true,
}

// baseName returns a sanitized base filename without an extension.
func baseName(name, rawURL string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		if u, err := url.Parse(rawURL); err == nil {
			base := filepath.Base(u.Path)
			name = strings.TrimSuffix(base, filepath.Ext(base))
		}
	}
	for _, c := range illegalNameChars {
		name = strings.ReplaceAll(name, c, "_")
	}
	name = strings.Trim(name, ". ")
	if ext := filepath.Ext(name); mediaExts[ext] {
		name = strings.TrimSuffix(name, ext)
	}
	if name == "" {
		name = "video"
	}
	return name
}

// finalOutputPath builds the final output path with the format's extension.
func finalOutputPath(saveDir, name, rawURL, format string) string {
	ext := ".ts"
	if format == "mp4" {
		ext = ".mp4"
	}
	return filepath.Join(saveDir, baseName(name, rawURL)+ext)
}

// ffmpegPath resolves the ffmpeg binary, preferring one bundled next to the
// executable (or in a bin/ subdirectory) before falling back to PATH.
func ffmpegPath(exeDir string) string {
	name := "ffmpeg"
	if runtime.GOOS == "windows" {
		name = "ffmpeg.exe"
	}
	for _, cand := range []string{
		filepath.Join(exeDir, name),
		filepath.Join(exeDir, "bin", name),
	} {
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			return cand
		}
	}
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p
	}
	return ""
}

func executableDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	return "."
}

// ffmpegAvailable reports whether a usable ffmpeg binary can be found.
func ffmpegAvailable() bool {
	return ffmpegPath(executableDir()) != ""
}

func remuxToMp4(ctx context.Context, input, output string) error {
	ffmpeg := ffmpegPath(executableDir())
	if ffmpeg == "" {
		return errors.New("ffmpeg not found")
	}
	cmd := exec.CommandContext(ctx, ffmpeg,
		"-y", "-hide_banner", "-loglevel", "error",
		"-i", input, "-c", "copy", "-movflags", "+faststart", output)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// writeFinalFile merges segments and, when format is "mp4", remuxes the TS
// stream to MP4 (falling back to a plain TS file if ffmpeg is missing/fails).
func writeFinalFile(t *dlTask, workDir string, n int) (string, error) {
	if t.req.Format != "mp4" {
		finalPath := finalOutputPath(t.req.SaveDir, t.req.Name, t.req.URL, "ts")
		return finalPath, mergeSegments(workDir, finalPath, n)
	}

	mp4Path := finalOutputPath(t.req.SaveDir, t.req.Name, t.req.URL, "mp4")
	if !ffmpegAvailable() {
		tsPath := finalOutputPath(t.req.SaveDir, t.req.Name, t.req.URL, "ts")
		return tsPath, mergeSegments(workDir, tsPath, n)
	}

	mergedTs := filepath.Join(workDir, "merged.ts")
	if err := mergeSegments(workDir, mergedTs, n); err != nil {
		return "", err
	}
	if err := remuxToMp4(t.ctx, mergedTs, mp4Path); err != nil {
		os.Remove(mp4Path)
		tsPath := finalOutputPath(t.req.SaveDir, t.req.Name, t.req.URL, "ts")
		if rerr := os.Rename(mergedTs, tsPath); rerr != nil {
			os.Remove(mergedTs)
			return "", err
		}
		return tsPath, nil
	}
	os.Remove(mergedTs)
	return mp4Path, nil
}
