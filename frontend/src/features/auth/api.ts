import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api, isApiError } from '@/api/client'
import type { Me, Role } from '@/api/types'

export const meKey = ['me'] as const

/** The signed-in user, or null when there is no valid session. */
export function useMe() {
  return useQuery({
    queryKey: meKey,
    queryFn: async () => {
      try {
        return await api.get<Me>('/auth/me')
      } catch (e) {
        if (isApiError(e, 'UNAUTHORIZED')) return null
        throw e
      }
    },
    staleTime: 5 * 60_000,
    retry: false,
  })
}

export function useLogin() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { employeeCode: string; password: string }) => api.post<void>('/auth/login', body),
    onSuccess: () => qc.invalidateQueries({ queryKey: meKey }),
  })
}

export function useLogout() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<void>('/auth/logout'),
    onSettled: () => qc.clear(),
  })
}

export function useChangePassword() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { currentPassword: string; newPassword: string }) =>
      api.post<void>('/auth/change-password', body),
    onSuccess: () => qc.invalidateQueries({ queryKey: meKey }),
  })
}

/** Only allow same-app paths as ?next= targets (no open redirects). */
/** Where each role lands: admins have no profile page of their own. */
export function homeFor(role: Role | undefined): string {
  return role === 'admin' ? '/admin' : '/profile'
}

export function safeNext(next: string | null, fallback = '/profile'): string {
  if (!next || !next.startsWith('/') || next.startsWith('//') || next.startsWith('/login')) return fallback
  return next
}
