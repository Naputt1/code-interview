import { ApiError } from '@/api/client'

/** Turn any thrown value into a message we can show in a toast. */
export function errorMessage(error: unknown): string {
  if (error instanceof ApiError) return error.message
  if (error instanceof Error) return error.message
  return 'Something went wrong'
}
