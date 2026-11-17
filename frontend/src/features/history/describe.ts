import type { HistoryField, HistoryItem } from '@/api/types'
import { formatPhone } from '@/lib/format'

export const fieldLabels: Record<HistoryField, string> = {
  first_name: 'First name',
  last_name: 'Last name',
  mobile_phone: 'Mobile',
  contact_email: 'Email',
  address: 'Address',
  photo: 'Photo',
  skills: 'Skills',
}

/** Filter chips on the history page (the API takes the stored field name). */
export const historyFilters: { label: string; field: HistoryField | null }[] = [
  { label: 'All', field: null },
  { label: 'Photo', field: 'photo' },
  { label: 'Mobile', field: 'mobile_phone' },
  { label: 'Email', field: 'contact_email' },
  { label: 'Address', field: 'address' },
  { label: 'Skills', field: 'skills' },
]

export type SkillDiff = { added: string[]; removed: string[] }

export function skillDiff(item: HistoryItem): SkillDiff {
  const before = Array.isArray(item.oldValue) ? item.oldValue : []
  const after = Array.isArray(item.newValue) ? item.newValue : []
  return {
    added: after.filter((s) => !before.includes(s)),
    removed: before.filter((s) => !after.includes(s)),
  }
}

/** Text value for display; phones are formatted, photos are not shown as file names. */
export function displayValue(field: HistoryField, v: HistoryItem['oldValue']): string | null {
  if (v === null || Array.isArray(v)) return null
  if (field === 'photo') return null
  if (field === 'mobile_phone') return formatPhone(v)
  return v
}
