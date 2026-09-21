package backend

import (
	"context"
	"log"
	"os/exec"
	"runtime"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the backend service bound to the frontend. Every exported method
// becomes callable from JavaScript through Wails' generated bindings.
type App struct {
	ctx   context.Context
	store *Store
	dl    *DownloadManager
}

// NewApp creates the application struct and opens the local database.
func NewApp() *App {
	store, err := NewStore()
	if err != nil {
		log.Printf("warning: failed to open database (settings/history disabled): %v", err)
	}
	return &App{store: store, dl: NewDownloadManager()}
}

// OnStartup saves the Wails context used by runtime helpers (dialogs, window).
func (a *App) OnStartup(ctx context.Context) {
	a.ctx = ctx
	a.dl.SetEmitter(func(eventName string, data interface{}) {
		wruntime.EventsEmit(ctx, eventName, data)
	})
}

// OnShutdown closes the database when the app quits.
func (a *App) OnShutdown(ctx context.Context) {
	if a.store != nil {
		_ = a.store.Close()
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
