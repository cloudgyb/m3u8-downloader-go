# Downloads a static ffmpeg build and places ffmpeg.exe next to the built app
# binary, so MP4 remuxing works without a system-wide ffmpeg install.
#
# The app auto-detects ffmpeg in this order:
#   1. next to the executable (build\bin\ffmpeg.exe)
#   2. build\bin\bin\ffmpeg.exe
#   3. anywhere on PATH
#
# Usage (from the project root — use forward slashes so Git Bash doesn't eat
# the backslash):
#   powershell -ExecutionPolicy Bypass -File build/fetch-ffmpeg.ps1

$ErrorActionPreference = "Stop"
$url = "https://github.com/BtbN/FFmpeg-Builds/releases/latest/download/ffmpeg-master-latest-win64-gpl.zip"

$root   = Split-Path -Parent $PSScriptRoot      # build\ -> project root
$binDir = Join-Path $root "build\bin"
$dest   = Join-Path $binDir "ffmpeg.exe"

if (Test-Path $dest) {
    Write-Host "ffmpeg.exe already present at $dest"
    exit 0
}

New-Item -ItemType Directory -Force -Path $binDir | Out-Null
$tmp = Join-Path $env:TEMP ("ffmpeg-" + [guid]::NewGuid().ToString("N"))
$zip = Join-Path $tmp "ffmpeg.zip"
New-Item -ItemType Directory -Force -Path $tmp | Out-Null

try {
    Write-Host "Downloading ffmpeg..."
    Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing
    Write-Host "Extracting..."
    Expand-Archive -Path $zip -DestinationPath $tmp
    $exe = Get-ChildItem -Path $tmp -Recurse -Filter "ffmpeg.exe" | Select-Object -First 1
    if (-not $exe) { throw "ffmpeg.exe not found in archive" }
    Copy-Item -Path $exe.FullName -Destination $dest
    Write-Host "Installed ffmpeg.exe -> $dest"
}
finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
