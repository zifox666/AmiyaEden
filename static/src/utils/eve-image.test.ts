import assert from 'node:assert/strict'
import test from 'node:test'

import { buildEveCharacterPortraitUrl } from './eve-image'

test('buildEveCharacterPortraitUrl returns the standard portrait URL', () => {
  assert.equal(
    buildEveCharacterPortraitUrl(90000001),
    'https://images.evetech.net/characters/90000001/portrait?size=128'
  )
})

test('buildEveCharacterPortraitUrl returns an empty string for non-positive ids', () => {
  assert.equal(buildEveCharacterPortraitUrl(0), '')
})
