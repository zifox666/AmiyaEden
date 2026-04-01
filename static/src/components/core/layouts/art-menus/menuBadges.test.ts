import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const sidebarSubmenuSource = readFileSync(
  new URL('./art-sidebar-menu/widget/SidebarSubmenu.vue', import.meta.url),
  'utf8'
)
const horizontalSubmenuSource = readFileSync(
  new URL('./art-horizontal-menu/widget/HorizontalSubmenu.vue', import.meta.url),
  'utf8'
)
const appStylesSource = readFileSync(
  new URL('../../../../assets/styles/core/app.scss', import.meta.url),
  'utf8'
)

test('submenu titles render numeric badges using the submenu badge slot', () => {
  assert.match(
    sidebarSubmenuSource,
    /v-if="item\.meta\.showTextBadge && \(level > 0 \|\| menuOpen\)"[\s\S]*class="art-text-badge art-text-badge-submenu"/
  )
  assert.match(
    horizontalSubmenuSource,
    /v-if="item\.meta\.showTextBadge"[\s\S]*class="art-text-badge art-text-badge-submenu"/
  )
})

test('submenu badge slot uses inline layout instead of absolute positioning', () => {
  assert.match(appStylesSource, /\.art-text-badge-submenu\s*\{[\s\S]*position:\s*static/)
  assert.match(appStylesSource, /\.art-text-badge-submenu\s*\{[\s\S]*margin-left:\s*8px/)
})
