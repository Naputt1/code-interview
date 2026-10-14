import { TanStackDevtools } from '@tanstack/react-devtools'
import { ReactQueryDevtoolsPanel } from '@tanstack/react-query-devtools'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'

/**
 * The one place the devtools are configured, so adding another plugin means
 * editing this file rather than hunting for a second floating button.
 *
 * Mounted from the root route because that is the only place inside both the
 * QueryClient provider (set up in main.tsx) and the router context.
 */
export function AppDevtools() {
  if (!import.meta.env.DEV) return null

  return (
    <TanStackDevtools
      plugins={[
        {
          id: 'tanstack-query',
          name: 'Query',
          render: <ReactQueryDevtoolsPanel />,
        },
        {
          id: 'tanstack-router',
          name: 'Router',
          render: <TanStackRouterDevtoolsPanel />,
        },
      ]}
    />
  )
}
