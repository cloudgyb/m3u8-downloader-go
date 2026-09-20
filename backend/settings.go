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
	}
}
