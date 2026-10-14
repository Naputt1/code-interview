import { XIcon } from 'lucide-react'

import { useCancelOrder } from '@/api'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { Button } from '@/components/ui/button'

/**
 * Cancel an order, from whichever screen lists it. An order may only be
 * cancelled before any of its steps has started, so the button is disabled with
 * the reason rather than letting the request fail.
 */
export function CancelOrderButton({
  orderId,
  reference,
  cancellable,
}: {
  orderId: number
  reference: string
  cancellable: boolean
}) {
  const cancel = useCancelOrder()

  return (
    <ConfirmDialog
      trigger={
        <Button
          size="icon"
          variant="ghost"
          className="size-7"
          aria-label={`Cancel ${reference}`}
          title={
            cancellable
              ? 'Cancel order'
              : 'Only an order that has not started can be cancelled'
          }
          disabled={!cancellable}
        >
          <XIcon className="size-4" />
        </Button>
      }
      title={`Cancel ${reference}?`}
      description="The order will be stopped and no further steps can run. This cannot be undone."
      confirmLabel="Cancel order"
      onConfirm={() => cancel.mutate({ id: orderId })}
      pending={cancel.isPending}
    />
  )
}
