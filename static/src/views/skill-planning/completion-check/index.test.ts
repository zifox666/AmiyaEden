import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('completion check missing-skills list renders the shared copy button for each skill name', () => {
  const tooltipBlock = source.match(
    /<ElTooltip\s+v-if="!plan\.fully_satisfied\s*&&\s*plan\.missing_skills\.length"[\s\S]*?<\/ElTooltip>/
  )

  assert.ok(tooltipBlock, 'expected tooltip content block')
  assert.match(
    source,
    /import ArtCopyButton from '@\/components\/core\/forms\/art-copy-button\/index.vue'/
  )
  assert.match(source, /missing-skills__name-row/)
  assert.match(source, /<ArtCopyButton\s+:text="skill\.skill_name"/)
  assert.doesNotMatch(tooltipBlock[0], /ArtCopyButton/)
})
