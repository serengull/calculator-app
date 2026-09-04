export type Fixity = 'infix' | 'prefix' | 'postfix'

export interface OperationSpec {
  symbol: string
  arity: 1 | 2
  fixity: Fixity
}

export const OPERATIONS = {
  sum: { symbol: '+', arity: 2, fixity: 'infix' },
  subtract: { symbol: '−', arity: 2, fixity: 'infix' },
  multiply: { symbol: '×', arity: 2, fixity: 'infix' },
  division: { symbol: '÷', arity: 2, fixity: 'infix' },
  exponentiation: { symbol: '^', arity: 2, fixity: 'infix' },
  sqrt: { symbol: '√', arity: 1, fixity: 'prefix' },
  percentage: { symbol: '%', arity: 1, fixity: 'postfix' },
} as const satisfies Record<string, OperationSpec>

export type Operation = keyof typeof OPERATIONS

export interface CalculateRequest {
  first: number
  second: number
  op: Operation
}

export interface CalculateSuccess {
  result: number
}

export interface ApiErrorBody {
  message: string
  requestId?: string
}

export function isUnary(op: Operation): boolean {
  return OPERATIONS[op].arity === 1
}
