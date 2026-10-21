import { describe, expect, it } from 'vitest'

import {
  calculatorReducer,
  initialState,
  keyToAction,
  MAX_DIGITS,
  MISSING_OPERAND,
  selectView,
  SQRT_MID_EXPRESSION,
  type CalculatorAction,
  type CalculatorState,
} from './calculatorReducer'

/** Feeds keyboard keys through the reducer, e.g. press('12+3='). */
function press(keys: string, from: CalculatorState = initialState): CalculatorState {
  return [...keys].reduce((state, key) => {
    const action = keyToAction(key)
    if (!action) throw new Error(`unbound key ${key}`)
    return calculatorReducer(state, action)
  }, from)
}

function apply(state: CalculatorState, ...actions: CalculatorAction[]): CalculatorState {
  return actions.reduce(calculatorReducer, state)
}

const resolved = (result: number): CalculatorAction => ({ type: 'resolved', result })

describe('entering numbers', () => {
  it('starts at 0', () => {
    expect(selectView(initialState)).toEqual({ value: '0', expression: '', isLoading: false, error: null })
  })

  it('appends digits and replaces a lone leading zero', () => {
    expect(press('0012').entry).toBe('12')
  })

  it('allows only one decimal point and prefixes a bare point with 0', () => {
    expect(press('.5.2').entry).toBe('0.52')
    expect(press('3..').entry).toBe('3.')
  })

  it(`caps input at ${MAX_DIGITS} digits, not counting sign and point`, () => {
    const full = press('1234567.89012345')
    expect(full.entry).toBe('1234567.89012345')
    expect(press('6', full).entry).toBe(full.entry)
    expect(press('n', full).entry).toBe('-1234567.89012345')
  })

  it('ignores digits that are not 0-9', () => {
    expect(apply(initialState, { type: 'digit', digit: 'a' })).toEqual(initialState)
  })

  it('toggles the sign of the entry, including before any digit', () => {
    expect(press('5n').entry).toBe('-5')
    expect(press('5nn').entry).toBe('5')
    expect(press('n7').entry).toBe('-7')
    expect(press('n.5').entry).toBe('-0.5')
  })

  it('deletes the last character, dropping a lone minus sign', () => {
    const backspace: CalculatorAction = { type: 'backspace' }
    expect(apply(press('123'), backspace).entry).toBe('12')
    expect(apply(press('5n'), backspace).entry).toBe('')
  })

  it('clears everything', () => {
    expect(apply(press('12+3'), { type: 'clear' })).toEqual(initialState)
  })
})

describe('binary operations', () => {
  it('builds a request on =', () => {
    const state = press('12+3=')
    expect(state.pending).toEqual({ request: { operation: 'add', a: 12, b: 3 }, expression: '12 + 3', next: null })
    expect(selectView(state)).toMatchObject({ value: '3', expression: '12 + 3', isLoading: true })
  })

  it.each([
    ['-', 'subtract', '−'],
    ['*', 'multiply', '×'],
    ['/', 'divide', '÷'],
    ['^', 'power', '^'],
    ['%', 'percentage', '% of'],
  ])('maps %s to %s', (key, operation, symbol) => {
    const state = press(`15${key}200=`)
    expect(state.pending?.request).toEqual({ operation, a: 15, b: 200 })
    expect(state.pending?.expression).toBe(`15 ${symbol} 200`)
  })

  it('shows the pending operator after the first operand', () => {
    expect(selectView(press('7*'))).toMatchObject({ value: '7', expression: '7 ×' })
  })

  it('replaces the operator when another one is pressed before the second operand', () => {
    expect(press('7*-').operator).toBe('subtract')
  })

  it('uses 0 as the first operand when an operator comes first', () => {
    expect(press('+5=').pending?.request).toEqual({ operation: 'add', a: 0, b: 5 })
  })

  it('shows the result and the evaluated expression once resolved', () => {
    const state = apply(press('12+3='), resolved(15))
    expect(selectView(state)).toEqual({ value: '15', expression: '12 + 3 =', isLoading: false, error: null })
  })

  it('formats float noise away for display only', () => {
    const state = apply(press('.1+.2='), resolved(0.1 + 0.2))
    expect(selectView(state).value).toBe('0.3')
    expect(state.accumulator).toBe(0.1 + 0.2)
  })

  it('evaluates a chain before continuing with the next operator', () => {
    const chained = press('2+3*')
    expect(chained.pending).toMatchObject({ request: { operation: 'add', a: 2, b: 3 }, next: 'multiply' })

    const continued = apply(chained, resolved(5))
    expect(continued).toMatchObject({ accumulator: 5, operator: 'multiply', entry: '', showingResult: false })
    expect(selectView(continued).expression).toBe('5 ×')
    expect(press('4=', continued).pending?.request).toEqual({ operation: 'multiply', a: 5, b: 4 })
  })

  it('continues from a result when an operator is pressed', () => {
    const state = press('+4=', apply(press('2+3='), resolved(5)))
    expect(state.pending?.request).toEqual({ operation: 'add', a: 5, b: 4 })
  })

  it('starts a new calculation when a digit or point is typed after a result', () => {
    const result = apply(press('2+3='), resolved(5))
    expect(press('7', result)).toEqual({ ...initialState, entry: '7' })
    expect(press('.', result)).toEqual({ ...initialState, entry: '0.' })
  })

  it('negates a result locally', () => {
    const result = apply(press('2+3='), resolved(5))
    expect(press('n', result).accumulator).toBe(-5)
    expect(Object.is(press('n', apply(press('2-2='), resolved(0))).accumulator, 0)).toBe(true)
  })

  it('does nothing on = without an operator, or right after a result', () => {
    expect(press('5=')).toEqual(press('5'))
    const result = apply(press('2+3='), resolved(5))
    expect(press('=', result)).toEqual(result)
  })

  it('rejects = when the second operand is missing', () => {
    const state = press('5+=')
    expect(state.pending).toBeNull()
    expect(selectView(state).error).toBe(MISSING_OPERAND)
  })
})

describe('square root', () => {
  it('takes the root of the entry', () => {
    expect(press('16r').pending).toEqual({ request: { operation: 'sqrt', a: 16 }, expression: '√16', next: null })
  })

  it('takes the root of the previous result', () => {
    const result = apply(press('8*2='), resolved(16))
    expect(press('r', result).pending?.request).toEqual({ operation: 'sqrt', a: 16 })
  })

  it('takes the root of 0 when nothing is entered', () => {
    expect(press('r').pending?.request).toEqual({ operation: 'sqrt', a: 0 })
  })

  it('is rejected in the middle of a binary expression', () => {
    const state = press('2+9r')
    expect(state.pending).toBeNull()
    expect(state.error).toBe(SQRT_MID_EXPRESSION)
  })

  it('shows the result with its expression', () => {
    expect(selectView(apply(press('16r'), resolved(4)))).toMatchObject({ value: '4', expression: '√16 =' })
  })
})

describe('requests in flight', () => {
  it('ignores input until the request settles', () => {
    const loading = press('5/2=')
    expect(press('7+.nr=', loading)).toBe(loading)
    expect(apply(loading, { type: 'backspace' })).toBe(loading)
  })

  it('keeps the inputs and shows the error when the request fails', () => {
    const state = apply(press('5/0='), { type: 'failed', message: "Can't divide by zero." })
    expect(state).toMatchObject({ pending: null, accumulator: 5, operator: 'divide', entry: '0' })
    expect(selectView(state)).toMatchObject({ value: '0', expression: '5 ÷', error: "Can't divide by zero." })
  })

  it('lets the user fix the input after an error, clearing the message', () => {
    const failed = apply(press('5/0='), { type: 'failed', message: 'x' })
    const fixed = apply(failed, { type: 'backspace' }, { type: 'digit', digit: '2' }, { type: 'equals' })
    expect(fixed.error).toBeNull()
    expect(fixed.pending?.request).toEqual({ operation: 'divide', a: 5, b: 2 })
  })

  it('drops the queued operator when a chained request fails', () => {
    const failed = apply(press('5/0*'), { type: 'failed', message: 'x' })
    expect(failed).toMatchObject({ operator: 'divide', entry: '0', pending: null })
  })

  it('ignores outcomes that arrive with nothing pending (e.g. after clear)', () => {
    expect(apply(initialState, resolved(1))).toBe(initialState)
    expect(apply(initialState, { type: 'failed', message: 'x' })).toBe(initialState)
  })
})

describe('keyToAction', () => {
  it.each<[string, CalculatorAction]>([
    ['7', { type: 'digit', digit: '7' }],
    ['.', { type: 'decimal' }],
    [',', { type: 'decimal' }],
    ['x', { type: 'operator', operator: 'multiply' }],
    ['X', { type: 'operator', operator: 'multiply' }],
    ['R', { type: 'sqrt' }],
    ['Enter', { type: 'equals' }],
    ['Backspace', { type: 'backspace' }],
    ['Escape', { type: 'clear' }],
    ['Delete', { type: 'clear' }],
  ])('maps %s', (key, action) => {
    expect(keyToAction(key)).toEqual(action)
  })

  it.each(['a', 'Shift', 'F5', 'constructor', '__proto__'])('ignores %s', (key) => {
    expect(keyToAction(key)).toBeNull()
  })
})
