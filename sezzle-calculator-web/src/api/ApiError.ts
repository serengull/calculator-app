export class ApiError extends Error {
  readonly status: number
  readonly requestId: string | undefined

  constructor(message: string, status: number, requestId?: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.requestId = requestId
  }
}

export class NetworkError extends Error {
  constructor(message = 'Could not reach the calculator service.') {
    super(message)
    this.name = 'NetworkError'
  }
}

export class ProtocolError extends Error {
  constructor(message = 'Unexpected response from the calculator service.') {
    super(message)
    this.name = 'ProtocolError'
  }
}

export function toDisplayError(error: unknown): string {
  if (
    error instanceof ApiError ||
    error instanceof NetworkError ||
    error instanceof ProtocolError
  ) {
    return error.message
  }
  return 'Something went wrong. Please try again.'
}
