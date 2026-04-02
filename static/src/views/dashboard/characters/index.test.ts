import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('characters page renders a dedicated expired-esi alert', () => {
  assert.match(source, /hasInvalidCharacterToken/)
  assert.match(source, /enforceCharacterESIRestriction/)
  assert.match(source, /characters\.tokenHealth/)
  assert.match(source, /type="error"/)
})
