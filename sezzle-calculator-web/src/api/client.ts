import { ApiError, NetworkError, ProtocolError } from './ApiError'
import type { ApiErrorBody, CalculateRequest, CalculateSuccess } from './types'

export const DEFAULT_BASE_URL = '/api'

export function resolveBaseUrl(): string {
  const configured = import.meta.env.VITE_API_BASE_URL
  const base = configured && configured.trim() !== '' ? configured : DEFAULT_BASE_URL
  return base.endsWith('/') ? base.slice(0, -1) : base
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function parseErrorBody(payload: unknown, status: number): ApiErrorBody {
  if (isRecord(payload) && typeof payload.message === 'string' && payload.message !== '') {
    return {
      message: payload.message,
      requestId: typeof payload.requestId === 'string' ? payload.requestId : undefined,
    }
  }
  return { message: `Request failed with status ${status}.` }
}

function parseSuccessBody(payload: unknown): CalculateSuccess {
  if (isRecord(payload) && typeof payload.result === 'number' && Number.isFinite(payload.result)) {
    return { result: payload.result }
  }
  throw new ProtocolError()
}

export async function calculate(
  request: CalculateRequest,
  options: { baseUrl?: string; signal?: AbortSignal } = {},
): Promise<CalculateSuccess> {
  const baseUrl = options.baseUrl ?? resolveBaseUrl()
  const params = new URLSearchParams({
    first: String(request.first),
    second: String(request.second),
    op: request.op,
  })

  let response: Response
  try {
    response = await fetch(`${baseUrl}/calculate?${params.toString()}`, {
      method: 'GET',
      headers: { Accept: 'application/json' },
      signal: options.signal,
    })
  } catch (error) {
    if (options.signal?.aborted || (error instanceof Error && error.name === 'AbortError')) {
      throw error
    }
    throw new NetworkError()
  }

  let payload: unknown = null
  try {
    payload = await response.json()
  } catch {
    payload = null
  }

  if (!response.ok) {
    const body = parseErrorBody(payload, response.status)
    throw new ApiError(body.message, response.status, body.requestId)
  }

  return parseSuccessBody(payload)
}
