import { useQueryClient } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import { toast } from 'sonner'

import type { ServiceBoardColumn } from '@/api/gen'
import { usePostOrdersIdStepsStepIdComplete } from '@/api/gen'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { errorMessage } from '@/lib/errors'
import { formatDuration, formatNumber } from '@/lib/format'

export function BoardColumns({ columns }: { columns: ServiceBoardColumn[] }) {
  const queryClient = useQueryClient()
  const complete = usePostOrdersIdStepsStepIdComplete({
    mutation: {
      onSuccess: () => {
        toast.success('Step completed')
        queryClient.invalidateQueries()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })

  if (columns.length === 0) {
    return (
      <Card className="text-muted-foreground items-center py-12 text-center text-sm">
        No steps defined yet. Add resources and processes in Setup.
      </Card>
    )
  }

  return (
    <div className="flex gap-4 overflow-x-auto pb-2">
      {columns.map((column) => {
        const process = column.process
        const running = column.running ?? []

        return (
          <Card key={process?.id ?? 0} className="w-72 shrink-0 gap-0 py-0">
            <div className="border-b px-4 py-3">
              <div className="flex items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                  <span className="bg-muted text-muted-foreground flex size-5 items-center justify-center rounded text-xs font-semibold tabular-nums">
                    {process?.position}
                  </span>
                  <span className="font-medium">{process?.name}</span>
                </div>
                <Badge variant="outline" className="text-xs">
                  {column.queued ?? 0} queued
                </Badge>
              </div>
              <p className="text-muted-foreground mt-1 text-xs">
                {process?.wipLimit ? `WIP limit ${process.wipLimit}` : 'No WIP limit'}
              </p>
            </div>
            <div className="flex flex-col gap-2 p-3">
              {running.length === 0 ? (
                <p className="text-muted-foreground py-6 text-center text-xs">No orders running</p>
              ) : (
                running.map((step) => (
                  <div key={step.stepId} className="bg-background rounded-md border p-3">
                    <div className="flex items-center justify-between gap-2">
                      <Link
                        to="/orders/$orderId"
                        params={{ orderId: String(step.orderId ?? 0) }}
                        className="text-sm font-medium hover:underline"
                      >
                        {step.reference}
                      </Link>
                      <span className="text-muted-foreground text-xs tabular-nums">
                        ×{formatNumber(step.quantity)}
                      </span>
                    </div>
                    <div className="mt-2 flex items-center justify-between gap-2">
                      <span className="text-muted-foreground text-xs">
                        {formatDuration(step.startedAt)}
                      </span>
                      <Button
                        size="sm"
                        variant="outline"
                        disabled={complete.isPending}
                        onClick={() =>
                          complete.mutate({
                            id: step.orderId ?? 0,
                            stepId: step.stepId ?? 0,
                          })
                        }
                      >
                        Complete
                      </Button>
                    </div>
                  </div>
                ))
              )}
            </div>
          </Card>
        )
      })}
    </div>
  )
}
