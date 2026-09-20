# README

## About

This is the official Wails Vue-TS template.

You can configure the project by editing `wails.json`. More information about the project settings can be found
here: https://wails.io/docs/reference/project-config

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use `wails build`.

## FFmpeg / MP4 output

By default downloads are remuxed to MP4 (no re-encode) after merging. The app
auto-detects `ffmpeg` in this order:

1. next to the executable — `ffmpeg.exe` (or a `bin/ffmpeg.exe` subfolder),
2. on `PATH`.

If ffmpeg is not found, the raw `.ts` stream is kept instead (the Settings page
shows a hint).

To bundle ffmpeg with the app:

```powershell
# from the project root — use forward slashes (Git Bash / PowerShell both accept them)
powershell -ExecutionPolicy Bypass -File build/fetch-ffmpeg.ps1
```

This downloads a static build into `build/bin/ffmpeg.exe`, which ships next to
the built `.exe` and is included by the Windows installer automatically.

## M3U8 URL for test
https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8
