import * as RD from '@radix-ui/react-dialog'
import { X } from 'lucide-react'
import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'
import { useIsDesktop } from '@/lib/useMediaQuery'

type Props = {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: ReactNode
  children: ReactNode
  /** Phones: full-height sheet (e.g. a searchable list) instead of a short one. */
  tall?: boolean
}

/**
 * Centered dialog on tablet/desktop, bottom sheet on phones (thumb-friendly and
 * the on-screen keyboard does not cover the form).
 */
export function ResponsiveDialog({ open, onOpenChange, title, description, children, tall }: Props) {
  const desktop = useIsDesktop()
  return (
    <RD.Root open={open} onOpenChange={onOpenChange}>
      <RD.Portal>
        <RD.Overlay className="fixed inset-0 z-40 bg-black/40" />
        <RD.Content
          className={cn(
            'fixed z-50 flex flex-col bg-surface shadow-xl focus:outline-none',
            desktop
              ? 'top-1/2 left-1/2 max-h-[85dvh] w-[min(32rem,calc(100vw-2rem))] -translate-x-1/2 -translate-y-1/2 rounded-2xl'
              : cn('inset-x-0 bottom-0 rounded-t-2xl pb-[env(safe-area-inset-bottom)]', tall ? 'h-[92dvh]' : 'max-h-[92dvh]'),
          )}
        >
          {!desktop && <div className="mx-auto mt-2 h-1.5 w-10 rounded-full bg-border" aria-hidden />}
          <div className="flex items-start justify-between gap-4 px-5 pt-4">
            <div>
              <RD.Title className="text-lg font-semibold">{title}</RD.Title>
              {description ? (
                <RD.Description className="mt-1 text-sm text-muted">{description}</RD.Description>
              ) : (
                <RD.Description className="sr-only">{title}</RD.Description>
              )}
            </div>
            <RD.Close className="-mr-2 grid size-11 shrink-0 place-items-center rounded-lg text-muted hover:bg-bg" aria-label="Close">
              <X className="size-5" />
            </RD.Close>
          </div>
          <div className="min-h-0 flex-1 overflow-y-auto px-5 pt-3 pb-5">{children}</div>
        </RD.Content>
      </RD.Portal>
    </RD.Root>
  )
}
