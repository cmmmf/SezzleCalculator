import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { App } from '../App'
import { Display } from './Display'
import { ErrorBanner } from './ErrorBanner'
import { Keypad } from './Keypad'

describe('Display', () => {
  it('shows the value and expression', () => {
    render(<Display value="5" expression="2 + 3 =" isLoading={false} />)

    expect(screen.getByRole('status', { name: 'result' })).toHaveTextContent('5')
    expect(screen.getByTestId('expression')).toHaveTextContent('2 + 3 =')
    expect(screen.queryByText(/calculating/)).not.toBeInTheDocument()
  })

  it('shrinks long values', () => {
    const { rerender } = render(<Display value="12345678901" expression="" isLoading={false} />)
    expect(screen.getByRole('status')).not.toHaveClass('display__value--long')

    rerender(<Display value="-1234567.89012345" expression="" isLoading={false} />)
    expect(screen.getByRole('status')).toHaveClass('display__value--long')
  })

  it('marks itself busy while loading', () => {
    const { container } = render(<Display value="2" expression="8 ÷ 4" isLoading />)

    expect(container.firstChild).toHaveAttribute('aria-busy', 'true')
    expect(screen.getByText(/calculating/)).toBeInTheDocument()
  })
})

describe('ErrorBanner', () => {
  it('renders nothing without a message', () => {
    const { container } = render(<ErrorBanner message={null} />)
    expect(container).toBeEmptyDOMElement()
  })

  it('announces the message', () => {
    render(<ErrorBanner message="Can't divide by zero." />)
    expect(screen.getByRole('alert')).toHaveTextContent("Can't divide by zero.")
  })
})

describe('Keypad', () => {
  it('has every key, with accessible names for symbols', () => {
    render(<Keypad onAction={vi.fn()} isLoading={false} />)

    const names = screen.getAllByRole('button').map((b) => b.getAttribute('aria-label') ?? b.textContent)
    expect(names).toEqual([
      'clear', 'backspace', 'percent of', 'divide',
      'square root', 'power', 'toggle sign', 'multiply',
      '7', '8', '9', 'subtract',
      '4', '5', '6', 'add',
      '1', '2', '3', 'equals',
      '0', 'decimal point',
    ])
  })

  it('sends the key action', async () => {
    const onAction = vi.fn()
    render(<Keypad onAction={onAction} isLoading={false} />)

    await userEvent.click(screen.getByRole('button', { name: '7' }))
    await userEvent.click(screen.getByRole('button', { name: 'divide' }))

    expect(onAction.mock.calls).toEqual([
      [{ type: 'digit', digit: '7' }],
      [{ type: 'operator', operator: 'divide' }],
    ])
  })

  it('disables only = while loading', () => {
    render(<Keypad onAction={vi.fn()} isLoading />)

    expect(screen.getByRole('button', { name: 'equals' })).toBeDisabled()
    expect(screen.getByRole('button', { name: 'clear' })).toBeEnabled()
  })
})

describe('App', () => {
  it('renders the calculator', () => {
    render(<App />)

    expect(screen.getByRole('heading', { name: 'Calculator' })).toBeInTheDocument()
    expect(screen.getByRole('region', { name: 'calculator' })).toBeInTheDocument()
  })
})
