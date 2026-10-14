import { useState } from 'react'
import type { FormEvent } from 'react'
import { Plus } from 'lucide-react'
import { toast } from 'sonner'

import { useCreateOrder, useProcesses } from '@/api'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

export function CreateOrderDialog() {
  const processes = useProcesses()
  const [open, setOpen] = useState(false)
  const [reference, setReference] = useState('')
  const [quantity, setQuantity] = useState('1')

  const create = useCreateOrder({
    onSuccess: () => {
      setOpen(false)
      setReference('')
      setQuantity('1')
    },
  })

  const processCount = processes.data?.length ?? 0

  function submit(event: FormEvent) {
    event.preventDefault()
    const parsed = Number(quantity)
    if (!reference.trim() || !Number.isFinite(parsed) || parsed <= 0) {
      toast.error('Enter a reference and a quantity greater than zero')
      return
    }
    create.mutate({ data: { reference: reference.trim(), quantity: parsed } })
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button disabled={processCount === 0}>
          <Plus /> New order
        </Button>
      </DialogTrigger>
      <DialogContent>
        <form onSubmit={submit} className="flex flex-col gap-4">
          <DialogHeader>
            <DialogTitle>Create order</DialogTitle>
            <DialogDescription>
              Snapshots the current {processCount}-step sequence onto the new
              order.
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-2">
            <Label htmlFor="order-reference">Reference</Label>
            <Input
              id="order-reference"
              value={reference}
              onChange={(event) => setReference(event.target.value)}
              placeholder="BRK-1004"
            />
          </div>
          <div className="flex flex-col gap-2">
            <Label htmlFor="order-quantity">Quantity</Label>
            <Input
              id="order-quantity"
              type="number"
              min="1"
              step="1"
              value={quantity}
              onChange={(event) => setQuantity(event.target.value)}
            />
          </div>
          <DialogFooter>
            <Button type="submit" disabled={create.isPending}>
              Create order
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
