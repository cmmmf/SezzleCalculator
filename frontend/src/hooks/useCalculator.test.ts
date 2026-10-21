import { act, renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { CalculatorError, type CalculationRequest, type CalculatorService } from '../services/calculatorService'
import { useCalculator } from './useCalculator'

interface Call {
  request: CalculationRequest
  signal?: AbortSignal
  resolve: (result: number) => void
  reject: (err: unknown) => void
}

/** A service whose calls stay pending until the test settles them. */
function deferredService() {
  const calls: Call[] = []
  const service: CalculatorService = {
    calculate: vi.fn(
      (request: CalculationRequest, signal?: AbortSignal) =>
        new Promise<number>((resolve, reject) => calls.push({ request, signal, resolve, reject })),
    ),
  }
  return { service, calls }
}

function typeKeys(result: { current: ReturnType<typeof useCalculator> }, keys: string) {
  act(() => {
    for (const key of keys) result.current.handleKey(key)
  })
}

describe('useCalculator', () => {
  it('sends the pending request and shows the result', async () => {
    const { service, calls } = deferredService()
    const { result } = renderHook(() => useCalculator(service))

    typeKeys(result, '2+3=')
    expect(result.current.view.isLoading).toBe(true)
    expect(calls).toHaveLength(1)
    expect(calls[0].request).toEqual({ operation: 'add', a: 2, b: 3 })

    await act(async () => calls[0].resolve(5))
    expect(result.current.view).toEqual({ value: '5', expression: '2 + 3 =', isLoading: false, error: null })
  })

  it('shows the error message of a CalculatorError', async () => {
    const { service, calls } = deferredService()
    const { result } = renderHook(() => useCalculator(service))

    typeKeys(result, '5/0=')
    await act(async () => calls[0].reject(new CalculatorError('DIVISION_BY_ZERO')))

    expect(result.current.view).toMatchObject({ error: "Can't divide by zero.", isLoading: false, value: '0' })
  })

  it('shows a generic message for unexpected errors', async () => {
    const { service, calls } = deferredService()
    const { result } = renderHook(() => useCalculator(service))

    typeKeys(result, '5/1=')
    await act(async () => calls[0].reject(new Error('boom')))

    expect(result.current.view.error).toBe('Something went wrong. Please try again.')
  })

  it('aborts the request on clear and ignores its late outcome', async () => {
    const { service, calls } = deferredService()
    const { result } = renderHook(() => useCalculator(service))

    const cleared = { value: '0', expression: '', isLoading: false, error: null }

    typeKeys(result, '9r')
    act(() => {
      result.current.handleKey('Escape')
    })
    expect(calls[0].signal?.aborted).toBe(true)
    await act(async () => calls[0].resolve(3))
    expect(result.current.view).toEqual(cleared)

    typeKeys(result, '4r')
    act(() => {
      result.current.handleKey('Escape')
    })
    await act(async () => calls[1].reject(new CalculatorError('NETWORK_ERROR')))
    expect(result.current.view).toEqual(cleared)
  })

  it('runs chained calculations one after another', async () => {
    const { service, calls } = deferredService()
    const { result } = renderHook(() => useCalculator(service))

    typeKeys(result, '2+3*')
    await act(async () => calls[0].resolve(5))
    typeKeys(result, '4=')

    await waitFor(() => expect(calls).toHaveLength(2))
    expect(calls[1].request).toEqual({ operation: 'multiply', a: 5, b: 4 })
    await act(async () => calls[1].resolve(20))
    expect(result.current.view.value).toBe('20')
  })

  it('aborts the request in flight on unmount', () => {
    const { service, calls } = deferredService()
    const { result, unmount } = renderHook(() => useCalculator(service))

    typeKeys(result, '1+1=')
    unmount()

    expect(calls[0].signal?.aborted).toBe(true)
  })

  it('reports whether a key was handled', () => {
    const { result } = renderHook(() => useCalculator(deferredService().service))
    let handled: boolean[] = []

    act(() => {
      handled = [result.current.handleKey('7'), result.current.handleKey('Shift')]
    })

    expect(handled).toEqual([true, false])
    expect(result.current.view.value).toBe('7')
  })

  it('exposes dispatch for keypad buttons', () => {
    const { result } = renderHook(() => useCalculator(deferredService().service))

    act(() => result.current.dispatch({ type: 'digit', digit: '4' }))

    expect(result.current.view.value).toBe('4')
  })
})
