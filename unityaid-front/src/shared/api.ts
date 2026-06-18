export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

const API_ASSET_ORIGIN = (() => {
  try {
    const fallbackOrigin = typeof window === 'undefined' ? 'http://localhost:8080' : window.location.origin
    return new URL(API_BASE_URL, fallbackOrigin).origin
  } catch {
    return ''
  }
})()

function normalizeAssetUrl(value: string) {
  if (!API_ASSET_ORIGIN) return value
  if (value.startsWith('/uploads/')) return `${API_ASSET_ORIGIN}${value}`

  try {
    const url = new URL(value)
    if ((url.hostname === 'localhost' || url.hostname === '127.0.0.1') && url.pathname.startsWith('/uploads/')) {
      return `${API_ASSET_ORIGIN}${url.pathname}${url.search}${url.hash}`
    }
  } catch {
    return value
  }

  return value
}

function normalizePayloadAssets<T>(payload: T): T {
  if (typeof payload === 'string') return normalizeAssetUrl(payload) as T
  if (!payload || typeof payload !== 'object') return payload
  if (Array.isArray(payload)) return payload.map((item) => normalizePayloadAssets(item)) as T

  return Object.fromEntries(
    Object.entries(payload).map(([key, value]) => [key, normalizePayloadAssets(value)])
  ) as T
}

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.status = status
  }
}

type RequestOptions = RequestInit & {
  token?: string | null
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers)
  headers.set('Accept', 'application/json')

  if (options.body && !(options.body instanceof FormData) && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  if (options.token) {
    headers.set('Authorization', `Bearer ${options.token}`)
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers
  })

  const payload = await response.json().catch(() => null)

  if (!response.ok) {
    throw new ApiError(payload?.message ?? 'Ошибка запроса', response.status)
  }

  return normalizePayloadAssets(payload as T)
}
