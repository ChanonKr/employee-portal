import { ArrowRight, ImageOff } from 'lucide-react'
import { useState } from 'react'

/** Old → new photo thumbnails. Old files are deleted by the retention job, so a
 * missing file shows a placeholder instead of a broken image. */
export function PhotoChange({ oldUrl, newUrl }: { oldUrl: string | null; newUrl: string | null }) {
  return (
    <div className="mt-1.5 flex items-center gap-2">
      {oldUrl ? (
        <>
          <Thumb url={oldUrl} label="Previous photo" />
          <ArrowRight className="size-3.5 text-muted" aria-label="changed to" />
        </>
      ) : (
        <span className="text-sm text-muted">Added</span>
      )}
      {newUrl && <Thumb url={newUrl} label="New photo" />}
    </div>
  )
}

function Thumb({ url, label }: { url: string; label: string }) {
  const [missing, setMissing] = useState(false)
  if (missing) {
    return (
      <span
        className="grid size-12 place-items-center rounded-lg border border-dashed border-border text-muted"
        title={`${label} is no longer kept`}
      >
        <ImageOff className="size-4" aria-hidden />
        <span className="sr-only">{label} is no longer kept</span>
      </span>
    )
  }
  return (
    <img
      src={url}
      alt={label}
      loading="lazy"
      onError={() => setMissing(true)}
      className="size-12 rounded-lg border border-border object-cover"
    />
  )
}
