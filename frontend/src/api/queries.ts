import { useQuery } from '@tanstack/react-query'

import {
  getGetBoardQueryOptions,
  getGetOrdersIdQueryOptions,
  getGetOrdersQueryOptions,
  getGetOverviewQueryOptions,
  getGetProcessesQueryOptions,
  getGetResourcesQueryOptions,
} from './gen'

/** How often the live views re-read the server while auto-refresh is on. */
export const POLL_MS = 4000

function interval(live: boolean): number | false {
  return live ? POLL_MS : false
}

/**
 * Query definitions live here rather than in components, so a route loader and
 * the components that render the same data always share one key and one set of
 * options.
 */
export const boardQueryOptions = (live = false) =>
  getGetBoardQueryOptions({ query: { refetchInterval: interval(live) } })

export const overviewQueryOptions = (live = false) =>
  getGetOverviewQueryOptions({ query: { refetchInterval: interval(live) } })

export const ordersQueryOptions = (live = false) =>
  getGetOrdersQueryOptions({ query: { refetchInterval: interval(live) } })

export const orderQueryOptions = (id: number, live = false) =>
  getGetOrdersIdQueryOptions(id, { query: { refetchInterval: interval(live) } })

export const resourcesQueryOptions = () => getGetResourcesQueryOptions()

export const processesQueryOptions = () => getGetProcessesQueryOptions()

/** Orders in progress at each step, plus the queue waiting behind them. */
export function useBoard(live = false) {
  return useQuery(boardQueryOptions(live))
}

/** Counters and the resource-consumption panel. */
export function useOverview(live = false) {
  return useQuery(overviewQueryOptions(live))
}

export function useOrders(live = false) {
  return useQuery(ordersQueryOptions(live))
}

export function useOrder(id: number, live = false) {
  return useQuery(orderQueryOptions(id, live))
}

export function useResources() {
  return useQuery(resourcesQueryOptions())
}

export function useProcesses() {
  return useQuery(processesQueryOptions())
}
