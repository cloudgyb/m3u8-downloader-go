<h1 align="center">M3U8 Downloader</h1>

<p align="center">A fast, multi-threaded M3U8 stream downloader with resume and MP4 remux support</p>

<p align="center">
  <a href="./README.md">简体中文</a> · <b>English</b>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Wails-v2.16-DF0000?logo=go&logoColor=white" alt="Wails v2" />
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Vue-3.5-4FC08D?logo=vuedotjs&logoColor=white" alt="Vue 3" />
  <img src="https://img.shields.io/badge/TypeScript-5.6-3178C6?logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/Naive%20UI-2.45-18a058" alt="Naive UI" />
</p>

## About

M3U8 Downloader is a cross-platform desktop app built on [Wails v2](https://wails.io): a Go backend parses and downloads the stream, while a Vue 3 + TypeScript frontend provides a native-feeling UI. It turns an M3U8 playlist (master or media) into a single merged video file.

## Features

- 📺 **M3U8 parsing** — master playlists (`#EXT-X-STREAM-INF`) and media playlists, with automatic relative-path resolution.
- 🎚️ **Variant selection** — fetches available renditions (`BANDWIDTH` / `RESOLUTION`) on URL input, defaults to highest bandwidth, allows manual choice.
- ⚡ **Multi-threaded download** — per-task thread count (default 8, up to 64), concurrent segment fetching with live progress and speed.
- 🔁 **Resume support** — downloaded segments and a `manifest.json` live under `<saveDir>/.m3u8dl-<hash>/`; re-downloads skip completed segments.
- 🔐 **AES-128 decryption** — `#EXT-X-KEY` (including `IV`, derived from the media sequence when absent).
- 📦 **`#EXT-X-BYTERANGE` support** — requests the correct byte range for each segment.
- 🎬 **MP4 remux** — the merged TS stream is losslessly remuxed to MP4 with ffmpeg (`-c copy -movflags +faststart`), falling back to raw `.ts` when ffmpeg is missing.
- 🌐 **HTTP / SOCKS5 proxy** — optional proxy for restricted sources.
- 🗂️ **Task queue** — schedules by concurrency, with pause / resume / cancel.
- 🕘 **Download history** — completed / failed / canceled records persist locally, with one-click "open in folder".
- 🌙 **Bilingual + themes** — Chinese / English UI, light / dark / system theme.
- 🔔 **Notifications & updates** — completion system notification; automatic or manual update checks.

## Tech Stack

| Layer | Technology |
| --- | --- |
| Desktop framework | [Wails v2](https://wails.io) |
| Backend | Go 1.25 |
| Frontend | Vue 3 · TypeScript · Vite |
| UI components | [Naive UI](https://www.naiveui.com) |
| Local storage | SQLite (pure-Go `modernc.org/sqlite`, no CGO) |
| Video remux | FFmpeg (optional, for MP4 output) |

## Getting Started

### Prerequisites

- Go 1.25+
- Node.js 18+ (with npm)
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)
- (Optional) FFmpeg — for MP4 output; falls back to `.ts` when absent

### Build

```bash
# Production build → build/bin/m3u8-downloader.exe
wails build
```

### Development

```bash
# Live development mode with Vite hot reload
wails dev
```

To preview the frontend standalone in a browser (backend mocked via `localStorage`):

```bash
cd frontend && npm run dev
```

## FFmpeg / MP4 Output

By default the merged TS is remuxed to MP4 (no re-encode). The app auto-detects `ffmpeg` in this order:

1. next to the executable — `ffmpeg.exe` (or a `bin/ffmpeg.exe` subfolder);
2. on `PATH`.

If ffmpeg is not found, the raw `.ts` stream is kept instead (the Settings page shows a hint).

To bundle ffmpeg with the app:

```powershell
# from the project root — forward slashes (Git Bash / PowerShell both accept them)
powershell -ExecutionPolicy Bypass -File build/fetch-ffmpeg.ps1
```

This downloads a static build into `build/bin/ffmpeg.exe`, which ships next to the built `.exe` and is included by the NSIS installer automatically.

## Usage

1. Paste an M3U8 URL on the Download page; available renditions load on blur;
2. Pick a resolution (auto-selected for single-variant playlists) and thread count;
3. Click "Start"; the task is queued and scheduled by the concurrency limit;
4. Watch progress and pause / resume / cancel on the Tasks page;
5. Locate finished files from the History page.

Configurable settings (Settings page):

| Setting | Description | Default |
| --- | --- | --- |
| Threads | Concurrent segments per task | 8 |
| Concurrent tasks | Tasks running at once | 2 |
| Save directory | Output location | `C:/Downloads/M3U8` |
| Format | `mp4` (ffmpeg remux) or `ts` (raw stream) | `mp4` |
| Proxy | HTTP / SOCKS5 proxy toggle and address | Off |
| Theme | Light / dark / system | System |
| Language | 中文 / English | Chinese |
| Notify | Completion system notification | On |
| Update mode | Automatic / manual update check | Auto |

## Project Structure

```
m3u8-downloader-go/
├── backend/                 # Go backend
│   ├── app.go               # App struct bound to the frontend (settings / history / download / window)
│   ├── downloader.go        # M3U8 engine (parse, AES decrypt, concurrency, resume, MP4 remux)
│   ├── downloader_test.go   # unit tests for parser / decrypt / naming / resume / ffmpeg path
│   ├── store.go             # SQLite persistence (settings + history)
│   └── settings.go          # settings & history data models
├── frontend/                # Vue 3 + TypeScript frontend
│   └── src/
│       ├── store.ts         # reactive app state and logic
│       ├── api.ts           # backend binding wrapper (localStorage mock in browser)
│       ├── i18n.ts          # zh/en dictionaries
│       ├── theme.ts         # Naive UI theme mapping
│       └── components/      # Shell / TitleBar / SideBar / four views, etc.
├── build/                   # build assets (icon, NSIS installer, ffmpeg fetch script)
├── main.go                  # Wails entry (frameless window, embedded frontend/dist)
└── wails.json               # Wails project config
```

## Test URL

An M3U8 URL for testing:

```
https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8
```
