package backend

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"

	"github.com/wailsapp/wails/v2/pkg/options"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the backend service bound to the frontend. Every exported method
// becomes callable from JavaScript through Wails' generated bindings.
type App struct {
	ctx             context.Context
	store           *Store
	dl              *DownloadManager
	clipboardWatch  atomic.Bool
	pendingProtocol *ProtocolLink
	httpServer      *http.Server
}

// NewApp creates the application struct and opens the local database.
func NewApp() *App {
	store, err := NewStore()
	if err != nil {
		log.Printf("warning: failed to open database (settings/history disabled): %v", err)
	}
	app := &App{store: store, dl: NewDownloadManager()}
	if link, ok := protocolLinkFromArgs(os.Args); ok {
		app.pendingProtocol = &link
	}
	return app
}

// OnStartup saves the Wails context used by runtime helpers (dialogs, window).
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.dl.SetEmitter(func(eventName string, data any) {
		wruntime.EventsEmit(ctx, eventName, data)
	})
	a.clipboardWatch.Store(a.GetSettings().ClipboardWatch)
	a.startClipboardWatcher()
	a.httpServer = &http.Server{Addr: "127.0.0.1:65530"}
	a.httpServer.Handler = http.HandlerFunc(a.reqHandle)
	// start http server in order to handle request from crx
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()
}

// OnShutdown closes the database when the app quits.
func (a *App) OnShutdown(_ context.Context) {
	if a.store != nil {
		_ = a.store.Close()
	}
	if a.httpServer != nil {
		_ = a.httpServer.Close()
	}
}

// ── settings ────────────────────────────────────────────────────────────

func (a *App) GetSettings() Settings {
	if a.store == nil {
		return defaultSettings()
	}
	cfg, err := a.store.GetSettings()
	if err != nil {
		log.Printf("GetSettings: %v", err)
		return defaultSettings()
	}
	return cfg
}

func (a *App) SaveSettings(cfg Settings) error {
	a.clipboardWatch.Store(cfg.ClipboardWatch)
	if a.store == nil {
		return nil
	}
	return a.store.SaveSettings(cfg)
}

// ── history ─────────────────────────────────────────────────────────────

func (a *App) GetHistory() []HistoryItem {
	if a.store == nil {
		return []HistoryItem{}
	}
	items, err := a.store.GetHistory()
	if err != nil {
		log.Printf("GetHistory: %v", err)
		return []HistoryItem{}
	}
	return items
}

func (a *App) AddHistory(h HistoryItem) int64 {
	if a.store == nil {
		return 0
	}
	id, err := a.store.AddHistory(h)
	if err != nil {
		log.Printf("AddHistory: %v", err)
	}
	return id
}

func (a *App) RemoveHistory(id int64) error {
	if a.store == nil {
		return nil
	}
	_, err := a.store.DeleteHistory(id)
	if err != nil {
		log.Printf("DeleteHistory: %v", err)
	}
	return nil
}

func (a *App) ClearHistory() error {
	if a.store == nil {
		return nil
	}
	return a.store.ClearHistory()
}

// ── crx http request handle ─────────────────────────────────────────────────────────────
type videoInfo struct {
	DiscoveryTime uint64 `json:"discoveryTime"`
	URL           string `json:"url"`
	Title         string `json:"title"`
}

func (a *App) reqHandle(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			log.Println(err)
		}
	}(body)
	okResp := `{"code": 200,"msg": "success"}`
	errorResp := `{"code": 400,"msg": "failed"}`
	resp := okResp
	defer func() {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, err := fmt.Fprint(w, resp)
		if err != nil {
			return
		}
	}()
	videoInfo := &videoInfo{}
	err := json.UnmarshalRead(body, &videoInfo)
	if err != nil {
		log.Println(err)
		resp = errorResp
		return
	}

	link := ProtocolLink{URL: videoInfo.URL, Title: videoInfo.Title}
	wruntime.EventsEmit(a.ctx, "protocol:link", link)
	wruntime.WindowUnminimise(a.ctx)
	wruntime.Show(a.ctx)
}

// ── downloads ────────────────────────────────────────────────────────────

// FetchVariants lists the renditions of a master playlist (empty for media).
func (a *App) FetchVariants(rawURL string, p ProxyConfig) []Variant {
	v, err := a.dl.FetchVariants(rawURL, p)
	if err != nil {
		log.Printf("FetchVariants: %v", err)
		return []Variant{}
	}
	return v
}

// HasFFmpeg reports whether ffmpeg is available for MP4 remuxing.
func (a *App) HasFFmpeg() bool {
	return ffmpegAvailable()
}

// StartDownload launches a background download and returns its task id.
func (a *App) StartDownload(req DownloadRequest) (int64, error) {
	return a.dl.Start(req)
}

func (a *App) PauseDownload(id int64) {
	a.dl.Pause(id)
}

func (a *App) ResumeDownload(id int64) {
	a.dl.Resume(id)
}

func (a *App) CancelDownload(id int64) {
	a.dl.Cancel(id)
}

// ── meta / dialogs / window ─────────────────────────────────────────────

func (a *App) GetVersion() string {
	return defaultVersion
}

// BrowseDir opens the native directory picker and returns the chosen path.
func (a *App) BrowseDir() string {
	if a.ctx == nil {
		return ""
	}
	dir, err := wruntime.OpenDirectoryDialog(a.ctx, wruntime.OpenDialogOptions{
		Title: "选择保存目录",
	})
	if err != nil {
		log.Printf("BrowseDir: %v", err)
		return ""
	}
	return dir
}

// ShowInFolder reveals a file or folder in the OS file explorer (best effort).
func (a *App) ShowInFolder(path string) {
	if path == "" {
		return
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,"+path)
	case "darwin":
		cmd = exec.Command("open", "-R", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("ShowInFolder: %v", err)
	}
}

func (a *App) Minimize() {
	if a.ctx != nil {
		wruntime.WindowMinimise(a.ctx)
	}
}

func (a *App) ToggleMaximize() {
	if a.ctx != nil {
		wruntime.WindowToggleMaximise(a.ctx)
	}
}

func (a *App) Quit() {
	if a.ctx != nil {
		wruntime.Quit(a.ctx)
	}
}

func (a *App) OnSecondInstanceLaunch(data options.SecondInstanceData) {
	wruntime.WindowUnminimise(a.ctx)
	wruntime.Show(a.ctx)
	if link, ok := protocolLinkFromArgs(data.Args); ok {
		wruntime.EventsEmit(a.ctx, "protocol:link", link)
	}
}

// TakePendingProtocolLink returns the m3u8:// launch payload the app started
// with (first launch only), clearing it so it is delivered exactly once. The
// frontend pulls this on init because the first launch fires before it can
// subscribe to the "protocol:link" event.
func (a *App) TakePendingProtocolLink() ProtocolLink {
	var link ProtocolLink
	if a.pendingProtocol != nil {
		link = *a.pendingProtocol
		a.pendingProtocol = nil
	}
	return link
}
