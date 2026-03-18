<template>
  <div class="art-card p-5 mb-5 max-sm:mb-4">
    <div class="art-card-header mb-4">
      <div class="title">
        <h4>{{ $t('dashboardConsole.srpList.title') }}</h4>
        <p>{{ $t('dashboardConsole.srpList.recentCount', { count: list.length }) }}</p>
      </div>
    </div>
    <div v-if="list.length === 0" class="flex-cc h-30 text-g-500 text-sm">
      {{ $t('dashboardConsole.srpList.empty') }}
    </div>
    <ArtTable
      v-else
      :data="list"
      size="large"
      :border="false"
      :stripe="false"
      :header-cell-style="{ background: 'transparent' }"
    >
      <template #default>
        <ElTableColumn
          :label="$t('srp.apply.columns.character')"
          prop="character_name"
          min-width="120"
        />
        <ElTableColumn :label="$t('srp.apply.columns.ship')" prop="ship_name" min-width="140" />
        <ElTableColumn
          :label="$t('srp.apply.columns.system')"
          prop="solar_system_name"
          min-width="120"
        />
        <ElTableColumn :label="$t('dashboardConsole.srpList.lossTime')" min-width="160">
          <template #default="{ row }">
            {{ formatTime(row.killmail_time) }}
          </template>
        </ElTableColumn>
        <ElTableColumn
          :label="$t('dashboardConsole.srpList.recommendedAmount')"
          min-width="120"
          align="right"
        >
          <template #default="{ row }">
            <span class="text-g-700">{{ formatISK(row.recommended_amount) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn
          :label="$t('dashboardConsole.srpList.finalAmount')"
          min-width="120"
          align="right"
        >
          <template #default="{ row }">
            <span class="font-medium">{{ formatISK(row.final_amount) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('srp.apply.columns.reviewStatus')" min-width="90" align="center">
          <template #default="{ row }">
            <ElTag :type="reviewStatusType(row.review_status)" size="small" effect="plain">
              {{ reviewStatusLabel(row.review_status) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('srp.apply.columns.payoutStatus')" min-width="90" align="center">
          <template #default="{ row }">
            <ElTag :type="payoutStatusType(row.payout_status)" size="small" effect="plain">
              {{ payoutStatusLabel(row.payout_status) }}
            </ElTag>
          </template>
        </ElTableColumn>
      </template>
    </ArtTable>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'

  const { t } = useI18n()
  defineProps<{
    list: Api.Dashboard.SrpItem[]
  }>()

  const formatTime = (time: string): string => {
    if (!time) return ''
    const d = new Date(time)
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  }

  const formatISK = (amount: number): string => {
    if (!amount) return '0'
    return amount.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  }

  const reviewStatusLabel = (status: string): string => {
    const map: Record<string, string> = {
      pending: t('srp.status.pending'),
      approved: t('srp.status.approved'),
      rejected: t('srp.status.rejected')
    }
    return map[status] ?? status
  }

  const reviewStatusType = (
    status: string
  ): 'primary' | 'success' | 'warning' | 'danger' | 'info' => {
    const map: Record<string, 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
      pending: 'warning',
      approved: 'success',
      rejected: 'danger'
    }
    return map[status] ?? 'info'
  }

  const payoutStatusLabel = (status: string): string => {
    const map: Record<string, string> = {
      pending: t('srp.status.unpaid'),
      paid: t('srp.status.paid')
    }
    return map[status] ?? status
  }

  const payoutStatusType = (
    status: string
  ): 'primary' | 'success' | 'warning' | 'danger' | 'info' => {
    const map: Record<string, 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
      pending: 'info',
      paid: 'success'
    }
    return map[status] ?? 'info'
  }
</script>
