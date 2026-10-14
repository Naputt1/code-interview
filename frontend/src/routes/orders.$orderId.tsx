import { useQueryClient } from '@tanstack/react-query'
import { Link, createFileRoute } from '@tanstack/react-router'
import { toast } from 'sonner'

import { useGetOrdersId, useGetResources, usePostOrdersIdCancel } from '@/api/gen'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { OrderStepList } from '@/features/orders/order-step-list'
import { StatusBadge } from '@/features/orders/status-badge'
import { errorMessage } from '@/lib/errors'
import { formatNumber } from '@/lib/format'

export const Route = createFileRoute('/orders/$orderId')({
  component: OrderDetailPage,
})

function OrderDetailPage() {
  const { orderId } = Route.useParams()
  const id = Number(orderId)
  const queryClient = useQueryClient()

  const order = useGetOrdersId(id, { query: { refetchInterval: 4000 } })
  const resources = useGetResources()

  const cancel = usePostOrdersIdCancel({
    mutation: {
      onSuccess: () => {
        toast.success('Order cancelled')
        queryClient.invalidateQueries()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })

  if (order.isLoading) {
    return <Skeleton className="h-64 w-full" />
  }

  if (order.isError || !order.data) {
    return (
      <div className="flex flex-col gap-4">
        <Link to="/orders" className="text-muted-foreground text-sm hover:underline">
          ← Back to orders
        </Link>
        <p className="text-destructive text-sm">Order not found.</p>
      </div>
    )
  }

  const data = order.data
  const steps = data.steps ?? []
  const doneCount = steps.filter((step) => step.status === 'done').length
  const cancellable =
    data.status !== 'done' &&
    data.status !== 'cancelled' &&
    !steps.some((step) => step.status === 'running')

  const resourcesById = new Map((resources.data ?? []).map((resource) => [resource.id ?? 0, resource]))

  return (
    <div className="flex flex-col gap-6">
      <Link to="/orders" className="text-muted-foreground text-sm hover:underline">
        ← Back to orders
      </Link>

      <header className="flex flex-wrap items-start justify-between gap-4">
        <div className="flex flex-col gap-1">
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-semibold">{data.reference}</h1>
            <StatusBadge status={data.status} />
          </div>
          <p className="text-muted-foreground text-sm">
            Quantity {formatNumber(data.quantity)} · {doneCount}/{steps.length} steps complete
          </p>
        </div>

        <ConfirmDialog
          trigger={
            <Button variant="outline" disabled={!cancellable}>
              Cancel order
            </Button>
          }
          title={`Cancel ${data.reference}?`}
          description="The order will be stopped and no further steps can run. This cannot be undone."
          confirmLabel="Cancel order"
          onConfirm={() => cancel.mutate({ id })}
          pending={cancel.isPending}
        />
      </header>

      <section className="flex flex-col gap-3">
        <h2 className="text-lg font-medium">Steps</h2>
        <OrderStepList order={data} resourcesById={resourcesById} />
      </section>
    </div>
  )
}
