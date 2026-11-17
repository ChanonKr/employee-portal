import { Camera, Loader2, Lock } from 'lucide-react'
import { useRef, useState, type DragEvent } from 'react'
import { toast } from 'sonner'
import { errorMessage } from '@/api/client'
import type { Profile } from '@/api/types'
import { Avatar } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { cn } from '@/lib/cn'
import { useUploadPhoto } from './api'

const MAX_BYTES = 5 * 1024 * 1024
const TYPES = ['image/jpeg', 'image/png', 'image/webp']

/** Quick client-side check; the server checks the real file type again. */
export function checkPhoto(file: File): string | null {
  if (!TYPES.includes(file.type)) return 'Choose a JPEG, PNG or WebP image.'
  if (file.size > MAX_BYTES) return 'The photo must be 5 MB or smaller.'
  return null
}

export function ProfileHeader({ profile }: { profile: Profile }) {
  const input = useRef<HTMLInputElement>(null)
  const upload = useUploadPhoto()
  const [dragging, setDragging] = useState(false)

  const send = (file: File | undefined) => {
    if (!file) return
    const problem = checkPhoto(file)
    if (problem) {
      toast.error(problem)
      return
    }
    upload.mutate(file, {
      onSuccess: () => toast.success('Photo updated'),
      onError: (e) => toast.error(errorMessage(e)),
    })
  }

  const onDrop = (e: DragEvent) => {
    e.preventDefault()
    setDragging(false)
    send(e.dataTransfer.files[0])
  }

  const hasPhoto = !!profile.photoUrl

  return (
    <Card className="flex flex-col items-center gap-4 text-center">
      <div
        className={cn('relative rounded-full', dragging && 'ring-4 ring-accent/40')}
        onDragOver={(e) => {
          e.preventDefault()
          setDragging(true)
        }}
        onDragLeave={() => setDragging(false)}
        onDrop={onDrop}
      >
        {hasPhoto ? (
          <Avatar url={profile.photoUrl} firstName={profile.firstName} lastName={profile.lastName} className="size-28" />
        ) : (
          <button
            type="button"
            onClick={() => input.current?.click()}
            className="grid size-28 place-items-center rounded-full border-2 border-dashed border-border text-sm text-muted hover:border-accent hover:text-accent"
          >
            <span className="flex flex-col items-center gap-1">
              <Camera className="size-5" aria-hidden />
              Add photo
            </span>
          </button>
        )}
        {upload.isPending && (
          <div className="absolute inset-0 grid place-items-center rounded-full bg-black/40" role="status" aria-label="Uploading photo">
            <Loader2 className="size-6 animate-spin text-white" />
          </div>
        )}
        {hasPhoto && (
          <Button
            variant="secondary"
            size="icon"
            className="absolute right-0 bottom-0 rounded-full shadow"
            aria-label="Change photo"
            onClick={() => input.current?.click()}
            disabled={upload.isPending}
          >
            <Camera className="size-4" />
          </Button>
        )}
        <input
          ref={input}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          className="sr-only"
          tabIndex={-1}
          aria-label="Profile photo"
          onChange={(e) => {
            send(e.target.files?.[0])
            e.target.value = '' // allow choosing the same file again
          }}
        />
      </div>

      <div>
        <h1 className="flex items-center justify-center gap-1.5 text-xl font-semibold break-words">
          {profile.firstName} {profile.lastName}
          <Lock className="size-4 text-muted" aria-label="Read-only" />
        </h1>
        <p className="mt-0.5 text-sm text-muted">{profile.employeeCode}</p>
        <p className="mt-2 text-xs text-muted">Your name comes from HR records. Contact HR to change it.</p>
      </div>
      <p className="hidden text-xs text-muted md:block">Drag a photo onto the picture to upload. JPEG, PNG or WebP, up to 5 MB.</p>
    </Card>
  )
}
