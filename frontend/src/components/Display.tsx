import type { CalculatorView } from '../state/calculatorReducer'

/** Values longer than this get a smaller font so they fit on narrow screens. */
const LONG_VALUE = 11

type DisplayProps = Pick<CalculatorView, 'value' | 'expression' | 'isLoading'>

export function Display({ value, expression, isLoading }: DisplayProps) {
  const valueClass = value.length > LONG_VALUE ? 'display__value display__value--long' : 'display__value'
  return (
    <div className="display" aria-busy={isLoading}>
      <div className="display__expression" data-testid="expression">
        {expression}
        {isLoading && <span className="display__loading"> calculating…</span>}
      </div>
      <output className={valueClass} aria-live="polite" aria-label="result">
        {value}
      </output>
    </div>
  )
}
