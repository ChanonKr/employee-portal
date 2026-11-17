import { ArrowLeft } from 'lucide-react'
import { useState } from 'react'
import { Link } from 'react-router'
import { api } from '@/api/client'
import type { HistoryPage as Page } from '@/api/types'
import { useMe } from '@/features/auth/api'
import { historyKey } from '@/features/profile/api'
import { HistoryTimeline, PAGE_SIZE } from './HistoryTimeline'

export function HistoryPage() {
  const { data: me } = useMe()
  const [days, setDays] = useState<number | null>(null)

  return (
    <div className="mx-auto max-w-3xl space-y-4">
      <Link to="/profile" className="inline-flex min-h-11 items-center gap-1 text-sm text-muted hover:text-fg sm:min-h-0">
        <ArrowLeft className="size-4" aria-hidden /> Back to profile
      </Link>
      <div>
        <h1 className="text-2xl font-semibold">Change history</h1>
        {/* The window comes from the API (admins can change it); never hard-code "7". */}
        {days !== null && <p className="mt-1 text-sm text-muted">Showing changes from the last {days} days.</p>}
      </div>
      <HistoryTimeline
        queryKey={[...historyKey, 'timeline']}
        selfUserId={me?.userId}
        onFirstPage={(p) => setDays(p.days)}
        fetchPage={(field, page) => {
          const q = new URLSearchParams({ page: String(page), pageSize: String(PAGE_SIZE) })
          if (field) q.set('field', field)
          return api.get<Page>(`/profile/history?${q}`)
        }}
      />
    </div>
  )
}
