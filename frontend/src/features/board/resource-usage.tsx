import type { ServiceResourceUsage } from '@/api/gen'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { formatNumber } from '@/lib/format'

export function ResourceUsagePanel({ resources }: { resources: ServiceResourceUsage[] }) {
  return (
    <Card className="gap-3">
      <CardHeader>
        <CardTitle>Resource consumption</CardTitle>
        <CardDescription>Recorded when a step starts, across all orders.</CardDescription>
      </CardHeader>
      <CardContent className="px-0">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="pl-6">Resource</TableHead>
              <TableHead>Kind</TableHead>
              <TableHead className="text-right">Consumed</TableHead>
              <TableHead className="text-right">Capacity</TableHead>
              <TableHead className="pr-6">Utilisation</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {resources.map((usage) => {
              const resource = usage.resource
              const consumed = usage.consumed ?? 0
              const capacity = resource?.capacity ?? null
              const pct =
                capacity && capacity > 0 ? Math.min(100, Math.round((consumed / capacity) * 100)) : null

              return (
                <TableRow key={resource?.id ?? 0}>
                  <TableCell className="pl-6 font-medium">{resource?.name}</TableCell>
                  <TableCell className="text-muted-foreground capitalize">{resource?.kind}</TableCell>
                  <TableCell className="text-right tabular-nums">
                    {formatNumber(consumed)} {resource?.unit}
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {capacity === null
                      ? 'Unlimited'
                      : `${formatNumber(capacity)} ${resource?.unit}`}
                  </TableCell>
                  <TableCell className="pr-6">
                    {pct === null ? (
                      <span className="text-muted-foreground text-xs">—</span>
                    ) : (
                      <div className="flex items-center gap-2">
                        <div className="bg-muted h-2 w-24 overflow-hidden rounded-full">
                          <div className="bg-primary h-full rounded-full" style={{ width: `${pct}%` }} />
                        </div>
                        <span className="text-muted-foreground text-xs tabular-nums">{pct}%</span>
                      </div>
                    )}
                  </TableCell>
                </TableRow>
              )
            })}
            {resources.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-muted-foreground py-6 text-center">
                  No resources defined yet.
                </TableCell>
              </TableRow>
            ) : null}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
