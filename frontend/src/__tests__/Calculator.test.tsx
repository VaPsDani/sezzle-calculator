import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { calculate as requestCalculation } from '../api/calculatorClient'
import { Calculator } from '../components/Calculator'
import type { ApiResponse } from '../types'

vi.mock('../api/calculatorClient', () => ({ calculate: vi.fn() }))

const requestMock = vi.mocked(requestCalculation)

beforeEach(() => {
  requestMock.mockReset()
})

/** Buttons are found by accessible name, so restyling cannot break the tests. */
function key(name: string): HTMLButtonElement {
  return screen.getByRole('button', { name }) as HTMLButtonElement
}

function displayValue(): string {
  return screen.getByRole('status').textContent ?? ''
}

describe('Calculator', () => {
  it('shows 15 after the user presses 7, +, 8, =', async () => {
    const user = userEvent.setup()
    requestMock.mockResolvedValue({
      ok: true,
      data: { operation: 'add', a: 7, b: 8, result: 15 },
    })

    render(<Calculator />)

    await user.click(key('Seven'))
    await user.click(key('Add'))
    await user.click(key('Eight'))
    await user.click(key('Equals'))

    expect(requestMock).toHaveBeenCalledExactlyOnceWith('add', 7, 8)
    expect(displayValue()).toBe('15')
  })

  it('announces a readable message in the alert region when the API rejects the operation', async () => {
    const user = userEvent.setup()
    requestMock.mockResolvedValue({
      ok: false,
      error: { code: 'DIVISION_BY_ZERO', message: 'division by zero' },
    })

    render(<Calculator />)

    expect(screen.getByRole('alert').textContent).toBe('')

    await user.click(key('One'))
    await user.click(key('Divide'))
    await user.click(key('Zero'))
    await user.click(key('Equals'))

    expect(screen.getByRole('alert').textContent).toContain('division by zero')
  })

  it('disables every button while a request is in flight', async () => {
    const user = userEvent.setup()

    let settle!: (response: ApiResponse) => void
    requestMock.mockReturnValue(
      new Promise<ApiResponse>((resolve) => {
        settle = resolve
      }),
    )

    render(<Calculator />)

    await user.click(key('Nine'))
    expect(key('Nine').disabled).toBe(false)

    await user.click(key('Square root'))

    const buttons = screen.getAllByRole('button') as HTMLButtonElement[]
    expect(buttons.every((button) => button.disabled)).toBe(true)
    expect(screen.getByLabelText('Calculator').getAttribute('aria-busy')).toBe('true')

    await act(async () => {
      settle({ ok: true, data: { operation: 'sqrt', a: 9, result: 3 } })
    })

    expect(screen.getAllByRole('button').some((b) => (b as HTMLButtonElement).disabled)).toBe(false)
    expect(displayValue()).toBe('3')
  })

  it('resets the display when the user presses C', async () => {
    const user = userEvent.setup()

    render(<Calculator />)

    await user.click(key('Four'))
    await user.click(key('Two'))
    expect(displayValue()).toBe('42')

    await user.click(key('Clear'))

    expect(displayValue()).toBe('0')
    expect(requestMock).not.toHaveBeenCalled()
  })
})
