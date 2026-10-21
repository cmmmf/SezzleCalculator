// Pure keypad state machine. It never calls the API: pressing `=` (or an
// operator that completes an expression, or √) sets `pending` to the request
// to send, and useCalculator performs it and dispatches `resolved`/`failed`.

import type { BinaryOperation, CalculationRequest } from '../services/calculatorService'
import { formatNumber } from './formatNumber'

/** 15 digits always round-trip exactly through a float64. */
export const MAX_DIGITS = 15

export const operatorSymbols: Record<BinaryOperation, string> = {
  add: '+',
  subtract: '−',
  multiply: '×',
  divide: '÷',
  power: '^',
  percentage: '% of',
}

export interface PendingCalculation {
  request: CalculationRequest
  /** Expression shown while waiting and, once resolved, above the result. */
  expression: string
  /** Operator to continue with after the result arrives (`2 + 3 ×` → `5 ×`). */
  next: BinaryOperation | null
}

export interface CalculatorState {
  /** The number being typed, exactly as typed ('' when nothing is typed). */
  entry: string
  /** Left operand, or the last result. */
  accumulator: number | null
  operator: BinaryOperation | null
  /** True while `accumulator` holds a result that hasn't been used yet. */
  showingResult: boolean
  /** Completed expression shown above a result, e.g. "2 + 3 =". */
  history: string
  pending: PendingCalculation | null
  error: string | null
}

export type CalculatorAction =
  | { type: 'digit'; digit: string }
  | { type: 'decimal' }
  | { type: 'operator'; operator: BinaryOperation }
  | { type: 'sqrt' }
  | { type: 'toggleSign' }
  | { type: 'backspace' }
  | { type: 'equals' }
  | { type: 'clear' }
  | { type: 'resolved'; result: number }
  | { type: 'failed'; message: string }

export const initialState: CalculatorState = {
  entry: '',
  accumulator: null,
  operator: null,
  showingResult: false,
  history: '',
  pending: null,
  error: null,
}

// Client-side validation messages.
export const MISSING_OPERAND = 'Enter a second number first.'
export const SQRT_MID_EXPRESSION = 'Press = before taking a square root.'

export function calculatorReducer(state: CalculatorState, action: CalculatorAction): CalculatorState {
  switch (action.type) {
    case 'clear':
      return initialState
    case 'resolved':
      return state.pending ? resolve(state, state.pending, action.result) : state
    case 'failed':
      // Keep the inputs so the user can fix them (e.g. change the 0 in 5 ÷ 0).
      return state.pending ? { ...state, pending: null, error: action.message } : state
  }

  // Input is ignored while a request is in flight; only `clear` cancels it.
  if (state.pending) return state
  const s: CalculatorState = { ...state, error: null }

  switch (action.type) {
    case 'digit':
      return inputDigit(s, action.digit)
    case 'decimal':
      return inputDecimal(s)
    case 'operator':
      return chooseOperator(s, action.operator)
    case 'sqrt':
      return squareRoot(s)
    case 'toggleSign':
      return toggleSign(s)
    case 'backspace':
      return backspace(s)
    case 'equals':
      return equals(s)
  }
}

function inputDigit(s: CalculatorState, digit: string): CalculatorState {
  if (!/^\d$/.test(digit)) return s
  if (s.showingResult) return { ...initialState, entry: digit }
  if (countDigits(s.entry) >= MAX_DIGITS) return s
  // Replace a lone leading zero: "0" + "5" → "5", "-0" + "5" → "-5".
  const entry = s.entry === '0' || s.entry === '-0' ? s.entry.slice(0, -1) + digit : s.entry + digit
  return { ...s, entry }
}

function inputDecimal(s: CalculatorState): CalculatorState {
  if (s.showingResult) return { ...initialState, entry: '0.' }
  if (s.entry.includes('.')) return s
  if (s.entry === '' || s.entry === '-') return { ...s, entry: s.entry + '0.' }
  return { ...s, entry: s.entry + '.' }
}

function chooseOperator(s: CalculatorState, operator: BinaryOperation): CalculatorState {
  if (s.entry === '') {
    // Change the operator, or start from the previous result (or 0).
    return { ...s, operator, accumulator: s.accumulator ?? 0, showingResult: false }
  }
  if (s.operator !== null && s.accumulator !== null) {
    // `2 + 3 ×`: evaluate 2 + 3 first, then continue with ×.
    return withPending(s, binaryCalculation(s.accumulator, s.operator, s.entry), operator)
  }
  return { ...s, accumulator: Number(s.entry), entry: '', operator, showingResult: false }
}

function squareRoot(s: CalculatorState): CalculatorState {
  if (s.operator !== null) return { ...s, error: SQRT_MID_EXPRESSION }
  const a = s.entry !== '' ? Number(s.entry) : (s.accumulator ?? 0)
  return withPending(s, {
    request: { operation: 'sqrt', a },
    expression: `√${formatNumber(a)}`,
  })
}

function toggleSign(s: CalculatorState): CalculatorState {
  if (s.entry === '' && s.showingResult && s.accumulator !== null) {
    // Negating is exact, so the result can be flipped locally.
    return { ...s, accumulator: s.accumulator === 0 ? 0 : -s.accumulator, history: '' }
  }
  if (s.entry === '') return { ...s, entry: '-0' }
  return { ...s, entry: s.entry.startsWith('-') ? s.entry.slice(1) : '-' + s.entry }
}

function backspace(s: CalculatorState): CalculatorState {
  const entry = s.entry.slice(0, -1)
  return { ...s, entry: entry === '-' ? '' : entry }
}

function equals(s: CalculatorState): CalculatorState {
  if (s.operator === null || s.accumulator === null) return s
  if (s.entry === '') return { ...s, error: MISSING_OPERAND }
  return withPending(s, binaryCalculation(s.accumulator, s.operator, s.entry))
}

function binaryCalculation(
  a: number,
  operator: BinaryOperation,
  entry: string,
): Omit<PendingCalculation, 'next'> {
  const b = Number(entry)
  return {
    request: { operation: operator, a, b },
    expression: `${formatNumber(a)} ${operatorSymbols[operator]} ${formatNumber(b)}`,
  }
}

function withPending(
  s: CalculatorState,
  calc: Omit<PendingCalculation, 'next'>,
  next: BinaryOperation | null = null,
): CalculatorState {
  return { ...s, pending: { ...calc, next } }
}

function resolve(s: CalculatorState, pending: PendingCalculation, result: number): CalculatorState {
  const base = { ...s, accumulator: result, entry: '', pending: null, error: null }
  if (pending.next) {
    return { ...base, operator: pending.next, showingResult: false, history: '' }
  }
  return { ...base, operator: null, showingResult: true, history: `${pending.expression} =` }
}

function countDigits(entry: string): number {
  return entry.replace(/\D/g, '').length
}

export interface CalculatorView {
  /** The main display line. */
  value: string
  /** The smaller line above it: the expression being built or just evaluated. */
  expression: string
  isLoading: boolean
  error: string | null
}

/** Derives what the display shows from the state. */
export function selectView(s: CalculatorState): CalculatorView {
  const value = s.entry !== '' ? s.entry : s.accumulator !== null ? formatNumber(s.accumulator) : '0'

  let expression = s.history
  if (s.pending) {
    expression = s.pending.expression
  } else if (s.operator !== null && s.accumulator !== null) {
    expression = `${formatNumber(s.accumulator)} ${operatorSymbols[s.operator]}`
  }

  return { value, expression, isLoading: s.pending !== null, error: s.error }
}

const keyActions = new Map<string, CalculatorAction>([
  ['.', { type: 'decimal' }],
  [',', { type: 'decimal' }],
  ['+', { type: 'operator', operator: 'add' }],
  ['-', { type: 'operator', operator: 'subtract' }],
  ['*', { type: 'operator', operator: 'multiply' }],
  ['x', { type: 'operator', operator: 'multiply' }],
  ['/', { type: 'operator', operator: 'divide' }],
  ['^', { type: 'operator', operator: 'power' }],
  ['%', { type: 'operator', operator: 'percentage' }],
  ['r', { type: 'sqrt' }],
  ['n', { type: 'toggleSign' }],
  ['=', { type: 'equals' }],
  ['Enter', { type: 'equals' }],
  ['Backspace', { type: 'backspace' }],
  ['Escape', { type: 'clear' }],
  ['Delete', { type: 'clear' }],
])

/** Maps a KeyboardEvent.key to an action, or null if the key isn't bound. */
export function keyToAction(key: string): CalculatorAction | null {
  if (/^\d$/.test(key)) return { type: 'digit', digit: key }
  return keyActions.get(key) ?? keyActions.get(key.toLowerCase()) ?? null
}
