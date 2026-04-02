import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('srp prices keeps the page readable for SRP officers but gates mutation controls to admins and senior fcs', () => {
  assert.match(
    source,
    /const\s+canManagePrices\s*=\s*computed\(\(\)\s*=>[\s\S]*\['super_admin', 'admin', 'senior_fc'\]\.includes\(role\)/
  )
  assert.match(source, /<ElButton\s+v-if="canManagePrices"\s+type="primary"\s+:icon="Plus"/)
  assert.match(source, /<ArtExcelImport\s+v-if="canManagePrices"\s+@import-success="handleImport">/)
  assert.match(
    source,
    /const\s+actionColumn(?::\s*ColumnOption<ShipPrice>\[\])?\s*=\s*canManagePrices\.value\s*\?[\s\S]*prop:\s*'actions'[\s\S]*:\s*\[\]/
  )
})
