import { cn } from '@/lib/cn'
import { initials } from '@/lib/format'

export function Avatar({
  url,
  firstName,
  lastName,
  className,
}: {
  url: string | null
  firstName: string
  lastName: string
  className?: string
}) {
  return url ? (
    <img src={url} alt="" className={cn('aspect-square rounded-full object-cover', className)} />
  ) : (
    <div
      className={cn('grid aspect-square place-items-center rounded-full bg-accent-soft font-semibold text-accent', className)}
      aria-hidden
    >
      {initials(firstName, lastName)}
    </div>
  )
}
