import { useEffect } from 'react'

/**
 * Sends key presses anywhere on the page to `handleKey`. Handled keys have
 * their default action prevented, so Enter on a focused keypad button
 * evaluates once instead of also "clicking" the button. Shortcuts with
 * Ctrl/Alt/Meta (e.g. Ctrl+R) are left to the browser.
 */
export function useKeyboardShortcuts(handleKey: (key: string) => boolean): void {
  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.ctrlKey || event.altKey || event.metaKey) return
      if (handleKey(event.key)) event.preventDefault()
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [handleKey])
}
