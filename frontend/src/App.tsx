import { Calculator } from './components/Calculator'

export function App() {
  return (
    <main className="app">
      <h1 className="app__title">Calculator</h1>
      <Calculator />
      <p className="app__hint">
        Keyboard: digits, <kbd>+</kbd> <kbd>-</kbd> <kbd>*</kbd> <kbd>/</kbd> <kbd>^</kbd> <kbd>%</kbd>,{' '}
        <kbd>r</kbd> √, <kbd>n</kbd> ±, <kbd>Enter</kbd> =, <kbd>Esc</kbd> clear
      </p>
    </main>
  )
}
