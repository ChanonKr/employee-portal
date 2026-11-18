import * as RD from '@radix-ui/react-dialog'
import { ChevronLeft, ChevronRight, X } from 'lucide-react'
import { useEffect, useState, type KeyboardEvent, type ReactNode } from 'react'
import { cn } from '@/lib/cn'

export type ViewerPhoto = { url: string; label: string }

/**
 * Full-screen photo popup. With several photos (e.g. a history change: previous
 * and new), arrows and ←/→ switch between them. Esc, the close button or a tap
 * outside the photo closes it.
 */
export function PhotoViewer({
  photos,
  index,
  onClose,
}: {
  photos: ViewerPhoto[]
  index: number | null // null = closed
  onClose: () => void
}) {
  const [current, setCurrent] = useState(index ?? 0)
  useEffect(() => {
    if (index !== null) setCurrent(index)
  }, [index])

  const open = index !== null && photos.length > 0
  const photo = photos[Math.min(current, photos.length - 1)]
  const many = photos.length > 1
  const go = (step: number) => setCurrent((i) => (i + step + photos.length) % photos.length)
  const onKeyDown = (e: KeyboardEvent) => {
    if (!many) return
    if (e.key === 'ArrowLeft') go(-1)
    if (e.key === 'ArrowRight') go(1)
  }

  return (
    <RD.Root open={open} onOpenChange={(o) => !o && onClose()}>
      <RD.Portal>
        <RD.Overlay className="fixed inset-0 z-40 bg-black/90" />
        <RD.Content
          onKeyDown={onKeyDown}
          onClick={(e) => e.target === e.currentTarget && onClose()}
          className="fixed inset-0 z-50 flex flex-col items-center justify-center gap-3 p-4 focus:outline-none"
        >
          {photo && (
            <>
              <RD.Title className="text-sm font-medium text-white">
                {photo.label}
                {many && <span className="text-white/70"> · {current + 1} of {photos.length}</span>}
              </RD.Title>
              <RD.Description className="sr-only">Enlarged photo</RD.Description>
              <img
                src={photo.url}
                alt={photo.label}
                className="max-h-[80dvh] w-[min(32rem,100%)] rounded-2xl bg-surface object-contain shadow-2xl"
              />
            </>
          )}
          <RD.Close
            className="absolute top-[max(1rem,env(safe-area-inset-top))] right-4 grid size-11 place-items-center rounded-full bg-white/15 text-white hover:bg-white/25"
            aria-label="Close"
          >
            <X className="size-5" />
          </RD.Close>
          {many && (
            // Phones: a row under the photo (keeps the photo clear). Larger screens: the screen edges.
            <div className="flex gap-6 sm:contents">
              <NavButton side="left" onClick={() => go(-1)} label="Previous photo">
                <ChevronLeft className="size-6" />
              </NavButton>
              <NavButton side="right" onClick={() => go(1)} label="Next photo">
                <ChevronRight className="size-6" />
              </NavButton>
            </div>
          )}
        </RD.Content>
      </RD.Portal>
    </RD.Root>
  )
}

function NavButton({ side, onClick, label, children }: { side: 'left' | 'right'; onClick: () => void; label: string; children: ReactNode }) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      className={cn(
        'grid size-11 place-items-center rounded-full bg-white/15 text-white hover:bg-white/25 sm:absolute sm:top-1/2 sm:-translate-y-1/2',
        side === 'left' ? 'sm:left-3' : 'sm:right-3',
      )}
    >
      {children}
    </button>
  )
}

/** A single photo that opens the viewer when tapped. */
export function ViewablePhoto({
  url,
  label,
  buttonLabel = `View ${label.toLowerCase()}`,
  className,
  children,
}: {
  url: string
  label: string // shown above the enlarged photo
  buttonLabel?: string
  className?: string
  children: ReactNode
}) {
  const [open, setOpen] = useState(false)
  return (
    <>
      <button
        type="button"
        onClick={() => setOpen(true)}
        aria-label={buttonLabel}
        className={cn('cursor-zoom-in rounded-full focus-visible:ring-2 focus-visible:ring-accent focus-visible:outline-none', className)}
      >
        {children}
      </button>
      <PhotoViewer photos={[{ url, label }]} index={open ? 0 : null} onClose={() => setOpen(false)} />
    </>
  )
}
