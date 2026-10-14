import { Link, createFileRoute } from '@tanstack/react-router'

import { useOrders } from '@/api'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { CancelOrderButton } from '@/features/orders/cancel-order-button'
import { CreateOrderDialog } from '@/features/orders/create-order-dialog'
import { StatusBadge } from '@/features/orders/status-badge'
import { formatNumber } from '@/lib/format'

export const Route = createFileRoute('/orders/')({
  component: OrdersPage,
})

function OrdersPage() {
  const orders = useOrders(true)
  const rows = orders.data ?? []

  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">Orders</h1>
          <p className="text-muted-foreground text-sm">
            Every order and its progress through the defined steps.
          </p>
        </div>
        <CreateOrderDialog />
      </header>

      {orders.isLoading ? (
        <Skeleton className="h-64 w-full" />
      ) : (
        <div className="bg-background rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="pl-6">Reference</TableHead>
                <TableHead className="text-right">Quantity</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Progress</TableHead>
                <TableHead>Current step</TableHead>
                <TableHead className="pr-6 text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((order) => {
                const steps = order.steps ?? []
                const done = steps.filter(
                  (step) => step.status === 'done',
                ).length
                const current = steps.find((step) => step.status === 'running')

                return (
                  <TableRow key={order.id}>
                    <TableCell className="pl-6 font-medium">
                      <Link
                        to="/orders/$orderId"
                        params={{ orderId: String(order.id) }}
                        className="hover:underline"
                      >
                        {order.reference}
                      </Link>
                    </TableCell>
                    <TableCell className="text-right tabular-nums">
                      {formatNumber(order.quantity)}
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={order.status} />
                    </TableCell>
                    <TableCell className="text-muted-foreground tabular-nums">
                      {done}/{steps.length}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {current?.process?.name ?? '—'}
                    </TableCell>
                    <TableCell className="pr-6 text-right">
                      <CancelOrderButton
                        orderId={order.id ?? 0}
                        reference={order.reference ?? ''}
                        cancellable={order.status === 'pending'}
                      />
                    </TableCell>
                  </TableRow>
                )
              })}
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={6}
                    className="text-muted-foreground py-8 text-center"
                  >
                    No orders yet. Create one to start tracking.
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
