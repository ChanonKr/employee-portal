import { ArrowDown, ArrowUp, Check, Pencil, Plus, Search, X } from 'lucide-react'
import { useState, type FormEvent } from 'react'
import { toast } from 'sonner'
import { errorMessage, isApiError } from '@/api/client'
import type { AdminSkill } from '@/api/types'
import { ErrorAlert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
import { cn } from '@/lib/cn'
import { PageTitle } from './AdminLayout'
import { useAdminSkills, useCreateSkill, useReorderSkills, useUpdateSkill } from './api'

const duplicateMessage = 'A skill with this name already exists.'

export function SkillsAdminPage() {
  const skills = useAdminSkills()
  const [q, setQ] = useState('')
  const reorder = useReorderSkills()

  if (skills.isPending) return <Skeleton className="h-64" />
  if (skills.isError) return <ErrorAlert>Could not load skills. {skills.error.message}</ErrorAlert>

  const all = skills.data
  const filtering = q.trim() !== ''
  const shown = filtering ? all.filter((s) => s.name.toLowerCase().includes(q.trim().toLowerCase())) : all

  // Move one row and renumber the whole list, so equal sort orders never get stuck.
  const move = (index: number, delta: -1 | 1) => {
    const next = [...all]
    const [item] = next.splice(index, 1)
    next.splice(index + delta, 0, item!)
    reorder.mutate(
      next.map((s, i) => ({ id: s.id, sortOrder: (i + 1) * 10 })),
      { onError: (e) => toast.error(errorMessage(e)) },
    )
  }

  return (
    <>
      <PageTitle title="Skills" />
      <p className="mb-4 text-sm text-muted">
        Inactive skills disappear from the picker, but employees who already have them keep them. Skills are never deleted.
      </p>
      <AddSkill />
      <div className="relative my-4">
        <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted" aria-hidden />
        <Input value={q} onChange={(e) => setQ(e.target.value)} type="search" placeholder="Search skills" aria-label="Search skills" className="pl-9" />
      </div>
      <Card className="p-0 sm:p-0">
        <ul className="divide-y divide-border">
          {shown.map((s) => {
            const index = all.indexOf(s)
            return (
              <SkillRow
                key={s.id}
                skill={s}
                // Reordering only makes sense on the full list.
                onUp={!filtering && index > 0 ? () => move(index, -1) : undefined}
                onDown={!filtering && index < all.length - 1 ? () => move(index, 1) : undefined}
                busy={reorder.isPending}
              />
            )
          })}
          {shown.length === 0 && <li className="p-6 text-center text-sm text-muted">No skills match “{q}”.</li>}
        </ul>
      </Card>
    </>
  )
}

function AddSkill() {
  const create = useCreateSkill()
  const [name, setName] = useState('')
  const [error, setError] = useState<string | null>(null)

  const submit = (e: FormEvent) => {
    e.preventDefault()
    if (!name.trim()) return setError('Enter a skill name.')
    create.mutate(name, {
      onSuccess: () => {
        toast.success(`Added ${name.trim()}`)
        setName('')
        setError(null)
      },
      onError: (err) => setError(isApiError(err, 'ALREADY_EXISTS') ? duplicateMessage : errorMessage(err)),
    })
  }

  return (
    <form onSubmit={submit} noValidate className="space-y-1.5">
      <div className="flex gap-2">
        <Input
          value={name}
          onChange={(e) => {
            setName(e.target.value)
            setError(null)
          }}
          placeholder="New skill name"
          aria-label="New skill name"
          maxLength={50}
          aria-invalid={!!error}
          aria-describedby={error ? 'add-skill-error' : undefined}
        />
        <Button type="submit" loading={create.isPending} className="shrink-0">
          <Plus className="size-4" aria-hidden /> Add skill
        </Button>
      </div>
      {error && (
        <p id="add-skill-error" role="alert" className="text-sm text-danger">
          {error}
        </p>
      )}
    </form>
  )
}

function SkillRow({ skill, onUp, onDown, busy }: { skill: AdminSkill; onUp?: () => void; onDown?: () => void; busy: boolean }) {
  const update = useUpdateSkill()
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState(skill.name)
  const [error, setError] = useState<string | null>(null)

  const save = (e: FormEvent) => {
    e.preventDefault()
    if (name.trim() === skill.name) return setEditing(false)
    update.mutate(
      { id: skill.id, name },
      {
        onSuccess: () => {
          setEditing(false)
          setError(null)
        },
        onError: (err) => setError(isApiError(err, 'ALREADY_EXISTS') ? duplicateMessage : errorMessage(err)),
      },
    )
  }

  return (
    <li className={cn('flex flex-wrap items-center gap-3 px-4 py-3', !skill.isActive && 'bg-bg/60')}>
      <div className="flex flex-col">
        <Button variant="ghost" size="icon" className="size-8 sm:size-7" aria-label={`Move ${skill.name} up`} disabled={!onUp || busy} onClick={onUp}>
          <ArrowUp className="size-3.5" />
        </Button>
        <Button variant="ghost" size="icon" className="size-8 sm:size-7" aria-label={`Move ${skill.name} down`} disabled={!onDown || busy} onClick={onDown}>
          <ArrowDown className="size-3.5" />
        </Button>
      </div>

      <div className="min-w-0 flex-1">
        {editing ? (
          <form onSubmit={save} className="flex items-center gap-1">
            <Input value={name} onChange={(e) => setName(e.target.value)} aria-label="Skill name" maxLength={50} autoFocus aria-invalid={!!error} />
            <Button type="submit" variant="ghost" size="icon" aria-label="Save name" loading={update.isPending}>
              {!update.isPending && <Check className="size-4" />}
            </Button>
            <Button
              variant="ghost"
              size="icon"
              aria-label="Cancel"
              onClick={() => {
                setEditing(false)
                setName(skill.name)
                setError(null)
              }}
            >
              <X className="size-4" />
            </Button>
          </form>
        ) : (
          <div className="flex items-center gap-1">
            <span className={cn('font-medium break-words', !skill.isActive && 'text-muted')}>{skill.name}</span>
            <Button variant="ghost" size="icon" aria-label={`Rename ${skill.name}`} onClick={() => setEditing(true)}>
              <Pencil className="size-3.5" />
            </Button>
          </div>
        )}
        {error && (
          <p role="alert" className="mt-1 text-sm text-danger">
            {error}
          </p>
        )}
        <p className="text-xs text-muted">
          Used by {skill.usedBy} employee{skill.usedBy === 1 ? '' : 's'}
        </p>
      </div>

      <label className="flex min-h-11 items-center gap-2 text-sm text-muted">
        {skill.isActive ? 'Active' : 'Inactive'}
        <Switch
          checked={skill.isActive}
          aria-label={`${skill.name} active`}
          disabled={update.isPending}
          onCheckedChange={(isActive) =>
            update.mutate({ id: skill.id, isActive }, { onError: (e) => toast.error(errorMessage(e)) })
          }
        />
      </label>
    </li>
  )
}
