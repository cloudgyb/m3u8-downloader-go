import {computed, reactive, ref} from 'vue'
import type {DialogApi, MessageApi} from 'naive-ui'
import {lang, t} from './i18n'
import type {DownloadEvent} from './api'
import * as api from './api'
import {deriveName, fmtDuration, nowStr} from './utils'
import type {HistFilter, HistoryItem, Lang, Settings, Task, TaskStatus, Theme, View} from './types'
import {DEFAULT_SETTINGS} from './types'

// Injected by the app shell once it mounts inside the Naive UI providers.
export const ui: { message: MessageApi | null; dialog: DialogApi | null } = {
    message: null,
    dialog: null,
}

export function toast(content: string, type: 'info' | 'success' | 'warning' | 'error' = 'success') {
    if (ui.message) ui.message[type](content)
}

const SEED_KEY = 'm3u8dl.seeded'
let seq = 0
let histSeq = 0

export const state = reactive({
    lang: 'zh' as Lang,
    theme: 'system' as Theme,
    settings: {...DEFAULT_SETTINGS} as Settings,
    tasks: [] as Task[],
    history: [] as HistoryItem[],
    histFilter: 'all' as HistFilter,
    currentView: 'download' as View,
})

// ── theme ────────────────────────────────────────────────────────────────

export function resolvedTheme(): 'light' | 'dark' {
    if (state.theme === 'system') {
        return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    }
    return state.theme
}

export const isDark = computed(() => resolvedTheme() === 'dark')

function syncThemeAttr() {
    document.documentElement.setAttribute('data-theme', resolvedTheme())
}

export function setTheme(theme: Theme) {
    state.theme = theme
    localStorage.setItem('m3u8dl.theme', theme)
    syncThemeAttr()
    saveSettings()
}

export function toggleTheme() {
    setTheme(resolvedTheme() === 'dark' ? 'light' : 'dark')
}

// ── language ─────────────────────────────────────────────────────────────

export function setLang(l: Lang) {
    state.lang = l
    lang.value = l
    localStorage.setItem('m3u8dl.lang', l)
    document.documentElement.lang = l === 'zh' ? 'zh-CN' : 'en'
    document.title = t('app.name')
}

export function toggleLang() {
    setLang(state.lang === 'zh' ? 'en' : 'zh')
}

// ── settings ─────────────────────────────────────────────────────────────

export function saveSettings() {
    void api.saveSettings({...state.settings})
}

export function updateSettings(patch: Partial<Settings>, notify = false) {
    Object.assign(state.settings, patch)
    saveSettings()
    if (notify) toast(t('toast.saved'))
}

// ── navigation ───────────────────────────────────────────────────────────

export function navigate(view: View) {
    state.currentView = view
}

// ── update dialog ────────────────────────────────────────────────────────

export const updateDialogVisible = ref(false)

export function showUpdateDialog() {
    updateDialogVisible.value = true
}

// ── download lifecycle ───────────────────────────────────────────────────

function currentProxy(): api.ProxyConfig {
    return {
        enabled: state.settings.proxyEnabled,
        type: state.settings.proxyType,
        host: state.settings.proxyHost,
        port: state.settings.proxyPort,
    }
}

function buildRequest(task: Task): api.StartDownloadRequest {
    return {
        url: task.url,
        name: task.name,
        saveDir: state.settings.saveDir,
        threads: task.threads,
        variantUri: task.variantUri,
        format: state.settings.format,
        proxy: currentProxy(),
    }
}

async function startTask(task: Task) {
    task.status = 'downloading'
    task.startedAt = Date.now()
    try {
        const backendId = await api.startDownload(buildRequest(task))
        task.backendId = backendId
    } catch (e) {
        finalize(task, 'failed', {done: 0, error: e instanceof Error ? e.message : String(e), file: ''})
    }
}

function schedule(task: Task) {
    if (task.status !== 'queued') return
    if (activeCount.value >= state.settings.concurrent) return
    void startTask(task)
}

export function promoteQueued() {
    for (const task of state.tasks) {
        if (task.status !== 'queued') continue
        if (activeCount.value >= state.settings.concurrent) break
        void startTask(task)
    }
}

export async function addTask(url: string, name: string, variantUri = ''): Promise<string | null> {
    url = (url || '').trim()
    if (!url) return 'dl.urlRequired'
    if (!/^https?:\/\//i.test(url)) return 'dl.urlInvalid'

    const task: Task = {
        id: ++seq,
        backendId: null,
        name: (name || '').trim() || deriveName(url),
        url,
        threads: state.settings.threads,
        variantUri,
        status: 'queued',
        total: 0,
        done: 0,
        speed: 0,
        startedAt: 0,
    }
    state.tasks.push(task)
    schedule(task)
    toast(t('toast.added'))
    navigate('tasks')
    return null
}

export function pauseTask(id: number) {
    const t = state.tasks.find((x) => x.id === id)
    if (!t) return
    if (t.backendId != null) void api.pauseDownload(t.backendId)
    t.status = 'paused'
    t.speed = 0
}

export function resumeTask(id: number) {
    const t = state.tasks.find((x) => x.id === id)
    if (!t) return
    if (t.backendId != null) void api.resumeDownload(t.backendId)
    t.status = 'downloading'
}

export function cancelTask(id: number) {
    const t = state.tasks.find((x) => x.id === id)
    if (!t) return
    if (t.backendId != null) {
        // Backend emits "canceled"; finalize happens in the event handler.
        void api.cancelDownload(t.backendId)
    } else {
        finalize(t, 'canceled', {done: 0, error: '', file: ''})
    }
}

export function pauseAll() {
    state.tasks.forEach((t) => {
        if (t.status === 'downloading') {
            if (t.backendId != null) void api.pauseDownload(t.backendId)
            t.status = 'paused'
            t.speed = 0
        }
    })
}

export function resumeAll() {
    state.tasks.forEach((t) => {
        if (t.status === 'paused') {
            if (t.backendId != null) void api.resumeDownload(t.backendId)
            t.status = 'downloading'
        }
    })
}

function finalize(task: Task, status: TaskStatus, info: { done: number; error: string; file: string }) {
    const idx = state.tasks.indexOf(task)
    if (idx !== -1) state.tasks.splice(idx, 1)
    pushHistory({
        name: task.name,
        url: task.url,
        size: info.done,
        status,
        duration: task.startedAt ? fmtDuration(Date.now() - task.startedAt) : '--',
        error: status === 'failed' ? info.error : status === 'canceled' ? 'err.canceled' : '',
        file: status === 'completed' ? info.file : '',
    })
    if (status === 'completed' && state.settings.notify) toast(t('toast.done') + task.name)
    if (status === 'canceled') toast(t('toast.canceled'), 'info')
    promoteQueued()
}

function handleDownloadEvent(e: DownloadEvent) {
    const task = state.tasks.find((x) => x.backendId === e.id)
    if (!task) return
    if (e.status === 'downloading') {
        task.done = e.done
        task.total = e.total
        task.speed = e.speed
        return
    }
    finalize(task, e.status, {done: e.done, error: e.error, file: e.file})
}

// ── history ──────────────────────────────────────────────────────────────

function pushHistory(h: Omit<HistoryItem, 'id' | 'finishedAt'>) {
    const item: HistoryItem = {...h, id: Date.now() + ++histSeq, finishedAt: nowStr()}
    state.history.unshift(item)
    void api.addHistory(item)
}

export const filteredHistory = computed<HistoryItem[]>(() => {
    if (state.histFilter === 'done') return state.history.filter((h) => h.status === 'completed')
    if (state.histFilter === 'fail') return state.history.filter((h) => h.status !== 'completed')
    return state.history
})

export function setHistFilter(f: HistFilter) {
    state.histFilter = f
}

export async function removeHistoryItem(id: number) {
    const doDelete = async () => {
        void api.removeHistory(id)
        toast(t('toast.deleted'), 'info')
        if (state.history) {
            const index = state.history.findIndex((item) => {
                return item.id === id
            })
            if (index === -1) {
                return
            }
            state.history.splice(index, 1)
        }
    }
    if (ui.dialog) {
        ui.dialog.warning({
            title: t('hist.delete'),
            content: t('toast.confirm.delete'),
            positiveText: t('hist.delete'),
            negativeText: t('upd.later'),
            onPositiveClick: doDelete,
        })
    } else if (window.confirm(t('toast.confirm.delete'))) {
        await doDelete()
    }
}

export async function refreshHistory() {
    state.history = await api.getHistory()
    toast(t('toast.refreshed'), 'info')
}

export function clearHistory() {
    if (!state.history.length) return
    const doClear = () => {
        state.history = []
        void api.clearHistory()
        toast(t('toast.cleared'), 'info')
    }
    if (ui.dialog) {
        ui.dialog.warning({
            title: t('hist.clear'),
            content: t('toast.cleared') + '？',
            positiveText: t('hist.clear'),
            negativeText: t('upd.later'),
            onPositiveClick: doClear,
        })
    } else if (window.confirm(t('toast.cleared') + '？')) {
        doClear()
    }
}

export function redoTask(h: HistoryItem) {
    void addTask(h.url, h.name)
}

export function openFolder(h: HistoryItem) {
    if (api.wailsAvailable) {
        void api.showInFolder(h.file || state.settings.saveDir)
    } else {
        toast(t('toast.openFolder'), 'info')
    }
}

// ── stats (status bar / tasks header) ────────────────────────────────────

export const activeCount = computed(() => state.tasks.filter((x) => x.status === 'downloading').length)
export const queuedCount = computed(() => state.tasks.filter((x) => x.status === 'queued').length)
export const doneCount = computed(() => state.history.filter((h) => h.status === 'completed').length)
export const totalSpeed = computed(() =>
    state.tasks.reduce((sum, x) => sum + (x.status === 'downloading' ? x.speed : 0), 0),
)

export const statusText = computed(() => {
    if (activeCount.value > 0) {
        return t('status.downloading') + ' · ' + activeCount.value + (queuedCount.value > 0 ? ' (+' + queuedCount.value + ')' : '')
    }
    return t('status.ready')
})

// ── init ─────────────────────────────────────────────────────────────────

const DEMO_HISTORY: Array<Omit<HistoryItem, 'id'>> = [
    {
        name: '蓝色星球 · 第二季 E01.mkv',
        url: 'https://demo.cdn.example/blue-planet-s2e1.m3u8',
        size: 1331 * 1024 * 1024,
        status: 'completed',
        finishedAt: '2026-09-19 14:30',
        duration: '12:34',
        error: '',
        file: ''
    },
    {
        name: '数据结构 · 第 3 讲.mp4',
        url: 'https://demo.cdn.example/ds03.m3u8',
        size: 0,
        status: 'failed',
        finishedAt: '2026-09-18 09:15',
        duration: '00:42',
        error: 'err.timeout',
        file: ''
    },
    {
        name: '国风音乐会 · 2023 跨年.mkv',
        url: 'https://demo.cdn.example/concert.m3u8',
        size: 843 * 1024 * 1024,
        status: 'completed',
        finishedAt: '2026-09-15 21:02',
        duration: '08:11',
        error: '',
        file: ''
    },
]

export async function init() {
    try {
        const s = await api.getSettings()
        if (s && typeof s.threads === 'number') state.settings = {...DEFAULT_SETTINGS, ...s}
    } catch {
        /* keep defaults */
    }

    try {
        state.history = await api.getHistory()
    } catch {
        state.history = []
    }

    // Seed demo history exactly once, so the History view matches the prototype on first run.
    if (!state.history.length && !localStorage.getItem(SEED_KEY)) {
        localStorage.setItem(SEED_KEY, '1')
        for (const h of DEMO_HISTORY) {
            void api.addHistory({...h, id: 0}).then((id) => {
                state.history.push({...h, id})
            })
        }
    }

    state.lang = localStorage.getItem('m3u8dl.lang') === 'en' ? 'en' : 'zh'
    lang.value = state.lang
    document.documentElement.lang = state.lang === 'zh' ? 'zh-CN' : 'en'

    const savedTheme = localStorage.getItem('m3u8dl.theme')
    if (savedTheme === 'light' || savedTheme === 'dark' || savedTheme === 'system') {
        state.theme = savedTheme
    }
    syncThemeAttr()

    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        if (state.theme === 'system') syncThemeAttr()
    })

    api.onDownloadEvent(handleDownloadEvent)
}
