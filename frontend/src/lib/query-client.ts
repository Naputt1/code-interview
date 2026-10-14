import { QueryClient } from '@tanstack/react-query'

/**
 * One QueryClient for the app. Queries are cheap and frequently refreshed so the
 * board reflects work happening in other tabs; the board/orders views opt into a
 * short refetch interval on top of this.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5_000,
      retry: 1,
      refetchOnWindowFocus: true,
    },
  },
})
