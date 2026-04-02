import assert from 'node:assert/strict'
import test from 'node:test'

import { formatFuxiCoinAmount, formatFuxiCoinWhole } from './fuxiCoin'

test('formatFuxiCoinWhole rounds and groups integer-valued shop prices', () => {
  assert.equal(formatFuxiCoinWhole(12_500), '12,500')
  assert.equal(formatFuxiCoinWhole(12_500.4), '12,500')
  assert.equal(formatFuxiCoinWhole(12_500.5), '12,501')
  assert.equal(formatFuxiCoinWhole(-12_500.5), '-12,500')
})

test('formatFuxiCoinAmount keeps grouping and two decimals for wallet values', () => {
  assert.equal(formatFuxiCoinAmount(12_500), '12,500.00')
  assert.equal(formatFuxiCoinAmount(12_500.5), '12,500.50')
  assert.equal(formatFuxiCoinAmount(-12_500.5), '-12,500.50')
  assert.equal(formatFuxiCoinAmount(0), '0.00')
})
