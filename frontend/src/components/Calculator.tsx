import { useCalculator } from '../hooks/useCalculator'
import { Display } from './Display'
import { Keypad } from './Keypad'
import './Calculator.css'

export function Calculator() {
  const {
    display,
    pendingOp,
    error,
    isLoading,
    inputDigit,
    inputDecimal,
    chooseOperation,
    calculate,
    clear,
    toggleSign,
    applyUnary,
  } = useCalculator()

  return (
    <section className="calculator" aria-label="Calculator" aria-busy={isLoading}>
      <Display value={display} pendingOp={pendingOp} error={error} />

      <Keypad
        disabled={isLoading}
        activeOperation={pendingOp}
        onDigit={inputDigit}
        onDecimal={inputDecimal}
        onOperation={chooseOperation}
        onUnary={applyUnary}
        onEquals={calculate}
        onClear={clear}
        onToggleSign={toggleSign}
      />
    </section>
  )
}
