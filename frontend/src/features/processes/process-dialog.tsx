import { useState } from 'react'
import type { FormEvent } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'

import type { DomainProcess, ServiceProcessInput } from '@/api/gen'
import { useGetResources, usePostProcesses, usePutProcessesId } from '@/api/gen'
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
import { errorMessage } from '@/lib/errors'

type InputRow = { resourceId: number; quantity: string }

export function ProcessDialog({
  process,
  onOpenChange,
}: {
  process: DomainProcess | null
  onOpenChange: (open: boolean) => void
}) {
  const queryClient = useQueryClient()
  const resources = useGetResources()
  const [name, setName] = useState(process?.name ?? '')
  const [wipLimit, setWipLimit] = useState(
    process?.wipLimit === undefined || process?.wipLimit === null ? '' : String(process.wipLimit),
  )
  const [rows, setRows] = useState<InputRow[]>(
    (process?.inputs ?? []).map((input) => ({
      resourceId: input.resourceId ?? 0,
      quantity: String(input.quantity ?? ''),
    })),
  )

  const isEdit = Boolean(process?.id)

  const create = usePostProcesses({
    mutation: {
      onSuccess: () => {
        toast.success('Process created')
        queryClient.invalidateQueries()
        onOpenChange(false)
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })

  const update = usePutProcessesId({
    mutation: {
      onSuccess: () => {
        toast.success('Process updated')
        queryClient.invalidateQueries()
        onOpenChange(false)
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })

  const resourceOptions = resources.data ?? []

  function addRow() {
    const firstResource = resourceOptions[0]
    if (!firstResource) {
      toast.error('Create a resource first')
      return
    }
    setRows((current) => [...current, { resourceId: firstResource.id ?? 0, quantity: '1' }])
  }

  function submit(event: FormEvent) {
    event.preventDefault()

    if (!name.trim()) {
      toast.error('Name is required')
      return
    }

    const parsedWip = wipLimit.trim() === '' ? undefined : Number(wipLimit)
    if (parsedWip !== undefined && (!Number.isInteger(parsedWip) || parsedWip < 1)) {
      toast.error('WIP limit must be a whole number of one or more, or blank')
      return
    }

    const inputs = rows.map((row) => ({
      resourceId: row.resourceId,
      quantity: Number(row.quantity),
    }))

    if (inputs.some((input) => !input.resourceId || !Number.isFinite(input.quantity) || input.quantity <= 0)) {
      toast.error('Each input needs a resource and a quantity greater than zero')
      return
    }

    const payload: ServiceProcessInput = {
      name: name.trim(),
      wipLimit: parsedWip,
      inputs,
      // Keep the step's place in the sequence on edit; new steps append.
      position: isEdit ? process?.position : undefined,
    }

    if (isEdit && process?.id) {
      update.mutate({ id: process.id, data: payload })
    } else {
      create.mutate({ data: payload })
    }
  }

  const pending = create.isPending || update.isPending

  return (
    <Dialog open onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <form onSubmit={submit} className="flex flex-col gap-4">
          <DialogHeader>
            <DialogTitle>{isEdit ? 'Edit process' : 'New process'}</DialogTitle>
            <DialogDescription>
              A step in the sequence. Steps run in the order they are created.
            </DialogDescription>
          </DialogHeader>

          <div className="grid grid-cols-2 gap-4">
            <div className="flex flex-col gap-2">
              <Label htmlFor="process-name">Name</Label>
              <Input
                id="process-name"
                value={name}
                onChange={(event) => setName(event.target.value)}
                placeholder="Cut"
              />
            </div>
            <div className="flex flex-col gap-2">
              <Label htmlFor="process-wip">WIP limit</Label>
              <Input
                id="process-wip"
                type="number"
                min="1"
                step="1"
                value={wipLimit}
                onChange={(event) => setWipLimit(event.target.value)}
                placeholder="Blank for unlimited"
              />
            </div>
          </div>

          <div className="flex flex-col gap-3">
            <div className="flex items-center justify-between">
              <Label>Resources consumed per unit</Label>
              <Button type="button" size="sm" variant="outline" onClick={addRow}>
                <Plus /> Add resource
              </Button>
            </div>

            {rows.length === 0 ? (
              <p className="text-muted-foreground text-xs">
                No resources on this step. It will not consume anything.
              </p>
            ) : (
              rows.map((row, index) => (
                <div key={index} className="flex items-center gap-2">
                  <Select
                    value={row.resourceId ? String(row.resourceId) : undefined}
                    onValueChange={(value) =>
                      setRows((current) =>
                        current.map((item, i) =>
                          i === index ? { ...item, resourceId: Number(value) } : item,
                        ),
                      )
                    }
                  >
                    <SelectTrigger className="flex-1">
                      <SelectValue placeholder="Select resource" />
                    </SelectTrigger>
                    <SelectContent>
                      {resourceOptions.map((resource) => (
                        <SelectItem key={resource.id} value={String(resource.id)}>
                          {resource.name} ({resource.unit})
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <Input
                    className="w-24"
                    type="number"
                    min="0"
                    step="any"
                    value={row.quantity}
                    onChange={(event) =>
                      setRows((current) =>
                        current.map((item, i) =>
                          i === index ? { ...item, quantity: event.target.value } : item,
                        ),
                      )
                    }
                  />
                  <Button
                    type="button"
                    size="icon"
                    variant="ghost"
                    aria-label="Remove resource"
                    onClick={() => setRows((current) => current.filter((_, i) => i !== index))}
                  >
                    <Trash2 />
                  </Button>
                </div>
              ))
            )}
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={pending}>
              {isEdit ? 'Save changes' : 'Create process'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
