package backend

// Settings holds user preferences, persisted to the local SQLite database.
type Settings struct {
	Threads        int    `json:"threads"`
	Concurrent     int    `json:"concurrent"`
	SaveDir        string `json:"saveDir"`
	Notify         bool   `json:"notify"`
	Format         string `json:"format"` // "mp4" (remux via ffmpeg) | "ts" (raw stream)
	ProxyEnabled   bool   `json:"proxyEnabled"`
	ProxyType      string `json:"proxyType"`
	ProxyHost      string `json:"proxyHost"`
	ProxyPort      int    `json:"proxyPort"`
	UpdateMode     string `json:"updateMode"`
	CurrentVersion string `json:"currentVersion"`
	ClipboardWatch bool   `json:"clipboardWatch"`
}

// HistoryItem is a completed, failed or canceled download record.
type HistoryItem struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Size       int64  `json:"size"`
	Status     string `json:"status"`
	FinishedAt string `json:"finishedAt"`
	Duration   string `json:"duration"`
	Error      string `json:"error"`
	File       string `json:"file"`
}

// TaskRecord is a download task in the queue (queued, downloading or paused)
// persisted so it survives an app restart and can be resumed. `done`/`total`
// are best-effort (the authoritative resume progress lives in the download's
// manifest.json on disk).
type TaskRecord struct {
	Name       string      `json:"name"`
	URL        string      `json:"url"`
	SaveDir    string      `json:"saveDir"`
	Threads    int         `json:"threads"`
	VariantURI string      `json:"variantUri"`
	Format     string      `json:"format"` // "mp4" | "ts"
	Proxy      ProxyConfig `json:"proxy"`
	Status     string      `json:"status"` // queued | downloading | paused
	Done       int64       `json:"done"`
	Total      int64       `json:"total"`
}

const defaultVersion = "v1.2.0"

func defaultSettings() Settings {
	return Settings{
		Threads:        8,
		Concurrent:     2,
		SaveDir:        "C:/Downloads/M3U8",
		Notify:         true,
		Format:         "mp4",
		ProxyEnabled:   false,
		ProxyType:      "http",
		ProxyHost:      "127.0.0.1",
		ProxyPort:      7890,
		UpdateMode:     "auto",
		CurrentVersion: defaultVersion,
		ClipboardWatch: true,
	}
}
