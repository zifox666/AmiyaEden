<template>
  <div class="art-card p-5 mb-5 max-sm:mb-4">
    <div class="art-card-header mb-4">
      <div class="title">
        <h4>补损申请</h4>
        <p>
          最近
          <span class="text-theme font-medium">{{ list.length }}</span>
          条
        </p>
      </div>
    </div>
    <div v-if="list.length === 0" class="flex-cc h-30 text-g-500 text-sm">
      暂无补损申请记录
    </div>
    <ArtTable v-else :data="list" size="large" :border="false" :stripe="false"
      :header-cell-style="{ background: 'transparent' }">
      <template #default>
        <ElTableColumn label="角色" prop="character_name" min-width="120" />
        <ElTableColumn label="舰船" prop="ship_name" min-width="140" />
        <ElTableColumn label="星系" prop="solar_system_name" min-width="120" />
        <ElTableColumn label="损失时间" min-width="160">
          <template #default="{ row }">
            {{ formatTime(row.killmail_time) }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="建议金额" min-width="120" align="right">
          <template #default="{ row }">
            <span class="text-g-700">{{ formatISK(row.recommended_amount) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="最终金额" min-width="120" align="right">
          <template #default="{ row }">
            <span class="font-medium">{{ formatISK(row.final_amount) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn label="审批" min-width="90" align="center">
          <template #default="{ row }">
            <ElTag
              :type="reviewStatusType(row.review_status)"
              size="small"
              effect="plain"
            >
              {{ reviewStatusLabel(row.review_status) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="发放" min-width="90" align="center">
          <template #default="{ row }">
            <ElTag
              :type="payoutStatusType(row.payout_status)"
              size="small"
              effect="plain"
            >
              {{ payoutStatusLabel(row.payout_status) }}
            </ElTag>
          </template>
        </ElTableColumn>
      </template>
    </ArtTable>
  </div>
</template>

<script setup lang="ts">
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
    const map: Record<string, string> = { pending: '待审批', approved: '已通过', rejected: '已拒绝' }
    return map[status] ?? status
  }

  const reviewStatusType = (status: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' => {
    const map: Record<string, 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
      pending: 'warning',
      approved: 'success',
      rejected: 'danger'
    }
    return map[status] ?? 'info'
  }

  const payoutStatusLabel = (status: string): string => {
    const map: Record<string, string> = { pending: '未发放', paid: '已发放' }
    return map[status] ?? status
  }

  const payoutStatusType = (status: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' => {
    const map: Record<string, 'primary' | 'success' | 'warning' | 'danger' | 'info'> = {
      pending: 'info',
      paid: 'success'
    }
    return map[status] ?? 'info'
  }
</script>
