// Inline SVG path data (24×24, stroke = currentColor) matching the design prototype.
export const ICONS: Record<string, string> = {
  download: '<path d="M12 4v11"/><path d="m7 11 5 5 5-5"/><path d="M5 20h14"/>',
  list: '<path d="M9 6h11"/><path d="M9 12h11"/><path d="M9 18h11"/><path d="M4 6h.01"/><path d="M4 12h.01"/><path d="M4 18h.01"/>',
  clock: '<circle cx="12" cy="12" r="8.5"/><path d="M12 7.5V12l3 2"/>',
  settings: '<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09a1.65 1.65 0 0 0 1.51-1 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33h.01a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51h.01a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82v.01a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.9 4.9 1.4 1.4"/><path d="m17.7 17.7 1.4 1.4"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m4.9 19.1 1.4-1.4"/><path d="m17.7 6.3 1.4-1.4"/>',
  moon: '<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>',
  globe: '<circle cx="12" cy="12" r="8.5"/><path d="M3 12h18"/><path d="M12 3a14.5 14.5 0 0 1 0 18"/><path d="M12 3a14.5 14.5 0 0 0 0 18"/>',
  pause: '<path d="M9 6v12"/><path d="M15 6v12"/>',
  play: '<path d="M8 5l11 6.5L8 18z"/>',
  x: '<path d="M6 6l12 12"/><path d="M18 6 6 18"/>',
  check: '<path d="M4.5 12.5l5 5L19.5 6.5"/>',
  alert: '<circle cx="12" cy="12" r="8.5"/><path d="M12 8v4.5"/><path d="M12 16h.01"/>',
  folder: '<path d="M3.5 6.5a2 2 0 0 1 2-2h4l2 2.5h7a2 2 0 0 1 2 2V17a2 2 0 0 1-2 2h-13a2 2 0 0 1-2-2z"/>',
  refresh: '<path d="M20 12a8 8 0 1 1-2.3-5.6"/><path d="M20 4v4h-4"/>',
  trash: '<path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6"/><path d="M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>',
  up: '<circle cx="12" cy="12" r="9"/><path d="M16 12l-4-4-4 4"/><path d="M12 16V8"/>',
  minimize: '<path d="M5 12h14"/>',
  maximize: '<rect x="5" y="5" width="14" height="14" rx="2"/>',
}

export function iconNameForStatus(status: string): string {
  switch (status) {
    case 'downloading': return 'download'
    case 'paused': return 'pause'
    case 'queued': return 'clock'
    case 'completed': return 'check'
    default: return 'alert'
  }
}

export function toneForStatus(status: string): string {
  switch (status) {
    case 'downloading': return 'downloading'
    case 'paused':
    case 'queued': return 'queued'
    case 'completed': return 'success'
    case 'failed':
    case 'canceled': return 'danger'
    default: return 'muted'
  }
}
