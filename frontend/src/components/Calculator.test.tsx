import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { createCalculatorService } from '../services/calculatorService'
import { Calculator } from './Calculator'

// User flows through the real client, hook, reducer and components. Only the
// network is faked: fakeBackend answers like the Go API does.

function json(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function apiError(code: string, message: string): Response {
  return json(422, { error: { code, message } })
}

const operations: Record<string, (a: number, b: number) => number> = {
  add: (a, b) => a + b,
  subtract: (a, b) => a - b,
  multiply: (a, b) => a * b,
  divide: (a, b) => a / b,
  power: (a, b) => a ** b,
  percentage: (a, b) => (a / 100) * b,
  sqrt: (a) => Math.sqrt(a),
}

async function fakeBackend(input: string, init: RequestInit): Promise<Response> {
  const operation = input.split('/').pop() ?? ''
  const { a, b } = JSON.parse(init.body as string) as { a: number; b: number }
  if (operation === 'divide' && b === 0) return apiError('DIVISION_BY_ZERO', 'division by zero is not allowed')
  if (operation === 'sqrt' && a < 0) return apiError('NEGATIVE_SQUARE_ROOT', 'negative')
  return json(200, { operation, result: operations[operation](a, b) })
}

function setup(fetchFn = vi.fn(fakeBackend)) {
  const user = userEvent.setup()
  render(<Calculator service={createCalculatorService(fetchFn)} />)
  const click = async (...names: string[]) => {
    for (const name of names) await user.click(screen.getByRole('button', { name }))
  }
  return { user, click, fetchFn }
}

const result = () => screen.getByRole('status', { name: 'result' })
const expression = () => screen.getByTestId('expression')

describe('Calculator', () => {
  it('shows 0 initially', () => {
    setup()
    expect(result()).toHaveTextContent('0')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('2 + 3 = shows 5', async () => {
    const { click, fetchFn } = setup()

    await click('2', 'add', '3', 'equals')

    await waitFor(() => expect(result()).toHaveTextContent('5'))
    expect(expression()).toHaveTextContent('2 + 3 =')
    expect(fetchFn).toHaveBeenCalledOnce()
  })

  it('5 ÷ 0 = shows the error, and fixing the input recovers', async () => {
    const { click } = setup()

    await click('5', 'divide', '0', 'equals')

    expect(await screen.findByRole('alert')).toHaveTextContent("Can't divide by zero.")
    expect(result()).toHaveTextContent('0')

    await click('backspace', '2', 'equals')

    await waitFor(() => expect(result()).toHaveTextContent('2.5'))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it.each([
    [['1', '0', 'subtract', '4'], '6'],
    [['1', '0', 'multiply', '4'], '40'],
    [['2', 'power', '1', '0'], '1024'],
    [['1', '5', 'percent of', '2', '0', '0'], '30'],
    [['0', 'decimal point', '1', 'add', '0', 'decimal point', '2'], '0.3'],
  ])('%j = shows %s', async (names, expected) => {
    const { click } = setup()

    await click(...names, 'equals')

    await waitFor(() => expect(result()).toHaveTextContent(expected))
  })

  it('takes square roots and reports negative ones', async () => {
    const { click } = setup()

    await click('1', '6', 'square root')
    await waitFor(() => expect(result()).toHaveTextContent('4'))
    expect(expression()).toHaveTextContent('√16 =')

    await click('clear', '4', 'toggle sign', 'square root')
    expect(await screen.findByRole('alert')).toHaveTextContent("Can't take the square root of a negative number.")
  })

  it('works from the keyboard', async () => {
    const { user } = setup()

    await user.keyboard('12*3{Enter}')

    await waitFor(() => expect(result()).toHaveTextContent('36'))
    expect(expression()).toHaveTextContent('12 × 3 =')
  })

  it('leaves browser shortcuts with modifiers alone', async () => {
    const { user, fetchFn } = setup()

    await user.keyboard('9{Control>}r{/Control}')

    expect(result()).toHaveTextContent('9')
    expect(fetchFn).not.toHaveBeenCalled()
  })

  it('validates a missing operand without calling the API', async () => {
    const { click, fetchFn } = setup()

    await click('5', 'add', 'equals')

    expect(screen.getByRole('alert')).toHaveTextContent('Enter a second number first.')
    expect(fetchFn).not.toHaveBeenCalled()
  })

  it('disables = and shows progress while the request is in flight', async () => {
    let respond: (response: Response) => void = () => {}
    const fetchFn = vi.fn(() => new Promise<Response>((resolve) => (respond = resolve)))
    const { click } = setup(fetchFn)

    await click('8', 'divide', '2', 'equals')

    expect(screen.getByRole('button', { name: 'equals' })).toBeDisabled()
    expect(expression()).toHaveTextContent('8 ÷ 2 calculating…')

    respond(json(200, { operation: 'divide', result: 4 }))

    await waitFor(() => expect(result()).toHaveTextContent('4'))
    expect(screen.getByRole('button', { name: 'equals' })).toBeEnabled()
  })

  it('explains network failures', async () => {
    const { click } = setup(vi.fn(() => Promise.reject(new TypeError('Failed to fetch'))))

    await click('1', 'add', '1', 'equals')

    expect(await screen.findByRole('alert')).toHaveTextContent("Can't reach the calculator service")
  })

  it('clears the display and the error', async () => {
    const { click } = setup()

    await click('5', 'divide', '0', 'equals')
    await screen.findByRole('alert')
    await click('clear')

    expect(result()).toHaveTextContent('0')
    expect(expression()).toBeEmptyDOMElement()
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})
