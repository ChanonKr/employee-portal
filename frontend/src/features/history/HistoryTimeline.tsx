import { useInfiniteQuery } from '@tanstack/react-query'
import { RefreshCw } from 'lucide-react'
import { useState } from 'react'
import type { HistoryField, HistoryItem, HistoryPage } from '@/api/types'
import { ErrorAlert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { HistoryEmpty } from '@/features/profile/RecentChangesCard'
import { cn } from '@/lib/cn'
import { dayLabel, timeOfDay } from '@/lib/format'
import { historyFilters } from './describe'
import { HistoryEntry } from './HistoryEntry'

const PAGE_SIZE = 20

type Props = {
  /** Cache key prefix; the filter is appended. */
  queryKey: readonly unknown[]
  fetchPage: (field: HistoryField | null, page: number) => Promise<HistoryPage>
  selfUserId?: number
  /** Lets the parent show the window ("last 7 days") from the first page. */
  onFirstPage?: (p: HistoryPage) => void
}

/** Filter chips + changes grouped by day + "Load more". Used by /history and the admin view. */
export function HistoryTimeline({ queryKey, fetchPage, selfUserId, onFirstPage }: Props) {
  const [field, setField] = useState<HistoryField | null>(null)
  const query = useInfiniteQuery({
    queryKey: [...queryKey, { field }],
    queryFn: async ({ pageParam }) => {
      const p = await fetchPage(field, pageParam)
      if (pageParam === 1) onFirstPage?.(p)
      return p
    },
    initialPageParam: 1,
    getNextPageParam: (last) => (last.page.hasNext ? last.page.page + 1 : undefined),
  })

  const items = query.data?.pages.flatMap((p) => p.items) ?? []
  const first = query.data?.pages[0]

  return (
    <div className="space-y-4">
      <div className="-mx-4 overflow-x-auto px-4" role="group" aria-label="Filter by field">
        <div className="flex w-max gap-2">
          {historyFilters.map((f) => (
            <button
              key={f.label}
              type="button"
              aria-pressed={field === f.field}
              onClick={() => setField(f.field)}
              className={cn(
                'min-h-11 rounded-full border px-4 text-sm font-medium sm:min-h-9',
                field === f.field ? 'border-accent bg-accent text-accent-fg' : 'border-border bg-surface text-muted hover:text-fg',
              )}
            >
              {f.label}
            </button>
          ))}
        </div>
      </div>

      {query.isPending ? (
        <Card className="space-y-4">
          <Skeleton className="h-12" />
          <Skeleton className="h-12" />
          <Skeleton className="h-12" />
        </Card>
      ) : query.isError ? (
        <div className="space-y-3">
          <ErrorAlert>Could not load the history. {query.error.message}</ErrorAlert>
          <Button variant="secondary" onClick={() => query.refetch()}>
            <RefreshCw className="size-4" /> Try again
          </Button>
        </div>
      ) : items.length === 0 ? (
        <Card>
          {field && first?.days !== null ? (
            <p className="py-6 text-center text-sm text-muted">No changes to this field in the last {first?.days} days.</p>
          ) : first?.days === null ? (
            <p className="py-6 text-center text-sm text-muted">No changes recorded.</p>
          ) : (
            <HistoryEmpty days={first?.days ?? 0} hasOlder={first?.hasOlder ?? false} />
          )}
        </Card>
      ) : (
        <>
          {groupByDay(items).map(([day, dayItems]) => (
            <section key={day} aria-labelledby={`day-${day}`}>
              <h2 id={`day-${day}`} className="mb-2 text-sm font-semibold text-muted">
                {day}
              </h2>
              <Card className="p-0 sm:p-0">
                <ul className="divide-y divide-border">
                  {dayItems.map((item) => (
                    <li key={item.id} className="p-4">
                      <HistoryEntry item={item} time={timeOfDay(item.changedAt)} selfUserId={selfUserId} />
                    </li>
                  ))}
                </ul>
              </Card>
            </section>
          ))}
          {query.hasNextPage && (
            <div className="flex justify-center">
              <Button variant="secondary" onClick={() => query.fetchNextPage()} loading={query.isFetchingNextPage}>
                Load more
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  )
}

function groupByDay(items: HistoryItem[]): [string, HistoryItem[]][] {
  const groups = new Map<string, HistoryItem[]>()
  for (const item of items) {
    const key = dayLabel(item.changedAt)
    groups.set(key, [...(groups.get(key) ?? []), item])
  }
  return [...groups]
}

export { PAGE_SIZE }
