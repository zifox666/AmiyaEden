import assert from 'node:assert/strict'
import test from 'node:test'

import { createClipboardCopy } from './useClipboardCopy'

test('createClipboardCopy reports success and failure through shared callbacks', async () => {
  const successEvents: string[] = []
  const success = createClipboardCopy({
    writeText: async () => {},
    successMessage: 'Copied',
    failureMessage: 'Copy failed',
    notifySuccess: (message) => successEvents.push(`success:${message}`),
    notifyFailure: (message) => successEvents.push(`failure:${message}`)
  })

  await success.copyText('A3KM9ZQ2')
  assert.deepEqual(successEvents, ['success:Copied'])

  const failureEvents: string[] = []
  const failure = createClipboardCopy({
    writeText: async () => {
      throw new Error('denied')
    },
    successMessage: 'Copied',
    failureMessage: 'Copy failed',
    notifySuccess: (message) => failureEvents.push(`success:${message}`),
    notifyFailure: (message) => failureEvents.push(`failure:${message}`)
  })

  await failure.copyText('A3KM9ZQ2')
  assert.deepEqual(failureEvents, ['failure:Copy failed'])
})
