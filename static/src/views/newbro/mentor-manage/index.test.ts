import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('pending mentor applications use cancel-specific admin copy', () => {
  assert.match(source, /function revokeActionLabel/)
  assert.match(source, /function revokeActionSuccessMessage/)
  assert.match(source, /status === 'pending'/)
  assert.match(source, /newbro\.mentorManage\.cancelPending/)
  assert.match(source, /newbro\.mentorManage\.cancelPendingSuccess/)
  assert.match(source, /newbro\.mentorManage\.revoke/)
  assert.match(source, /newbro\.mentorManage\.revokeSuccess/)
})
