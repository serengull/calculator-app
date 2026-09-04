import type { KeyAction } from './calculator'

export type KeyKind = 'digit' | 'function' | 'operator'

export interface KeySpec {
  id: string
  label: string
  ariaLabel: string
  kind: KeyKind
  action: KeyAction
  chars?: readonly string[]
  wide?: boolean
  keyboardOnly?: boolean
}

function digit(value: string, wide?: boolean): KeySpec {
  return {
    id: value,
    label: value,
    ariaLabel: value,
    kind: 'digit',
    action: { type: 'digit', value },
    chars: [value],
    wide,
  }
}

export const KEYS: readonly KeySpec[] = [
  {
    id: 'exponentiation',
    label: 'xʸ',
    ariaLabel: 'Exponentiation',
    kind: 'function',
    wide: true,
    action: { type: 'operation', op: 'exponentiation' },
    chars: ['^'],
  },
  {
    id: 'sqrt',
    label: '√x',
    ariaLabel: 'Square root',
    kind: 'function',
    wide: true,
    action: { type: 'operation', op: 'sqrt' },
    chars: ['r'],
  },
  {
    id: 'clear',
    label: 'AC',
    ariaLabel: 'All clear',
    kind: 'function',
    action: { type: 'clear' },
    chars: ['Escape', 'Delete'],
  },
  { id: 'negate', label: '+/−', ariaLabel: 'Negate', kind: 'function', action: { type: 'negate' } },
  {
    id: 'percentage',
    label: '%',
    ariaLabel: 'Percent',
    kind: 'function',
    action: { type: 'operation', op: 'percentage' },
    chars: ['%'],
  },
  {
    id: 'division',
    label: '÷',
    ariaLabel: 'Divide',
    kind: 'operator',
    action: { type: 'operation', op: 'division' },
    chars: ['/'],
  },
  digit('7'),
  digit('8'),
  digit('9'),
  {
    id: 'multiply',
    label: '×',
    ariaLabel: 'Multiply',
    kind: 'operator',
    action: { type: 'operation', op: 'multiply' },
    chars: ['*', 'x'],
  },
  digit('4'),
  digit('5'),
  digit('6'),
  {
    id: 'subtract',
    label: '−',
    ariaLabel: 'Subtract',
    kind: 'operator',
    action: { type: 'operation', op: 'subtract' },
    chars: ['-'],
  },
  digit('1'),
  digit('2'),
  digit('3'),
  {
    id: 'sum',
    label: '+',
    ariaLabel: 'Add',
    kind: 'operator',
    action: { type: 'operation', op: 'sum' },
    chars: ['+'],
  },
  digit('0', true),
  {
    id: 'dot',
    label: '.',
    ariaLabel: 'Decimal point',
    kind: 'digit',
    action: { type: 'dot' },
    chars: ['.', ','],
  },
  {
    id: 'equals',
    label: '=',
    ariaLabel: 'Equals',
    kind: 'operator',
    action: { type: 'equals' },
    chars: ['=', 'Enter'],
  },
  {
    id: 'backspace',
    label: '⌫',
    ariaLabel: 'Backspace',
    kind: 'function',
    action: { type: 'backspace' },
    chars: ['Backspace'],
    keyboardOnly: true,
  },
]

export const PAD_KEYS: readonly KeySpec[] = KEYS.filter((key) => !key.keyboardOnly)

export const KEY_BY_CHARACTER: Record<string, KeySpec> = Object.fromEntries(
  KEYS.flatMap((key) => (key.chars ?? []).map((char) => [char, key] as const)),
)
