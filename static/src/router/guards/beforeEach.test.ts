import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./beforeEach.ts', import.meta.url), 'utf8')

test('dynamic route initialization fetches badge counts after storing the menu list', () => {
  assert.match(source, /menuStore\.setMenuList\(menuList\)[\s\S]*loadBadgeCounts\(badgeStore\)/)
})
