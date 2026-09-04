import type { Operation } from '../types'

const OPERATION_SYMBOLS: Record<Operation, string> = {
  add: '+',
  subtract: '\u2212',
  multiply: '\u00d7',
  divide: '\u00f7',
  power: 'x\u02b8',
  percentage: '%',
  sqrt: '\u221a',
}

interface DisplayProps {
  value: string
  pendingOp: Operation | null
  error: string | null
}

export function Display({ value, pendingOp, error }: DisplayProps) {
  return (
    <div className="display">
      <span className="display__pending" aria-hidden="true">
        {pendingOp ? OPERATION_SYMBOLS[pendingOp] : ''}
      </span>

      <output
        className="display__value"
        aria-live="polite"
        aria-label={`Current value ${value}`}
        title={value}
      >
        {value}
      </output>

      {/* Always mounted: a live region has to exist before it changes for a
          screen reader to announce it. */}
      <p className="display__error" role="alert">
        {error}
      </p>
    </div>
  )
}
