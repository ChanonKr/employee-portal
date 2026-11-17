// Thin fetch wrapper. Every error from the API uses one envelope:
// { error: { code, message, fields, requestId } }

export type ApiErrorCode =
  | 'VALIDATION_ERROR'
  | 'UNAUTHORIZED'
  | 'FORBIDDEN'
  | 'PASSWORD_CHANGE_REQUIRED'
  | 'NOT_FOUND'
  | 'METHOD_NOT_ALLOWED'
  | 'VERSION_CONFLICT'
  | 'ALREADY_EXISTS'
  | 'PAYLOAD_TOO_LARGE'
  | 'UNSUPPORTED_MEDIA_TYPE'
  | 'PRECONDITION_REQUIRED'
  | 'RATE_LIMITED'
  | 'INTERNAL_ERROR'
  | 'NETWORK_ERROR'

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: ApiErrorCode,
    message: string,
    readonly fields: Record<string, string> | null = null,
    readonly requestId?: string,
  ) {
    super(message)
    this.name = 'ApiError'
  }
}

type Options = {
  body?: unknown // JSON-encoded unless it is FormData
  version?: number // sent as If-Match for optimistic locking
  signal?: AbortSignal
}

// Called on any 401 except from login / me (those handle it themselves).
let onUnauthorized: (() => void) | null = null
export function setUnauthorizedHandler(fn: (() => void) | null) {
  onUnauthorized = fn
}

const BASE = '/api/v1'

export async function request<T>(method: string, path: string, opts: Options = {}): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json' }
  let body: BodyInit | undefined
  if (opts.body instanceof FormData) {
    body = opts.body
  } else if (opts.body !== undefined) {
    headers['Content-Type'] = 'application/json'
    body = JSON.stringify(opts.body)
  }
  if (opts.version !== undefined) headers['If-Match'] = `"${opts.version}"`

  let res: Response
  try {
    res = await fetch(BASE + path, { method, headers, body, credentials: 'same-origin', signal: opts.signal })
  } catch (e) {
    if (e instanceof DOMException && e.name === 'AbortError') throw e
    throw new ApiError(0, 'NETWORK_ERROR', 'Cannot reach the server. Check your connection and try again.')
  }

  if (res.status === 204) return undefined as T
  const data = await res.json().catch(() => null)
  if (res.ok) return data as T

  const err = data?.error
  const apiError = new ApiError(
    res.status,
    err?.code ?? 'INTERNAL_ERROR',
    err?.message ?? 'Something went wrong. Please try again.',
    err?.fields ?? null,
    err?.requestId,
  )
  if (res.status === 401 && !path.startsWith('/auth/login') && !path.startsWith('/auth/me')) {
    onUnauthorized?.()
  }
  throw apiError
}

export const api = {
  get: <T>(path: string, opts?: Options) => request<T>('GET', path, opts),
  post: <T>(path: string, body?: unknown, opts?: Options) => request<T>('POST', path, { ...opts, body }),
  put: <T>(path: string, body?: unknown, opts?: Options) => request<T>('PUT', path, { ...opts, body }),
  patch: <T>(path: string, body?: unknown, opts?: Options) => request<T>('PATCH', path, { ...opts, body }),
}

export function isApiError(e: unknown, code?: ApiErrorCode): e is ApiError {
  return e instanceof ApiError && (code === undefined || e.code === code)
}

/** The message to show for a failed request, preferring the field message if there is one. */
export function errorMessage(e: unknown): string {
  if (e instanceof ApiError) {
    const first = e.fields ? Object.values(e.fields)[0] : undefined
    return first ? `${capitalize(Object.keys(e.fields!)[0]!)} ${first}` : e.message
  }
  return 'Something went wrong. Please try again.'
}

function capitalize(s: string) {
  const words = s.replace(/([A-Z])/g, ' $1').toLowerCase()
  return words.charAt(0).toUpperCase() + words.slice(1)
}
