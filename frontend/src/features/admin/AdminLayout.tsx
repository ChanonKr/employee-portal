import { LayoutDashboard, Settings, Sparkles, Users, type LucideIcon } from 'lucide-react'
import { NavLink, Outlet } from 'react-router'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/cn'

const items: { to: string; label: string; icon: LucideIcon; end?: boolean }[] = [
  { to: '/admin', label: 'Dashboard', icon: LayoutDashboard, end: true },
  { to: '/admin/employees', label: 'Employees', icon: Users },
  { to: '/admin/skills', label: 'Skills', icon: Sparkles },
  { to: '/admin/settings', label: 'Settings', icon: Settings },
]

/** Sidebar on desktop; a scrollable tab bar on smaller screens. */
export function AdminLayout() {
  return (
    <div className="lg:grid lg:grid-cols-[13rem_1fr] lg:gap-8">
      <aside className="mb-4 lg:mb-0">
        <Badge tone="accent" className="mb-3 hidden lg:inline-flex">
          Admin
        </Badge>
        <nav aria-label="Admin" className="-mx-4 overflow-x-auto px-4 lg:mx-0 lg:px-0">
          <ul className="flex w-max gap-1 lg:w-auto lg:flex-col">
            {items.map(({ to, label, icon: Icon, end }) => (
              <li key={to}>
                <NavLink
                  to={to}
                  end={end}
                  className={({ isActive }) =>
                    cn(
                      'flex min-h-11 items-center gap-2 rounded-lg px-3 text-sm font-medium whitespace-nowrap sm:min-h-9',
                      isActive ? 'bg-accent-soft text-accent' : 'text-muted hover:bg-surface hover:text-fg',
                    )
                  }
                >
                  <Icon className="size-4" aria-hidden />
                  {label}
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>
      </aside>
      <div className="min-w-0">
        <Outlet />
      </div>
    </div>
  )
}

export function PageTitle({ title, children }: { title: string; children?: React.ReactNode }) {
  return (
    <div className="mb-4 flex flex-wrap items-end justify-between gap-3">
      <h1 className="text-2xl font-semibold">{title}</h1>
      {children}
    </div>
  )
}
