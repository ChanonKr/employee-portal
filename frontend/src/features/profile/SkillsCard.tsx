import * as Popover from '@radix-ui/react-popover'
import { Check, Plus, Search, X } from 'lucide-react'
import { useMemo, useState } from 'react'
import { toast } from 'sonner'
import type { Profile, Skill } from '@/api/types'
import { Button } from '@/components/ui/button'
import { Card, CardTitle } from '@/components/ui/card'
import { ResponsiveDialog } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/cn'
import { useIsDesktop } from '@/lib/useMediaQuery'
import { useSetSkills, useSkillCatalog } from './api'

const MAX_SKILLS = 20

/**
 * Chips are a local draft; Save sends the whole set in one request.
 * Skills an admin turned off stay on the profile (and in this list) but are not
 * offered in the picker.
 */
export function SkillsCard({ profile }: { profile: Profile }) {
  const saved = useMemo(() => profile.skills.map((s) => s.id), [profile.skills])
  const [draft, setDraft] = useState<number[] | null>(null) // null = not editing
  const ids = draft ?? saved
  const setSkills = useSetSkills()
  const catalog = useSkillCatalog()

  // Names for every id we might show: catalog + what the profile already has.
  const byId = useMemo(() => {
    const m = new Map<number, Skill>()
    for (const s of profile.skills) m.set(s.id, s)
    for (const s of catalog.data ?? []) m.set(s.id, s)
    return m
  }, [profile.skills, catalog.data])

  const changed = draft !== null && (draft.length !== saved.length || draft.some((id) => !saved.includes(id)))

  const toggle = (id: number) => {
    const cur = draft ?? saved
    if (cur.includes(id)) setDraft(cur.filter((x) => x !== id))
    else if (cur.length >= MAX_SKILLS) toast.error(`You can choose up to ${MAX_SKILLS} skills.`)
    else setDraft([...cur, id])
  }

  const save = () =>
    setSkills.mutate(ids, {
      onSuccess: () => {
        toast.success('Skills saved')
        setDraft(null)
      },
      onError: () => setDraft(null),
    })

  return (
    <Card>
      <div className="flex items-center justify-between gap-2">
        <CardTitle>Skills</CardTitle>
        <SkillPicker catalog={catalog.data ?? []} loading={catalog.isPending} selected={ids} onToggle={toggle} />
      </div>

      {ids.length === 0 ? (
        <p className="mt-3 text-sm text-muted italic">No skills added yet.</p>
      ) : (
        <ul className="mt-3 flex flex-wrap gap-2" aria-label="Your skills">
          {ids.map((id) => (
            <li key={id} className="inline-flex items-center rounded-full border border-border bg-bg pl-3 text-sm">
              {byId.get(id)?.name ?? '…'}
              <button
                type="button"
                onClick={() => toggle(id)}
                className="grid size-9 place-items-center rounded-full text-muted hover:text-danger"
                aria-label={`Remove ${byId.get(id)?.name ?? 'skill'}`}
              >
                <X className="size-3.5" />
              </button>
            </li>
          ))}
        </ul>
      )}

      {changed && (
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="secondary" size="sm" onClick={() => setDraft(null)} disabled={setSkills.isPending}>
            Cancel
          </Button>
          <Button size="sm" onClick={save} loading={setSkills.isPending}>
            Save skills
          </Button>
        </div>
      )}
    </Card>
  )
}

function SkillPicker(props: { catalog: Skill[]; loading: boolean; selected: number[]; onToggle: (id: number) => void }) {
  const desktop = useIsDesktop()
  const [open, setOpen] = useState(false)
  const trigger = (
    <Button variant="secondary" size="sm" onClick={desktop ? undefined : () => setOpen(true)}>
      <Plus className="size-4" aria-hidden />
      Add
    </Button>
  )

  if (desktop) {
    return (
      <Popover.Root open={open} onOpenChange={setOpen}>
        <Popover.Trigger asChild>{trigger}</Popover.Trigger>
        <Popover.Portal>
          <Popover.Content align="end" sideOffset={8} className="z-50 w-72 rounded-xl border border-border bg-surface p-3 shadow-lg">
            <SkillList {...props} />
          </Popover.Content>
        </Popover.Portal>
      </Popover.Root>
    )
  }
  return (
    <>
      {trigger}
      <ResponsiveDialog open={open} onOpenChange={setOpen} title="Choose skills" tall>
        <SkillList {...props} />
        <Button className="mt-4 w-full" onClick={() => setOpen(false)}>
          Done
        </Button>
      </ResponsiveDialog>
    </>
  )
}

function SkillList({ catalog, loading, selected, onToggle }: { catalog: Skill[]; loading: boolean; selected: number[]; onToggle: (id: number) => void }) {
  const [q, setQ] = useState('')
  const shown = catalog.filter((s) => s.name.toLowerCase().includes(q.trim().toLowerCase()))
  return (
    <div className="space-y-2">
      <div className="relative">
        <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted" aria-hidden />
        <Input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Search skills" aria-label="Search skills" className="pl-9" autoFocus />
      </div>
      {loading ? (
        <p className="py-4 text-center text-sm text-muted">Loading…</p>
      ) : shown.length === 0 ? (
        <p className="py-4 text-center text-sm text-muted">No skills match “{q}”.</p>
      ) : (
        <ul className="max-h-72 overflow-y-auto" role="listbox" aria-multiselectable aria-label="Skills">
          {shown.map((s) => {
            const on = selected.includes(s.id)
            return (
              <li key={s.id} role="option" aria-selected={on}>
                <button
                  type="button"
                  onClick={() => onToggle(s.id)}
                  className={cn('flex min-h-11 w-full items-center justify-between rounded-lg px-3 text-left hover:bg-bg', on && 'font-medium text-accent')}
                >
                  {s.name}
                  {on && <Check className="size-4" aria-hidden />}
                </button>
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}
