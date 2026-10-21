import { useCalculator } from '../hooks/useCalculator'
import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcuts'
import type { CalculatorService } from '../services/calculatorService'
import { Display } from './Display'
import { ErrorBanner } from './ErrorBanner'
import { Keypad } from './Keypad'

interface CalculatorProps {
  /** Injectable for tests; defaults to the real API client. */
  service?: CalculatorService
}

export function Calculator({ service }: CalculatorProps) {
  const { view, dispatch, handleKey } = useCalculator(service)
  useKeyboardShortcuts(handleKey)

  return (
    <section className="calculator" aria-label="calculator">
      <Display value={view.value} expression={view.expression} isLoading={view.isLoading} />
      <ErrorBanner message={view.error} />
      <Keypad onAction={dispatch} isLoading={view.isLoading} />
    </section>
  )
}
