import { Link } from '@tanstack/react-router'

import type { ServiceBoardColumn } from '@/api'
import { useCompleteStep, useStartStep } from '@/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { CancelOrderButton } from '@/features/orders/cancel-order-button'
import { formatDuration, formatNumber } from '@/lib/format'

export function BoardColumns({ columns }: { columns: ServiceBoardColumn[] }) {
  const start = useStartStep()
  const complete = useCompleteStep()

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
        const queued = column.queued ?? []
        const wipLimit = process?.wipLimit

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
                  {queued.length} queued
                </Badge>
              </div>
              <p className="text-muted-foreground mt-1 text-xs">
                {wipLimit ? `WIP limit ${wipLimit}` : 'No WIP limit'}
              </p>
            </div>

            <div className="flex max-h-[32rem] flex-col gap-2 overflow-y-auto p-3">
              {running.length === 0 ? (
                <p className="text-muted-foreground py-4 text-center text-xs">Nothing running</p>
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
                      <div className="flex items-center gap-1">
                        <span className="text-muted-foreground text-xs tabular-nums">
                          ×{formatNumber(step.quantity)}
                        </span>
                        <CancelOrderButton
                          orderId={step.orderId ?? 0}
                          reference={step.reference ?? ''}
                          cancellable={step.cancellable ?? false}
                        />
                      </div>
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

              {queued.length > 0 ? (
                <>
                  <p className="text-muted-foreground mt-1 text-xs font-medium tracking-wide uppercase">
                    Queued
                  </p>
                  {queued.map((step, index) => (
                    <div
                      key={step.stepId}
                      className="bg-muted/40 rounded-md border border-dashed p-3"
                    >
                      <div className="flex items-center justify-between gap-2">
                        <div className="flex min-w-0 items-center gap-2">
                          <span className="text-muted-foreground text-xs tabular-nums">
                            #{index + 1}
                          </span>
                          <Link
                            to="/orders/$orderId"
                            params={{ orderId: String(step.orderId ?? 0) }}
                            className="truncate text-sm font-medium hover:underline"
                          >
                            {step.reference}
                          </Link>
                        </div>
                        <div className="flex items-center gap-1">
                          <span className="text-muted-foreground text-xs tabular-nums">
                            ×{formatNumber(step.quantity)}
                          </span>
                          <CancelOrderButton
                            orderId={step.orderId ?? 0}
                            reference={step.reference ?? ''}
                            cancellable={step.cancellable ?? false}
                          />
                        </div>
                      </div>
                      <div className="mt-2 flex items-center justify-between gap-2">
                        <span className="text-muted-foreground text-xs">
                          {step.startable ? 'Ready to start' : 'Waiting for a free slot'}
                        </span>
                        <Button
                          size="sm"
                          variant="outline"
                          disabled={!step.startable || start.isPending}
                          onClick={() =>
                            start.mutate({
                              id: step.orderId ?? 0,
                              stepId: step.stepId ?? 0,
                            })
                          }
                        >
                          Start
                        </Button>
                      </div>
                    </div>
                  ))}
                </>
              ) : null}
            </div>
          </Card>
        )
      })}
    </div>
  )
}
