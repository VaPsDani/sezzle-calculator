import { afterEach, describe, expect, it, vi } from 'vitest'

import { calculate } from '../api/calculatorClient'

afterEach(() => {
  vi.unstubAllGlobals()
})

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function stubFetch(response: Response | Error) {
  const mock = vi.fn((_input: RequestInfo | URL, _init?: RequestInit) =>
    response instanceof Error ? Promise.reject(response) : Promise.resolve(response),
  )
  vi.stubGlobal('fetch', mock)
  return mock
}

describe('calculate', () => {
  it('returns the result of a 200 response', async () => {
    stubFetch(jsonResponse({ operation: 'divide', a: 10, b: 4, result: 2.5 }))

    const response = await calculate('divide', 10, 4)

    expect(response).toEqual({
      ok: true,
      data: { operation: 'divide', a: 10, b: 4, result: 2.5 },
    })
  })

  it('returns the typed error of a 422 response', async () => {
    stubFetch(
      jsonResponse(
        { error: { code: 'DIVISION_BY_ZERO', message: 'division by zero' } },
        422,
      ),
    )

    const response = await calculate('divide', 10, 0)

    expect(response).toEqual({
      ok: false,
      error: { code: 'DIVISION_BY_ZERO', message: 'division by zero' },
    })
  })

  it('handles a network failure without throwing', async () => {
    stubFetch(new TypeError('Failed to fetch'))

    const response = await calculate('add', 1, 2)

    expect(response).toEqual({
      ok: false,
      error: { code: 'NETWORK_ERROR', message: expect.any(String) },
    })
  })

  it('builds the URL and body of a binary operation', async () => {
    const fetchMock = stubFetch(
      jsonResponse({ operation: 'power', a: 2, b: 10, result: 1024 }),
    )

    await calculate('power', 2, 10)

    expect(fetchMock).toHaveBeenCalledTimes(1)
    const [url, init] = fetchMock.mock.calls[0]

    expect(url).toBe('/api/v1/power')
    expect(init?.method).toBe('POST')
    expect(init?.headers).toEqual({ 'Content-Type': 'application/json' })
    expect(JSON.parse(init?.body as string)).toEqual({ a: 2, b: 10 })
  })

  it('omits the second operand for a unary operation', async () => {
    const fetchMock = stubFetch(jsonResponse({ operation: 'sqrt', a: 9, result: 3 }))

    await calculate('sqrt', 9)

    const [url, init] = fetchMock.mock.calls[0]

    expect(url).toBe('/api/v1/sqrt')
    expect(JSON.parse(init?.body as string)).toEqual({ a: 9 })
    expect(JSON.parse(init?.body as string)).not.toHaveProperty('b')
  })
})
