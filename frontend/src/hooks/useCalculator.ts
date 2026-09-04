import { useCallback, useState } from 'react'

import { calculate as requestCalculation } from '../api/calculatorClient'
import type { Operation } from '../types'

const MAX_DISPLAY_LENGTH = 15
const PRECISION_DIGITS = 12

export interface CalculatorState {
  display: string
  storedValue: number | null
  pendingOp: Operation | null
  error: string | null
  isLoading: boolean
}

export interface CalculatorActions {
  inputDigit: (digit: string) => void
  inputDecimal: () => void
  chooseOperation: (operation: Operation) => Promise<void>
  calculate: () => Promise<void>
  clear: () => void
  toggleSign: () => void
  applyUnary: (operation: Operation) => Promise<void>
}

export function useCalculator(): CalculatorState & CalculatorActions {
  const [display, setDisplay] = useState('0')
  const [storedValue, setStoredValue] = useState<number | null>(null)
  const [pendingOp, setPendingOp] = useState<Operation | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  // True right after a result or an operator, so the next digit starts a new
  // number instead of appending to the one already shown.
  const [overwrite, setOverwrite] = useState(true)

  const inputDigit = useCallback(
    (digit: string) => {
      setError(null)
      setOverwrite(false)
      setDisplay((current) => {
        if (overwrite || current === '0') return digit
        if (current === '-0') return `-${digit}`
        if (current.length >= MAX_DISPLAY_LENGTH) return current
        return current + digit
      })
    },
    [overwrite],
  )

  const inputDecimal = useCallback(() => {
    setError(null)
    setOverwrite(false)
    setDisplay((current) => {
      if (overwrite) return '0.'
      if (current.includes('.')) return current
      if (current.length >= MAX_DISPLAY_LENGTH) return current
      return `${current}.`
    })
  }, [overwrite])

  const toggleSign = useCallback(() => {
    setError(null)
    setDisplay((current) =>
      current.startsWith('-') ? current.slice(1) : `-${current}`,
    )
  }, [])

  const clear = useCallback(() => {
    setDisplay('0')
    setStoredValue(null)
    setPendingOp(null)
    setError(null)
    setOverwrite(true)
  }, [])

  /** Returns the result, or null when the request failed and error is set. */
  const runOperation = useCallback(
    async (operation: Operation, a: number, b?: number): Promise<number | null> => {
      setIsLoading(true)
      setError(null)

      const response = await requestCalculation(operation, a, b)

      setIsLoading(false)

      if (!response.ok) {
        setError(response.error.message)
        return null
      }
      return response.data.result
    },
    [],
  )

  const chooseOperation = useCallback(
    async (operation: Operation) => {
      if (isLoading) return
      setError(null)

      // Nothing typed since the last operator, so there is nothing to resolve
      // and this press only replaces the operator.
      if (overwrite || storedValue === null || pendingOp === null) {
        setStoredValue(Number(display))
        setPendingOp(operation)
        setOverwrite(true)
        return
      }

      // Both operands are in, so the pending operation is resolved before the
      // next one starts. That is what makes 2 + 3 + 4 = yield 9 and not 7.
      const result = await runOperation(pendingOp, storedValue, Number(display))
      if (result === null) return

      setDisplay(formatResult(result))
      setStoredValue(result)
      setPendingOp(operation)
      setOverwrite(true)
    },
    [display, isLoading, overwrite, pendingOp, runOperation, storedValue],
  )

  const calculate = useCallback(async () => {
    if (isLoading) return

    if (storedValue === null || pendingOp === null) {
      setError('choose an operation first')
      return
    }
    if (overwrite) {
      setError('the second operand is missing')
      return
    }

    const result = await runOperation(pendingOp, storedValue, Number(display))
    if (result === null) return

    setDisplay(formatResult(result))
    setStoredValue(null)
    setPendingOp(null)
    setOverwrite(true)
  }, [display, isLoading, overwrite, pendingOp, runOperation, storedValue])

  const applyUnary = useCallback(
    async (operation: Operation) => {
      if (isLoading) return

      const result = await runOperation(operation, Number(display))
      if (result === null) return

      setDisplay(formatResult(result))
      setOverwrite(true)
    },
    [display, isLoading, runOperation],
  )

  return {
    display,
    storedValue,
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
  }
}

// A float64 result can need 17 digits to round trip, which does not fit the
// display, so long values are shortened before losing the number entirely.
function formatResult(value: number): string {
  const plain = String(value)
  if (plain.length <= MAX_DISPLAY_LENGTH) return plain

  const rounded = String(Number(value.toPrecision(PRECISION_DIGITS)))
  return rounded.length <= MAX_DISPLAY_LENGTH ? rounded : value.toExponential(8)
}
