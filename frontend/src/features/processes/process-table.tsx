import { useState } from 'react'
import { Pencil, Plus, Trash2 } from 'lucide-react'

import type { DomainProcess } from '@/api'
import { useDeleteProcess, useProcesses } from '@/api'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { formatNumber } from '@/lib/format'

import { ProcessDialog } from './process-dialog'

function inputSummary(process: DomainProcess): string {
  const inputs = process.inputs ?? []
  if (inputs.length === 0) return '—'
  return inputs
    .map((input) => `${input.resource?.name ?? `#${input.resourceId}`} ×${formatNumber(input.quantity)}`)
    .join(', ')
}

export function ProcessTable() {
  const processes = useProcesses()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<DomainProcess | null>(null)

  const remove = useDeleteProcess()

  const rows = processes.data ?? []

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-muted-foreground text-sm">
          Steps run in the order below, one after another.
        </p>
        <Button
          onClick={() => {
            setEditing(null)
            setDialogOpen(true)
          }}
        >
          <Plus /> New process
        </Button>
      </div>

      {processes.isLoading ? (
        <Skeleton className="h-48 w-full" />
      ) : (
        <div className="bg-background rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="pl-6 w-12">#</TableHead>
                <TableHead>Process</TableHead>
                <TableHead className="text-right">WIP limit</TableHead>
                <TableHead>Consumes</TableHead>
                <TableHead className="pr-6 text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((process) => (
                <TableRow key={process.id}>
                  <TableCell className="pl-6 tabular-nums">{process.position}</TableCell>
                  <TableCell className="font-medium">{process.name}</TableCell>
                  <TableCell className="text-right tabular-nums">
                    {process.wipLimit ?? 'Unlimited'}
                  </TableCell>
                  <TableCell className="text-muted-foreground">{inputSummary(process)}</TableCell>
                  <TableCell className="pr-6">
                    <div className="flex justify-end gap-1">
                      <Button
                        size="icon"
                        variant="ghost"
                        aria-label="Edit process"
                        onClick={() => {
                          setEditing(process)
                          setDialogOpen(true)
                        }}
                      >
                        <Pencil />
                      </Button>
                      <ConfirmDialog
                        trigger={
                          <Button size="icon" variant="ghost" aria-label="Delete process">
                            <Trash2 />
                          </Button>
                        }
                        title={`Delete ${process.name}?`}
                        description="This removes the step from the sequence. It is blocked if any order already uses it."
                        confirmLabel="Delete"
                        onConfirm={() => remove.mutate({ id: process.id ?? 0 })}
                        pending={remove.isPending}
                      />
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-muted-foreground py-8 text-center">
                    No processes yet. Define the first step of the sequence.
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </div>
      )}

      {dialogOpen ? (
        <ProcessDialog process={editing} onOpenChange={setDialogOpen} />
      ) : null}
    </div>
  )
}
