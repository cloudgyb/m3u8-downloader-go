import { ref, computed } from 'vue'
import { zhCN, enUS } from 'naive-ui'
import type { Lang } from './types'

const zh: Record<string, string> = {
  'app.name': 'M3U8 下载器',
  'nav.download': '新建下载', 'nav.tasks': '下载列表', 'nav.history': '历史记录', 'nav.settings': '设置',
  'status.ready': '就绪', 'status.downloading': '正在下载', 'status.proxy': '代理', 'status.noProxy': '无代理', 'status.total': '总速率',
  'dl.title': '一键下载 M3U8',
  'dl.subtitle': '粘贴 M3U8 链接即可开始下载，支持多线程与代理加速。',
  'dl.url': 'M3U8 链接', 'dl.name': '文件名（可选）', 'dl.namePh': '留空则自动识别',
  'dl.threads': '下载线程', 'dl.saveTo': '保存到', 'dl.start': '一键下载', 'dl.summaryTitle': '当前配置',
  'dl.variant': '清晰度', 'dl.variant.unknown': '未知清晰度',
  'dl.urlRequired': '请粘贴 M3U8 链接', 'dl.urlInvalid': '链接格式不正确，请以 http(s):// 开头',
  'tasks.title': '下载列表', 'tasks.active': '进行中', 'tasks.queued': '排队中', 'tasks.totalSpeed': '总速率',
  'tasks.finished': '已完成', 'tasks.pauseAll': '全部暂停', 'tasks.resumeAll': '全部继续',
  'tasks.empty': '暂无下载任务', 'tasks.emptyHint': '去新建下载，粘贴 M3U8 链接开始。', 'tasks.go': '去新建下载',
  'st.downloading': '下载中', 'st.paused': '已暂停', 'st.queued': '排队中',
  'st.completed': '已完成', 'st.failed': '失败', 'st.canceled': '已取消',
  'row.speed': '速率', 'row.eta': '剩余', 'row.pause': '暂停', 'row.resume': '继续', 'row.cancel': '取消',
  'hist.title': '下载历史', 'hist.clear': '清空历史', 'hist.empty': '暂无历史记录',
  'hist.filter.all': '全部', 'hist.filter.done': '已完成', 'hist.filter.fail': '失败',
  'hist.redownload': '重新下载', 'hist.openFolder': '打开文件夹',
  'hist.size': '大小', 'hist.time': '完成时间', 'hist.duration': '用时',
  'set.title': '设置',
  'set.sec.download': '下载设置', 'set.sec.network': '网络代理', 'set.sec.appearance': '外观', 'set.sec.update': '更新',
  'set.threads': '默认线程数', 'set.threadsHint': '1–32，数值越高占用带宽越多',
  'set.concurrent': '并发下载数', 'set.saveDir': '默认保存目录', 'set.browse': '浏览', 'set.notify': '下载完成后通知',
  'set.format': '输出格式', 'set.format.mp4': 'MP4（转封装）', 'set.format.ts': 'TS（原始流）',
  'set.format.ffmpegHint': '未检测到 ffmpeg，将输出 TS 原始流',
  'set.proxy.enable': '启用本地代理', 'set.proxy.type': '代理类型', 'set.proxy.host': '代理地址', 'set.proxy.port': '端口',
  'set.theme': '主题', 'theme.light': '明亮', 'theme.dark': '暗黑', 'theme.system': '跟随系统',
  'set.lang': '语言', 'lang.zh': '中文',
  'set.update.mode': '更新模式', 'upd.auto': '自动更新', 'upd.manual': '手动更新',
  'set.update.current': '当前版本', 'set.update.check': '检查更新',
  'upd.checking': '正在检查更新…', 'upd.found': '发现新版本', 'upd.latest': '已是最新版本',
  'upd.install': '立即更新', 'upd.later': '稍后',
  'toast.added': '已添加下载任务', 'toast.done': '下载完成：', 'toast.saved': '设置已保存',
  'toast.cleared': '历史记录已清空', 'toast.canceled': '任务已取消',
  'toast.demo': '原型演示', 'toast.window': '演示环境不支持窗口操作',
  'toast.browse': '真实应用中将调用系统目录选择器', 'toast.update': '已开始下载更新…',
  'toast.openFolder': '已在资源管理器中打开',
  'win.min': '最小化', 'win.max': '最大化', 'win.close': '关闭',
  'err.timeout': '网络超时', 'err.canceled': '任务已取消',
}

const en: Record<string, string> = {
  'app.name': 'M3U8 Downloader',
  'nav.download': 'New Download', 'nav.tasks': 'Downloads', 'nav.history': 'History', 'nav.settings': 'Settings',
  'status.ready': 'Ready', 'status.downloading': 'Downloading', 'status.proxy': 'Proxy', 'status.noProxy': 'No proxy', 'status.total': 'Total',
  'dl.title': 'Download M3U8 in one click',
  'dl.subtitle': 'Paste an M3U8 link to start downloading, with multi-threading and proxy support.',
  'dl.url': 'M3U8 URL', 'dl.name': 'File name (optional)', 'dl.namePh': 'Leave empty to auto-detect',
  'dl.threads': 'Threads', 'dl.saveTo': 'Save to', 'dl.start': 'Download', 'dl.summaryTitle': 'Current setup',
  'dl.variant': 'Quality', 'dl.variant.unknown': 'Unknown quality',
  'dl.urlRequired': 'Please paste an M3U8 link', 'dl.urlInvalid': 'Invalid URL — it should start with http(s)://',
  'tasks.title': 'Downloads', 'tasks.active': 'Active', 'tasks.queued': 'Queued', 'tasks.totalSpeed': 'Total speed',
  'tasks.finished': 'Finished', 'tasks.pauseAll': 'Pause all', 'tasks.resumeAll': 'Resume all',
  'tasks.empty': 'No downloads yet', 'tasks.emptyHint': 'Create a download and paste an M3U8 link to start.', 'tasks.go': 'New download',
  'st.downloading': 'Downloading', 'st.paused': 'Paused', 'st.queued': 'Queued',
  'st.completed': 'Completed', 'st.failed': 'Failed', 'st.canceled': 'Canceled',
  'row.speed': 'Speed', 'row.eta': 'ETA', 'row.pause': 'Pause', 'row.resume': 'Resume', 'row.cancel': 'Cancel',
  'hist.title': 'History', 'hist.clear': 'Clear', 'hist.empty': 'No history yet',
  'hist.filter.all': 'All', 'hist.filter.done': 'Completed', 'hist.filter.fail': 'Failed',
  'hist.redownload': 'Re-download', 'hist.openFolder': 'Open folder',
  'hist.size': 'Size', 'hist.time': 'Finished', 'hist.duration': 'Duration',
  'set.title': 'Settings',
  'set.sec.download': 'Download', 'set.sec.network': 'Network proxy', 'set.sec.appearance': 'Appearance', 'set.sec.update': 'Update',
  'set.threads': 'Default threads', 'set.threadsHint': '1–32, higher values use more bandwidth',
  'set.concurrent': 'Concurrent downloads', 'set.saveDir': 'Default save directory', 'set.browse': 'Browse', 'set.notify': 'Notify when download completes',
  'set.format': 'Output format', 'set.format.mp4': 'MP4 (remux)', 'set.format.ts': 'TS (raw stream)',
  'set.format.ffmpegHint': 'ffmpeg not found; will output raw TS',
  'set.proxy.enable': 'Enable local proxy', 'set.proxy.type': 'Proxy type', 'set.proxy.host': 'Proxy host', 'set.proxy.port': 'Port',
  'set.theme': 'Theme', 'theme.light': 'Light', 'theme.dark': 'Dark', 'theme.system': 'System',
  'set.lang': 'Language', 'lang.zh': '中文',
  'set.update.mode': 'Update mode', 'upd.auto': 'Automatic', 'upd.manual': 'Manual',
  'set.update.current': 'Current version', 'set.update.check': 'Check for updates',
  'upd.checking': 'Checking for updates…', 'upd.found': 'New version available', 'upd.latest': 'You are up to date',
  'upd.install': 'Update now', 'upd.later': 'Later',
  'toast.added': 'Download task added', 'toast.done': 'Download finished: ', 'toast.saved': 'Settings saved',
  'toast.cleared': 'History cleared', 'toast.canceled': 'Task canceled',
  'toast.demo': 'Demo', 'toast.window': 'Window controls unavailable in demo',
  'toast.browse': 'The real app uses the OS directory picker', 'toast.update': 'Update download started…',
  'toast.openFolder': 'Opened in file explorer',
  'win.min': 'Minimize', 'win.max': 'Maximize', 'win.close': 'Close',
  'err.timeout': 'Connection timed out', 'err.canceled': 'Task canceled',
}

export const releaseNotes: Record<Lang, string[]> = {
  zh: ['修复并发下载时的偶发崩溃', '新增 SOCKS5 代理支持', '优化大文件下载的内存占用', '界面细节与多语言完善'],
  en: ['Fixed occasional crash during concurrent downloads', 'Added SOCKS5 proxy support', 'Reduced memory usage for large files', 'UI polish and i18n improvements'],
}

const dict: Record<Lang, Record<string, string>> = { zh, en }

export const lang = ref<Lang>('zh')

export function t(key: string): string {
  return dict[lang.value]?.[key] ?? dict.zh[key] ?? key
}

export const naiveLocale = computed(() => (lang.value === 'zh' ? zhCN : enUS))
