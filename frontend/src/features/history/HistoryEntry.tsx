import { ArrowRight, Camera, Mail, MapPin, Phone, ShieldCheck, Sparkles, UserRound, type LucideIcon } from 'lucide-react'
import type { HistoryField, HistoryItem } from '@/api/types'
import { displayValue, fieldLabels, skillDiff } from './describe'

const icons: Record<HistoryField, LucideIcon> = {
  first_name: UserRound,
  last_name: UserRound,
  mobile_phone: Phone,
  contact_email: Mail,
  address: MapPin,
  photo: Camera,
  skills: Sparkles,
}

/** One change: field, old → new (or skill diff), and who made it. */
export function HistoryEntry({ item, time, selfUserId }: { item: HistoryItem; time: string; selfUserId?: number }) {
  const Icon = icons[item.field]
  const byAdmin = item.changedBy.role === 'admin' && item.changedBy.id !== selfUserId

  return (
    <div className="flex gap-3">
      <div className="grid size-9 shrink-0 place-items-center rounded-full bg-bg text-muted">
        <Icon className="size-4" aria-hidden />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex flex-wrap items-baseline justify-between gap-x-3">
          <p className="font-medium">{fieldLabels[item.field]}</p>
          <p className="text-xs text-muted">{time}</p>
        </div>
        <ChangeDetail item={item} />
        {byAdmin && (
          <p className="mt-1 inline-flex items-center gap-1 text-xs text-muted">
            <ShieldCheck className="size-3.5" aria-hidden /> Changed by an admin
          </p>
        )}
      </div>
    </div>
  )
}

function ChangeDetail({ item }: { item: HistoryItem }) {
  if (item.field === 'skills') {
    const { added, removed } = skillDiff(item)
    if (added.length + removed.length === 0) return null
    return (
      <p className="mt-1 flex flex-wrap gap-1.5 text-sm">
        {added.map((s) => (
          <span key={`+${s}`} className="rounded-full bg-success-soft px-2 py-0.5 text-success">
            <span aria-hidden>+</span>
            <span className="sr-only">Added </span>
            {s}
          </span>
        ))}
        {removed.map((s) => (
          <span key={`-${s}`} className="rounded-full bg-danger-soft px-2 py-0.5 text-danger">
            <span aria-hidden>−</span>
            <span className="sr-only">Removed </span>
            {s}
          </span>
        ))}
      </p>
    )
  }
  if (item.field === 'photo') {
    return <p className="mt-0.5 text-sm text-muted">{item.oldValue ? 'Photo changed' : 'Photo added'}</p>
  }
  const before = displayValue(item.field, item.oldValue)
  const after = displayValue(item.field, item.newValue)
  return (
    <p className="mt-0.5 flex flex-wrap items-center gap-1.5 text-sm break-words">
      {before ? (
        <>
          <del className="text-muted">{before}</del>
          <ArrowRight className="size-3.5 text-muted" aria-label="changed to" />
        </>
      ) : (
        <span className="text-muted">Added</span>
      )}
      <span>{after}</span>
    </p>
  )
}
