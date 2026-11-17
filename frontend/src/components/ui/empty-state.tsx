import type { LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'

export function EmptyState({ icon: Icon, title, children }: { icon: LucideIcon; title: string; children?: ReactNode }) {
  return (
    <div className="flex flex-col items-center gap-2 px-4 py-8 text-center">
      <div className="grid size-10 place-items-center rounded-full bg-bg text-muted">
        <Icon className="size-5" aria-hidden />
      </div>
      <p className="font-medium">{title}</p>
      {children && <p className="max-w-sm text-sm text-muted">{children}</p>}
    </div>
  )
}
