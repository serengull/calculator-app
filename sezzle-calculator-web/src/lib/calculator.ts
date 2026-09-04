import { isUnary, OPERATIONS } from '../api/types'
import type { Operation } from '../api/types'

export const MAX_DIGITS = 15

export interface CalculatorState {
  display: string
  accumulator: number | null
  pendingOp: Operation | null
  overwrite: boolean
  lastOp: Operation | null
  lastOperand: number | null
  cleared: boolean
  completed: CompletedExpression | null
  error: string | null
  inFlight: ComputeRequest | null
}

export interface CompletedExpression {
  first: number
  op: Operation
  second: number
}

export const INITIAL_STATE: CalculatorState = {
  display: '0',
  accumulator: null,
  pendingOp: null,
  overwrite: true,
  lastOp: null,
  lastOperand: null,
  cleared: true,
  completed: null,
  error: null,
  inFlight: null,
}

export type Follow = { kind: 'chain'; op: Operation } | { kind: 'equals' }

export interface ComputeRequest {
  first: number
  second: number
  op: Operation
  follow: Follow
}

export type KeyAction =
  | { type: 'digit'; value: string }
  | { type: 'dot' }
  | { type: 'negate' }
  | { type: 'backspace' }
  | { type: 'clear' }
  | { type: 'operation'; op: Operation }
  | { type: 'equals' }

export type CalculatorAction =
  | KeyAction
  | { type: 'result'; result: number }
  | { type: 'failure'; message: string }

export function interrupts(action: CalculatorAction): boolean {
  return action.type === 'clear'
}

export function reduce(state: CalculatorState, action: CalculatorAction): CalculatorState {
  if (action.type === 'result') return applyResult(state, action.result)
  if (action.type === 'failure') return applyError(state, action.message)
  if (state.inFlight !== null && !interrupts(action)) return state

  switch (action.type) {
    case 'digit':
      return inputDigit(state, action.value)
    case 'dot':
      return inputDot(state)
    case 'negate':
      return negate(state)
    case 'backspace':
      return backspace(state)
    case 'clear':
      return clear(state)
    case 'equals':
      return equals(state)
    case 'operation':
      return isUnary(action.op) ? applyUnary(state, action.op) : chooseOperation(state, action.op)
  }
}

export function currentValue(state: CalculatorState): number {
  const parsed = Number(state.display)
  return Number.isFinite(parsed) ? parsed : 0
}

export const MAX_DISPLAY_CHARS = 21

const EXPONENTIAL_DIGITS = 8

export function numberToDisplay(value: number): string {
  if (!Number.isFinite(value)) return 'Error'

  const plain =
    Number.isInteger(value) && Math.abs(value) < 1e21
      ? String(value)
      : String(Number(value.toPrecision(12)))

  if (formatDisplay(plain).length <= MAX_DISPLAY_CHARS) return plain

  return value.toExponential(EXPONENTIAL_DIGITS).replace(/\.?0+e/, 'e')
}

function countDigits(display: string): number {
  return display.replace(/[^0-9]/g, '').length
}

export function inputDigit(state: CalculatorState, digit: string): CalculatorState {
  if (state.error) return state
  if (state.overwrite) {
    return { ...state, display: digit, overwrite: false, cleared: false }
  }
  if (countDigits(state.display) >= MAX_DIGITS) return state
  const display = state.display === '0' ? digit : state.display + digit
  return { ...state, display, cleared: false }
}

export function inputDot(state: CalculatorState): CalculatorState {
  if (state.error) return state
  if (state.overwrite) {
    return { ...state, display: '0.', overwrite: false, cleared: false }
  }
  if (state.display.includes('.')) return state
  return { ...state, display: `${state.display}.`, cleared: false }
}

export function backspace(state: CalculatorState): CalculatorState {
  if (state.error) return clear(state)
  if (state.overwrite) return state
  const next = state.display.slice(0, -1)
  if (next === '' || next === '-') {
    return { ...state, display: '0', overwrite: true, cleared: true }
  }
  return { ...state, display: next }
}

export function clear(state: CalculatorState): CalculatorState {
  if (state.cleared) return INITIAL_STATE
  return {
    ...state,
    display: '0',
    overwrite: true,
    cleared: true,
    completed: null,
    error: null,
    inFlight: null,
  }
}

export function negate(state: CalculatorState): CalculatorState {
  if (state.error) return state
  if (state.display === '0') return state
  const display = state.display.startsWith('-')
    ? state.display.slice(1)
    : `-${state.display}`
  return { ...state, display, cleared: false }
}

export function chooseOperation(state: CalculatorState, op: Operation): CalculatorState {
  if (state.error) return state

  if (state.pendingOp !== null && state.accumulator !== null && !state.overwrite) {
    return {
      ...state,
      overwrite: true,
      cleared: true,
      inFlight: {
        first: state.accumulator,
        second: currentValue(state),
        op: state.pendingOp,
        follow: { kind: 'chain', op },
      },
    }
  }

  return {
    ...state,
    accumulator: currentValue(state),
    pendingOp: op,
    overwrite: true,
    cleared: true,
  }
}

export function equals(state: CalculatorState): CalculatorState {
  if (state.error) return state

  if (state.pendingOp !== null && state.accumulator !== null) {
    return {
      ...state,
      overwrite: true,
      inFlight: {
        first: state.accumulator,
        second: currentValue(state),
        op: state.pendingOp,
        follow: { kind: 'equals' },
      },
    }
  }

  if (state.lastOp !== null && state.lastOperand !== null) {
    return {
      ...state,
      inFlight: {
        first: currentValue(state),
        second: state.lastOperand,
        op: state.lastOp,
        follow: { kind: 'equals' },
      },
    }
  }

  return state
}

export function applyResult(state: CalculatorState, result: number): CalculatorState {
  const request = state.inFlight
  if (!request) return state

  return {
    ...state,
    display: numberToDisplay(result),
    overwrite: true,
    cleared: true,
    completed: { first: request.first, op: request.op, second: request.second },
    error: null,
    inFlight: null,
    ...(request.follow.kind === 'chain'
      ? { accumulator: result, pendingOp: request.follow.op, lastOp: null, lastOperand: null }
      : { accumulator: null, pendingOp: null, lastOp: request.op, lastOperand: request.second }),
  }
}

export function applyError(state: CalculatorState, message: string): CalculatorState {
  return {
    ...state,
    display: 'Error',
    accumulator: null,
    pendingOp: null,
    overwrite: true,
    cleared: true,
    lastOp: null,
    lastOperand: null,
    error: message,
    inFlight: null,
  }
}

export function formatDisplay(display: string): string {
  if (display === 'Error') return display

  const negative = display.startsWith('-')
  const unsigned = negative ? display.slice(1) : display
  const [integer, fraction] = unsigned.split('.')
  const grouped = integer.replace(/\B(?=(\d{3})+(?!\d))/g, ',')
  const rebuilt = fraction === undefined ? grouped : `${grouped}.${fraction}`
  return negative ? `-${rebuilt}` : rebuilt
}

export function formatValue(value: number): string {
  return formatDisplay(numberToDisplay(value))
}


export function formatExpression(first: number, op: Operation, second: number): string {
  const { symbol, fixity } = OPERATIONS[op]
  const left = formatValue(first)

  if (fixity === 'prefix') return `${symbol}${left}`
  if (fixity === 'postfix') return `${left}${symbol}`
  return `${left} ${symbol} ${formatValue(second)}`
}

export function applyUnary(state: CalculatorState, op: Operation): CalculatorState {
  if (state.error) return state

  return {
    ...state,
    overwrite: true,
    cleared: true,
    inFlight: { first: currentValue(state), second: 0, op, follow: { kind: 'equals' } },
  }
}

export function inputLine(state: CalculatorState): string {
  if (state.error) return state.display

  if (state.pendingOp !== null && state.accumulator !== null) {
    const left = formatValue(state.accumulator)
    const symbol = OPERATIONS[state.pendingOp].symbol
    return state.overwrite ? `${left} ${symbol}` : `${left} ${symbol} ${formatDisplay(state.display)}`
  }

  return formatDisplay(state.display)
}

export function historyLine(state: CalculatorState): string {
  if (!state.completed) return ''

  const { first, op, second } = state.completed
  return `${formatExpression(first, op, second)} =`
}

export function armedOperation(state: CalculatorState): Operation | null {
  return state.pendingOp !== null && state.overwrite ? state.pendingOp : null
}

export function clearLabel(state: CalculatorState): string {
  return state.cleared ? 'AC' : 'C'
}

export function clearAriaLabel(state: CalculatorState): string {
  return state.cleared ? 'All clear' : 'Clear'
}

export const MIN_VALUE_FONT_PX = 22
export const MIN_HISTORY_FONT_PX = 10

export function fitFontSize(
  base: number,
  min: number,
  textWidth: number,
  available: number,
): number {
  if (!(base > 0) || !(textWidth > 0) || !(available > 0)) return base
  if (textWidth <= available) return base

  const scaled = Math.floor((available / textWidth) * base * 100) / 100
  return Math.max(min, Math.min(base, scaled))
}
