function pad(n: number): string {
  return n < 10 ? '0' + n : '' + n
}

export function fmtBytes(b: number): string {
  if (b < 1024) return b + ' B'
  const kb = b / 1024
  if (kb < 1024) return kb.toFixed(1) + ' KB'
  const mb = kb / 1024
  if (mb < 1024) return mb.toFixed(1) + ' MB'
  return (mb / 1024).toFixed(2) + ' GB'
}

export function fmtSpeed(bps: number): string {
  if (bps < 1024) return bps.toFixed(0) + ' B/s'
  const kb = bps / 1024
  if (kb < 1024) return kb.toFixed(1) + ' KB/s'
  const mb = kb / 1024
  if (mb < 1024) return mb.toFixed(1) + ' MB/s'
  return (mb / 1024).toFixed(2) + ' GB/s'
}

export function fmtEta(sec: number): string {
  if (!isFinite(sec) || sec < 0) return '--'
  sec = Math.round(sec)
  const h = Math.floor(sec / 3600)
  const m = Math.floor((sec % 3600) / 60)
  const s = sec % 60
  if (h > 0) return h + ':' + pad(m) + ':' + pad(s)
  return m + ':' + pad(s)
}

export function fmtDuration(ms: number): string {
  const s = Math.round(ms / 1000)
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const sec = s % 60
  if (h > 0) return h + ':' + pad(m) + ':' + pad(sec)
  return m + ':' + pad(sec)
}

export function nowStr(): string {
  const d = new Date()
  return (
    d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
    ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes())
  )
}

export function deriveName(url: string): string {
  try {
    const u = new URL(url)
    const seg = u.pathname.split('/').filter(Boolean).pop() || ''
    const name = seg.replace(/\.m3u8$/i, '')
    return name ? decodeURIComponent(name) : 'video'
  } catch {
    return 'video'
  }
}

export function clampInt(v: number | string, lo: number, hi: number): number {
  const n = typeof v === 'string' ? parseInt(v, 10) : v
  if (isNaN(n)) return lo
  return Math.max(lo, Math.min(hi, n))
}
