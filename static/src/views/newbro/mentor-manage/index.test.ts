import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')
const apiSource = readFileSync(new URL('../../../api/mentor.ts', import.meta.url), 'utf8')
const typeSource = readFileSync(new URL('../../../types/api/api.d.ts', import.meta.url), 'utf8')
const docSource = readFileSync(
  new URL('../../../../../docs/features/current/mentor-system.md', import.meta.url),
  'utf8'
)
const zhLocaleSource = readFileSync(
  new URL('../../../locales/langs/zh.json', import.meta.url),
  'utf8'
)
const enLocaleSource = readFileSync(
  new URL('../../../locales/langs/en.json', import.meta.url),
  'utf8'
)

test('pending mentor applications use cancel-specific admin copy', () => {
  assert.match(source, /function revokeActionLabel/)
  assert.match(source, /function revokeActionSuccessMessage/)
  assert.match(source, /status === 'pending'/)
  assert.match(source, /newbro\.mentorManage\.cancelPending/)
  assert.match(source, /newbro\.mentorManage\.cancelPendingSuccess/)
  assert.match(source, /newbro\.mentorManage\.revoke/)
  assert.match(source, /newbro\.mentorManage\.revokeSuccess/)
})

test('mentor manage page includes a reward distribution records tab with ledger pagination', () => {
  assert.match(source, /<ElTabs v-model="activeTab"/)
  assert.match(source, /newbro\.mentorManage\.relationshipsTab/)
  assert.match(source, /newbro\.mentorManage\.rewardRecordsTab/)
  assert.match(source, /fetchAdminMentorRewardDistributions/)
  assert.match(source, /visual-variant="ledger"/)
  assert.match(
    source,
    /rewardHistoryPaginationOptions = \{\s*pageSizes: \[50, 100, 200, 500, 1000\]/
  )
})

test('mentor manage reward records support mentor character and nickname filtering across contract and docs', () => {
  assert.match(source, /rewardHistoryKeyword/)
  assert.match(source, /newbro\.mentorManage\.rewardKeyword/)
  assert.match(source, /mentor_character_name/)
  assert.match(source, /mentor_nickname/)

  assert.match(apiSource, /export function fetchAdminMentorRewardDistributions\(/)
  assert.match(typeSource, /interface RewardDistributionView\s*\{/)
  assert.match(typeSource, /mentor_character_name: string/)
  assert.match(typeSource, /mentor_nickname: string/)
  assert.match(typeSource, /type AdminRewardDistributionsParams = Partial<\{/)
  assert.match(
    typeSource,
    /type AdminRewardDistributionsResponse = Api\.Common\.PaginatedResponse<RewardDistributionView>/
  )

  assert.match(docSource, /GET \/api\/v1\/system\/mentor\/reward-distributions/)
  assert.match(
    docSource,
    /奖励发放记录 tab：按 ledger 方式分页显示导师奖励发放记录，并支持按导师人物名或昵称搜索/
  )

  assert.match(zhLocaleSource, /"rewardRecordsTab"\s*:/)
  assert.match(zhLocaleSource, /"rewardKeyword"\s*:/)
  assert.match(enLocaleSource, /"rewardRecordsTab"\s*:/)
  assert.match(enLocaleSource, /"rewardKeyword"\s*:/)
})
