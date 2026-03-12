<!-- 我的兑换码面板 -->
<template>
  <ElCard class="art-table-card" shadow="never">
    <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

    <ArtTable
      :loading="loading"
      :data="data"
      :columns="columns"
      :pagination="pagination"
      :empty-text="$t('shop.noRedeemCodes')"
      @pagination:size-change="handleSizeChange"
      @pagination:current-change="handleCurrentChange"
    />
  </ElCard>
</template>

<script setup lang="ts">
  import { ElTag } from 'element-plus'
  import { fetchMyRedeemCodes } from '@/api/shop'
  import { useTable } from '@/hooks/core/useTable'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'ShopRedeem' })

  const { t } = useI18n()

  type RedeemCode = Api.Shop.RedeemCode

  // ─── 兑换码状态映射 ───
  const REDEEM_STATUS_MAP: Record<string, { label: string; type: string }> = {
    unused: { label: t('shopAdmin.redeem.status.unused'), type: 'success' },
    used: { label: t('shopAdmin.redeem.status.used'), type: 'info' },
    expired: { label: t('shopAdmin.redeem.status.expired'), type: 'danger' }
  }

  const formatTime = (v: string | null) => (v ? new Date(v).toLocaleString() : '-')

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchMyRedeemCodes,
      apiParams: { current: 1, size: 20 },
      immediate: false,
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'code',
          label: t('shop.redeemCode'),
          minWidth: 220,
          formatter: (row: RedeemCode) => h('code', { class: 'text-sm font-mono' }, row.code)
        },
        {
          prop: 'status',
          label: t('shop.status'),
          minWidth: 100,
          formatter: (row: RedeemCode) => {
            const cfg = REDEEM_STATUS_MAP[row.status] ?? { label: row.status, type: 'info' }
            return h(
              ElTag,
              { type: cfg.type as any, size: 'small', effect: 'plain' },
              () => cfg.label
            )
          }
        },
        {
          prop: 'created_at',
          label: t('shop.orderTime'),
          minWidth: 180,
          formatter: (row: RedeemCode) => h('span', {}, formatTime(row.created_at))
        },
        {
          prop: 'expires_at',
          label: t('shop.expiresAt'),
          minWidth: 180,
          formatter: (row: RedeemCode) => h('span', {}, formatTime(row.expires_at))
        }
      ]
    }
  })

  // 供父页面按需触发首次加载
  defineExpose({ load: getData, refresh: refreshData })
</script>
