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

test('corp km enable button calls the zero-argument handler without passing a character', () => {
  assert.match(source, /const handleEnableCorpKm = async \(\) =>/)
  assert.doesNotMatch(source, /@click="handleEnableCorpKm\(char\)"/)
  assert.match(source, /@click="handleEnableCorpKm"/)
})

test('corp km controls stay limited to admin roles', () => {
  assert.match(source, /const canManageCorpKm = computed\(\(\) => \{/)
  assert.match(source, /roles\.some\(\(r\) => \['super_admin', 'admin'\]\.includes\(r\)\)/)
})
