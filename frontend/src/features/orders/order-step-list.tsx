import type { DomainOrder, DomainResource } from '@/api'
import { useCompleteStep, useStartStep } from '@/api'
import { Button } from '@/components/ui/button'
import { formatDateTime, formatDuration, formatNumber } from '@/lib/format'

import { StatusBadge } from './status-badge'

export function OrderStepList({
  order,
  resourcesById,
}: {
  order: DomainOrder
  resourcesById: Map<number, DomainResource>
}) {
  const start = useStartStep()
  const complete = useCompleteStep()

  const orderId = order.id ?? 0
  const steps = [...(order.steps ?? [])].sort(
    (a, b) => (a.position ?? 0) - (b.position ?? 0),
  )

  return (
    <ol className="flex flex-col gap-3">
      {steps.map((step, index) => {
        const previousDone = steps
          .slice(0, index)
          .every((previous) => previous.status === 'done')
        const canStart =
          step.status === 'queued' &&
          previousDone &&
          order.status !== 'cancelled' &&
          order.status !== 'done'
        const canComplete = step.status === 'running'
        const inputs = step.process?.inputs ?? []

        return (
          <li key={step.id} className="bg-background rounded-lg border p-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <span className="bg-muted text-muted-foreground flex size-6 items-center justify-center rounded text-xs font-semibold tabular-nums">
                  {index + 1}
                </span>
                <div>
                  <div className="font-medium">
                    {step.process?.name ?? 'Step'}
                  </div>
                  <div className="text-muted-foreground text-xs">
                    {step.status === 'running'
                      ? `Running for ${formatDuration(step.startedAt)}`
                      : step.status === 'done'
                        ? `Finished ${formatDateTime(step.finishedAt)}`
                        : 'Waiting in queue'}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <StatusBadge status={step.status} kind="step" />
                {canStart ? (
                  <Button
                    size="sm"
                    disabled={start.isPending}
                    onClick={() =>
                      start.mutate({ id: orderId, stepId: step.id ?? 0 })
                    }
                  >
                    Start
                  </Button>
                ) : null}
                {canComplete ? (
                  <Button
                    size="sm"
                    variant="outline"
                    disabled={complete.isPending}
                    onClick={() =>
                      complete.mutate({ id: orderId, stepId: step.id ?? 0 })
                    }
                  >
                    Complete
                  </Button>
                ) : null}
              </div>
            </div>

            <div className="text-muted-foreground mt-3 text-xs">
              Requires{' '}
              {inputs.length === 0
                ? 'no resources'
                : inputs
                    .map((input) =>
                      `${input.resource?.name ?? `#${input.resourceId}`} ×${formatNumber(input.quantity)} ${input.resource?.unit ?? ''}`.trim(),
                    )
                    .join(', ')}
            </div>

            {(step.consumptions ?? []).length > 0 ? (
              <ul className="mt-3 flex flex-wrap gap-2">
                {(step.consumptions ?? []).map((consumption) => {
                  const resource = resourcesById.get(
                    consumption.resourceId ?? 0,
                  )
                  return (
                    <li
                      key={consumption.id}
                      className="bg-muted text-muted-foreground rounded px-2 py-1 text-xs tabular-nums"
                    >
                      {resource?.name ?? `#${consumption.resourceId}`} ×
                      {formatNumber(consumption.quantity)} {resource?.unit}
                    </li>
                  )
                })}
              </ul>
            ) : null}
          </li>
        )
      })}
    </ol>
  )
}
