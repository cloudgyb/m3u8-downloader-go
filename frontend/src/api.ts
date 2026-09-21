import * as backend from '../wailsjs/go/backend/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import { DEFAULT_SETTINGS } from './types'
import type { Settings, HistoryItem, Variant } from './types'

// When running under `wails dev`/`wails build`, Wails injects `window.go`.
// In a plain browser (vite preview) it is absent, so we fall back to
// localStorage-backed mocks so the UI remains usable for preview.
const w = window as unknown as { go?: { backend?: { App?: unknown } } }
const isWails = typeof w.go !== 'undefined' && !!w.go?.backend?.App

const LS_SETTINGS = 'm3u8dl.settings'
const LS_HISTORY = 'm3u8dl.history'

// ── download event & request types ────────────────────────────────────────

export interface ProxyConfig {
  enabled: boolean
  type: string
  host: string
  port: number
}

export interface StartDownloadRequest {
  url: string
  name: string
  saveDir: string
  threads: number
  variantUri: string
  format: string
  proxy: ProxyConfig
}

export interface DownloadEvent {
  id: number
  status: 'downloading' | 'completed' | 'failed' | 'canceled'
  done: number
  total: number
  speed: number
  file: string
  error: string
}

// ── settings / history ────────────────────────────────────────────────────

function readJSON<T>(key: string, fallback: T): T {
  try {
    const raw = localStorage.getItem(key)
    return raw ? (JSON.parse(raw) as T) : fallback
  } catch {
    return fallback
  }
}

export async function getSettings(): Promise<Settings> {
  if (isWails) return (await backend.GetSettings()) as Settings
  return readJSON<Settings>(LS_SETTINGS, DEFAULT_SETTINGS)
}

export async function saveSettings(s: Settings): Promise<void> {
  if (isWails) return backend.SaveSettings(s as never)
  localStorage.setItem(LS_SETTINGS, JSON.stringify(s))
}

export async function getHistory(): Promise<HistoryItem[]> {
  if (isWails) return (await backend.GetHistory()) as HistoryItem[]
  return readJSON<HistoryItem[]>(LS_HISTORY, [])
}

export async function addHistory(h: HistoryItem): Promise<number> {
  if (isWails) return backend.AddHistory(h as never)
  const list = readJSON<HistoryItem[]>(LS_HISTORY, [])
  const id = Date.now()
  list.unshift({ ...h, id })
  localStorage.setItem(LS_HISTORY, JSON.stringify(list))
  return id
}

export async function removeHistory(id: number): Promise<void> {
  if(isWails) {
    return backend.RemoveHistory(id)
  }
}

export async function clearHistory(): Promise<void> {
  if (isWails) return backend.ClearHistory()
  localStorage.removeItem(LS_HISTORY)
}

// ── downloads ─────────────────────────────────────────────────────────────

export async function startDownload(req: StartDownloadRequest): Promise<number> {
  if (isWails) return backend.StartDownload(req as never)
  return mockStartDownload(req)
}

export async function fetchVariants(url: string, proxy: ProxyConfig): Promise<Variant[]> {
  if (isWails) return backend.FetchVariants(url, proxy as never)
  return []
}

export async function hasFFmpeg(): Promise<boolean> {
  if (isWails) return backend.HasFFmpeg()
  return false
}

export async function pauseDownload(id: number): Promise<void> {
  if (isWails) return backend.PauseDownload(id)
  const m = mockTasks.get(id)
  if (m) m.paused = true
}

export async function resumeDownload(id: number): Promise<void> {
  if (isWails) return backend.ResumeDownload(id)
  const m = mockTasks.get(id)
  if (m) m.paused = false
}

export async function cancelDownload(id: number): Promise<void> {
  if (isWails) return backend.CancelDownload(id)
  const m = mockTasks.get(id)
  if (m) {
    clearInterval(m.timer)
    mockTasks.delete(id)
    emitMock({ id, status: 'canceled', done: m.done, total: m.total, speed: 0, file: '', error: '' })
  }
}

// ── download events ───────────────────────────────────────────────────────

const mockHandlers = new Set<(e: DownloadEvent) => void>()

export function onDownloadEvent(h: (e: DownloadEvent) => void): () => void {
  if (isWails) return EventsOn('download:event', h)
  mockHandlers.add(h)
  return () => {
    mockHandlers.delete(h)
  }
}

// ── browser mock (no Wails backend) ───────────────────────────────────────

interface MockTask {
  id: number
  timer: ReturnType<typeof setInterval>
  paused: boolean
  done: number
  total: number
  speed: number
}

const mockTasks = new Map<number, MockTask>()
let mockSeq = 0

function emitMock(e: DownloadEvent) {
  mockHandlers.forEach((h) => h(e))
}

function mockStartDownload(req: StartDownloadRequest): number {
  const id = ++mockSeq
  const total = (150 + Math.random() * 2400) * 1024 * 1024
  const m: MockTask = {
    id,
    timer: 0 as unknown as ReturnType<typeof setInterval>,
    paused: false,
    done: 0,
    total,
    speed: req.threads * 2.2 * 1024 * 1024,
  }
  m.timer = setInterval(() => {
    if (m.paused) {
      emitMock({ id, status: 'downloading', done: Math.floor(m.done), total: m.total, speed: 0, file: '', error: '' })
      return
    }
    const target = req.threads * (1.6 + Math.random() * 2.4) * 1024 * 1024
    m.speed += (target - m.speed) * 0.35
    m.done += m.speed * 0.5
    if (m.done >= m.total) {
      clearInterval(m.timer)
      mockTasks.delete(id)
      emitMock({ id, status: 'completed', done: m.total, total: m.total, speed: 0, file: '', error: '' })
    } else {
      emitMock({ id, status: 'downloading', done: Math.floor(m.done), total: m.total, speed: Math.floor(m.speed), file: '', error: '' })
    }
  }, 500)
  mockTasks.set(id, m)
  return id
}

// ── meta / dialogs / window ───────────────────────────────────────────────

export async function getVersion(): Promise<string> {
  if (isWails) return backend.GetVersion()
  return 'v1.2.0'
}

export async function browseDir(): Promise<string> {
  if (isWails) return backend.BrowseDir()
  return ''
}

export async function showInFolder(path: string): Promise<void> {
  if (isWails) return backend.ShowInFolder(path)
}

export async function minimize(): Promise<void> {
  if (isWails) return backend.Minimize()
}

export async function toggleMaximize(): Promise<void> {
  if (isWails) return backend.ToggleMaximize()
}

export async function quit(): Promise<void> {
  if (isWails) return backend.Quit()
}

export const wailsAvailable = isWails
