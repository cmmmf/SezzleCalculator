import { describe, expect, it } from 'vitest'

import { formatNumber } from './formatNumber'

describe('formatNumber', () => {
  it.each([
    [0, '0'],
    [-0, '0'],
    [14, '14'],
    [-6, '-6'],
    [2.5, '2.5'],
    [0.1 + 0.2, '0.3'],
    [1 / 3, '0.333333333333333'],
    [123456789012345 + 0.5, '123456789012346'],
    [123456789012345, '123456789012345'],
    [1e15, '1e+15'],
    [1.5e21, '1.5e+21'],
    [-(2 ** 60), '-1.1529215e+18'],
    [0.000001, '0.000001'],
    [1e-10, '1e-10'],
    [1.23456789123e-12, '1.23456789e-12'],
    [Number.POSITIVE_INFINITY, 'Infinity'],
    [Number.NaN, 'NaN'],
  ])('formats %d as %s', (value, expected) => {
    expect(formatNumber(value)).toBe(expected)
  })
})
