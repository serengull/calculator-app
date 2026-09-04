import { describe, expect, it } from 'vitest'
import {
  applyError,
  applyUnary,
  applyResult,
  backspace,
  chooseOperation,
  clear,
  equals,
  fitFontSize,
  formatDisplay,
  formatExpression,
  formatValue,
  historyLine,
  INITIAL_STATE,
  interrupts,
  inputLine,
  inputDigit,
  inputDot,
  MAX_DIGITS,
  MAX_DISPLAY_CHARS,
  negate,
  numberToDisplay,
  reduce,
} from './calculator'
import type { CalculatorState } from './calculator'

function type(keys: string, from: CalculatorState = INITIAL_STATE): CalculatorState {
  return keys.split('').reduce((state, key) => {
    return key === '.' ? inputDot(state) : inputDigit(state, key)
  }, from)
}

describe('digit entry', () => {
  it('replaces the leading zero', () => {
    expect(type('7').display).toBe('7')
  })

  it('appends further digits', () => {
    expect(type('123').display).toBe('123')
  })

  it('starts a decimal from zero', () => {
    expect(inputDot(INITIAL_STATE).display).toBe('0.')
  })

  it('allows only one decimal point', () => {
    expect(type('1.2.3').display).toBe('1.23')
  })

  it('stops at the digit ceiling', () => {
    const long = type('9'.repeat(MAX_DIGITS + 4))
    expect(long.display).toHaveLength(MAX_DIGITS)
  })

  it('deletes the last character', () => {
    expect(backspace(type('123')).display).toBe('12')
  })

  it('returns to zero when the last character is deleted', () => {
    expect(backspace(type('5')).display).toBe('0')
  })
})

describe('clear', () => {
  it('shows C after entry and clears only the entry', () => {
    const entered = chooseOperation(type('12'), 'sum')
    const withSecond = type('34', entered)
    expect(withSecond.cleared).toBe(false)

    const cleared = clear(withSecond)
    expect(cleared.display).toBe('0')
    expect(cleared.pendingOp).toBe('sum')
    expect(cleared.accumulator).toBe(12)
  })

  it('resets everything on a second press', () => {
    const cleared = clear(clear(type('34', chooseOperation(type('12'), 'sum'))))
    expect(cleared).toEqual(INITIAL_STATE)
  })
})

describe('local operations', () => {
  it('negates the entry', () => {
    expect(negate(type('42')).display).toBe('-42')
    expect(negate(negate(type('42'))).display).toBe('42')
  })

  it('leaves a bare zero alone', () => {
    expect(negate(INITIAL_STATE).display).toBe('0')
  })
})

describe('operations dispatched to the API', () => {
  it('stores the accumulator without a request', () => {
    const step = chooseOperation(type('6'), 'multiply')
    expect(step.inFlight).toBeNull()
    expect(step).toMatchObject({ accumulator: 6, pendingOp: 'multiply', overwrite: true })
  })

  it('requests the pending evaluation when a second operator arrives', () => {
    const armed = chooseOperation(type('6'), 'multiply')
    const step = chooseOperation(type('7', armed), 'sum')

    expect(step.inFlight).toEqual({
      first: 6,
      second: 7,
      op: 'multiply',
      follow: { kind: 'chain', op: 'sum' },
    })
  })

  it('swaps the operator when none was entered in between', () => {
    const armed = chooseOperation(type('6'), 'multiply')
    const step = chooseOperation(armed, 'subtract')

    expect(step.inFlight).toBeNull()
    expect(step.pendingOp).toBe('subtract')
    expect(step.accumulator).toBe(6)
  })

  it('requests the evaluation on equals', () => {
    const armed = chooseOperation(type('6'), 'multiply')
    const step = equals(type('7', armed))

    expect(step.inFlight).toEqual({
      first: 6,
      second: 7,
      op: 'multiply',
      follow: { kind: 'equals' },
    })
  })

  it('does nothing on equals with no history', () => {
    expect(equals(type('6')).inFlight).toBeNull()
  })

  it('replays the last operation on a repeated equals', () => {
    const armed = chooseOperation(type('6'), 'multiply')
    const first = equals(type('7', armed))
    const settled = applyResult(first, 42)

    const repeat = equals(settled)
    expect(repeat.inFlight).toEqual({
      first: 42,
      second: 7,
      op: 'multiply',
      follow: { kind: 'equals' },
    })
  })

  it('keeps the chained operator pending once the result lands', () => {
    const settled = applyResult(
      {
        ...INITIAL_STATE,
        inFlight: { first: 6, second: 7, op: 'multiply', follow: { kind: 'chain', op: 'sum' } },
      },
      42,
    )
    expect(settled).toMatchObject({
      display: '42',
      accumulator: 42,
      pendingOp: 'sum',
      overwrite: true,
    })
  })

  it('clears the pending state on an error and blocks further input', () => {
    const errored = applyError(type('6'), 'cannot divide by zero')
    expect(errored.display).toBe('Error')
    expect(errored.error).toBe('cannot divide by zero')
    expect(errored.inFlight).toBeNull()
    expect(inputDigit(errored, '5')).toBe(errored)
    expect(equals(errored).inFlight).toBeNull()
    expect(clear(errored)).toEqual(INITIAL_STATE)
  })
})

describe('formatting', () => {
  it('trims floating point noise', () => {
    expect(numberToDisplay(0.30000000000000004)).toBe('0.3')
  })

  it('keeps integers plain', () => {
    expect(numberToDisplay(42)).toBe('42')
  })

  it('groups thousands', () => {
    expect(formatDisplay('1234567')).toBe('1,234,567')
  })

  it('leaves the fraction ungrouped', () => {
    expect(formatDisplay('-1234.5678')).toBe('-1,234.5678')
  })

  it('preserves a trailing decimal point mid-entry', () => {
    expect(formatDisplay('1000.')).toBe('1,000.')
  })

  it('passes the error text through', () => {
    expect(formatDisplay('Error')).toBe('Error')
  })
})

describe('input line', () => {
  it('shows the entry before an operator is chosen', () => {
    expect(inputLine(INITIAL_STATE)).toBe('0')
    expect(inputLine(type('7'))).toBe('7')
  })

  it('shows the left operand and operator once chosen', () => {
    expect(inputLine(chooseOperation(type('7'), 'multiply'))).toBe('7 ×')
  })

  it('grows into the full problem as the second operand is typed', () => {
    const armed = chooseOperation(type('7'), 'multiply')
    expect(inputLine(type('2', armed))).toBe('7 × 2')
  })

  it('shows the result once it lands', () => {
    const armed = chooseOperation(type('7'), 'multiply')
    const step = equals(type('2', armed))
    const settled = applyResult(step, 14)

    expect(inputLine(settled)).toBe('14')
  })

  it('carries the result into the chained problem', () => {
    const settled = applyResult(
      {
        ...INITIAL_STATE,
        inFlight: { first: 2, second: 3, op: 'sum', follow: { kind: 'chain', op: 'sum' } },
      },
      5,
    )
    expect(inputLine(settled)).toBe('5 +')
  })

  it('groups thousands on both sides', () => {
    const armed = chooseOperation(type('1234'), 'sum')
    expect(inputLine(type('5678', armed))).toBe('1,234 + 5,678')
  })

  it('shows the error text', () => {
    expect(inputLine(applyError(type('8'), 'boom'))).toBe('Error')
  })
})

describe('history line', () => {
  function completed(): CalculatorState {
    const armed = chooseOperation(type('7'), 'multiply')
    const step = equals(type('2', armed))
    return applyResult(step, 14)
  }

  it('is empty when nothing has been calculated', () => {
    expect(historyLine(INITIAL_STATE)).toBe('')
    expect(historyLine(type('7'))).toBe('')
    expect(historyLine(chooseOperation(type('7'), 'multiply'))).toBe('')
  })

  it('shows the finished problem above the result', () => {
    expect(historyLine(completed())).toBe('7 × 2 =')
  })

  it('stays put while the next problem is typed', () => {
    const next = chooseOperation(inputDigit(completed(), '9'), 'sum')
    expect(historyLine(next)).toBe('7 × 2 =')
    expect(inputLine(next)).toBe('9 +')
  })

  it('is replaced by the next completed problem', () => {
    const armed = chooseOperation(inputDigit(completed(), '9'), 'sum')
    const step = equals(type('1', armed))

    expect(historyLine(applyResult(step, 10))).toBe('9 + 1 =')
  })

  it('records a chained evaluation', () => {
    const armed = chooseOperation(type('2'), 'sum')
    const step = chooseOperation(type('3', armed), 'sum')

    expect(historyLine(applyResult(step, 5))).toBe('2 + 3 =')
  })

  it('is cleared by a full reset', () => {
    expect(historyLine(clear(clear(completed())))).toBe('')
  })
})

describe('unary operations', () => {
  it('sends the current entry as the first operand', () => {
    const step = applyUnary(type('144'), 'sqrt')

    expect(step.inFlight).toEqual({
      first: 144,
      second: 0,
      op: 'sqrt',
      follow: { kind: 'equals' },
    })
  })

  it('sends percentage the same way', () => {
    expect(applyUnary(type('50'), 'percentage').inFlight).toMatchObject({
      first: 50,
      op: 'percentage',
    })
  })

  it('leaves the entry ready to be replaced', () => {
    expect(applyUnary(type('9'), 'sqrt').overwrite).toBe(true)
  })

  it('does nothing while an error is shown', () => {
    expect(applyUnary(applyError(type('9'), 'boom'), 'sqrt').inFlight).toBeNull()
  })

  it('shows the result on the input line', () => {
    const step = applyUnary(type('144'), 'sqrt')
    expect(inputLine(applyResult(step, 12))).toBe('12')
  })
})

describe('expression formatting', () => {
  it('puts the root symbol before its operand', () => {
    expect(formatExpression(144, 'sqrt', 0)).toBe('√144')
  })

  it('puts the percent symbol after its operand', () => {
    expect(formatExpression(50, 'percentage', 0)).toBe('50%')
  })

  it('renders exponentiation between its operands', () => {
    expect(formatExpression(2, 'exponentiation', 10)).toBe('2 ^ 10')
  })

  it('omits the unused second operand from the history line', () => {
    const step = applyUnary(type('144'), 'sqrt')
    expect(historyLine(applyResult(step, 12))).toBe('√144 =')
  })
})

describe('display budget', () => {
  const shown = formatValue

  it('keeps a value that fits in plain notation', () => {
    expect(shown(1234567)).toBe('1,234,567')
  })

  it('keeps the longest exact entry in plain notation', () => {
    expect(shown(999999999999999)).toBe('999,999,999,999,999')
    expect(shown(1000000000000001)).toBe('1,000,000,000,000,001')
    expect(shown(1000000000000001)).toHaveLength(MAX_DISPLAY_CHARS)
  })

  it('switches to exponential once the plain form is too long', () => {
    expect(shown(123456789 * 987654321)).toBe('1.21932631e+17')
    expect(shown(2 ** 60)).toBe('1.1529215e+18')
  })

  it('keeps the sign in exponential notation', () => {
    expect(shown(-(2 ** 60))).toBe('-1.1529215e+18')
  })

  it('trims trailing zeros from the mantissa', () => {
    expect(shown(1e21)).toBe('1e+21')
    expect(shown(1.5e30)).toBe('1.5e+30')
  })

  it('never exceeds the budget for any magnitude', () => {
    const values = [
      1e21,
      -1e21,
      Number.MAX_SAFE_INTEGER,
      2 ** 60,
      -(2 ** 60),
      1.7976931348623157e308,
      -1.7976931348623157e308,
      5e-324,
      1 / 3,
      -1 / 7,
      123456789 * 987654321,
    ]
    for (const value of values) {
      expect(shown(value).length).toBeLessThanOrEqual(MAX_DISPLAY_CHARS)
    }
  })

  it('still reports a non-finite value as an error', () => {
    expect(numberToDisplay(Number.POSITIVE_INFINITY)).toBe('Error')
    expect(numberToDisplay(Number.NaN)).toBe('Error')
  })
})

describe('fitFontSize', () => {
  it('keeps the base size when the text already fits', () => {
    expect(fitFontSize(54, 22, 200, 256)).toBe(54)
    expect(fitFontSize(54, 22, 256, 256)).toBe(54)
  })

  it('scales down in proportion to the overflow', () => {
    expect(fitFontSize(54, 22, 512, 256)).toBe(27)
  })

  it('never goes below the floor', () => {
    expect(fitFontSize(54, 22, 5000, 256)).toBe(22)
  })

  it('never scales up past the base', () => {
    expect(fitFontSize(54, 22, 10, 256)).toBe(54)
  })

  it('leaves the base alone when nothing can be measured', () => {
    expect(fitFontSize(54, 22, 0, 256)).toBe(54)
    expect(fitFontSize(54, 22, 300, 0)).toBe(54)
  })
})

describe('reduce', () => {
  function armed(): CalculatorState {
    return reduce(
      reduce(reduce(INITIAL_STATE, { type: 'digit', value: '6' }), {
        type: 'operation',
        op: 'multiply',
      }),
      { type: 'digit', value: '7' },
    )
  }

  it('defers a binary operation to the server on equals', () => {
    expect(reduce(armed(), { type: 'equals' }).inFlight).toEqual({
      first: 6,
      second: 7,
      op: 'multiply',
      follow: { kind: 'equals' },
    })
  })

  it('routes a unary operation without arming an operator', () => {
    const busy = reduce(reduce(INITIAL_STATE, { type: 'digit', value: '9' }), {
      type: 'operation',
      op: 'sqrt',
    })

    expect(busy.inFlight).toEqual({
      first: 9,
      second: 0,
      op: 'sqrt',
      follow: { kind: 'equals' },
    })
    expect(busy.pendingOp).toBeNull()
  })

  it('ignores further keys while a request is in flight', () => {
    const busy = reduce(armed(), { type: 'equals' })

    expect(reduce(busy, { type: 'digit', value: '5' })).toBe(busy)
    expect(reduce(busy, { type: 'equals' })).toBe(busy)
    expect(reduce(busy, { type: 'operation', op: 'sum' })).toBe(busy)
  })

  it('lets clear through while a request is in flight', () => {
    expect(interrupts({ type: 'clear' })).toBe(true)
    expect(interrupts({ type: 'equals' })).toBe(false)

    const cleared = reduce(reduce(armed(), { type: 'equals' }), { type: 'clear' })
    expect(cleared.inFlight).toBeNull()
    expect(cleared.pendingOp).toBe('multiply')
    expect(cleared.display).toBe('0')
  })

  it('settles an in-flight request with its result', () => {
    const settled = reduce(reduce(armed(), { type: 'equals' }), { type: 'result', result: 42 })

    expect(settled.display).toBe('42')
    expect(settled.inFlight).toBeNull()
    expect(historyLine(settled)).toBe('6 × 7 =')
  })

  it('settles an in-flight request with a failure', () => {
    const failed = reduce(reduce(armed(), { type: 'equals' }), {
      type: 'failure',
      message: 'cannot divide by zero',
    })

    expect(failed.display).toBe('Error')
    expect(failed.error).toBe('cannot divide by zero')
    expect(failed.inFlight).toBeNull()
  })

})
