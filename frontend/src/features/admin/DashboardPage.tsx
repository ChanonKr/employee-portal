import { ShieldCheck } from 'lucide-react'
import { Link } from 'react-router'
import { ErrorAlert } from '@/components/ui/alert'
import { Card, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { fieldLabels } from '@/features/history/describe'
import { timeAgo } from '@/lib/format'
import { PageTitle } from './AdminLayout'
import { useStats } from './api'

export function DashboardPage() {
  const stats = useStats()
  if (stats.isPending) return <Skeleton className="h-64" />
  if (stats.isError) return <ErrorAlert>Could not load the dashboard. {stats.error.message}</ErrorAlert>

  const s = stats.data
  const pct = s.employees ? Math.round((s.completeProfiles / s.employees) * 100) : 0
  const maxUse = Math.max(1, ...s.topSkills.map((t) => t.usedBy))

  return (
    <>
      <PageTitle title="Dashboard" />
      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <p className="text-sm text-muted">Employees</p>
          <p className="mt-1 text-3xl font-semibold tabular-nums">{s.employees}</p>
        </Card>
        <Card>
          <p className="text-sm text-muted">Complete profiles</p>
          <p className="mt-1 text-3xl font-semibold tabular-nums">
            {pct}%<span className="ml-2 text-base font-normal text-muted">({s.completeProfiles} of {s.employees})</span>
          </p>
          <p className="mt-1 text-xs text-muted">Photo, all contact details and at least one skill.</p>
        </Card>
        <Card>
          <CardTitle>Top skills</CardTitle>
          {s.topSkills.length === 0 ? (
            <p className="mt-3 text-sm text-muted">No skills selected yet.</p>
          ) : (
            <ul className="mt-3 space-y-2">
              {s.topSkills.map((t) => (
                <li key={t.id} className="grid grid-cols-[7rem_1fr_2rem] items-center gap-3 text-sm">
                  <span className="truncate">{t.name}</span>
                  <span className="h-2 rounded-full bg-accent" style={{ width: `${(t.usedBy / maxUse) * 100}%` }} aria-hidden />
                  <span className="text-right text-muted tabular-nums">{t.usedBy}</span>
                </li>
              ))}
            </ul>
          )}
        </Card>
        <Card>
          <CardTitle>Recent changes</CardTitle>
          {s.recentChanges.length === 0 ? (
            <p className="mt-3 text-sm text-muted">No changes yet.</p>
          ) : (
            <ul className="mt-3 divide-y divide-border">
              {s.recentChanges.map((c) => (
                <li key={c.id} className="flex items-center justify-between gap-3 py-2 text-sm">
                  <span className="min-w-0">
                    <Link to={`/admin/employees/${c.employeeId}`} className="font-medium hover:underline">
                      {c.name}
                    </Link>{' '}
                    <span className="text-muted">· {fieldLabels[c.field]}</span>
                    {c.changedByRole === 'admin' && <ShieldCheck className="ml-1 inline size-3.5 text-muted" aria-label="by an admin" />}
                  </span>
                  <span className="shrink-0 text-xs text-muted">{timeAgo(c.changedAt)}</span>
                </li>
              ))}
            </ul>
          )}
        </Card>
      </div>
    </>
  )
}
