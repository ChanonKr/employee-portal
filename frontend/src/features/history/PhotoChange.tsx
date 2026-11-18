import { ArrowRight, ImageOff } from 'lucide-react'
import { useState } from 'react'
import { PhotoViewer, type ViewerPhoto } from '@/components/ui/photo-viewer'

/** Old → new photo thumbnails; tap one to see it larger (with the other one a swipe away).
 * Old files are deleted by the retention job, so a missing file shows a placeholder. */
export function PhotoChange({ oldUrl, newUrl }: { oldUrl: string | null; newUrl: string | null }) {
  const [missing, setMissing] = useState<Record<string, boolean>>({})
  const [viewing, setViewing] = useState<number | null>(null)

  const shown: ViewerPhoto[] = [
    oldUrl && !missing[oldUrl] ? { url: oldUrl, label: 'Previous photo' } : null,
    newUrl && !missing[newUrl] ? { url: newUrl, label: 'New photo' } : null,
  ].filter((p): p is ViewerPhoto => p !== null)

  const thumb = (url: string, label: string) =>
    missing[url] ? (
      <span
        className="grid size-12 place-items-center rounded-lg border border-dashed border-border text-muted"
        title={`${label} is no longer kept`}
      >
        <ImageOff className="size-4" aria-hidden />
        <span className="sr-only">{label} is no longer kept</span>
      </span>
    ) : (
      <button
        type="button"
        onClick={() => setViewing(shown.findIndex((p) => p.url === url))}
        aria-label={`View ${label.toLowerCase()}`}
        className="cursor-zoom-in rounded-lg focus-visible:ring-2 focus-visible:ring-accent focus-visible:outline-none"
      >
        <img
          src={url}
          alt={label}
          loading="lazy"
          onError={() => setMissing((m) => ({ ...m, [url]: true }))}
          className="size-12 rounded-lg border border-border object-cover"
        />
      </button>
    )

  return (
    <div className="mt-1.5 flex items-center gap-2">
      {oldUrl ? (
        <>
          {thumb(oldUrl, 'Previous photo')}
          <ArrowRight className="size-3.5 text-muted" aria-label="changed to" />
        </>
      ) : (
        <span className="text-sm text-muted">Added</span>
      )}
      {newUrl && thumb(newUrl, 'New photo')}
      <PhotoViewer photos={shown} index={viewing} onClose={() => setViewing(null)} />
    </div>
  )
}
