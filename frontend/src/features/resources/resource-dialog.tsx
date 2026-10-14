import { useState } from 'react'
import type { FormEvent } from 'react'
import { toast } from 'sonner'

import type { DomainResource, ServiceResourceInput } from '@/api'
import { useCreateResource, useUpdateResource } from '@/api'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

const KINDS = ['material', 'machine', 'labour'] as const

export function ResourceDialog({
  resource,
  onOpenChange,
}: {
  resource: DomainResource | null
  onOpenChange: (open: boolean) => void
}) {
  const [name, setName] = useState(resource?.name ?? '')
  const [kind, setKind] = useState<string>(resource?.kind ?? 'material')
  const [unit, setUnit] = useState(resource?.unit ?? '')
  const [capacity, setCapacity] = useState(
    resource?.capacity === undefined || resource?.capacity === null
      ? ''
      : String(resource.capacity),
  )

  const isEdit = Boolean(resource?.id)

  const create = useCreateResource({ onSuccess: () => onOpenChange(false) })
  const update = useUpdateResource({ onSuccess: () => onOpenChange(false) })

  function submit(event: FormEvent) {
    event.preventDefault()

    if (!name.trim() || !unit.trim()) {
      toast.error('Name and unit are required')
      return
    }

    const parsedCapacity = capacity.trim() === '' ? undefined : Number(capacity)
    if (parsedCapacity !== undefined && (!Number.isFinite(parsedCapacity) || parsedCapacity < 0)) {
      toast.error('Capacity must be zero or more, or blank for unlimited')
      return
    }

    const payload: ServiceResourceInput = {
      name: name.trim(),
      kind,
      unit: unit.trim(),
      capacity: parsedCapacity,
    }

    if (isEdit && resource?.id) {
      update.mutate({ id: resource.id, data: payload })
    } else {
      create.mutate({ data: payload })
    }
  }

  const pending = create.isPending || update.isPending

  return (
    <Dialog open onOpenChange={onOpenChange}>
      <DialogContent>
        <form onSubmit={submit} className="flex flex-col gap-4">
          <DialogHeader>
            <DialogTitle>{isEdit ? 'Edit resource' : 'New resource'}</DialogTitle>
            <DialogDescription>
              Resources are materials, machines or labour that steps consume.
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-2">
            <Label htmlFor="resource-name">Name</Label>
            <Input
              id="resource-name"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="Aluminium Sheet"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="flex flex-col gap-2">
              <Label>Kind</Label>
              <Select value={kind} onValueChange={setKind}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {KINDS.map((option) => (
                    <SelectItem key={option} value={option} className="capitalize">
                      {option}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="resource-unit">Unit</Label>
              <Input
                id="resource-unit"
                value={unit}
                onChange={(event) => setUnit(event.target.value)}
                placeholder="kg"
              />
            </div>
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="resource-capacity">Capacity</Label>
            <Input
              id="resource-capacity"
              type="number"
              min="0"
              step="any"
              value={capacity}
              onChange={(event) => setCapacity(event.target.value)}
              placeholder="Blank for unlimited"
            />
            <p className="text-muted-foreground text-xs">
              Leave blank for unlimited capacity (typical for labour).
            </p>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={pending}>
              {isEdit ? 'Save changes' : 'Create resource'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
