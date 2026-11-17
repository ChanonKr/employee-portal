import { useEffect, useState, type FormEvent } from 'react'
import { toast } from 'sonner'
import { errorMessage } from '@/api/client'
import { ErrorAlert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardTitle } from '@/components/ui/card'
import { Field } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/cn'
import { timeAgo } from '@/lib/format'
import { PageTitle } from './AdminLayout'
import { useSetHistoryDays, useSettings } from './api'

const presets = [7, 14, 30, 90]

export function SettingsPage() {
  const settings = useSettings()
  const save = useSetHistoryDays()
  const [value, setValue] = useState('')
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (settings.data) setValue(String(settings.data.historyDays))
  }, [settings.data])

  if (settings.isPending) return <Skeleton className="h-48" />
  if (settings.isError) return <ErrorAlert>Could not load settings. {settings.error.message}</ErrorAlert>

  const s = settings.data
  const n = Number(value)
  const valid = Number.isInteger(n) && n >= 1 && n <= 365

  const submit = (e: FormEvent) => {
    e.preventDefault()
    if (!valid) return setError('Enter a whole number from 1 to 365.')
    save.mutate(n, {
      onSuccess: () => toast.success(`Employees now see the last ${n} days of changes`),
      onError: (err) => setError(errorMessage(err)),
    })
  }

  return (
    <>
      <PageTitle title="Settings" />
      <Card className="max-w-xl">
        <CardTitle>Change history</CardTitle>
        <p className="mt-1 text-sm text-muted">How many days of changes employees see on their profile.</p>
        <form onSubmit={submit} noValidate className="mt-4 space-y-4">
          <Field id="historyDays" label="Days shown" error={error ?? undefined} hint="1–365 days. Older records are kept, just hidden.">
            <Input
              id="historyDays"
              type="number"
              inputMode="numeric"
              min={1}
              max={365}
              value={value}
              onChange={(e) => {
                setValue(e.target.value)
                setError(null)
              }}
              aria-invalid={!!error}
              aria-describedby={error ? 'historyDays-error' : 'historyDays-hint'}
              className="max-w-32"
            />
          </Field>
          <div className="flex flex-wrap gap-2" role="group" aria-label="Presets">
            {presets.map((p) => (
              <button
                key={p}
                type="button"
                onClick={() => {
                  setValue(String(p))
                  setError(null)
                }}
                aria-pressed={n === p}
                className={cn(
                  'min-h-11 rounded-full border px-4 text-sm sm:min-h-9',
                  n === p ? 'border-accent bg-accent-soft text-accent' : 'border-border text-muted hover:text-fg',
                )}
              >
                {p} days
              </button>
            ))}
          </div>
          <div className="flex flex-wrap items-center justify-between gap-3">
            <p className="text-xs text-muted">
              {s.updatedBy && s.updatedAt ? `Last changed by ${s.updatedBy.employeeCode} ${timeAgo(s.updatedAt)}.` : 'Using the default value.'}
            </p>
            <Button type="submit" loading={save.isPending} disabled={n === s.historyDays}>
              Save
            </Button>
          </div>
        </form>
      </Card>
    </>
  )
}
