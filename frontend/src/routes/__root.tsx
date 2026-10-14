import { Link, Outlet, createRootRoute } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools'
import {
  ClipboardList,
  Factory,
  LayoutDashboard,
  Settings2,
} from 'lucide-react'

import { cn } from '@/lib/utils'

export const Route = createRootRoute({
  component: RootLayout,
})

const navItems = [
  { to: '/board', label: 'Board', icon: LayoutDashboard },
  { to: '/orders', label: 'Orders', icon: ClipboardList },
  { to: '/setup', label: 'Setup', icon: Settings2 },
] as const

function RootLayout() {
  return (
    <div className="bg-muted/30 flex min-h-svh flex-col md:flex-row">
      <aside className="border-border bg-background md:sticky md:top-0 md:h-svh md:w-60 md:shrink-0 md:border-r">
        <div className="flex items-center gap-2 px-5 py-4">
          <Factory className="size-6" />
          <div className="leading-tight">
            <div className="text-sm font-semibold">Manufacturing Tracker</div>
            <div className="text-muted-foreground text-xs">Prototype</div>
          </div>
        </div>
        <nav className="flex gap-1 overflow-x-auto px-3 pb-3 md:flex-col md:overflow-visible">
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              activeProps={{ 'data-status': 'active' }}
              className={cn(
                'text-muted-foreground flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium whitespace-nowrap transition-colors',
                'hover:bg-accent hover:text-accent-foreground',
                'data-[status=active]:bg-accent data-[status=active]:text-accent-foreground',
              )}
            >
              <item.icon className="size-4" />
              {item.label}
            </Link>
          ))}
        </nav>
      </aside>
      <main className="min-w-0 flex-1 px-4 py-6 md:px-8">
        <Outlet />
      </main>
      <TanStackRouterDevtools position="bottom-right" />
    </div>
  )
}
