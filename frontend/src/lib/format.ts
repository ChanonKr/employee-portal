/** 0812345678 → 081 234 5678 (other formats are shown as stored). */
export function formatPhone(v: string | null): string | null {
  if (!v) return v
  const m = /^0(\d{2})(\d{3})(\d{4})$/.exec(v)
  return m ? `0${m[1]} ${m[2]} ${m[3]}` : v
}

export function initials(first: string, last: string): string {
  return ((first[0] ?? '') + (last[0] ?? '')).toUpperCase()
}

const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' })

/** "just now", "5 min ago", "2 hours ago", "yesterday", or a date. */
export function timeAgo(iso: string, now = new Date()): string {
  const then = new Date(iso)
  const sec = Math.round((then.getTime() - now.getTime()) / 1000)
  if (sec > -60) return 'just now'
  if (sec > -3600) return rtf.format(Math.round(sec / 60), 'minute')
  if (sec > -86400) return rtf.format(Math.round(sec / 3600), 'hour')
  if (sec > -2 * 86400) return rtf.format(-1, 'day')
  return then.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
}

/** Group label for the history timeline: Today / Yesterday / 12 Nov 2026. */
export function dayLabel(iso: string, now = new Date()): string {
  const d = new Date(iso)
  const day = (x: Date) => new Date(x.getFullYear(), x.getMonth(), x.getDate()).getTime()
  const diff = Math.round((day(now) - day(d)) / 86400000)
  if (diff === 0) return 'Today'
  if (diff === 1) return 'Yesterday'
  return d.toLocaleDateString('en-GB', { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' })
}

export function timeOfDay(iso: string): string {
  return new Date(iso).toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
}
