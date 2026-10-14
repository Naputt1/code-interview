import { createFileRoute } from '@tanstack/react-router'
import { PauseIcon, PlayIcon } from 'lucide-react'

import { useGetBoard, useGetOverview } from '@/api/gen'
import { BoardColumns } from '@/features/board/board-columns'
import { OverviewCards } from '@/features/board/overview-cards'
import { ResourceUsagePanel } from '@/features/board/resource-usage'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { useBoardPrefs } from '@/state/board-prefs'

export const Route = createFileRoute('/board')({
  component: BoardPage,
})

const POLL_MS = 4000

function BoardPage() {
  // Client-only view preference (zustand); the data itself belongs to TanStack
  // Query. `false` turns the interval off without touching the cached data.
  const live = useBoardPrefs((state) => state.live)
  const setLive = useBoardPrefs((state) => state.setLive)
  const refetchInterval = live ? POLL_MS : false

  const board = useGetBoard({ query: { refetchInterval } })
  const overview = useGetOverview({ query: { refetchInterval } })

  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 className="text-2xl font-semibold">Production board</h1>
          <p className="text-muted-foreground text-sm">
            Orders in progress at each step,{' '}
            {live ? `refreshed every ${POLL_MS / 1000}s` : 'auto-refresh paused'}.
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={() => setLive(!live)}>
          {live ? <PauseIcon /> : <PlayIcon />}
          {live ? 'Pause live' : 'Go live'}
        </Button>
      </header>

      {overview.data ? (
        <OverviewCards counts={overview.data.orders} />
      ) : (
        <Skeleton className="h-24 w-full" />
      )}

      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-medium">Work in progress</h2>
        {board.data ? (
          <BoardColumns columns={board.data.columns ?? []} />
        ) : (
          <Skeleton className="h-64 w-full" />
        )}
      </section>

      {overview.data ? (
        <ResourceUsagePanel resources={overview.data.resources ?? []} />
      ) : (
        <Skeleton className="h-48 w-full" />
      )}
    </div>
  )
}
