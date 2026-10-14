import { useState } from 'react'
import { Pencil, Plus, Trash2 } from 'lucide-react'

import type { DomainResource } from '@/api'
import { useDeleteResource, useResources } from '@/api'
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

import { ResourceDialog } from './resource-dialog'

export function ResourceTable() {
  const resources = useResources()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<DomainResource | null>(null)

  const remove = useDeleteResource()

  const rows = resources.data ?? []

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <p className="text-muted-foreground text-sm">
          Materials, machines and labour consumed while work progresses.
        </p>
        <Button
          onClick={() => {
            setEditing(null)
            setDialogOpen(true)
          }}
        >
          <Plus /> New resource
        </Button>
      </div>

      {resources.isLoading ? (
        <Skeleton className="h-48 w-full" />
      ) : (
        <div className="bg-background rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="pl-6">Name</TableHead>
                <TableHead>Kind</TableHead>
                <TableHead>Unit</TableHead>
                <TableHead className="text-right">Capacity</TableHead>
                <TableHead className="pr-6 text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {rows.map((resource) => (
                <TableRow key={resource.id}>
                  <TableCell className="pl-6 font-medium">{resource.name}</TableCell>
                  <TableCell className="text-muted-foreground capitalize">{resource.kind}</TableCell>
                  <TableCell className="text-muted-foreground">{resource.unit}</TableCell>
                  <TableCell className="text-right tabular-nums">
                    {resource.capacity === undefined || resource.capacity === null
                      ? 'Unlimited'
                      : formatNumber(resource.capacity)}
                  </TableCell>
                  <TableCell className="pr-6">
                    <div className="flex justify-end gap-1">
                      <Button
                        size="icon"
                        variant="ghost"
                        aria-label="Edit resource"
                        onClick={() => {
                          setEditing(resource)
                          setDialogOpen(true)
                        }}
                      >
                        <Pencil />
                      </Button>
                      <ConfirmDialog
                        trigger={
                          <Button size="icon" variant="ghost" aria-label="Delete resource">
                            <Trash2 />
                          </Button>
                        }
                        title={`Delete ${resource.name}?`}
                        description="This removes the resource. It is blocked if any step or recorded consumption still uses it."
                        confirmLabel="Delete"
                        onConfirm={() => remove.mutate({ id: resource.id ?? 0 })}
                        pending={remove.isPending}
                      />
                    </div>
                  </TableCell>
                </TableRow>
              ))}
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={5} className="text-muted-foreground py-8 text-center">
                    No resources yet. Create one to get started.
                  </TableCell>
                </TableRow>
              ) : null}
            </TableBody>
          </Table>
        </div>
      )}

      {dialogOpen ? (
        <ResourceDialog resource={editing} onOpenChange={setDialogOpen} />
      ) : null}
    </div>
  )
}
