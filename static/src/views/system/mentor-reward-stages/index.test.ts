import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('mentor reward stage number inputs are configured for integers only', () => {
  const inputNumbers = source.match(/<ElInputNumber[\s\S]*?\/>/g) ?? []

  assert.equal(inputNumbers.length, 3)

  for (const inputNumber of inputNumbers) {
    assert.match(inputNumber, /:step="1"/)
    assert.match(inputNumber, /step-strictly/)
    assert.doesNotMatch(inputNumber, /:precision="/)
    assert.doesNotMatch(inputNumber, /0\.01/)
  }
})
