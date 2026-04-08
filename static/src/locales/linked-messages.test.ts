import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import test from 'node:test'

interface LocaleTree {
  [key: string]: LocaleTree | string
}

const enMessages = JSON.parse(
  readFileSync(new URL('./langs/en.json', import.meta.url), 'utf8')
) as LocaleTree
const zhMessages = JSON.parse(
  readFileSync(new URL('./langs/zh.json', import.meta.url), 'utf8')
) as LocaleTree

// MAINTENANCE: when new shared locale copy is introduced via @:canonical links,
// add the canonical key and its aliases here so the test enforces the link.
const linkedMessages = {
  'common.status': [
    'welfareSettings.filterStatus',
    'shop.status',
    'shop.manage.filterStatus',
    'shop.manage.status',
    'alliancePap.columns.status',
    'shopAdmin.products.statusPlaceholder',
    'shopAdmin.products.fields.status',
    'shopAdmin.redeem.statusPlaceholder'
  ],
  'common.type': [
    'info.journalType',
    'npcKill.journalRefType',
    'shop.manage.type',
    'shop.manage.colType',
    'fleet.wallet.refType',
    'shopAdmin.products.fields.type',
    'webhook.fields.type',
    'webhook.test.type'
  ],
  'common.amount': [
    'info.journalAmount',
    'npcKill.journalAmount',
    'newbro.common.amount',
    'fleet.wallet.amount',
    'srp.manage.batchPayoutAmount',
    'walletAdmin.fields.amount',
    'walletAdmin.transactions.amount'
  ],
  'common.reason': [
    'info.journalReason',
    'npcKill.journalReason',
    'fleet.wallet.reason',
    'walletAdmin.fields.reason'
  ],
  'common.search': [
    'npcKill.search',
    'shop.manage.search',
    'fleet.corporationPap.search',
    'srp.manage.searchBtn',
    'table.searchBar.search'
  ],
  'common.reset': [
    'npcKill.reset',
    'shop.manage.reset',
    'fleet.corporationPap.reset',
    'srp.manage.resetBtn',
    'table.form.reset',
    'table.searchBar.reset'
  ],
  'common.refresh': [
    'worktab.btn.refresh',
    'srp.manage.refresh',
    'srp.prices.refresh',
    'alliancePap.refresh'
  ],
  'common.time': ['fleet.wallet.time', 'srp.apply.fleetDetailTime', 'srp.apply.columns.time'],
  'common.cancel': ['srp.apply.cancelBtn', 'srp.prices.cancelBtn'],
  'common.close': ['search.exitKeydown'],
  'common.copied': ['welfareApproval.copied', 'srp.manage.copied'],
  'common.copyFailed': ['welfareApproval.copyFailed', 'srp.manage.copyFailed'],
  'common.createdAt': ['fleet.fields.createdAt', 'shopAdmin.redeem.table.createdAt'],
  'common.delete': ['shop.manage.deleteBtn', 'srp.prices.deleteBtn'],
  'common.edit': ['srp.manage.editBtn', 'srp.prices.editBtn'],
  'common.save': ['srp.prices.saveBtn', 'system.basicConfig.save'],
  'setting.basics.title': ['menus.system.basicConfig', 'system.basicConfig.title'],
  'console.alliancePap': [
    'menus.system.alliancePap',
    'fleet.pap.allianceCard',
    'alliancePap.shortTitle',
    'dashboardConsole.alliancePapTitle'
  ],
  'newbro.common.nickname': [
    'newbro.select.nicknameLabel',
    'newbro.selectMentor.nicknameLabel',
    'newbro.manage.nicknameLabel',
    'newbro.manage.nicknameColumn',
    'fleet.corporationPap.columns.nickname',
    'characters.profile.nickname',
    'srp.manage.columns.nickname',
    'shopAdmin.orders.table.nickname'
  ],
  'shop.redeemCode': [
    'shop.typeRedeem',
    'shopAdmin.products.typeRedeem',
    'shopAdmin.products.values.redeem',
    'shopAdmin.redeem.table.code'
  ],
  'welfareApproval.rejectBtn': [
    'newbro.mentor.reject',
    'srp.manage.rejectBtn',
    'shopAdmin.orders.rejectButton'
  ],
  'welfareApproval.contact': ['userAdmin.table.contact', 'shopAdmin.orders.table.contact'],
  'shop.orderNo': ['shopAdmin.orders.fields.orderNo', 'shopAdmin.orders.table.orderNo'],
  'fleet.fields.startAt': ['fleet.pap.allianceStartTime', 'alliancePap.columns.startTime'],
  'fleet.fields.endAt': ['fleet.pap.allianceEndTime', 'alliancePap.columns.endTime']
} as const

const getMessage = (messages: LocaleTree, path: string): string => {
  const value = path.split('.').reduce<LocaleTree | string | undefined>((current, segment) => {
    if (!current || typeof current === 'string') {
      return undefined
    }

    return current[segment]
  }, messages)

  if (typeof value !== 'string') {
    assert.fail(`Expected "${path}" to resolve to a locale string`)
  }

  return value
}

for (const [localeName, messages] of [
  ['en', enMessages],
  ['zh', zhMessages]
] as const) {
  test(`${localeName} locale links repeated copy to shared canonical messages`, () => {
    for (const [canonical, aliases] of Object.entries(linkedMessages)) {
      const canonicalValue = getMessage(messages, canonical)
      assert.ok(
        !canonicalValue.startsWith('@:'),
        `Expected "${canonical}" in ${localeName} to remain the canonical literal`
      )

      for (const alias of aliases) {
        assert.equal(
          getMessage(messages, alias),
          `@:${canonical}`,
          `Expected "${alias}" in ${localeName} to link to "${canonical}"`
        )
      }
    }
  })
}
