# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

A Wails v2 desktop app — **M3U8 下载器** (M3U8 downloader prototype). Stack: Go backend + Vue 3 / TypeScript / Vite frontend, Naive UI for components, SQLite (pure-Go `modernc.org/sqlite`) for local persistence. Go module: `github.com/cloudgyb/m3u8-downloader`.

## Commands

- `wails dev` — live development with Vite hot reload (primary dev workflow).
- `wails build` — production build; output `build/bin/m3u8-downloader.exe`.
- `wails generate module` — regenerate frontend Go bindings (`frontend/wailsjs/go/backend/App.*`), also run automatically by `dev`/`build` after Go method changes.
- `cd frontend && npm run build` — type-check (`vue-tsc --noEmit`) + Vite build.
- `cd frontend && npm run dev` — run the Vite dev server standalone in a browser (backend is mocked via `localStorage` when `window.go` is absent).
- `go test ./...` / `go build ./...` — test / compile the Go backend.

## Architecture

Wails bundles the Go backend into a native binary and serves the frontend from an embedded `frontend/dist` (`//go:embed all:frontend/dist` in `main.go`).

**Backend (`backend/`)**
- `app.go` — the `App` struct bound to the frontend. Every **exported method** is callable from JS. It also owns the Wails `ctx` for runtime helpers (`OpenDirectoryDialog`, window minimize/maximize/quit, `explorer /select`) and wires the downloader's event emitter to `runtime.EventsEmit`.
- `store.go` + `settings.go` — SQLite persistence (`m3u8dl.db` under the OS config dir). `kv` table holds a single JSON `settings` row; `history` table stores download records. `database/sql` with `modernc.org/sqlite` (no CGO, `SetMaxOpenConns(1)` for serialized access).
- `downloader.go` — the real m3u8 engine: fetches/parses playlists (master + media variants with `BANDWIDTH`/`RESOLUTION`, `#EXTINF`/`#EXT-X-KEY` AES-128/`#EXT-X-BYTERANGE`), downloads segments concurrently (per-task `threads`), decrypts + merges into a single file, and pushes progress via the `download:event` event (`{id,status,done,total,speed,file,error}`). Supports **variant selection** (`FetchVariants`; defaults to highest bandwidth), **resume/断点续传** (partial segments + a `manifest.json` under `<saveDir>/.m3u8dl-<hash>/` are reused on re-download; cancel discards partials, failure keeps them), and **MP4 remux** — when `settings.format == "mp4"`, `writeFinalFile` remuxes the merged TS to MP4 with ffmpeg (`-c copy -movflags +faststart`), falling back to raw `.ts` if ffmpeg is absent or errors. ffmpeg is auto-detected via `ffmpegPath` (next to the exe, a `bin/` subfolder, then PATH) so it can be bundled — `build/fetch-ffmpeg.ps1` downloads it into `build/bin/`, and the NSIS installer ships it (`File /nonfatal` in `project.nsi`). `downloader_test.go` unit-tests parser/decrypt/naming/resume/ffmpeg-path.

**Frontend (`frontend/src/`)**
- `store.ts` — reactive app singleton (`state`) + all logic: settings/history, i18n/theme toggling, navigation, and the download lifecycle (queueing by `concurrent`, starting via the backend, updating tasks from `download:event`). Exposes `ui` (Naive UI message/dialog) injected by `Shell.vue`; non-component code toasts via `toast()`.
- `api.ts` — thin wrapper over the generated `wailsjs/go/backend/App` bindings + `EventsOn`, with a `localStorage`/simulated fallback when `window.go` is absent (browser preview).
- `i18n.ts` — zh/en dictionaries + reactive `t()`; `types.ts` — shared interfaces; `theme.ts` — Naive UI `themeOverrides` (light/dark) mapped to the design tokens in `style.css`.
- `components/` — `App.vue` → providers (`n-config-provider` + message/dialog) → `Shell.vue` → `TitleBar` / `SideBar` / `StatusBar` + four views (`DownloadView`, `TasksView`, `HistoryView`, `SettingsView`), `UpdateDialog`, plus `SvgIcon` (inline SVG path map) and `Segmented`.

**Key patterns**
- The custom design system (colors, layout, spacing, dark mode) lives in `style.css` as CSS variables + structural classes (`.window`, `.card`, `.task-row`, `.badge`, `.segmented`, `.icon-btn`, …). Naive UI components are themed to match via `theme.ts`. Custom CSS handles layout; Naive UI handles inputs/buttons/switch/select/progress/modal/message/dialog.
- The window is **frameless** (`Frameless: true` in `main.go`). Dragging is driven by the Wails `--wails-draggable` CSS property: `.titlebar` sets `drag`, `.titlebar-controls` sets `no-drag`; `TitleBar.vue` also adds double-click-to-maximize. Edge resizing stays enabled because `DisableResize` is left unset.
- Binding path mirrors the Go package: package `backend` → `frontend/wailsjs/go/backend/App`. Regenerate with `wails generate module` after adding/renaming exported Go methods.
- Downloads are **real**: `addTask` queues by `settings.concurrent`, starts via `StartDownload`, and the backend emits `download:event` (progress/completed/failed/canceled) which `store.ts` consumes. The download form fetches variants on URL blur (`FetchVariants`) and lets the user pick a resolution. Completed history records store the output `file` path so "open folder" (`ShowInFolder`) locates the actual file. Only `settings` and `history` persist; active tasks are in-memory.
