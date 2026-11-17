import { KeyRound, LogOut, ShieldCheck, UserRound } from 'lucide-react'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router'
import { homeFor, useLogout, useMe } from '@/features/auth/api'
import { cn } from '@/lib/cn'
import { Button } from './ui/button'

const navLink = ({ isActive }: { isActive: boolean }) =>
  cn(
    'inline-flex min-h-11 items-center gap-2 rounded-lg px-3 text-sm font-medium sm:min-h-9',
    isActive ? 'bg-accent-soft text-accent' : 'text-muted hover:text-fg',
  )

export function AppShell() {
  const { data: me } = useMe()
  const logout = useLogout()
  const navigate = useNavigate()
  const { pathname } = useLocation()
  const inProfile = pathname === '/profile' || pathname === '/history' // history is part of the profile
  const isAdmin = me?.role === 'admin'

  return (
    <div className="min-h-dvh">
      <header className="sticky top-0 z-20 border-b border-border bg-surface/90 backdrop-blur">
        <div className="mx-auto flex h-14 max-w-6xl items-center gap-2 px-4">
          <NavLink to={homeFor(me?.role)} className="mr-2 flex items-center gap-2 font-semibold" aria-label="Employee portal home">
            <span className="grid size-8 place-items-center rounded-lg bg-accent text-accent-fg">
              <UserRound className="size-4" aria-hidden />
            </span>
            <span className="hidden sm:inline">Employee portal</span>
          </NavLink>
          {/* Few items on purpose: History lives under the profile ("View all"). */}
          <nav className="flex items-center gap-1" aria-label="Main">
            {/* Admins have no profile page, so they only see Admin. */}
            {!isAdmin && (
              <NavLink to="/profile" className={({ isActive }) => navLink({ isActive: isActive || inProfile })}>
                Profile
              </NavLink>
            )}
            {isAdmin && (
              <NavLink to="/admin" className={navLink}>
                <ShieldCheck className="size-4" aria-hidden />
                Admin
              </NavLink>
            )}
          </nav>
          <div className="ml-auto flex items-center gap-1">
            <span className="hidden text-sm text-muted md:inline">{me?.employeeCode}</span>
            <Button variant="ghost" size="icon" aria-label="Change password" title="Change password" onClick={() => navigate('/change-password')}>
              <KeyRound className="size-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              aria-label="Sign out"
              title="Sign out"
              loading={logout.isPending}
              onClick={() => logout.mutate(undefined, { onSettled: () => navigate('/login', { replace: true }) })}
            >
              {!logout.isPending && <LogOut className="size-4" />}
            </Button>
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-4 py-6 pb-[max(1.5rem,env(safe-area-inset-bottom))] sm:py-8">
        <Outlet />
      </main>
    </div>
  )
}
