import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

const orderLabels: Record<string, string> = {
  pending: 'Pending',
  in_progress: 'In progress',
  done: 'Done',
  cancelled: 'Cancelled',
}

const stepLabels: Record<string, string> = {
  queued: 'Queued',
  running: 'Running',
  done: 'Done',
}

const styles: Record<string, string> = {
  pending: 'border-slate-200 bg-slate-100 text-slate-700',
  queued: 'border-slate-200 bg-slate-100 text-slate-700',
  in_progress: 'border-blue-200 bg-blue-100 text-blue-700',
  running: 'border-blue-200 bg-blue-100 text-blue-700',
  done: 'border-emerald-200 bg-emerald-100 text-emerald-700',
  cancelled: 'border-rose-200 bg-rose-100 text-rose-700',
}

export function StatusBadge({
  status,
  kind = 'order',
}: {
  status?: string
  kind?: 'order' | 'step'
}) {
  const key = status ?? 'pending'
  const label =
    kind === 'step' ? (stepLabels[key] ?? key) : (orderLabels[key] ?? key)

  return (
    <Badge variant="outline" className={cn('font-medium', styles[key])}>
      {label}
    </Badge>
  )
}
