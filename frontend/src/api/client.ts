/**
 * The one place HTTP happens. Orval-generated hooks call into `customInstance`
 * (see mutator.ts), which calls `request` here.
 */

export const API_BASE = '/api'

export class ApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string, message: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

/** Ensure a path starts with the API base exactly once. */
export function withApiBase(url: string): string {
  if (/^https?:\/\//.test(url)) return url
  if (url.startsWith(`${API_BASE}/`) || url === API_BASE) return url
  return `${API_BASE}${url.startsWith('/') ? '' : '/'}${url}`
}

function safeJson(text: string): unknown {
  try {
    return JSON.parse(text)
  } catch {
    return text
  }
}

export async function request<T>(
  url: string,
  init: RequestInit = {},
): Promise<T> {
  const response = await fetch(withApiBase(url), {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init.headers ?? {}),
    },
  })

  if (response.status === 204) {
    return undefined as T
  }

  const text = await response.text()
  const payload = text ? safeJson(text) : undefined

  if (!response.ok) {
    const detail = (
      payload as { error?: { code?: string; message?: string } } | undefined
    )?.error
    throw new ApiError(
      response.status,
      detail?.code ?? 'http_error',
      detail?.message ?? response.statusText ?? 'Request failed',
    )
  }

  return payload as T
}
