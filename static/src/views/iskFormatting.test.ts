import assert from 'node:assert/strict'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import { fileURLToPath } from 'node:url'

const SRC_ROOT = fileURLToPath(new URL('..', import.meta.url))

function readSource(relativePath: string) {
  return readFileSync(join(SRC_ROOT, relativePath), 'utf8')
}

function collectTrackedFiles(entry: string): string[] {
  const fullPath = join(SRC_ROOT, entry)
  if (!statSync(fullPath).isDirectory()) return [entry]

  return readdirSync(fullPath).flatMap((child) => {
    const relativeChild = `${entry}/${child}`
    const childPath = join(SRC_ROOT, relativeChild)
    if (statSync(childPath).isDirectory()) return collectTrackedFiles(relativeChild)
    return childPath.endsWith('.vue') || childPath.endsWith('.ts') ? [relativeChild] : []
  })
}

const SRP_ROOTS = [
  'views/srp',
  'components/business/KmPreviewDialog.vue',
  'hooks/srp/useSrpManage.ts',
  'hooks/srp/useSrpWorkflow.ts',
  'views/operation/fleet-configs/modules/fleet-config-dialog.vue',
  'views/dashboard/console/modules/srp-list.vue'
] as const

const SRP_TRACKED_FILES = [...new Set(SRP_ROOTS.flatMap(collectTrackedFiles))]

const SRP_EXPECTATIONS = [
  {
    path: 'views/srp/apply/index.vue',
    mustContain: ['formatIskSmart']
  },
  {
    path: 'hooks/srp/useSrpManage.ts',
    mustContain: ['formatIskSmart']
  },
  {
    path: 'views/srp/manage/index.vue',
    mustContain: ['iskToMillionInput']
  },
  {
    path: 'views/srp/prices/index.vue',
    mustContain: ['formatIskSmart', 'iskToMillionInput', 'millionInputToIsk']
  },
  {
    path: 'components/business/KmPreviewDialog.vue',
    mustContain: ['formatIskSmart']
  },
  {
    path: 'views/dashboard/console/modules/srp-list.vue',
    mustContain: ['formatIskSmart']
  },
  {
    path: 'views/operation/fleet-configs/modules/fleet-config-dialog.vue',
    mustContain: ['iskToMillionInput', 'millionInputToIsk']
  },
  {
    path: 'hooks/srp/useSrpWorkflow.ts',
    mustContain: ['millionInputToIsk']
  }
] as const

for (const relativePath of SRP_TRACKED_FILES) {
  test(`${relativePath} does not define forbidden local ISK formatting`, () => {
    const source = readSource(relativePath)
    assert.doesNotMatch(source, /const\s+formatISK\s*=/)
    assert.doesNotMatch(source, /@\/utils\/iskUnits/)
    assert.doesNotMatch(source, /Intl\.NumberFormat\('en-US'/)
    assert.doesNotMatch(source, /toLocaleString\('en-US'/)
    assert.doesNotMatch(source, /\.toFixed\([^)]*\)\s*\+\s*['"`][^'"`]*[KMBTkmbt][^'"`]*['"`]/)
  })
}

for (const expectation of SRP_EXPECTATIONS) {
  test(`${expectation.path} uses the shared SRP/editor ISK helpers`, () => {
    const source = readSource(expectation.path)
    for (const helperName of expectation.mustContain) {
      assert.match(source, new RegExp(helperName))
    }
  })
}

const NON_SRP_ROOTS = [
  'views/info/contracts',
  'views/info/wallet/index.vue',
  'views/info/npc-kills/index.vue',
  'views/dashboard/npc-kills/index.vue',
  'views/dashboard/console/modules/card-list.vue',
  'views/newbro/manage/index.vue',
  'views/newbro/captain/index.vue',
  'hooks/newbro/useNewbroFormatters.ts'
] as const

const FUXI_ROOTS = ['views/shop', 'views/system/wallet'] as const

const NON_SRP_TRACKED_FILES = [...new Set(NON_SRP_ROOTS.flatMap(collectTrackedFiles))]
const FUXI_TRACKED_FILES = [...new Set(FUXI_ROOTS.flatMap(collectTrackedFiles))]

const NON_SRP_EXPECTATIONS = [
  ['views/info/wallet/index.vue', 'formatIskPlain'],
  ['views/info/npc-kills/index.vue', 'formatIskPlain'],
  ['views/dashboard/npc-kills/index.vue', 'formatIskPlain'],
  ['hooks/newbro/useNewbroFormatters.ts', 'formatIskPlain'],
  ['views/info/contracts/index.vue', 'formatIskSmart'],
  ['views/info/contracts/modules/contract-detail-dialog.vue', 'formatIskSmart'],
  ['views/dashboard/console/modules/card-list.vue', 'formatIskSmart']
] as const

for (const relativePath of NON_SRP_TRACKED_FILES) {
  test(`${relativePath} does not define forbidden local ISK formatting`, () => {
    const source = readSource(relativePath)
    assert.doesNotMatch(source, /const\s+formatISK\s*=/)
    assert.doesNotMatch(source, /function\s+formatISK\s*\(/)
    assert.doesNotMatch(source, /@\/utils\/iskUnits/)
    assert.doesNotMatch(source, /\.toFixed\([^)]*\)\s*\+\s*['"`][^'"`]*[KMBTkmbt][^'"`]*['"`]/)
  })
}

for (const relativePath of FUXI_TRACKED_FILES) {
  test(`${relativePath} does not use misleading local ISK formatter names for Fuxi Coin`, () => {
    const source = readSource(relativePath)
    assert.doesNotMatch(source, /const\s+formatISK\s*=/)
    assert.doesNotMatch(source, /function\s+formatISK\s*\(/)
  })
}

for (const [relativePath, helperName] of NON_SRP_EXPECTATIONS) {
  test(`${relativePath} uses ${helperName} for ISK formatting`, () => {
    const source = readSource(relativePath)
    assert.match(source, new RegExp(helperName))
  })
}

for (const relativePath of [
  'views/info/wallet/index.vue',
  'views/info/npc-kills/index.vue',
  'views/dashboard/npc-kills/index.vue',
  'views/info/contracts/index.vue',
  'views/info/contracts/modules/contract-detail-dialog.vue',
  'views/dashboard/console/modules/card-list.vue'
] as const) {
  test(`${relativePath} removes direct locale-based ISK formatting`, () => {
    const source = readSource(relativePath)
    assert.doesNotMatch(source, /Intl\.NumberFormat\('en-US'/)
    assert.doesNotMatch(source, /toLocaleString\('en-US'/)
  })
}

test('dashboard console keeps the animated wallet value plain while the summary string is smart-formatted', () => {
  const source = readSource('views/dashboard/console/modules/card-list.vue')
  assert.match(source, /value:\s*c\?\.eve_wallet_balance\s*\?\?\s*0/)
  assert.match(source, /decimals:\s*2/)
  assert.match(source, /separator=","/)
  assert.match(source, /desc:\s*formatIskSmart\(c\?\.eve_wallet_balance\s*\?\?\s*0\)/)
})
