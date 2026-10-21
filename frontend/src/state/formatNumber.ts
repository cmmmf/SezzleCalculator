// The API returns raw float64 results; rounding for display happens only here.

// 15 significant digits always round-trip through a float64 (DBL_DIG); float
// noise only shows up in the 16th–17th digit.
const SIGNIFICANT_DIGITS = 15
const EXPONENT_DIGITS = 8
const LARGE = 1e15
const SMALL = 1e-9

/**
 * Formats a result for the display: at most 15 significant digits, so float
 * noise such as 0.1 + 0.2 = 0.30000000000000004 shows as 0.3, and scientific
 * notation for very large or very small magnitudes.
 */
export function formatNumber(value: number): string {
  if (!Number.isFinite(value)) return String(value)
  if (value === 0) return '0' // also turns -0 into 0

  const magnitude = Math.abs(value)
  if (magnitude >= LARGE || magnitude < SMALL) {
    // 1.50000000e+21 → 1.5e+21
    return value.toExponential(EXPONENT_DIGITS).replace(/\.?0+e/, 'e')
  }
  return String(Number(value.toPrecision(SIGNIFICANT_DIGITS)))
}
