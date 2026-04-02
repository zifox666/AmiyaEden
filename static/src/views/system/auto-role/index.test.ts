import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('auto role mapping page allows mapping tables to grow with content', () => {
  assert.match(source, /<ElTabs v-model="activeTab" type="border-card">/)
  assert.match(source, /<ElTable v-loading="esiRoleLoading" :data="esiRoleMappings" border stripe>/)
  assert.match(source, /<ElTable v-loading="titleLoading" :data="titleMappings" border stripe>/)
  assert.doesNotMatch(source, /<template>\s*<div class="[^"]*\bart-full-height\b[^"]*">/)
})
