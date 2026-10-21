import type { CalculatorAction } from '../state/calculatorReducer'

interface Key {
  label: string
  /** Accessible name, for keys whose label is a symbol. */
  name?: string
  action: CalculatorAction
  variant?: 'operator' | 'function' | 'equals' | 'wide'
}

const digit = (d: string): Key => ({ label: d, action: { type: 'digit', digit: d } })

// Laid out row by row on a 4-column grid; `=` spans the last two rows.
const keys: Key[] = [
  { label: 'AC', name: 'clear', action: { type: 'clear' }, variant: 'function' },
  { label: '⌫', name: 'backspace', action: { type: 'backspace' }, variant: 'function' },
  { label: '%', name: 'percent of', action: { type: 'operator', operator: 'percentage' }, variant: 'operator' },
  { label: '÷', name: 'divide', action: { type: 'operator', operator: 'divide' }, variant: 'operator' },

  { label: '√', name: 'square root', action: { type: 'sqrt' }, variant: 'function' },
  { label: 'xʸ', name: 'power', action: { type: 'operator', operator: 'power' }, variant: 'operator' },
  { label: '±', name: 'toggle sign', action: { type: 'toggleSign' }, variant: 'function' },
  { label: '×', name: 'multiply', action: { type: 'operator', operator: 'multiply' }, variant: 'operator' },

  digit('7'),
  digit('8'),
  digit('9'),
  { label: '−', name: 'subtract', action: { type: 'operator', operator: 'subtract' }, variant: 'operator' },

  digit('4'),
  digit('5'),
  digit('6'),
  { label: '+', name: 'add', action: { type: 'operator', operator: 'add' }, variant: 'operator' },

  digit('1'),
  digit('2'),
  digit('3'),
  { label: '=', name: 'equals', action: { type: 'equals' }, variant: 'equals' },

  { ...digit('0'), variant: 'wide' },
  { label: '.', name: 'decimal point', action: { type: 'decimal' } },
]

interface KeypadProps {
  onAction: (action: CalculatorAction) => void
  /** Disables `=` while a request is in flight. */
  isLoading: boolean
}

export function Keypad({ onAction, isLoading }: KeypadProps) {
  return (
    <div className="keypad" role="group" aria-label="keypad">
      {keys.map((key) => (
        <button
          key={key.label}
          type="button"
          className={key.variant ? `key key--${key.variant}` : 'key'}
          aria-label={key.name}
          disabled={isLoading && key.action.type === 'equals'}
          onClick={() => onAction(key.action)}
        >
          {key.label}
        </button>
      ))}
    </div>
  )
}
