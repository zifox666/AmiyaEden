<!-- 军团管理 - 舰队作战激励配置 -->
<template>
  <div class="fleet-incentive-page art-full-height">
    <ElCard shadow="never" v-loading="loading">
      <template #header>
        <span class="section-title">{{ $t('fleetIncentive.title') }}</span>
      </template>

      <ElAlert
        :title="$t('fleetIncentive.hint')"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      />

      <ElTable :data="rows" border style="width: 100%">
        <!-- 舰队类型 -->
        <ElTableColumn :label="$t('fleetIncentive.fields.fleetType')" width="140">
          <template #default="{ row }">
            {{ $t(`fleetIncentive.types.${row.fleet_type}`) }}
          </template>
        </ElTableColumn>

        <!-- 启用 -->
        <ElTableColumn :label="$t('fleetIncentive.fields.enabled')" width="100" align="center">
          <template #default="{ row }">
            <ElSwitch v-model="row.enabled" />
          </template>
        </ElTableColumn>

        <!-- 倍率条件（0 = 无条件） -->
        <ElTableColumn :label="$t('fleetIncentive.fields.multiplier')" min-width="160">
          <template #default="{ row }">
            <ElInputNumber
              v-model="row.multiplier"
              :min="0"
              :precision="2"
              :step="0.5"
              style="width: 130px"
            />
          </template>
        </ElTableColumn>

        <!-- 成员奖励 -->
        <ElTableColumn :label="$t('fleetIncentive.fields.memberReward')" min-width="160">
          <template #default="{ row }">
            <ElInputNumber
              v-model="row.member_reward"
              :min="0"
              :precision="2"
              :step="100"
              style="width: 130px"
            />
          </template>
        </ElTableColumn>

        <!-- FC 奖励 -->
        <ElTableColumn :label="$t('fleetIncentive.fields.fcReward')" min-width="160">
          <template #default="{ row }">
            <ElInputNumber
              v-model="row.fc_reward"
              :min="0"
              :precision="2"
              :step="100"
              style="width: 130px"
            />
          </template>
        </ElTableColumn>

        <!-- 带队奖励（PAP 触发） -->
        <ElTableColumn
          :label="$t('fleetIncentive.fields.fcLeadEnabled')"
          width="120"
          align="center"
        >
          <template #default="{ row }">
            <ElSwitch v-model="row.fc_lead_enabled" />
          </template>
        </ElTableColumn>

        <ElTableColumn :label="$t('fleetIncentive.fields.fcLeadReward')" min-width="160">
          <template #default="{ row }">
            <ElInputNumber
              v-model="row.fc_lead_reward"
              :min="0"
              :precision="2"
              :step="100"
              style="width: 130px"
            />
          </template>
        </ElTableColumn>

        <!-- 操作 -->
        <ElTableColumn :label="$t('fleetIncentive.action')" width="120" align="center">
          <template #default="{ row }">
            <ElButton
              type="primary"
              size="small"
              :loading="row._saving"
              @click="handleSave(row as Row)"
            >
              {{ $t('fleetIncentive.save') }}
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ref, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { ElMessage } from 'element-plus'
  import { fetchBattleIncentives, updateBattleIncentive } from '@/api/fleet-incentive'

  const { t } = useI18n()

  type Row = Api.FleetIncentive.BattleIncentive & { _saving: boolean }

  const loading = ref(false)
  const rows = ref<Row[]>([])

  async function getData() {
    loading.value = true
    try {
      const res = await fetchBattleIncentives()
      rows.value = (res ?? []).map((item) => ({ ...item, _saving: false }))
    } finally {
      loading.value = false
    }
  }

  async function handleSave(row: Row) {
    row._saving = true
    try {
      await updateBattleIncentive(row.fleet_type, {
        enabled: row.enabled,
        multiplier: row.multiplier,
        member_reward: row.member_reward,
        fc_reward: row.fc_reward,
        fc_lead_enabled: row.fc_lead_enabled,
        fc_lead_reward: row.fc_lead_reward
      })
      ElMessage.success(t('fleetIncentive.saveSuccess'))
    } finally {
      row._saving = false
    }
  }

  onMounted(getData)
</script>

<style scoped>
  .section-title {
    font-size: 16px;
    font-weight: 600;
  }
</style>
