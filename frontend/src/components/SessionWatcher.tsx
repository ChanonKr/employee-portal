import { useQueryClient } from '@tanstack/react-query'
import { useEffect } from 'react'
import { useLocation, useNavigate } from 'react-router'
import { toast } from 'sonner'
import { setUnauthorizedHandler } from '@/api/client'

/** Any 401 while using the app → clear cached data and go to sign-in. */
export function SessionWatcher() {
  const qc = useQueryClient()
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    setUnauthorizedHandler(() => {
      if (location.pathname === '/login') return
      qc.clear()
      toast.info('Your session expired. Please sign in again.')
      navigate(`/login?next=${encodeURIComponent(location.pathname + location.search)}`, { replace: true })
    })
    return () => setUnauthorizedHandler(null)
  }, [qc, navigate, location])

  return null
}
