export type Lang = 'zh' | 'en'
export type Theme = 'light' | 'dark' | 'system'
export type View = 'download' | 'tasks' | 'history' | 'settings'
export type TaskStatus = 'downloading' | 'paused' | 'queued' | 'completed' | 'failed' | 'canceled'
export type HistFilter = 'all' | 'done' | 'fail'

export interface Settings {
  threads: number
  concurrent: number
  saveDir: string
  notify: boolean
  format: 'ts' | 'mp4'
  proxyEnabled: boolean
  proxyType: 'http' | 'socks5'
  proxyHost: string
  proxyPort: number
  updateMode: 'auto' | 'manual'
  currentVersion: string
  clipboardWatch: boolean
}

export interface Variant {
  uri: string
  bandwidth: number
  resolution: string
  codecs: string
  name: string
}

export interface ProxyConfig {
  enabled: boolean
  type: string
  host: string
  port: number
}

export interface Task {
  id: number
  backendId: number | null
  name: string
  url: string
  threads: number
  variantUri: string
  status: TaskStatus
  total: number
  done: number
  speed: number
  startedAt: number
  saveDir: string
  format: 'ts' | 'mp4'
  proxy: ProxyConfig
}

export interface HistoryItem {
  id: number
  name: string
  url: string
  size: number
  status: TaskStatus
  finishedAt: string
  duration: string
  error: string
  file: string
}

// TaskRecord is the persisted form of an active task, restored on app start.
export interface TaskRecord {
  name: string
  url: string
  saveDir: string
  threads: number
  variantUri: string
  format: 'ts' | 'mp4'
  proxy: ProxyConfig
  status: 'queued' | 'downloading' | 'paused'
  done: number
  total: number
}

export const DEFAULT_SETTINGS: Settings = {
  threads: 8,
  concurrent: 2,
  saveDir: 'C:/Downloads/M3U8',
  notify: true,
  format: 'mp4',
  proxyEnabled: false,
  proxyType: 'http',
  proxyHost: '127.0.0.1',
  proxyPort: 7890,
  updateMode: 'auto',
  currentVersion: 'v1.0.0',
  clipboardWatch: true,
}
