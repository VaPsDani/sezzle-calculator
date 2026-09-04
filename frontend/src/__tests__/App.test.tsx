import { render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import App from '../App'

vi.mock('../api/calculatorClient', () => ({ calculate: vi.fn() }))

describe('App', () => {
  it('mounts the calculator', () => {
    render(<App />)

    expect(screen.getByRole('heading', { name: 'Sezzle Calculator' })).toBeTruthy()
    expect(screen.getByLabelText('Calculator')).toBeTruthy()
    expect(screen.getAllByRole('button')).toHaveLength(21)
  })
})
