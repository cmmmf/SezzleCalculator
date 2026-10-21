import { useCallback, useEffect, useMemo, useReducer } from 'react'

import { CalculatorError, calculatorService, type CalculatorService } from '../services/calculatorService'
import {
  calculatorReducer,
  initialState,
  keyToAction,
  selectView,
  type CalculatorAction,
  type CalculatorView,
} from '../state/calculatorReducer'

export interface UseCalculator {
  view: CalculatorView
  dispatch: (action: CalculatorAction) => void
  /** Handles a KeyboardEvent.key; returns true if the key was used. */
  handleKey: (key: string) => boolean
}

const FALLBACK_ERROR = new CalculatorError('UNKNOWN_ERROR').message

/**
 * Connects the pure reducer to the API: whenever the reducer sets `pending`,
 * the request is sent and its outcome dispatched back. Clearing (or
 * unmounting) aborts a request in flight, so a late response is never applied.
 */
export function useCalculator(service: CalculatorService = calculatorService): UseCalculator {
  const [state, dispatch] = useReducer(calculatorReducer, initialState)
  const { pending } = state

  useEffect(() => {
    if (!pending) return
    const controller = new AbortController()
    service.calculate(pending.request, controller.signal).then(
      (result) => {
        if (!controller.signal.aborted) dispatch({ type: 'resolved', result })
      },
      (err: unknown) => {
        if (controller.signal.aborted) return
        dispatch({ type: 'failed', message: err instanceof CalculatorError ? err.message : FALLBACK_ERROR })
      },
    )
    return () => controller.abort()
  }, [pending, service])

  const handleKey = useCallback((key: string) => {
    const action = keyToAction(key)
    if (action) dispatch(action)
    return action !== null
  }, [])

  const view = useMemo(() => selectView(state), [state])
  return { view, dispatch, handleKey }
}
