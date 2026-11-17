import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api, errorMessage, isApiError } from '@/api/client'
import type { HistoryField, HistoryPage, Profile, Skill, Version } from '@/api/types'

export const profileKey = ['profile'] as const
export const historyKey = ['history'] as const

export function useProfile() {
  return useQuery({ queryKey: profileKey, queryFn: () => api.get<Profile>('/profile') })
}

export function useSkillCatalog() {
  return useQuery({ queryKey: ['skills'], queryFn: () => api.get<Skill[]>('/skills'), staleTime: 5 * 60_000 })
}

export function useHistory(params: { field?: HistoryField | null; page?: number; pageSize?: number } = {}) {
  const q = new URLSearchParams()
  if (params.field) q.set('field', params.field)
  if (params.page) q.set('page', String(params.page))
  if (params.pageSize) q.set('pageSize', String(params.pageSize))
  const qs = q.toString()
  return useQuery({
    queryKey: [...historyKey, params],
    queryFn: () => api.get<HistoryPage>(`/profile/history${qs ? `?${qs}` : ''}`),
  })
}

/**
 * Shared behaviour for every profile edit:
 * - sends the current version as If-Match
 * - on success, refreshes the profile and its history
 * - on 409, explains and reloads the latest profile
 */
function useProfileMutation<TVars>(send: (vars: TVars, version: number) => Promise<Version>) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: TVars) => {
      const current = qc.getQueryData<Profile>(profileKey)
      if (!current) throw new Error('Profile not loaded')
      return send(vars, current.version)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: profileKey })
      qc.invalidateQueries({ queryKey: historyKey })
    },
    onError: (e) => {
      if (isApiError(e, 'VERSION_CONFLICT')) {
        toast.warning('This profile was updated elsewhere. We loaded the latest version — please try again.')
        qc.invalidateQueries({ queryKey: profileKey })
      } else if (!isApiError(e, 'VALIDATION_ERROR')) {
        toast.error(errorMessage(e))
      }
    },
  })
}

export type ContactKey = 'mobile' | 'email' | 'address'

export function useUpdateContact(field: ContactKey) {
  return useProfileMutation((value: string, version) => api.put<Version>(`/profile/contact/${field}`, { value }, { version }))
}

export function useSetSkills() {
  return useProfileMutation((skillIds: number[], version) => api.put<Version>('/profile/skills', { skillIds }, { version }))
}

export function useUploadPhoto() {
  return useProfileMutation((file: File, version) => {
    const form = new FormData()
    form.append('file', file)
    return api.put<Version>('/profile/photo', form, { version })
  })
}
