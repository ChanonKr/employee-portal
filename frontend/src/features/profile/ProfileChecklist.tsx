import { Check, Circle } from 'lucide-react'
import type { Profile } from '@/api/types'
import { Card } from '@/components/ui/card'

export function checklist(p: Profile) {
  const c = p.contact
  return [
    { label: 'Name', done: true },
    { label: 'Photo', done: !!p.photoUrl },
    { label: 'Contact info', done: !!(c.mobilePhone && c.email && c.address) },
    { label: 'Skills', done: p.skills.length > 0 },
  ]
}

/** First-time guidance. Disappears by itself once everything is filled in. */
export function ProfileChecklist({ profile }: { profile: Profile }) {
  const items = checklist(profile)
  const done = items.filter((i) => i.done).length
  if (done === items.length) return null
  const pct = Math.round((done / items.length) * 100)

  return (
    <Card className="border-accent/30 bg-accent-soft/40">
      <div className="flex flex-wrap items-baseline justify-between gap-2">
        <h2 className="font-semibold">Welcome, {profile.firstName}</h2>
        <span className="text-sm text-muted">
          {done} of {items.length} done
        </span>
      </div>
      <p className="mt-1 text-sm text-muted">Complete your profile so colleagues can reach you.</p>
      <div
        className="mt-3 h-2 overflow-hidden rounded-full bg-border"
        role="progressbar"
        aria-label="Profile completeness"
        aria-valuenow={pct}
        aria-valuemin={0}
        aria-valuemax={100}
      >
        <div className="h-full rounded-full bg-accent transition-all" style={{ width: `${pct}%` }} />
      </div>
      <ul className="mt-3 grid grid-cols-2 gap-2 text-sm sm:grid-cols-4">
        {items.map((i) => (
          <li key={i.label} className="flex items-center gap-2">
            {i.done ? <Check className="size-4 text-success" aria-hidden /> : <Circle className="size-4 text-muted" aria-hidden />}
            <span className={i.done ? 'text-muted line-through' : ''}>{i.label}</span>
            <span className="sr-only">{i.done ? '(done)' : '(to do)'}</span>
          </li>
        ))}
      </ul>
    </Card>
  )
}
