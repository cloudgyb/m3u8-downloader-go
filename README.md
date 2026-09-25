<h1 align="center">M3U8 下载器</h1>

<p align="center">高速、多线程、支持断点续传与 MP4 封装的 M3U8 流媒体下载桌面应用</p>

<p align="center">
  <b>简体中文</b> · <a href="./README_EN.md">English</a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Wails-v2.16-DF0000?logo=go&logoColor=white" alt="Wails v2" />
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Vue-3.5-4FC08D?logo=vuedotjs&logoColor=white" alt="Vue 3" />
  <img src="https://img.shields.io/badge/TypeScript-5.6-3178C6?logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/Naive%20UI-2.45-18a058" alt="Naive UI" />
</p>

## 简介

M3U8 下载器是一款基于 [Wails v2](https://wails.io) 的跨平台桌面应用，Go 后端负责解析与下载，Vue 3 + TypeScript 前端提供原生体验的界面。它能把一个 M3U8 播放列表（主播放列表或分片媒体播放列表）下载并合并为单个视频文件。

## 功能特性

- 📺 **M3U8 解析** —— 支持主播放列表（`#EXT-X-STREAM-INF`）与媒体播放列表，自动处理相对路径。
- 🎚️ **清晰度 / 码率选择** —— 输入 URL 后自动拉取可选清晰度（`BANDWIDTH` / `RESOLUTION`），默认取最高码率，也可手动选择。
- ⚡ **多线程并发下载** —— 每个任务独立线程数（默认 8，最高 64），分片并发拉取，实时显示进度与速度。
- 🔁 **断点续传** —— 已下载分片与 `manifest.json` 保存在 `<保存目录>/.m3u8dl-<hash>/`，重新下载时跳过已完成分片。
- 🔐 **AES-128 解密** —— 支持 `#EXT-X-KEY`（含 `IV`，缺省时按媒体序号派生）。
- 📦 **`#EXT-X-BYTERANGE` 支持** —— 正确请求字节范围分片。
- 🎬 **MP4 封装** —— 合并后的 TS 流可用 ffmpeg 无损重封装为 MP4（`-c copy -movflags +faststart`），ffmpeg 缺失时回退为原始 `.ts`。
- 🌐 **HTTP / SOCKS5 代理** —— 可选代理，适合访问受限源。
- 🗂️ **多任务队列** —— 按并发数排队调度，支持暂停 / 继续 / 取消。
- 🕘 **下载历史** —— 完成、失败、取消的记录本地持久化，可一键打开所在文件夹。
- 🌙 **中英双语 + 明暗主题** —— 界面支持中文 / English 切换，浅色 / 深色 / 跟随系统主题。
- 🔔 **完成通知与更新检查** —— 下载完成系统通知；自动 / 手动检查新版本。

## 技术栈

| 层 | 技术 |
| --- | --- |
| 桌面框架 | [Wails v2](https://wails.io) |
| 后端 | Go 1.25 |
| 前端 | Vue 3 · TypeScript · Vite |
| UI 组件 | [Naive UI](https://www.naiveui.com) |
| 本地存储 | SQLite（纯 Go `modernc.org/sqlite`，无 CGO） |
| 视频封装 | FFmpeg（可选，用于 MP4 重封装） |

## 快速开始

### 环境要求

- Go 1.25+
- Node.js 18+（含 npm）
- Wails CLI（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）
- （可选）FFmpeg —— 用于输出 MP4，缺失时自动回退为 `.ts`

### 构建

```bash
# 生产构建，输出 build/bin/m3u8-downloader.exe
wails build
```

### 开发调试

```bash
# 实时开发模式，前端热更新
wails dev
```

如需在浏览器中独立预览前端（后端用 `localStorage` 模拟）：

```bash
cd frontend && npm run dev
```

## FFmpeg / MP4 输出

默认将合并后的 TS 重封装为 MP4（无转码）。应用按以下顺序自动探测 `ffmpeg`：

1. 可执行文件旁 —— `ffmpeg.exe`（或 `bin/ffmpeg.exe` 子目录）；
2. 系统 `PATH`。

若未找到 ffmpeg，则保留原始 `.ts` 流（设置页会显示提示）。

要将 ffmpeg 与程序一起打包：

```powershell
# 在项目根目录执行（正斜杠，Git Bash / PowerShell 均可用）
powershell -ExecutionPolicy Bypass -File build/fetch-ffmpeg.ps1
```

脚本会把静态构建下载到 `build/bin/ffmpeg.exe`，随产物 `.exe` 一起分发，并由 NSIS 安装器自动携带。

## 使用说明

1. 在「下载」页粘贴 M3U8 链接，失焦后自动拉取可用清晰度；
2. 选择清晰度（单清晰度自动选中）与线程数；
3. 点击「开始下载」，任务进入队列，按并发数调度；
4. 在「任务」页查看进度、暂停 / 继续 / 取消；
5. 完成后可在「历史」页定位文件。

常用设置（可在「设置」页调整）：

| 设置项 | 说明 | 默认值 |
| --- | --- | --- |
| 线程数 | 单任务分片并发数 | 8 |
| 并发任务数 | 同时进行的任务数 | 2 |
| 保存目录 | 下载文件保存位置 | `C:/Downloads/M3U8` |
| 输出格式 | `mp4`（ffmpeg 封装）或 `ts`（原始流） | `mp4` |
| 代理 | HTTP / SOCKS5 代理开关与地址 | 关闭 |
| 主题 | 浅色 / 深色 / 跟随系统 | 跟随系统 |
| 语言 | 中文 / English | 中文 |
| 完成通知 | 下载完成系统通知 | 开启 |
| 更新方式 | 自动 / 手动检查更新 | 自动 |

## 项目结构

```
m3u8-downloader-go/
├── backend/                 # Go 后端
│   ├── app.go               # 绑定到前端的 App 结构体（设置 / 历史 / 下载 / 窗口）
│   ├── downloader.go        # M3U8 下载引擎（解析、AES 解密、并发、续传、MP4 重封装）
│   ├── downloader_test.go   # 解析 / 解密 / 命名 / 续传 / ffmpeg 路径单测
│   ├── store.go             # SQLite 持久化（设置 + 历史）
│   └── settings.go          # 设置与历史数据模型
├── frontend/                # Vue 3 + TypeScript 前端
│   └── src/
│       ├── store.ts         # 响应式应用状态与业务逻辑
│       ├── api.ts           # 后端绑定封装（浏览器预览时用 localStorage 模拟）
│       ├── i18n.ts          # 中英文案字典
│       ├── theme.ts         # Naive UI 主题映射
│       └── components/      # Shell / 标题栏 / 侧边栏 / 四个视图等
├── build/                   # 构建资源（图标、NSIS 安装器、ffmpeg 拉取脚本）
├── main.go                  # Wails 入口（无边框窗口，内嵌 frontend/dist）
└── wails.json               # Wails 项目配置
```

## 测试链接

可用于测试的 M3U8 地址：

```
https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8
```
