import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router'
import type { Role } from '@/api/types'
import { FullPageSpinner } from '@/components/ui/spinner'
import { useMe } from './api'

/** Signed in → render; otherwise go to /login?next=<current path>. */
export function RequireAuth({ children }: { children: ReactNode }) {
  const { data: me, isPending } = useMe()
  const location = useLocation()

  if (isPending) return <FullPageSpinner />
  if (!me) {
    const next = encodeURIComponent(location.pathname + location.search)
    return <Navigate to={`/login?next=${next}`} replace />
  }
  if (me.mustChangePassword && location.pathname !== '/change-password') {
    return <Navigate to="/change-password" replace />
  }
  return <>{children}</>
}

/** UI-level role check. The API enforces it too; this only hides what cannot be used. */
export function RequireRole({ role, children }: { role: Role; children: ReactNode }) {
  const { data: me } = useMe()
  if (me?.role !== role) return <Navigate to="/profile" replace />
  return <>{children}</>
}

export function RedirectIfSignedIn({ children }: { children: ReactNode }) {
  const { data: me, isPending } = useMe()
  if (isPending) return <FullPageSpinner />
  if (me) return <Navigate to="/profile" replace />
  return <>{children}</>
}
