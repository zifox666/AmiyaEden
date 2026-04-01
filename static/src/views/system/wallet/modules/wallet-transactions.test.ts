import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./wallet-transactions.vue', import.meta.url), 'utf8')

test('wallet admin localizes mentor_reward ref types in the filter and tag map', () => {
  assert.match(source, /walletAdmin\.refTypes\.mentor_reward/)
  assert.match(source, /value="mentor_reward"/)
  assert.match(source, /mentor_reward:\s*\{/)
})
