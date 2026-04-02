import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const source = readFileSync(new URL('./manage-orders.vue', import.meta.url), 'utf8')

test('manage orders renders shared copy buttons for order number and main character', () => {
  assert.match(
    source,
    /import ArtCopyButton from '@\/components\/core\/forms\/art-copy-button\/index.vue'/
  )
  assert.match(source, /prop:\s*'order_no'[\s\S]*h\(ArtCopyButton,[\s\S]*text:\s*row\.order_no/)
  assert.match(
    source,
    /prop:\s*'main_character_name'[\s\S]*h\(ArtCopyButton,[\s\S]*text:\s*row\.main_character_name/
  )
})
