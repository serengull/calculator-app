import { beforeEach, describe, expect, it, vi } from 'vitest'
import { calculate, DEFAULT_BASE_URL, resolveBaseUrl } from './client'
import { ApiError, NetworkError, ProtocolError } from './ApiError'
import { jsonResponse } from '../test/http'

describe('resolveBaseUrl', () => {
  it('falls back to the dev proxy path', () => {
    expect(resolveBaseUrl()).toBe(DEFAULT_BASE_URL)
  })

  it('strips a trailing slash from a configured base url', () => {
    vi.stubEnv('VITE_API_BASE_URL', 'http://localhost:8081/')
    expect(resolveBaseUrl()).toBe('http://localhost:8081')
  })

  it('resolves the same-origin production value to an empty base', () => {
    vi.stubEnv('VITE_API_BASE_URL', '/')
    expect(resolveBaseUrl()).toBe('')
  })

  it('ignores a blank configured value', () => {
    vi.stubEnv('VITE_API_BASE_URL', '   ')
    expect(resolveBaseUrl()).toBe(DEFAULT_BASE_URL)
  })
})

describe('calculate', () => {
  let fetchMock: ReturnType<typeof vi.fn>

  beforeEach(() => {
    fetchMock = vi.fn()
    vi.stubGlobal('fetch', fetchMock)
  })

  it('builds the query string and returns the result', async () => {
    fetchMock.mockResolvedValue(jsonResponse({ result: 6 }))

    const result = await calculate({ first: 2, second: 4, op: 'sum' }, { baseUrl: '/api' })

    expect(result).toEqual({ result: 6 })
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/calculate?first=2&second=4&op=sum')
    expect(init).toMatchObject({ method: 'GET' })
  })

  it('sends the subtract wire value', async () => {
    fetchMock.mockResolvedValue(jsonResponse({ result: -2 }))

    await calculate({ first: 2, second: 4, op: 'subtract' }, { baseUrl: '/api' })

    expect(String(fetchMock.mock.calls[0][0])).toBe(
      '/api/calculate?first=2&second=4&op=subtract',
    )
  })

  it('calls the same origin when the base url is empty', async () => {
    fetchMock.mockResolvedValue(jsonResponse({ result: 6 }))

    await calculate({ first: 2, second: 4, op: 'sum' }, { baseUrl: '' })

    expect(String(fetchMock.mock.calls[0][0])).toBe('/calculate?first=2&second=4&op=sum')
  })

  it('surfaces the "must be numbers" message', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse(
        { message: 'first and second must be numbers', requestId: 'abc123' },
        400,
      ),
    )

    await expect(
      calculate({ first: 1, second: 2, op: 'sum' }, { baseUrl: '/api' }),
    ).rejects.toMatchObject({
      name: 'ApiError',
      message: 'first and second must be numbers',
      status: 400,
      requestId: 'abc123',
    })
  })

  it('surfaces the invalid-op message', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({ message: 'op must be one of: sum, subtract, multiply, division' }, 400),
    )

    await expect(
      calculate({ first: 1, second: 2, op: 'sum' }, { baseUrl: '/api' }),
    ).rejects.toThrow('op must be one of: sum, subtract, multiply, division')
  })

  it('surfaces the divide-by-zero message and the request id', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({ message: 'cannot divide by zero', requestId: 'XTAq' }, 400),
    )

    const error = await calculate(
      { first: 1, second: 0, op: 'division' },
      { baseUrl: '/api' },
    ).catch((e: unknown) => e)

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).status).toBe(400)
    expect((error as ApiError).requestId).toBe('XTAq')
  })

  it('handles a 5xx with the same body shape', async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({ message: 'internal server error', requestId: 'zzz' }, 500),
    )

    await expect(
      calculate({ first: 1, second: 2, op: 'sum' }, { baseUrl: '/api' }),
    ).rejects.toMatchObject({ status: 500, message: 'internal server error' })
  })

  it('falls back to a status message when the error body is unusable', async () => {
    fetchMock.mockResolvedValue(new Response('not json', { status: 502 }))

    await expect(
      calculate({ first: 1, second: 2, op: 'sum' }, { baseUrl: '/api' }),
    ).rejects.toThrow('Request failed with status 502.')
  })

  it('rejects a 200 response without a numeric result', async () => {
    fetchMock.mockResolvedValue(jsonResponse({ result: 'six' }))

    const error = await calculate(
      { first: 1, second: 2, op: 'sum' },
      { baseUrl: '/api' },
    ).catch((e: unknown) => e)

    expect(error).toBeInstanceOf(ProtocolError)
    expect(error).not.toBeInstanceOf(ApiError)
    expect((error as ProtocolError).message).toBe(
      'Unexpected response from the calculator service.',
    )
  })

  it('rethrows an abort instead of reporting a network failure', async () => {
    const controller = new AbortController()
    const abortError = new DOMException('The operation was aborted.', 'AbortError')
    fetchMock.mockRejectedValue(abortError)
    controller.abort()

    const error = await calculate(
      { first: 1, second: 2, op: 'sum' },
      { baseUrl: '/api', signal: controller.signal },
    ).catch((e: unknown) => e)

    expect(error).toBe(abortError)
    expect(error).not.toBeInstanceOf(NetworkError)
  })

  it('rejects a non-finite result', async () => {
    fetchMock.mockResolvedValue(new Response('{"result":1e999}', { status: 200 }))

    await expect(
      calculate({ first: 1, second: 2, op: 'sum' }, { baseUrl: '/api' }),
    ).rejects.toThrow('Unexpected response from the calculator service.')
  })

  it('wraps a network failure', async () => {
    fetchMock.mockRejectedValue(new TypeError('Failed to fetch'))

    const error = await calculate({ first: 1, second: 2, op: 'sum' }, { baseUrl: '/api' }).catch(
      (e: unknown) => e,
    )

    expect(error).toBeInstanceOf(NetworkError)
    expect((error as NetworkError).message).toBe('Could not reach the calculator service.')
  })
})
