import {computed, reactive, ref} from 'vue'
import type {DialogApi, MessageApi} from 'naive-ui'
import {lang, t} from './i18n'
import type {DownloadEvent} from './api'
import * as api from './api'
import {deriveName, fmtDuration, nowStr} from './utils'
import type {HistFilter, HistoryItem, Lang, Settings, Task, TaskRecord, TaskStatus, Theme, View} from './types'
import {DEFAULT_SETTINGS} from './types'

// Injected by the app shell once it mounts inside the Naive UI providers.
export const ui: { message: MessageApi | null; dialog: DialogApi | null } = {
    message: null,
    dialog: null,
}

export function toast(content: string, type: 'info' | 'success' | 'warning' | 'error' = 'success') {
    if (ui.message) ui.message[type](content)
}

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

// Reactive mirror of the OS color-scheme preference so `isDark` recomputes
// (and Naive UI re-themes) when the system theme changes.
const systemDark = ref(window.matchMedia('(prefers-color-scheme: dark)').matches)

export function resolvedTheme(): 'light' | 'dark' {
    if (state.theme === 'system') {
        return systemDark.value ? 'dark' : 'light'
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
        saveDir: task.saveDir,
        threads: task.threads,
        variantUri: task.variantUri,
        format: task.format,
        proxy: task.proxy,
    }
}

// persistTasks snapshots the active task list to the backend store so queued,
// paused and in-flight downloads survive an app restart. Progress is best-effort
// (the authoritative resume state lives in each download's manifest.json).
// Persists are chained so async Wails calls commit in snapshot order (otherwise
// an older snapshot could overwrite a newer one and drop a task).
let persistChain: Promise<unknown> = Promise.resolve()

function persistTasks() {
    const records: TaskRecord[] = state.tasks
        .filter((t) => t.status === 'queued' || t.status === 'downloading' || t.status === 'paused')
        .map((t) => ({
            name: t.name,
            url: t.url,
            saveDir: t.saveDir,
            threads: t.threads,
            variantUri: t.variantUri,
            format: t.format,
            proxy: t.proxy,
            status: t.status as TaskRecord['status'],
            done: t.done,
            total: t.total,
        }))
    persistChain = persistChain.then(() => api.saveTasks(records)).catch(() => {})
}

async function startTask(task: Task) {
    task.status = 'downloading'
    task.startedAt = Date.now()
    persistTasks()
    try {
        task.backendId = await api.startDownload(buildRequest(task))
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
        saveDir: state.settings.saveDir,
        format: state.settings.format,
        proxy: currentProxy(),
    }
    state.tasks.push(task)
    persistTasks()
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
    persistTasks()
    promoteQueued()
}

export function resumeTask(id: number) {
    const t = state.tasks.find((x) => x.id === id)
    if (!t) return
    if (t.backendId != null) {
        void api.resumeDownload(t.backendId)
        t.status = 'downloading'
        persistTasks()
        return
    }
    // Restored (or never-started) task: relaunch the download; the backend
    // resumes from the on-disk manifest and partial segments.
    void startTask(t)
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
    persistTasks()
}

export function resumeAll() {
    state.tasks.forEach((t) => {
        if (t.status !== 'paused') return
        if (t.backendId != null) {
            void api.resumeDownload(t.backendId)
            t.status = 'downloading'
        } else {
            void startTask(t)
        }
    })
    persistTasks()
}

function finalize(task: Task, status: TaskStatus, info: { done: number; error: string; file: string }) {
    const idx = state.tasks.indexOf(task)
    if (idx !== -1) state.tasks.splice(idx, 1)
    persistTasks()
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

let lastProgressPersist = 0

function handleDownloadEvent(e: DownloadEvent) {
    const task = state.tasks.find((x) => x.backendId === e.id)
    if (!task) return
    if (e.status === 'downloading') {
        task.done = e.done
        task.total = e.total
        task.speed = e.speed
        // Throttle progress persistence so a mid-download restart keeps a
        // near-current byte count without hammering the store on every tick.
        const now = Date.now()
        if (now - lastProgressPersist > 1000) {
            lastProgressPersist = now
            persistTasks()
        }
        return
    }
    finalize(task, e.status, {done: e.done, error: e.error, file: e.file})
}

// ── link detection (clipboard + m3u8downloader:// protocol) ────────────────────────

export interface PendingLink {
    url: string
    title: string
}

// Links accumulated across clipboard copies and protocol launches, awaiting
// the user's confirmation. `checkedLinks` holds the checked URLs (the checkbox
// group value); each entry's `title` prefills the save-name input.
export const pendingLinks = ref<PendingLink[]>([])
export const clipDialogVisible = ref(false)
export const checkedLinks = ref<string[]>([])

function addPendingLinks(links: PendingLink[]) {
    const added: string[] = []
    for (const link of links) {
        if (!pendingLinks.value.some((l) => l.url === link.url)) {
            pendingLinks.value.push({url: link.url, title: link.title || ''})
            added.push(link.url)
        }
    }
    if (added.length) checkedLinks.value = [...checkedLinks.value, ...added]
    if (pendingLinks.value.length) clipDialogVisible.value = true
}

function handleClipboardLinks(links: string[]) {
    addPendingLinks(links.map((url) => ({url, title: ''})))
}

// A browser launch via the m3u8downloader:// protocol carries an optional title that
// prefills the save name.
function handleProtocolLink(link: {url: string; title: string}) {
    if (link && link.url) addPendingLinks([link])
}

function closeClipboardDialog() {
    clipDialogVisible.value = false
    pendingLinks.value = []
    checkedLinks.value = []
}

export function confirmClipboardDownload() {
    const links = pendingLinks.value.filter((l) => checkedLinks.value.includes(l.url))
    closeClipboardDialog()
    for (const link of links) void addTask(link.url, link.title)
}

export function cancelClipboardDownload() {
    closeClipboardDialog()
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

    try {
        const tasks = await api.getTasks()
        for (const p of tasks) {
            if (!p.url) continue
            // Restore every unfinished task (downloading / paused / queued) as
            // "paused": the app quit mid-run, so downloads were stopped and the
            // user must explicitly resume them (progress is preserved in the
            // on-disk manifest). Nothing auto-starts on launch.
            state.tasks.push({
                id: ++seq,
                backendId: null,
                name: p.name || deriveName(p.url),
                url: p.url,
                threads: p.threads > 0 ? p.threads : state.settings.threads,
                variantUri: p.variantUri || '',
                status: 'paused',
                total: p.total || 0,
                done: p.done || 0,
                speed: 0,
                startedAt: 0,
                saveDir: p.saveDir || state.settings.saveDir,
                format: p.format === 'ts' || p.format === 'mp4' ? p.format : state.settings.format,
                proxy: p.proxy && typeof p.proxy === 'object' ? p.proxy : currentProxy(),
            })
        }
        persistTasks()
    } catch {
        /* keep empty */
    }

    state.lang = localStorage.getItem('m3u8dl.lang') === 'en' ? 'en' : 'zh'
    lang.value = state.lang
    document.documentElement.lang = state.lang === 'zh' ? 'zh-CN' : 'en'

    const savedTheme = localStorage.getItem('m3u8dl.theme')
    if (savedTheme === 'light' || savedTheme === 'dark' || savedTheme === 'system') {
        state.theme = savedTheme
    }
    syncThemeAttr()

    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        systemDark.value = e.matches
        if (state.theme === 'system') syncThemeAttr()
    })

    api.onDownloadEvent(handleDownloadEvent)
    api.onClipboardLinks(handleClipboardLinks)
    api.onProtocolLink(handleProtocolLink)

    // A first launch via m3u8downloader:// arrives before the frontend can subscribe to
    // the "protocol:link" event, so the backend stashes it and we pull it here.
    void api.takePendingProtocolLink().then((link) => {
        if (link && link.url) handleProtocolLink(link)
    })
}
