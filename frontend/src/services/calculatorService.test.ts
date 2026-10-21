import { describe, expect, it, vi } from 'vitest'

import { CalculatorError, createCalculatorService, type CalculatorErrorCode } from './calculatorService'

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })
}

function serviceReturning(response: Response | Promise<Response>) {
  const fetchFn = vi.fn((_input: string, _init: RequestInit) => Promise.resolve(response))
  return { service: createCalculatorService(fetchFn), fetchFn }
}

async function rejection(promise: Promise<unknown>): Promise<CalculatorError> {
  const err = await promise.then(
    () => {
      throw new Error('expected the promise to reject')
    },
    (e: unknown) => e,
  )
  expect(err).toBeInstanceOf(CalculatorError)
  return err as CalculatorError
}

describe('calculatorService', () => {
  it('POSTs binary operands as JSON to the operation route', async () => {
    const { service, fetchFn } = serviceReturning(jsonResponse(200, { operation: 'divide', result: 2.5 }))

    await expect(service.calculate({ operation: 'divide', a: 10, b: 4 })).resolves.toBe(2.5)

    expect(fetchFn).toHaveBeenCalledOnce()
    const [url, init] = fetchFn.mock.calls[0]
    expect(url).toBe('/api/v1/divide')
    expect(init.method).toBe('POST')
    expect(init.headers).toEqual({ 'Content-Type': 'application/json' })
    expect(JSON.parse(init.body as string)).toEqual({ a: 10, b: 4 })
  })

  it('sends only `a` for sqrt', async () => {
    const { service, fetchFn } = serviceReturning(jsonResponse(200, { operation: 'sqrt', result: 4 }))

    await expect(service.calculate({ operation: 'sqrt', a: 16 })).resolves.toBe(4)

    const [url, init] = fetchFn.mock.calls[0]
    expect(url).toBe('/api/v1/sqrt')
    expect(JSON.parse(init.body as string)).toEqual({ a: 16 })
  })

  it('prefixes the base URL and forwards the abort signal', async () => {
    const fetchFn = vi.fn((_input: string, _init: RequestInit) =>
      Promise.resolve(jsonResponse(200, { operation: 'add', result: 3 })),
    )
    const signal = new AbortController().signal

    await createCalculatorService(fetchFn, 'http://api.test').calculate({ operation: 'add', a: 1, b: 2 }, signal)

    expect(fetchFn.mock.calls[0][0]).toBe('http://api.test/api/v1/add')
    expect(fetchFn.mock.calls[0][1].signal).toBe(signal)
  })

  it.each<[CalculatorErrorCode, number, string]>([
    ['INVALID_REQUEST', 400, 'Please enter valid numbers.'],
    ['DIVISION_BY_ZERO', 422, "Can't divide by zero."],
    ['NEGATIVE_SQUARE_ROOT', 422, "Can't take the square root of a negative number."],
    ['RESULT_OUT_OF_RANGE', 422, 'The result is too large to calculate.'],
    ['UNDEFINED_RESULT', 422, 'The result is undefined for these numbers.'],
    ['NOT_FOUND', 404, "This operation isn't supported."],
    ['METHOD_NOT_ALLOWED', 405, "This operation isn't supported."],
    ['PAYLOAD_TOO_LARGE', 413, 'The numbers are too long.'],
    ['INTERNAL_ERROR', 500, 'Something went wrong on our side. Please try again.'],
  ])('maps %s to a user-friendly message', async (code, status, message) => {
    const { service } = serviceReturning(jsonResponse(status, { error: { code, message: 'raw server message' } }))

    const err = await rejection(service.calculate({ operation: 'divide', a: 1, b: 0 }))

    expect(err.code).toBe(code)
    expect(err.message).toBe(message)
  })

  it('shows the server message for an error code it does not know', async () => {
    const { service } = serviceReturning(
      jsonResponse(429, { error: { code: 'RATE_LIMITED', message: 'slow down' } }),
    )

    const err = await rejection(service.calculate({ operation: 'add', a: 1, b: 2 }))

    expect(err.code).toBe('UNKNOWN_ERROR')
    expect(err.message).toBe('slow down')
  })

  it('turns a fetch failure into NETWORK_ERROR', async () => {
    const cause = new TypeError('Failed to fetch')
    const service = createCalculatorService(() => Promise.reject(cause))

    const err = await rejection(service.calculate({ operation: 'add', a: 1, b: 2 }))

    expect(err.code).toBe('NETWORK_ERROR')
    expect(err.message).toMatch(/can't reach the calculator service/i)
    expect(err.cause).toBe(cause)
  })

  it('rethrows the abort error unchanged when the request is cancelled', async () => {
    const controller = new AbortController()
    const abortError = new DOMException('aborted', 'AbortError')
    const service = createCalculatorService(() => {
      controller.abort()
      return Promise.reject(abortError)
    })

    await expect(service.calculate({ operation: 'add', a: 1, b: 2 }, controller.signal)).rejects.toBe(abortError)
  })

  it.each([
    ['a non-JSON body', new Response('<html>Bad Gateway</html>', { status: 502 })],
    ['an error status without the error shape', jsonResponse(500, { message: 'oops' })],
    ['a success status without a numeric result', jsonResponse(200, { operation: 'add', result: '3' })],
    ['a success status with an error body', jsonResponse(200, { error: { code: 'INTERNAL_ERROR', message: 'x' } })],
  ])('reports UNEXPECTED_RESPONSE for %s', async (_name, response) => {
    const { service } = serviceReturning(response)

    const err = await rejection(service.calculate({ operation: 'add', a: 1, b: 2 }))

    expect(err.code).toBe('UNEXPECTED_RESPONSE')
  })

  it('uses the global fetch by default', async () => {
    const fetchSpy = vi
      .spyOn(globalThis, 'fetch')
      .mockResolvedValue(jsonResponse(200, { operation: 'multiply', result: 40 }))

    await expect(createCalculatorService().calculate({ operation: 'multiply', a: 10, b: 4 })).resolves.toBe(40)
    expect(fetchSpy).toHaveBeenCalledWith('/api/v1/multiply', expect.objectContaining({ method: 'POST' }))

    fetchSpy.mockRestore()
  })
})
