import { RefreshCw } from 'lucide-react'
import { ErrorAlert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useProfile } from './api'
import { ContactCard } from './ContactCard'
import { ProfileChecklist } from './ProfileChecklist'
import { ProfileHeader } from './ProfileHeader'
import { RecentChangesCard } from './RecentChangesCard'
import { SkillsCard } from './SkillsCard'

export function ProfilePage() {
  const profile = useProfile()

  if (profile.isPending) {
    return (
      <div className="grid gap-4 lg:grid-cols-[20rem_1fr]" aria-busy>
        <Skeleton className="h-72" />
        <Skeleton className="h-72" />
      </div>
    )
  }
  if (profile.isError) {
    return (
      <div className="space-y-3">
        <ErrorAlert>Could not load your profile. {profile.error.message}</ErrorAlert>
        <Button variant="secondary" onClick={() => profile.refetch()}>
          <RefreshCw className="size-4" /> Try again
        </Button>
      </div>
    )
  }

  const p = profile.data
  // Phones: header → contact → skills → recent changes (DOM order).
  // Desktop: left column header + skills, right column contact + recent changes.
  return (
    <div className="space-y-4">
      <ProfileChecklist profile={p} />
      <div className="grid gap-4 lg:grid-cols-[20rem_1fr] lg:items-start">
        <div className="lg:col-start-1 lg:row-start-1">
          <ProfileHeader profile={p} />
        </div>
        <div className="lg:col-start-2 lg:row-start-1">
          <ContactCard profile={p} />
        </div>
        <div className="lg:col-start-1 lg:row-start-2">
          <SkillsCard profile={p} />
        </div>
        <div className="lg:col-start-2 lg:row-start-2">
          <RecentChangesCard />
        </div>
      </div>
    </div>
  )
}
