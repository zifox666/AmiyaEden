import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('welfare settings exposes a dedicated admin-only auto-approve config tab', () => {
  assert.match(source, /fetchWelfareAutoApproveConfig/)
  assert.match(source, /updateWelfareAutoApproveConfig/)
  assert.match(
    source,
    /<ElTabPane\s+v-if="canManage"\s+:label="t\('welfareSettings\.autoApproveConfigTab'\)"\s+name="autoApproveConfig"/
  )
  assert.match(source, /t\('welfareSettings\.autoApproveThreshold'\)/)
  assert.match(source, /t\('welfareSettings\.autoApproveThresholdHint'\)/)
})

test('welfare settings wires the Fuxi Legion tenure threshold through the form payload', () => {
  assert.match(source, /t\('welfareSettings\.minimumFuxiLegionYears'\)/)
  assert.match(source, /v-model="formData\.minimum_fuxi_legion_years"/)
  assert.match(source, /minimum_fuxi_legion_years:\s*row\.minimum_fuxi_legion_years \?\? undefined/)
  assert.match(source, /minimum_fuxi_legion_years:\s*formData\.minimum_fuxi_legion_years \?\? null/)
  assert.match(source, /max_char_age_months:\s*formData\.max_char_age_months \?\? null/)
  assert.match(source, /minimum_pap:\s*formData\.minimum_pap \?\? null/)
})
