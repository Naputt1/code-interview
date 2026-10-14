import { request } from './client'

/**
 * Orval mutator. Orval calls this instead of `fetch` for every generated hook,
 * so all traffic funnels through `request` in client.ts.
 *
 * Orval passes an already-stringified `body`; hand-written callers may pass a
 * plain object as `data`. Both are forwarded, and `body` wins when present.
 */
type MutatorConfig = RequestInit & { data?: unknown }

export function customInstance<T>(
  url: string,
  config: MutatorConfig = {},
): Promise<T> {
  const { data, body, ...init } = config
  const payload = body ?? data
  return request<T>(url, {
    ...init,
    body:
      payload === undefined
        ? undefined
        : typeof payload === 'string'
          ? payload
          : JSON.stringify(payload),
  })
}
