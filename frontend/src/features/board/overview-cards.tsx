import type { ServiceOrderCounts } from '@/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function OverviewCards({ counts }: { counts?: ServiceOrderCounts }) {
  const items = [
    { label: 'Total', value: counts?.total ?? 0 },
    { label: 'Pending', value: counts?.pending ?? 0 },
    { label: 'In progress', value: counts?.inProgress ?? 0 },
    { label: 'Done', value: counts?.done ?? 0 },
    { label: 'Cancelled', value: counts?.cancelled ?? 0 },
  ]

  return (
    <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
      {items.map((item) => (
        <Card key={item.label} className="gap-2 py-4">
          <CardHeader className="px-4">
            <CardTitle className="text-muted-foreground text-xs font-medium tracking-wide uppercase">
              {item.label}
            </CardTitle>
          </CardHeader>
          <CardContent className="px-4">
            <div className="text-2xl font-semibold tabular-nums">{item.value}</div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
