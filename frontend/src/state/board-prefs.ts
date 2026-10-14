import { create } from 'zustand'

/**
 * Client-only UI preferences.
 *
 * Server data never lives here — it is owned by TanStack Query so that caching,
 * invalidation and polling stay in one place. This store only holds view state
 * that the server has no opinion about: whether the production board keeps
 * polling in the background.
 */
type BoardPrefs = {
  live: boolean
  setLive: (live: boolean) => void
}

export const useBoardPrefs = create<BoardPrefs>((set) => ({
  live: true,
  setLive: (live) => set({ live }),
}))
