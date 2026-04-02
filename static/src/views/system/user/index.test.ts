import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('system user page renders token health for nested characters', () => {
  assert.match(source, /token_invalid/)
  assert.match(source, /userAdmin\.characters\.tokenHealth/)
})

test('system user page renders a super-admin switch for the character esi restriction', () => {
  assert.match(source, /isSuperAdmin/)
  assert.match(source, /fetchCharacterESIRestrictionConfig/)
  assert.match(source, /updateCharacterESIRestrictionConfig/)
  assert.match(source, /ElSwitch/)
  assert.match(source, /userAdmin\.characterEsiRestriction\./)
})

test('system user page keeps the restriction card out of the flex table shell', () => {
  assert.doesNotMatch(
    source,
    /<ElCard v-if="isSuperAdmin" class="art-table-card mb-4" shadow="never">/
  )
})
