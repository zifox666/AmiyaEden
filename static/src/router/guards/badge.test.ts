import assert from 'node:assert/strict'
import test from 'node:test'
import { loadBadgeCounts } from './badge'

test('loadBadgeCounts leaves badge state intact when fetch succeeds', async () => {
  let clearCalls = 0
  const errors: unknown[][] = []

  await loadBadgeCounts(
    {
      async fetchBadgeCounts() {},
      clearBadgeCounts() {
        clearCalls += 1
      }
    },
    {
      error(...args: unknown[]) {
        errors.push(args)
      }
    }
  )

  assert.equal(clearCalls, 0)
  assert.deepEqual(errors, [])
})

test('loadBadgeCounts clears badge state and logs when fetch fails', async () => {
  let clearCalls = 0
  const errors: unknown[][] = []
  const failure = new Error('network down')

  await loadBadgeCounts(
    {
      async fetchBadgeCounts() {
        throw failure
      },
      clearBadgeCounts() {
        clearCalls += 1
      }
    },
    {
      error(...args: unknown[]) {
        errors.push(args)
      }
    }
  )

  assert.equal(clearCalls, 1)
  assert.equal(errors.length, 1)
  assert.equal(errors[0]?.[0], '[RouteGuard] 导航徽章加载失败:')
  assert.equal(errors[0]?.[1], failure)
})
