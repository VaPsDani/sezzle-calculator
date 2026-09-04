import type { Operation } from '../types'

type KeyAction =
  | { type: 'digit'; value: string }
  | { type: 'decimal' }
  | { type: 'operation'; operation: Operation }
  | { type: 'unary'; operation: Operation }
  | { type: 'equals' }
  | { type: 'clear' }
  | { type: 'toggleSign' }

interface KeyDefinition {
  label: string
  ariaLabel: string
  action: KeyAction
  variant: 'digit' | 'function' | 'operator' | 'equals'
}

const KEYS: readonly KeyDefinition[] = [
  { label: 'C', ariaLabel: 'Clear', action: { type: 'clear' }, variant: 'function' },
  { label: '\u00b1', ariaLabel: 'Toggle sign', action: { type: 'toggleSign' }, variant: 'function' },
  { label: '\u221a', ariaLabel: 'Square root', action: { type: 'unary', operation: 'sqrt' }, variant: 'function' },
  { label: '\u00f7', ariaLabel: 'Divide', action: { type: 'operation', operation: 'divide' }, variant: 'operator' },

  { label: '7', ariaLabel: 'Seven', action: { type: 'digit', value: '7' }, variant: 'digit' },
  { label: '8', ariaLabel: 'Eight', action: { type: 'digit', value: '8' }, variant: 'digit' },
  { label: '9', ariaLabel: 'Nine', action: { type: 'digit', value: '9' }, variant: 'digit' },
  { label: '\u00d7', ariaLabel: 'Multiply', action: { type: 'operation', operation: 'multiply' }, variant: 'operator' },

  { label: '4', ariaLabel: 'Four', action: { type: 'digit', value: '4' }, variant: 'digit' },
  { label: '5', ariaLabel: 'Five', action: { type: 'digit', value: '5' }, variant: 'digit' },
  { label: '6', ariaLabel: 'Six', action: { type: 'digit', value: '6' }, variant: 'digit' },
  { label: '\u2212', ariaLabel: 'Subtract', action: { type: 'operation', operation: 'subtract' }, variant: 'operator' },

  { label: '1', ariaLabel: 'One', action: { type: 'digit', value: '1' }, variant: 'digit' },
  { label: '2', ariaLabel: 'Two', action: { type: 'digit', value: '2' }, variant: 'digit' },
  { label: '3', ariaLabel: 'Three', action: { type: 'digit', value: '3' }, variant: 'digit' },
  { label: '+', ariaLabel: 'Add', action: { type: 'operation', operation: 'add' }, variant: 'operator' },

  { label: 'x\u02b8', ariaLabel: 'Power', action: { type: 'operation', operation: 'power' }, variant: 'operator' },
  { label: '%', ariaLabel: 'Percentage', action: { type: 'operation', operation: 'percentage' }, variant: 'operator' },
  { label: '0', ariaLabel: 'Zero', action: { type: 'digit', value: '0' }, variant: 'digit' },
  { label: '.', ariaLabel: 'Decimal point', action: { type: 'decimal' }, variant: 'digit' },

  { label: '=', ariaLabel: 'Equals', action: { type: 'equals' }, variant: 'equals' },
]

interface KeypadProps {
  disabled: boolean
  activeOperation: Operation | null
  onDigit: (digit: string) => void
  onDecimal: () => void
  onOperation: (operation: Operation) => void
  onUnary: (operation: Operation) => void
  onEquals: () => void
  onClear: () => void
  onToggleSign: () => void
}

export function Keypad({
  disabled,
  activeOperation,
  onDigit,
  onDecimal,
  onOperation,
  onUnary,
  onEquals,
  onClear,
  onToggleSign,
}: KeypadProps) {
  function run(action: KeyAction) {
    switch (action.type) {
      case 'digit':
        return onDigit(action.value)
      case 'decimal':
        return onDecimal()
      case 'operation':
        return onOperation(action.operation)
      case 'unary':
        return onUnary(action.operation)
      case 'equals':
        return onEquals()
      case 'clear':
        return onClear()
      case 'toggleSign':
        return onToggleSign()
    }
  }

  return (
    <div className="keypad">
      {KEYS.map((key) => (
        <button
          key={key.ariaLabel}
          type="button"
          className={`key key--${key.variant}`}
          aria-label={key.ariaLabel}
          aria-pressed={
            key.action.type === 'operation'
              ? activeOperation === key.action.operation
              : undefined
          }
          disabled={disabled}
          onClick={() => run(key.action)}
        >
          {key.label}
        </button>
      ))}
    </div>
  )
}
