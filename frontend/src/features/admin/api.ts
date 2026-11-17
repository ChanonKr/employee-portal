import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { api, errorMessage, isApiError } from '@/api/client'
import type { AdminSettings, AdminSkill, CreatedAccount, DashboardStats, EmployeeDetail, EmployeeList, Version } from '@/api/types'

const keys = {
  stats: ['admin', 'stats'] as const,
  skills: ['admin', 'skills'] as const,
  settings: ['admin', 'settings'] as const,
  employees: ['admin', 'employees'] as const,
  employee: (id: number) => ['admin', 'employee', id] as const,
  employeeHistory: (id: number) => ['admin', 'employee', id, 'history'] as const,
}
export { keys as adminKeys }

const toastError = (e: unknown) => toast.error(errorMessage(e))

export function useStats() {
  return useQuery({ queryKey: keys.stats, queryFn: () => api.get<DashboardStats>('/admin/stats') })
}

// ── skills

export function useAdminSkills() {
  return useQuery({ queryKey: keys.skills, queryFn: () => api.get<AdminSkill[]>('/admin/skills') })
}

function useSkillMutation<T>(fn: (v: T) => Promise<unknown>) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: fn,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: keys.skills })
      qc.invalidateQueries({ queryKey: ['skills'] }) // the employee picker
    },
  })
}

export const useCreateSkill = () => useSkillMutation((name: string) => api.post<{ id: number }>('/admin/skills', { name }))
export const useUpdateSkill = () =>
  useSkillMutation((v: { id: number; name?: string; isActive?: boolean }) =>
    api.patch<void>(`/admin/skills/${v.id}`, { name: v.name, isActive: v.isActive }),
  )
export const useReorderSkills = () =>
  useSkillMutation((order: { id: number; sortOrder: number }[]) => api.put<void>('/admin/skills/order', order))

// ── settings

export function useSettings() {
  return useQuery({ queryKey: keys.settings, queryFn: () => api.get<AdminSettings>('/admin/settings') })
}

export function useSetHistoryDays() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (value: number) => api.put<void>('/admin/settings/history-days', { value }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: keys.settings })
      qc.invalidateQueries({ queryKey: ['history'] })
    },
  })
}

// ── employees

export function useEmployees(params: { q: string; page: number; pageSize?: number }) {
  const qs = new URLSearchParams({ page: String(params.page), pageSize: String(params.pageSize ?? 20) })
  if (params.q.trim()) qs.set('q', params.q.trim())
  return useQuery({
    queryKey: [...keys.employees, params],
    queryFn: () => api.get<EmployeeList>(`/admin/employees?${qs}`),
    placeholderData: keepPreviousData, // keep the table while the next page loads
  })
}

export function useEmployee(id: number) {
  return useQuery({ queryKey: keys.employee(id), queryFn: () => api.get<EmployeeDetail>(`/admin/employees/${id}`) })
}

export function useCreateAccount() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: { employeeCode: string; firstName: string; lastName: string; role: 'employee' | 'admin' }) =>
      api.post<CreatedAccount>('/admin/employees', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: keys.employees })
      qc.invalidateQueries({ queryKey: keys.stats })
    },
  })
}

export function useResetPassword() {
  return useMutation({
    mutationFn: (id: number) => api.post<{ temporaryPassword: string }>(`/admin/employees/${id}/reset-password`),
    onError: toastError,
  })
}

export function useRenameEmployee(id: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (v: { firstName: string; lastName: string; version: number }) =>
      api.put<Version>(`/admin/employees/${id}/name`, { firstName: v.firstName, lastName: v.lastName }, { version: v.version }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: keys.employee(id) })
      qc.invalidateQueries({ queryKey: keys.employeeHistory(id) })
      qc.invalidateQueries({ queryKey: keys.employees })
    },
    onError: (e) => {
      if (isApiError(e, 'VERSION_CONFLICT')) {
        toast.warning('This profile was updated elsewhere. We loaded the latest version — please try again.')
        qc.invalidateQueries({ queryKey: keys.employee(id) })
      } else if (!isApiError(e, 'VALIDATION_ERROR')) {
        toastError(e)
      }
    },
  })
}
