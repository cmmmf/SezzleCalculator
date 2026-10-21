import react from '@vitejs/plugin-react'
import { defineConfig } from 'vitest/config'

// In local dev the Go backend runs on :8080; proxying keeps the browser on a
// single origin, so the app never needs CORS outside of tests.
const backendURL = process.env.VITE_BACKEND_URL ?? 'http://localhost:8080'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': backendURL,
      '/healthz': backendURL,
      '/swagger': backendURL,
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      include: ['src/**/*.{ts,tsx}'],
      exclude: ['src/main.tsx', 'src/test/**', 'src/**/*.test.{ts,tsx}', 'src/vite-env.d.ts'],
      reporter: ['text', 'html', 'lcov'],
    },
  },
})
