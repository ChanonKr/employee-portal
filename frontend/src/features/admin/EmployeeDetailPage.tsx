import { zodResolver } from '@hookform/resolvers/zod'
import { ArrowLeft, KeyRound, Mail, MapPin, Phone } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Link, useParams } from 'react-router'
import { toast } from 'sonner'
import { z } from 'zod'
import { api, isApiError } from '@/api/client'
import type { EmployeeDetail, HistoryPage } from '@/api/types'
import { ErrorAlert } from '@/components/ui/alert'
import { Avatar } from '@/components/ui/avatar'
import { ViewablePhoto } from '@/components/ui/photo-viewer'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardTitle } from '@/components/ui/card'
import { ResponsiveDialog } from '@/components/ui/dialog'
import { Field } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { useMe } from '@/features/auth/api'
import { HistoryTimeline, PAGE_SIZE } from '@/features/history/HistoryTimeline'
import { formatPhone } from '@/lib/format'
import { adminKeys, useEmployee, useRenameEmployee, useResetPassword } from './api'
import { TempPassword } from './TempPassword'

export function EmployeeDetailPage() {
  const id = Number(useParams().id)
  const employee = useEmployee(id)
  const { data: me } = useMe()

  return (
    <div className="space-y-4">
      <Link to="/admin/employees" className="inline-flex min-h-11 items-center gap-1 text-sm text-muted hover:text-fg sm:min-h-0">
        <ArrowLeft className="size-4" aria-hidden /> Employees
      </Link>
      {employee.isPending ? (
        <Skeleton className="h-64" />
      ) : employee.isError ? (
        <ErrorAlert>{isApiError(employee.error, 'NOT_FOUND') ? 'This employee does not exist.' : employee.error.message}</ErrorAlert>
      ) : (
        <>
          <Overview e={employee.data} />
          <div className="grid gap-4 lg:grid-cols-2">
            <LegalNameCard e={employee.data} />
            <AccessCard e={employee.data} />
          </div>
          <section aria-labelledby="history-title" className="space-y-3">
            <h2 id="history-title" className="text-lg font-semibold">
              Change history <span className="text-sm font-normal text-muted">· all time</span>
            </h2>
            <HistoryTimeline
              queryKey={adminKeys.employeeHistory(id)}
              selfUserId={me?.userId}
              fetchPage={(field, page) => {
                const q = new URLSearchParams({ page: String(page), pageSize: String(PAGE_SIZE) })
                if (field) q.set('field', field)
                return api.get<HistoryPage>(`/admin/employees/${id}/history?${q}`)
              }}
            />
          </section>
        </>
      )}
    </div>
  )
}

function Overview({ e }: { e: EmployeeDetail }) {
  const rows = [
    { icon: Phone, label: 'Mobile', value: formatPhone(e.contact.mobilePhone) },
    { icon: Mail, label: 'Email', value: e.contact.email },
    { icon: MapPin, label: 'Address', value: e.contact.address },
  ]
  return (
    <Card className="flex flex-col gap-4 sm:flex-row sm:items-start">
      {e.photoUrl ? (
        <ViewablePhoto url={e.photoUrl} label={`${e.firstName} ${e.lastName}`} buttonLabel="View photo" className="self-start">
          <Avatar url={e.photoUrl} firstName={e.firstName} lastName={e.lastName} className="size-20 text-xl" />
        </ViewablePhoto>
      ) : (
        <Avatar url={null} firstName={e.firstName} lastName={e.lastName} className="size-20 text-xl" />
      )}
      <div className="min-w-0 flex-1 space-y-3">
        <div>
          <h1 className="text-xl font-semibold break-words">
            {e.firstName} {e.lastName}
          </h1>
          <div className="mt-1 flex flex-wrap items-center gap-2 text-sm text-muted">
            {e.employeeCode}
            {e.role === 'admin' && <Badge tone="accent">Admin</Badge>}
            {!e.isActive && <Badge tone="danger">Inactive</Badge>}
          </div>
        </div>
        <dl className="grid gap-2 text-sm sm:grid-cols-[6rem_1fr]">
          {rows.map(({ icon: Icon, label, value }) => (
            <div key={label} className="contents">
              <dt className="flex items-center gap-2 text-muted">
                <Icon className="size-4" aria-hidden /> {label}
              </dt>
              <dd className={`break-words ${value ? '' : 'text-muted italic'}`}>{value ?? 'Not added yet'}</dd>
            </div>
          ))}
          <dt className="text-muted">Skills</dt>
          <dd className="flex flex-wrap gap-1.5">
            {e.skills.length === 0 ? (
              <span className="text-muted italic">None</span>
            ) : (
              e.skills.map((s) => (
                <Badge key={s.id} title={s.isActive ? undefined : 'Turned off in the catalog'}>
                  {s.name}
                  {!s.isActive && ' (off)'}
                </Badge>
              ))
            )}
          </dd>
        </dl>
      </div>
    </Card>
  )
}

const nameSchema = z.object({
  firstName: z.string().trim().min(1, 'Enter a first name').max(100, 'Use at most 100 characters'),
  lastName: z.string().trim().min(1, 'Enter a last name').max(100, 'Use at most 100 characters'),
})
type NameForm = z.infer<typeof nameSchema>

/** The only place a legal name can change. Each edited part is logged with the admin as author. */
function LegalNameCard({ e }: { e: EmployeeDetail }) {
  const rename = useRenameEmployee(e.id)
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isDirty },
  } = useForm<NameForm>({ resolver: zodResolver(nameSchema), values: { firstName: e.firstName, lastName: e.lastName } })

  const onSubmit = handleSubmit((v) =>
    rename.mutate(
      { ...v, version: e.version },
      {
        onSuccess: () => toast.success('Name updated'),
        onError: (err) => {
          if (isApiError(err, 'VALIDATION_ERROR') && err.fields) {
            for (const [f, m] of Object.entries(err.fields)) if (f === 'firstName' || f === 'lastName') setError(f, { message: m })
          }
        },
      },
    ),
  )

  return (
    <Card>
      <div className="flex items-center gap-2">
        <CardTitle>Legal name</CardTitle>
        <Badge tone="accent">Admin only</Badge>
      </div>
      <form onSubmit={onSubmit} noValidate className="mt-4 space-y-4">
        <div className="grid gap-4 sm:grid-cols-2">
          <Field id="admin-first" label="First name" error={errors.firstName?.message}>
            <Input id="admin-first" aria-invalid={!!errors.firstName} {...register('firstName')} />
          </Field>
          <Field id="admin-last" label="Last name" error={errors.lastName?.message}>
            <Input id="admin-last" aria-invalid={!!errors.lastName} {...register('lastName')} />
          </Field>
        </div>
        <Button type="submit" disabled={!isDirty} loading={rename.isPending}>
          Save name
        </Button>
      </form>
    </Card>
  )
}

function AccessCard({ e }: { e: EmployeeDetail }) {
  const reset = useResetPassword()
  const [confirming, setConfirming] = useState(false)
  const [temp, setTemp] = useState<string | null>(null)

  const close = (v: boolean) => {
    setConfirming(v)
    if (!v) setTemp(null)
  }

  return (
    <Card>
      <CardTitle>Sign-in</CardTitle>
      <p className="mt-2 text-sm text-muted">Reset the password if the employee forgot it. They will choose a new one at next sign-in.</p>
      <Button variant="secondary" className="mt-4" onClick={() => setConfirming(true)}>
        <KeyRound className="size-4" aria-hidden /> Reset password
      </Button>
      <ResponsiveDialog open={confirming} onOpenChange={close} title={temp ? 'Password reset' : 'Reset password?'}>
        {temp ? (
          <div className="space-y-4">
            <TempPassword code={e.employeeCode} password={temp} />
            <Button className="w-full" onClick={() => close(false)}>
              Done
            </Button>
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-sm">
              {e.firstName} {e.lastName}’s current password stops working immediately.
            </p>
            <div className="grid grid-cols-2 gap-2">
              <Button variant="secondary" onClick={() => close(false)}>
                Cancel
              </Button>
              <Button variant="danger" loading={reset.isPending} onClick={() => reset.mutate(e.id, { onSuccess: (r) => setTemp(r.temporaryPassword) })}>
                Reset
              </Button>
            </div>
          </div>
        )}
      </ResponsiveDialog>
    </Card>
  )
}
