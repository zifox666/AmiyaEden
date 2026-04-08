import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const source = readFileSync(new URL('./manage-orders.vue', import.meta.url), 'utf8')
const zhLocale = JSON.parse(
  readFileSync(new URL('../../../../locales/langs/zh.json', import.meta.url), 'utf8')
)
const enLocale = JSON.parse(
  readFileSync(new URL('../../../../locales/langs/en.json', import.meta.url), 'utf8')
)

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

test('manage orders adds the compact isk total column before quantity', () => {
  assert.match(source, /formatIskSmart/)
  assert.match(source, /resolveOrderIskTotal/)
  assert.match(source, /prop:\s*'product_name'[\s\S]*prop:\s*'isk_total'[\s\S]*prop:\s*'quantity'/)
  assert.match(
    source,
    /prop:\s*'isk_total'[\s\S]*label:\s*t\('shopAdmin\.orders\.table\.iskTotal'\)/
  )
  assert.match(source, /prop:\s*'isk_total'[\s\S]*formatIskSmart\(iskTotal\)/)
  assert.match(source, /prop:\s*'isk_total'[\s\S]*h\(ArtCopyButton,\s*\{\s*text:\s*iskTotal\s*\}\)/)
  assert.equal(zhLocale.shopAdmin.orders.table.iskTotal, 'ISK总和')
  assert.equal(enLocale.shopAdmin.orders.table.iskTotal, 'ISK Total')
})
