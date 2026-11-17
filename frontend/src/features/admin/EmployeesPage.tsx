import { zodResolver } from '@hookform/resolvers/zod'
import { ChevronLeft, ChevronRight, Plus, Search, UserX } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link } from 'react-router'
import { z } from 'zod'
import { isApiError } from '@/api/client'
import type { CreatedAccount } from '@/api/types'
import { ErrorAlert } from '@/components/ui/alert'
import { Avatar } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { ResponsiveDialog } from '@/components/ui/dialog'
import { EmptyState } from '@/components/ui/empty-state'
import { Field } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { useDebounced } from '@/lib/useDebounced'
import { PageTitle } from './AdminLayout'
import { useCreateAccount, useEmployees } from './api'
import { TempPassword } from './TempPassword'

export function EmployeesPage() {
  const [q, setQ] = useState('')
  const [page, setPage] = useState(1)
  const search = useDebounced(q)
  const list = useEmployees({ q: search, page })
  const [creating, setCreating] = useState(false)

  return (
    <>
      <PageTitle title="Employees">
        <Button onClick={() => setCreating(true)}>
          <Plus className="size-4" aria-hidden /> New account
        </Button>
      </PageTitle>

      <div className="relative mb-4">
        <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted" aria-hidden />
        <Input
          value={q}
          onChange={(e) => {
            setQ(e.target.value)
            setPage(1)
          }}
          placeholder="Search by name or employee code"
          aria-label="Search employees"
          className="pl-9"
          type="search"
        />
      </div>

      {list.isPending ? (
        <Skeleton className="h-64" />
      ) : list.isError ? (
        <ErrorAlert>Could not load employees. {list.error.message}</ErrorAlert>
      ) : list.data.items.length === 0 ? (
        <Card>
          <EmptyState icon={UserX} title={search ? `No employees match “${search}”` : 'No employees yet'} />
        </Card>
      ) : (
        <Card className="p-0 sm:p-0">
          <ul className="divide-y divide-border" aria-busy={list.isFetching}>
            {list.data.items.map((e) => (
              <li key={e.id}>
                <Link to={`/admin/employees/${e.id}`} className="flex min-h-16 items-center gap-3 px-4 py-3 hover:bg-bg">
                  <Avatar url={e.photoUrl} firstName={e.firstName} lastName={e.lastName} className="size-10 text-sm" />
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-medium">
                      {e.firstName} {e.lastName}
                    </p>
                    <p className="text-sm text-muted">{e.employeeCode}</p>
                  </div>
                  <div className="flex shrink-0 gap-1.5">
                    {e.role === 'admin' && <Badge tone="accent">Admin</Badge>}
                    {!e.isActive && <Badge tone="danger">Inactive</Badge>}
                  </div>
                  <ChevronRight className="size-4 shrink-0 text-muted" aria-hidden />
                </Link>
              </li>
            ))}
          </ul>
          <div className="flex items-center justify-between gap-2 border-t border-border px-4 py-3 text-sm">
            <span className="text-muted">
              {list.data.total} employee{list.data.total === 1 ? '' : 's'}
            </span>
            <div className="flex items-center gap-1">
              <Button variant="ghost" size="icon" aria-label="Previous page" disabled={page === 1} onClick={() => setPage((p) => p - 1)}>
                <ChevronLeft className="size-4" />
              </Button>
              <span className="px-2 tabular-nums">Page {page}</span>
              <Button variant="ghost" size="icon" aria-label="Next page" disabled={!list.data.page.hasNext} onClick={() => setPage((p) => p + 1)}>
                <ChevronRight className="size-4" />
              </Button>
            </div>
          </div>
        </Card>
      )}

      <CreateAccountDialog open={creating} onOpenChange={setCreating} />
    </>
  )
}

const accountSchema = z.object({
  employeeCode: z
    .string()
    .trim()
    .toUpperCase()
    .regex(/^[A-Z0-9-]{3,20}$/, '3–20 letters, digits or dashes'),
  firstName: z.string().trim().min(1, 'Enter a first name').max(100),
  lastName: z.string().trim().min(1, 'Enter a last name').max(100),
  role: z.enum(['employee', 'admin']),
})
type AccountForm = z.input<typeof accountSchema>
type AccountValues = z.output<typeof accountSchema>

function CreateAccountDialog({ open, onOpenChange }: { open: boolean; onOpenChange: (v: boolean) => void }) {
  const create = useCreateAccount()
  const [created, setCreated] = useState<CreatedAccount | null>(null)
  const {
    register,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<AccountForm, unknown, AccountValues>({ resolver: zodResolver(accountSchema), defaultValues: { role: 'employee' } })

  const close = (v: boolean) => {
    onOpenChange(v)
    if (!v) {
      setCreated(null)
      reset()
      create.reset()
    }
  }

  const onSubmit = handleSubmit((values) =>
    create.mutate(values, {
      onSuccess: setCreated,
      onError: (e) => {
        if (isApiError(e, 'ALREADY_EXISTS')) setError('employeeCode', { message: 'This employee code is already used' })
      },
    }),
  )

  return (
    <ResponsiveDialog open={open} onOpenChange={close} title={created ? 'Account created' : 'New account'}>
      {created ? (
        <div className="space-y-4">
          <TempPassword code={created.employeeCode} password={created.temporaryPassword} />
          <Button className="w-full" onClick={() => close(false)}>
            Done
          </Button>
        </div>
      ) : (
        <form onSubmit={onSubmit} noValidate className="space-y-4">
          {create.error && !isApiError(create.error, 'ALREADY_EXISTS') && <ErrorAlert>{create.error.message}</ErrorAlert>}
          <Field id="employeeCode" label="Employee code" error={errors.employeeCode?.message}>
            <Input id="employeeCode" autoCapitalize="characters" placeholder="EMP-00042" aria-invalid={!!errors.employeeCode} {...register('employeeCode')} />
          </Field>
          <div className="grid gap-4 sm:grid-cols-2">
            <Field id="firstName" label="First name" error={errors.firstName?.message}>
              <Input id="firstName" aria-invalid={!!errors.firstName} {...register('firstName')} />
            </Field>
            <Field id="lastName" label="Last name" error={errors.lastName?.message}>
              <Input id="lastName" aria-invalid={!!errors.lastName} {...register('lastName')} />
            </Field>
          </div>
          <fieldset className="space-y-2">
            <legend className="text-sm font-medium">Role</legend>
            <div className="flex gap-4">
              {(['employee', 'admin'] as const).map((r) => (
                <label key={r} className="flex min-h-11 items-center gap-2 text-sm capitalize">
                  <input type="radio" value={r} {...register('role')} className="size-4 accent-[var(--color-accent)]" />
                  {r}
                </label>
              ))}
            </div>
          </fieldset>
          <p className="text-sm text-muted">A temporary password is generated. The employee sets their own at first sign-in.</p>
          <Button type="submit" className="w-full" loading={create.isPending}>
            Create account
          </Button>
        </form>
      )}
    </ResponsiveDialog>
  )
}
