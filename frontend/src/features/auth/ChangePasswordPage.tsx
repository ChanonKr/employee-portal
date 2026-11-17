import { zodResolver } from '@hookform/resolvers/zod'
import { KeyRound } from 'lucide-react'
import { useForm } from 'react-hook-form'
import { useNavigate } from 'react-router'
import { toast } from 'sonner'
import { z } from 'zod'
import { isApiError } from '@/api/client'
import { ErrorAlert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Field } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { useChangePassword, useLogout, useMe } from './api'

// Mirrors the server policy so most mistakes are caught before submitting.
const schema = z
  .object({
    currentPassword: z.string().min(1, 'Enter your current password'),
    newPassword: z
      .string()
      .min(8, 'Use at least 8 characters')
      .refine((v) => /\p{L}/u.test(v) && /\d/.test(v), 'Use at least one letter and one number'),
    confirm: z.string(),
  })
  .refine((v) => v.newPassword === v.confirm, { path: ['confirm'], message: 'Passwords do not match' })
  .refine((v) => v.newPassword !== v.currentPassword, {
    path: ['newPassword'],
    message: 'Choose a password different from the current one',
  })
type Form = z.infer<typeof schema>

export function ChangePasswordPage() {
  const { data: me } = useMe()
  const change = useChangePassword()
  const logout = useLogout()
  const navigate = useNavigate()
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<Form>({ resolver: zodResolver(schema) })

  const onSubmit = handleSubmit(({ currentPassword, newPassword }) =>
    change.mutate(
      { currentPassword, newPassword },
      {
        onSuccess: () => {
          toast.success('Password updated')
          navigate('/profile', { replace: true })
        },
        onError: (e) => {
          if (isApiError(e, 'VALIDATION_ERROR') && e.fields) {
            for (const [field, message] of Object.entries(e.fields)) {
              if (field === 'currentPassword' || field === 'newPassword') setError(field, { message: `Password ${message}` })
            }
          }
        },
      },
    ),
  )

  const forced = me?.mustChangePassword
  const generalError = change.error && !(isApiError(change.error, 'VALIDATION_ERROR') && change.error.fields)

  return (
    <main className="grid min-h-dvh place-items-center px-4 py-10">
      <div className="w-full max-w-sm">
        <div className="mb-8 flex flex-col items-center gap-3 text-center">
          <div className="grid size-12 place-items-center rounded-2xl bg-accent-soft text-accent">
            <KeyRound className="size-6" aria-hidden />
          </div>
          <h1 className="text-2xl font-semibold">{forced ? 'Set a new password' : 'Change password'}</h1>
          {forced && (
            <p className="text-sm text-muted">You signed in with a temporary password. Choose your own to continue.</p>
          )}
        </div>
        <form onSubmit={onSubmit} noValidate className="space-y-4 rounded-2xl border border-border bg-surface p-6">
          {generalError && <ErrorAlert>{change.error.message}</ErrorAlert>}
          <Field id="currentPassword" label={forced ? 'Temporary password' : 'Current password'} error={errors.currentPassword?.message}>
            <Input id="currentPassword" type="password" autoComplete="current-password" aria-invalid={!!errors.currentPassword} aria-describedby={errors.currentPassword ? 'currentPassword-error' : undefined} {...register('currentPassword')} />
          </Field>
          <Field id="newPassword" label="New password" error={errors.newPassword?.message} hint="At least 8 characters with a letter and a number.">
            <Input id="newPassword" type="password" autoComplete="new-password" aria-invalid={!!errors.newPassword} aria-describedby={errors.newPassword ? 'newPassword-error' : undefined} {...register('newPassword')} />
          </Field>
          <Field id="confirm" label="Confirm new password" error={errors.confirm?.message}>
            <Input id="confirm" type="password" autoComplete="new-password" aria-invalid={!!errors.confirm} aria-describedby={errors.confirm ? 'confirm-error' : undefined} {...register('confirm')} />
          </Field>
          <Button type="submit" className="w-full" loading={change.isPending}>
            Save password
          </Button>
          {forced ? (
            <Button variant="ghost" className="w-full" onClick={() => logout.mutate(undefined, { onSettled: () => navigate('/login') })}>
              Sign out
            </Button>
          ) : (
            <Button variant="ghost" className="w-full" onClick={() => navigate(-1)}>
              Cancel
            </Button>
          )}
        </form>
      </div>
    </main>
  )
}
