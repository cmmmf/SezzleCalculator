// Typed client for the calculator REST API (see backend/internal/docs/openapi.yaml).

export type BinaryOperation = 'add' | 'subtract' | 'multiply' | 'divide' | 'power' | 'percentage'
export type UnaryOperation = 'sqrt'

export type CalculationRequest =
  | { operation: BinaryOperation; a: number; b: number }
  | { operation: UnaryOperation; a: number }

/** Error codes returned by the API. Clients branch on these, never on messages. */
export type ApiErrorCode =
  | 'INVALID_REQUEST'
  | 'DIVISION_BY_ZERO'
  | 'NEGATIVE_SQUARE_ROOT'
  | 'RESULT_OUT_OF_RANGE'
  | 'UNDEFINED_RESULT'
  | 'NOT_FOUND'
  | 'METHOD_NOT_ALLOWED'
  | 'PAYLOAD_TOO_LARGE'
  | 'INTERNAL_ERROR'

/** Failures detected on the client, plus API codes this client doesn't know yet. */
export type ClientErrorCode = 'NETWORK_ERROR' | 'UNEXPECTED_RESPONSE' | 'UNKNOWN_ERROR'

export type CalculatorErrorCode = ApiErrorCode | ClientErrorCode

const errorMessages: Record<CalculatorErrorCode, string> = {
  INVALID_REQUEST: 'Please enter valid numbers.',
  DIVISION_BY_ZERO: "Can't divide by zero.",
  NEGATIVE_SQUARE_ROOT: "Can't take the square root of a negative number.",
  RESULT_OUT_OF_RANGE: 'The result is too large to calculate.',
  UNDEFINED_RESULT: 'The result is undefined for these numbers.',
  NOT_FOUND: "This operation isn't supported.",
  METHOD_NOT_ALLOWED: "This operation isn't supported.",
  PAYLOAD_TOO_LARGE: 'The numbers are too long.',
  INTERNAL_ERROR: 'Something went wrong on our side. Please try again.',
  NETWORK_ERROR: "Can't reach the calculator service. Check your connection and try again.",
  UNEXPECTED_RESPONSE: 'The calculator service sent an unexpected response.',
  UNKNOWN_ERROR: 'Something went wrong. Please try again.',
}

/** A failed calculation. `message` is safe to show to the user. */
export class CalculatorError extends Error {
  readonly code: CalculatorErrorCode

  constructor(code: CalculatorErrorCode, message: string = errorMessages[code], options?: ErrorOptions) {
    super(message, options)
    this.name = 'CalculatorError'
    this.code = code
  }
}

export interface CalculatorService {
  /** Resolves with the result, or rejects with a CalculatorError (or the AbortError when `signal` aborts). */
  calculate(request: CalculationRequest, signal?: AbortSignal): Promise<number>
}

type FetchFn = (input: string, init: RequestInit) => Promise<Response>

/**
 * Creates a client. `fetchFn` is injectable for tests; `baseUrl` is empty
 * because the Vite dev server and nginx both proxy /api to the backend.
 */
export function createCalculatorService(
  fetchFn: FetchFn = (input, init) => fetch(input, init),
  baseUrl = '',
): CalculatorService {
  return {
    async calculate(request, signal) {
      const { operation, ...operands } = request

      let response: Response
      try {
        response = await fetchFn(`${baseUrl}/api/v1/${operation}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(operands),
          signal,
        })
      } catch (err) {
        if (signal?.aborted) throw err
        throw new CalculatorError('NETWORK_ERROR', undefined, { cause: err })
      }

      const body: unknown = await response.json().catch(() => undefined)
      if (response.ok && isOperationResponse(body)) {
        return body.result
      }
      if (!response.ok && isErrorResponse(body)) {
        const { code, message } = body.error
        throw isApiErrorCode(code) ? new CalculatorError(code) : new CalculatorError('UNKNOWN_ERROR', message)
      }
      throw new CalculatorError('UNEXPECTED_RESPONSE')
    },
  }
}

/** The client used by the app. */
export const calculatorService = createCalculatorService()

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function isOperationResponse(body: unknown): body is { operation: string; result: number } {
  return isRecord(body) && typeof body.operation === 'string' && typeof body.result === 'number'
}

function isErrorResponse(body: unknown): body is { error: { code: string; message: string } } {
  return (
    isRecord(body) &&
    isRecord(body.error) &&
    typeof body.error.code === 'string' &&
    typeof body.error.message === 'string'
  )
}

const apiErrorCodes: ReadonlySet<string> = new Set<ApiErrorCode>([
  'INVALID_REQUEST',
  'DIVISION_BY_ZERO',
  'NEGATIVE_SQUARE_ROOT',
  'RESULT_OUT_OF_RANGE',
  'UNDEFINED_RESULT',
  'NOT_FOUND',
  'METHOD_NOT_ALLOWED',
  'PAYLOAD_TOO_LARGE',
  'INTERNAL_ERROR',
])

function isApiErrorCode(code: string): code is ApiErrorCode {
  return apiErrorCodes.has(code)
}
