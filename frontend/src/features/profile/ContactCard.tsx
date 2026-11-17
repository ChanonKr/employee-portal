import { zodResolver } from '@hookform/resolvers/zod'
import { ChevronRight, Mail, MapPin, Pencil, Phone, Plus, type LucideIcon } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { toast } from 'sonner'
import { z } from 'zod'
import { isApiError } from '@/api/client'
import type { Profile } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Card, CardTitle } from '@/components/ui/card'
import { ResponsiveDialog } from '@/components/ui/dialog'
import { Input, Textarea } from '@/components/ui/input'
import { formatPhone } from '@/lib/format'
import { useIsDesktop } from '@/lib/useMediaQuery'
import { useUpdateContact, type ContactKey } from './api'
import { contactSchemas } from './contactRules'

type FieldDef = {
  key: ContactKey
  label: string
  icon: LucideIcon
  value: (p: Profile) => string | null
  display?: (v: string | null) => string | null
}

const fields: FieldDef[] = [
  { key: 'mobile', label: 'Mobile', icon: Phone, value: (p) => p.contact.mobilePhone, display: formatPhone },
  { key: 'email', label: 'Email', icon: Mail, value: (p) => p.contact.email },
  { key: 'address', label: 'Address', icon: MapPin, value: (p) => p.contact.address },
]

/** Each field is edited and saved on its own (one request, one history entry). */
export function ContactCard({ profile }: { profile: Profile }) {
  return (
    <Card>
      <CardTitle>Contact</CardTitle>
      <ul className="mt-2 divide-y divide-border">
        {fields.map((f) => (
          <ContactRow key={f.key} def={f} value={f.value(profile)} />
        ))}
      </ul>
    </Card>
  )
}

function ContactRow({ def, value }: { def: FieldDef; value: string | null }) {
  const desktop = useIsDesktop()
  const [editing, setEditing] = useState(false)
  const Icon = def.icon
  const shown = def.display ? def.display(value) : value
  const id = `contact-${def.key}`

  const label = (
    <span className="flex w-24 shrink-0 items-center gap-2 text-sm text-muted">
      <Icon className="size-4" aria-hidden />
      {def.label}
    </span>
  )

  if (desktop) {
    return (
      <li className="flex min-h-14 items-start gap-4 py-3">
        <span className="pt-2.5">{label}</span>
        {editing ? (
          <ContactForm def={def} value={value} id={id} onDone={() => setEditing(false)} inline />
        ) : (
          <>
            <span className={`min-w-0 flex-1 pt-2 break-words whitespace-pre-line ${shown ? '' : 'text-muted italic'}`}>
              {shown ?? 'Not added yet'}
            </span>
            <Button variant={value ? 'ghost' : 'secondary'} size="sm" onClick={() => setEditing(true)} aria-label={`${value ? 'Edit' : 'Add'} ${def.label.toLowerCase()}`}>
              {value ? <Pencil className="size-4" /> : <Plus className="size-4" />}
              {value ? 'Edit' : 'Add'}
            </Button>
          </>
        )}
      </li>
    )
  }

  // Phones: the whole row opens a bottom sheet.
  return (
    <li>
      <button
        type="button"
        onClick={() => setEditing(true)}
        className="flex min-h-14 w-full items-center gap-3 py-3 text-left"
        aria-label={`${value ? 'Edit' : 'Add'} ${def.label.toLowerCase()}`}
      >
        {label}
        <span className={`min-w-0 flex-1 break-words ${shown ? '' : 'text-muted italic'}`}>{shown ?? 'Not added yet'}</span>
        <ChevronRight className="size-4 shrink-0 text-muted" aria-hidden />
      </button>
      <ResponsiveDialog open={editing} onOpenChange={setEditing} title={`${value ? 'Edit' : 'Add'} ${def.label.toLowerCase()}`}>
        <ContactForm def={def} value={value} id={id} onDone={() => setEditing(false)} />
      </ResponsiveDialog>
    </li>
  )
}

function ContactForm({ def, value, id, onDone, inline }: { def: FieldDef; value: string | null; id: string; onDone: () => void; inline?: boolean }) {
  const update = useUpdateContact(def.key)
  const schema = z.object({ value: contactSchemas[def.key] })
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors },
  } = useForm<{ value: string }, unknown, { value: string }>({
    resolver: zodResolver(schema),
    defaultValues: { value: value ?? '' },
  })

  const onSubmit = handleSubmit(({ value: v }) =>
    update.mutate(v, {
      onSuccess: () => {
        toast.success(`${def.label} saved`)
        onDone()
      },
      onError: (e) => {
        if (isApiError(e, 'VALIDATION_ERROR')) {
          setError('value', { message: e.fields?.[def.key] ? `${def.label} ${e.fields[def.key]}` : e.message })
        } else if (isApiError(e, 'VERSION_CONFLICT')) {
          onDone()
        }
      },
    }),
  )

  const common = {
    id,
    'aria-label': def.label,
    'aria-invalid': !!errors.value,
    'aria-describedby': errors.value ? `${id}-error` : undefined,
    autoFocus: true,
    ...register('value'),
  }

  return (
    <form onSubmit={onSubmit} noValidate className={inline ? 'flex min-w-0 flex-1 flex-col gap-2' : 'space-y-4'}>
      <div className="space-y-1.5">
        {def.key === 'address' ? (
          <Textarea {...common} autoComplete="street-address" rows={3} />
        ) : (
          <Input
            {...common}
            type={def.key === 'mobile' ? 'tel' : 'email'}
            inputMode={def.key === 'mobile' ? 'tel' : 'email'}
            autoComplete={def.key === 'mobile' ? 'tel' : 'email'}
          />
        )}
        {errors.value && (
          <p id={`${id}-error`} role="alert" className="text-sm text-danger">
            {errors.value.message}
          </p>
        )}
      </div>
      <div className={inline ? 'flex justify-end gap-2' : 'grid grid-cols-2 gap-2'}>
        <Button variant="secondary" size={inline ? 'sm' : 'md'} onClick={onDone} disabled={update.isPending}>
          Cancel
        </Button>
        <Button type="submit" size={inline ? 'sm' : 'md'} loading={update.isPending}>
          Save
        </Button>
      </div>
    </form>
  )
}
