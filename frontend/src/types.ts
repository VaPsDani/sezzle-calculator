export type Operation =
  | 'add'
  | 'subtract'
  | 'multiply'
  | 'divide'
  | 'power'
  | 'percentage'
  | 'sqrt'

export const UNARY_OPERATIONS: readonly Operation[] = ['sqrt']

/** Codes the API is known to return, plus the two this client produces itself. */
export type KnownErrorCode =
  | 'DIVISION_BY_ZERO'
  | 'NEGATIVE_SQUARE_ROOT'
  | 'RESULT_OUT_OF_RANGE'
  | 'MISSING_FIELD'
  | 'UNEXPECTED_FIELD'
  | 'INVALID_FIELD_TYPE'
  | 'INVALID_JSON'
  | 'EMPTY_BODY'
  | 'UNKNOWN_FIELD'
  | 'INVALID_REQUEST'
  | 'NUMBER_OUT_OF_RANGE'
  | 'REQUEST_TOO_LARGE'
  | 'UNSUPPORTED_MEDIA_TYPE'
  | 'METHOD_NOT_ALLOWED'
  | 'NOT_FOUND'
  | 'INTERNAL_ERROR'
  | 'NETWORK_ERROR'
  | 'UNEXPECTED_RESPONSE'

/** The `string` arm keeps an unrecognised code assignable without a cast. */
export type ErrorCode = KnownErrorCode | (string & {})

export interface CalculationSuccess {
  operation: Operation
  a: number
  b?: number
  result: number
}

export interface CalculationError {
  code: ErrorCode
  message: string
}

/** Discriminated on `ok`, so `if (response.ok)` narrows to one arm or the other. */
export type ApiResponse =
  | { ok: true; data: CalculationSuccess }
  | { ok: false; error: CalculationError }
