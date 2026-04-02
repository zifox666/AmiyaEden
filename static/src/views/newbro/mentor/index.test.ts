import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'

const source = readFileSync(new URL('./index.vue', import.meta.url), 'utf8')

test('mentee summary cards show mentee contact details in both tables', () => {
  const menteeColumns =
    source.match(
      /<ElTableColumn :label="t\('newbro\.common\.mentee'\)"[\s\S]*?<\/ElTableColumn>/g
    ) ?? []

  assert.equal(menteeColumns.length, 2)

  for (const column of menteeColumns) {
    assert.match(column, /newbro\.mentor\.qq/)
    assert.match(column, /row\.mentee_qq/)
    assert.match(column, /newbro\.mentor\.discordId/)
    assert.match(column, /row\.mentee_discord_id/)
    assert.match(column, /v-if="row\.mentee_discord_id"/)
  }
})

test('mentor dashboard hides applied and graduated time columns', () => {
  assert.doesNotMatch(source, /<ElTableColumn prop="applied_at"/)
  assert.doesNotMatch(source, /<ElTableColumn prop="graduated_at"/)
})

test('mentor dashboard includes a read-only reward stage config tab', () => {
  assert.match(source, /newbro\.mentor\.rewardStagesTab/)
  assert.match(source, /newbro\.mentor\.rewardStagesTitle/)
  assert.match(source, /fetchMentorDashboardRewardStages/)
  assert.match(source, /newbro\.mentorConditionTypes\./)
  assert.match(source, /system\.mentorRewardStages\.stageOrder/)
  assert.match(source, /system\.mentorRewardStages\.stageName/)
  assert.match(source, /system\.mentorRewardStages\.rewardAmount/)
})

test('mentee list shows accumulated distributed reward amount', () => {
  assert.match(source, /newbro\.mentor\.distributedRewardAmount/)
  assert.match(source, /row\.distributed_reward_amount/)
})

test('distributed stages show stage names instead of raw ids', () => {
  assert.match(source, /formatDistributedStageName/)
  assert.match(source, /rewardStages\.value\.find/)
  assert.match(source, /\{\{\s*formatDistributedStageName\(stage\)\s*\}\}/)
  assert.doesNotMatch(source, /#\{\{\s*stage\s*\}\}/)
})

test('mentor dashboard includes the mentor responsibilities card', () => {
  assert.match(source, /<MentorResponsibilitiesCard\b/)
})
