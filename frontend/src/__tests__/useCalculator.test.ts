import { act, renderHook } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { calculate as requestCalculation } from '../api/calculatorClient'
import { useCalculator } from '../hooks/useCalculator'

vi.mock('../api/calculatorClient', () => ({ calculate: vi.fn() }))

const requestMock = vi.mocked(requestCalculation)

beforeEach(() => {
  requestMock.mockReset()
})

describe('useCalculator', () => {
  it('appends digits to the display', () => {
    const { result } = renderHook(() => useCalculator())

    act(() => result.current.inputDigit('1'))
    act(() => result.current.inputDigit('2'))
    act(() => result.current.inputDigit('3'))

    expect(result.current.display).toBe('123')
  })

  it('accepts only one decimal point', () => {
    const { result } = renderHook(() => useCalculator())

    act(() => result.current.inputDigit('3'))
    act(() => result.current.inputDecimal())
    act(() => result.current.inputDigit('1'))
    act(() => result.current.inputDecimal())
    act(() => result.current.inputDigit('4'))

    expect(result.current.display).toBe('3.14')
  })

  it('resets every piece of state on clear', async () => {
    const { result } = renderHook(() => useCalculator())

    act(() => result.current.inputDigit('7'))
    await act(() => result.current.chooseOperation('divide'))
    await act(() => result.current.calculate())

    expect(result.current.storedValue).toBe(7)
    expect(result.current.pendingOp).toBe('divide')
    expect(result.current.error).not.toBeNull()

    act(() => result.current.clear())

    expect(result.current).toMatchObject({
      display: '0',
      storedValue: null,
      pendingOp: null,
      error: null,
      isLoading: false,
    })
  })

  it('sends the right arguments when 2 + 3 is resolved with equals', async () => {
    requestMock.mockResolvedValue({
      ok: true,
      data: { operation: 'add', a: 2, b: 3, result: 5 },
    })

    const { result } = renderHook(() => useCalculator())

    act(() => result.current.inputDigit('2'))
    await act(() => result.current.chooseOperation('add'))
    act(() => result.current.inputDigit('3'))
    await act(() => result.current.calculate())

    expect(requestMock).toHaveBeenCalledExactlyOnceWith('add', 2, 3)
    expect(result.current.display).toBe('5')
    expect(result.current.pendingOp).toBeNull()
  })

  it('resolves the pending operation when a second operator is chained', async () => {
    requestMock.mockResolvedValue({
      ok: true,
      data: { operation: 'add', a: 2, b: 3, result: 5 },
    })

    const { result } = renderHook(() => useCalculator())

    act(() => result.current.inputDigit('2'))
    await act(() => result.current.chooseOperation('add'))
    act(() => result.current.inputDigit('3'))
    await act(() => result.current.chooseOperation('multiply'))

    expect(requestMock).toHaveBeenCalledExactlyOnceWith('add', 2, 3)
    expect(result.current.display).toBe('5')
    expect(result.current.storedValue).toBe(5)
    expect(result.current.pendingOp).toBe('multiply')
  })

  it('surfaces a failed request as an error and keeps the pending operation', async () => {
    requestMock.mockResolvedValue({
      ok: false,
      error: { code: 'DIVISION_BY_ZERO', message: 'division by zero' },
    })

    const { result } = renderHook(() => useCalculator())

    act(() => result.current.inputDigit('1'))
    act(() => result.current.inputDigit('0'))
    await act(() => result.current.chooseOperation('divide'))
    act(() => result.current.inputDigit('0'))
    await act(() => result.current.calculate())

    expect(result.current.error).toBe('division by zero')
    expect(result.current.pendingOp).toBe('divide')
    expect(result.current.storedValue).toBe(10)
    expect(result.current.isLoading).toBe(false)
  })
})
