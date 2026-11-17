import { ArrowRight, History } from 'lucide-react'
import { Link } from 'react-router'
import { Card, CardTitle } from '@/components/ui/card'
import { EmptyState } from '@/components/ui/empty-state'
import { Skeleton } from '@/components/ui/skeleton'
import { HistoryEntry } from '@/features/history/HistoryEntry'
import { useMe } from '@/features/auth/api'
import { timeAgo } from '@/lib/format'
import { useHistory } from './api'

export function HistoryEmpty({ days, hasOlder }: { days: number; hasOlder: boolean }) {
  return hasOlder ? (
    <EmptyState icon={History} title={`No changes in the last ${days} days`} />
  ) : (
    <EmptyState icon={History} title="No changes yet">
      Edits to your photo, contact info, and skills from the last {days} days will appear here.
    </EmptyState>
  )
}

export function RecentChangesCard() {
  const history = useHistory({ pageSize: 3 })
  const { data: me } = useMe()
  const days = history.data?.days

  return (
    <Card>
      <div className="flex items-center justify-between gap-2">
        <CardTitle>
          Recent changes{days ? <span className="font-normal text-muted"> · last {days} days</span> : null}
        </CardTitle>
        <Link to="/history" className="inline-flex min-h-11 items-center gap-1 text-sm font-medium text-accent hover:underline sm:min-h-0">
          View all <ArrowRight className="size-4" aria-hidden />
        </Link>
      </div>
      {history.isPending ? (
        <div className="mt-4 space-y-4">
          <Skeleton className="h-10" />
          <Skeleton className="h-10" />
        </div>
      ) : history.isError ? (
        <p className="mt-4 text-sm text-danger">Could not load recent changes.</p>
      ) : history.data.items.length === 0 ? (
        <HistoryEmpty days={history.data.days ?? 0} hasOlder={history.data.hasOlder} />
      ) : (
        <ul className="mt-4 space-y-4">
          {history.data.items.map((item) => (
            <li key={item.id}>
              <HistoryEntry item={item} time={timeAgo(item.changedAt)} selfUserId={me?.userId} />
            </li>
          ))}
        </ul>
      )}
    </Card>
  )
}
