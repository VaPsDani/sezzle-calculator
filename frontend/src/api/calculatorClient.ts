import type {
  ApiResponse,
  CalculationError,
  CalculationSuccess,
  ErrorCode,
  Operation,
} from '../types'

const BASE_PATH = '/api/v1'

/**
 * Relative URL on purpose: the Vite proxy in development and the reverse proxy
 * in production both forward /api, so the same build works in either.
 */
export async function calculate(
  operation: Operation,
  a: number,
  b?: number,
): Promise<ApiResponse> {
  let response: Response

  try {
    response = await fetch(`${BASE_PATH}/${operation}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(b === undefined ? { a } : { a, b }),
    })
  } catch {
    return failure(
      'NETWORK_ERROR',
      'No se pudo conectar con el servidor. Revisa tu conexión.',
    )
  }

  const body = await readJson(response)

  if (response.ok) {
    return isCalculationSuccess(body)
      ? { ok: true, data: body }
      : failure('UNEXPECTED_RESPONSE', 'El servidor devolvió una respuesta inesperada.')
  }

  if (isErrorEnvelope(body)) {
    return { ok: false, error: body.error }
  }

  return failure(
    'UNEXPECTED_RESPONSE',
    `El servidor respondió con el estado ${response.status}.`,
  )
}

function failure(code: ErrorCode, message: string): ApiResponse {
  return { ok: false, error: { code, message } }
}

/** A body that is empty or not JSON must not turn into a thrown exception. */
async function readJson(response: Response): Promise<unknown> {
  try {
    return await response.json()
  } catch {
    return undefined
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function isCalculationSuccess(body: unknown): body is CalculationSuccess {
  return (
    isRecord(body) &&
    typeof body.operation === 'string' &&
    typeof body.a === 'number' &&
    typeof body.result === 'number' &&
    (body.b === undefined || typeof body.b === 'number')
  )
}

function isErrorEnvelope(body: unknown): body is { error: CalculationError } {
  return (
    isRecord(body) &&
    isRecord(body.error) &&
    typeof body.error.code === 'string' &&
    typeof body.error.message === 'string'
  )
}
