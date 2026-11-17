import { zodResolver } from '@hookform/resolvers/zod'
import { Eye, EyeOff, UserRound } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { useNavigate, useSearchParams } from 'react-router'
import { z } from 'zod'
import { isApiError } from '@/api/client'
import { ErrorAlert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Field } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { safeNext, useLogin } from './api'

const schema = z.object({
  employeeCode: z.string().trim().min(1, 'Enter your employee code'),
  password: z.string().min(1, 'Enter your password'),
})
type Form = z.infer<typeof schema>

export function LoginPage() {
  const [params] = useSearchParams()
  const navigate = useNavigate()
  const login = useLogin()
  const [showPassword, setShowPassword] = useState(false)
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<Form>({ resolver: zodResolver(schema) })

  const onSubmit = handleSubmit((values) =>
    login.mutate(values, { onSuccess: () => navigate(safeNext(params.get('next')), { replace: true }) }),
  )

  const failure = login.error
    ? isApiError(login.error, 'RATE_LIMITED')
      ? 'Too many attempts. Try again in a minute.'
      : isApiError(login.error, 'UNAUTHORIZED')
        ? 'Employee code or password is incorrect.'
        : login.error.message
    : null

  return (
    <main className="grid min-h-dvh place-items-center px-4 py-10">
      <div className="w-full max-w-sm">
        <div className="mb-8 flex flex-col items-center gap-3 text-center">
          <div className="grid size-12 place-items-center rounded-2xl bg-accent text-accent-fg">
            <UserRound className="size-6" aria-hidden />
          </div>
          <h1 className="text-2xl font-semibold">Employee portal</h1>
          <p className="text-sm text-muted">Sign in with your employee code.</p>
        </div>

        <form onSubmit={onSubmit} noValidate className="space-y-4 rounded-2xl border border-border bg-surface p-6">
          {failure && <ErrorAlert>{failure}</ErrorAlert>}
          <Field id="employeeCode" label="Employee code" error={errors.employeeCode?.message}>
            <Input
              id="employeeCode"
              autoComplete="username"
              autoCapitalize="characters"
              placeholder="EMP-00001"
              aria-invalid={!!errors.employeeCode} aria-describedby={errors.employeeCode ? 'employeeCode-error' : undefined}
              {...register('employeeCode')}
            />
          </Field>
          <Field id="password" label="Password" error={errors.password?.message}>
            <div className="relative">
              <Input
                id="password"
                type={showPassword ? 'text' : 'password'}
                autoComplete="current-password"
                className="pr-12"
                aria-invalid={!!errors.password} aria-describedby={errors.password ? 'password-error' : undefined}
                {...register('password')}
              />
              <button
                type="button"
                onClick={() => setShowPassword((s) => !s)}
                className="absolute inset-y-0 right-0 grid w-11 place-items-center text-muted"
                aria-label={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
              </button>
            </div>
          </Field>
          <Button type="submit" className="w-full" loading={login.isPending}>
            Sign in
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-muted">
          Forgot your password or need access? Contact HR.
        </p>
      </div>
    </main>
  )
}
