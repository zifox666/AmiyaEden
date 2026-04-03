import assert from 'node:assert/strict'
import test from 'node:test'

import { formatMailAttemptSuccess, formatMailAttemptWarning } from './mailAttempt'

const messages: Record<string, string> = {
  'mailAttempt.sent': 'Mail accepted.',
  'mailAttempt.failed': 'Mail failed:',
  'mailAttempt.sender': 'sender',
  'mailAttempt.recipient': 'recipient',
  'mailAttempt.mailId': 'mail ID'
}

const t = (key: string) => messages[key] ?? key

test('formatMailAttemptSuccess includes sender recipient and mail id details', () => {
  assert.equal(
    formatMailAttemptSuccess(
      {
        mail_sender_character_name: 'Officer Main',
        mail_sender_character_id: 90000077,
        mail_recipient_character_name: 'Pilot Main',
        mail_recipient_character_id: 90000042,
        mail_id: 123456789
      },
      t
    ),
    'Mail accepted. sender: Officer Main (90000077), recipient: Pilot Main (90000042), mail ID: 123456789'
  )
})

test('formatMailAttemptWarning includes the error and available route details', () => {
  assert.equal(
    formatMailAttemptWarning(
      {
        mail_error: 'missing scope',
        mail_sender_character_name: 'Officer Main',
        mail_sender_character_id: 90000077,
        mail_recipient_character_id: 90000042
      },
      t
    ),
    'Mail failed: missing scope, sender: Officer Main (90000077), recipient: 90000042'
  )
})
