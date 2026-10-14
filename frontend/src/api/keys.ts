import {
  getGetBoardQueryKey,
  getGetConsumptionsQueryKey,
  getGetOrdersQueryKey,
  getGetOverviewQueryKey,
  getGetProcessesQueryKey,
  getGetResourcesQueryKey,
} from './gen'

/**
 * Every query key the app invalidates, taken straight from the Orval factories
 * so they cannot drift from the contract.
 *
 * The list keys are also prefixes of their detail keys (`/orders` matches
 * `/orders/1`), so invalidating `orders()` refreshes the order screens too.
 */
export const queryKeys = {
  board: getGetBoardQueryKey,
  overview: getGetOverviewQueryKey,
  orders: getGetOrdersQueryKey,
  resources: getGetResourcesQueryKey,
  processes: getGetProcessesQueryKey,
  consumptions: getGetConsumptionsQueryKey,
}
