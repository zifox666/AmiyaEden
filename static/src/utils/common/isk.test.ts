import assert from 'node:assert/strict'
import test from 'node:test'

import { formatIskPlain, formatIskSmart, iskToMillionInput, millionInputToIsk } from './isk'

test('formatIskPlain keeps grouping and two decimals', () => {
  assert.equal(formatIskPlain(711_103_702.38), '711,103,702.38')
  assert.equal(formatIskPlain(-14_500_000), '-14,500,000.00')
  assert.equal(formatIskPlain(0), '0.00')
  assert.equal(formatIskPlain(undefined), '-')
  assert.equal(formatIskPlain(null), '-')
})

test('formatIskSmart applies unit thresholds and promotion', () => {
  assert.equal(formatIskSmart(950), '950.00')
  assert.equal(formatIskSmart(12_500), '12.50 K')
  assert.equal(formatIskSmart(711_103_702.38), '711.10 M')
  assert.equal(formatIskSmart(1_250_000_000), '1.25 B')
  assert.equal(formatIskSmart(2_400_000_000_000), '2.40 T')
  assert.equal(formatIskSmart(999_995), '1.00 M')
  assert.equal(formatIskSmart(-12_500), '-12.50 K')
  assert.equal(formatIskSmart(null), '-')
  assert.equal(formatIskSmart(undefined), '-')
})

test('million conversion helpers preserve existing editor semantics', () => {
  assert.equal(iskToMillionInput(14_500_000), 14.5)
  assert.equal(millionInputToIsk(14.5), 14_500_000)
  assert.equal(millionInputToIsk(null), 0)
})
