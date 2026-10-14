import type { QueryClient } from '@tanstack/react-query'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'

import { errorMessage } from '@/lib/errors'

import {
  useDeleteProcessesId,
  useDeleteResourcesId,
  usePostOrders,
  usePostOrdersIdCancel,
  usePostOrdersIdStepsStepIdComplete,
  usePostOrdersIdStepsStepIdStart,
  usePostProcesses,
  usePostResources,
  usePutProcessesId,
  usePutResourcesId,
} from './gen'
import { queryKeys } from './keys'

/**
 * A step transition moves an order through the sequence, so it changes the
 * orders, the board, the counters and the consumption ledger. Refreshing these
 * four is still far narrower than clearing the whole cache.
 */
function invalidateWork(client: QueryClient) {
  void client.invalidateQueries({ queryKey: queryKeys.orders() })
  void client.invalidateQueries({ queryKey: queryKeys.board() })
  void client.invalidateQueries({ queryKey: queryKeys.overview() })
  void client.invalidateQueries({ queryKey: queryKeys.consumptions() })
}

/** Editing steps or resources changes the board columns and the stock panel. */
function invalidateConfig(client: QueryClient) {
  void client.invalidateQueries({ queryKey: queryKeys.processes() })
  void client.invalidateQueries({ queryKey: queryKeys.resources() })
  void client.invalidateQueries({ queryKey: queryKeys.board() })
  void client.invalidateQueries({ queryKey: queryKeys.overview() })
}

/** Components use this to run their own cleanup, such as closing a dialog. */
type AfterSuccess = { onSuccess?: () => void }

export function useStartStep({ onSuccess }: AfterSuccess = {}) {
  const client = useQueryClient()
  return usePostOrdersIdStepsStepIdStart({
    mutation: {
      onSuccess: () => {
        toast.success('Step started — consumption recorded')
        invalidateWork(client)
        onSuccess?.()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })
}

export function useCompleteStep({ onSuccess }: AfterSuccess = {}) {
  const client = useQueryClient()
  return usePostOrdersIdStepsStepIdComplete({
    mutation: {
      onSuccess: () => {
        toast.success('Step completed')
        invalidateWork(client)
        onSuccess?.()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })
}

export function useCreateOrder({ onSuccess }: AfterSuccess = {}) {
  const client = useQueryClient()
  return usePostOrders({
    mutation: {
      onSuccess: () => {
        toast.success('Order created')
        invalidateWork(client)
        onSuccess?.()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })
}

export function useCancelOrder({ onSuccess }: AfterSuccess = {}) {
  const client = useQueryClient()
  return usePostOrdersIdCancel({
    mutation: {
      onSuccess: () => {
        toast.success('Order cancelled')
        invalidateWork(client)
        onSuccess?.()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })
}

export function useCreateResource({ onSuccess }: AfterSuccess = {}) {
  const client = useQueryClient()
  return usePostResources({
    mutation: {
      onSuccess: () => {
        toast.success('Resource created')
        invalidateConfig(client)
        onSuccess?.()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })
}

export function useUpdateResource({ onSuccess }: AfterSuccess = {}) {
  const client = useQueryClient()
  return usePutResourcesId({
    mutation: {
      onSuccess: () => {
        toast.success('Resource updated')
        invalidateConfig(client)
        onSuccess?.()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })
}

export function useDeleteResource({ onSuccess }: AfterSuccess = {}) {
  const client = useQueryClient()
  return useDeleteResourcesId({
    mutation: {
      onSuccess: () => {
        toast.success('Resource deleted')
        invalidateConfig(client)
        onSuccess?.()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })
}

export function useCreateProcess({ onSuccess }: AfterSuccess = {}) {
  const client = useQueryClient()
  return usePostProcesses({
    mutation: {
      onSuccess: () => {
        toast.success('Process created')
        invalidateConfig(client)
        onSuccess?.()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })
}

export function useUpdateProcess({ onSuccess }: AfterSuccess = {}) {
  const client = useQueryClient()
  return usePutProcessesId({
    mutation: {
      onSuccess: () => {
        toast.success('Process updated')
        invalidateConfig(client)
        onSuccess?.()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })
}

export function useDeleteProcess({ onSuccess }: AfterSuccess = {}) {
  const client = useQueryClient()
  return useDeleteProcessesId({
    mutation: {
      onSuccess: () => {
        toast.success('Process deleted')
        invalidateConfig(client)
        onSuccess?.()
      },
      onError: (error) => toast.error(errorMessage(error)),
    },
  })
}
