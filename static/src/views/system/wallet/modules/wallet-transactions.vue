<!-- 钱包流水查询子模块 -->
<template>
  <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
    <template #left>
      <ElInput
        v-model="filterForm.user_id"
        placeholder="按用户 ID 筛选"
        clearable
        style="width: 160px"
        @clear="handleSearch"
        @keyup.enter="handleSearch"
      />
      <ElSelect
        v-model="filterForm.ref_type"
        placeholder="流水类型"
        clearable
        style="width: 160px"
        @change="handleSearch"
      >
        <ElOption label="PAP 奖励" value="pap_reward" />
        <ElOption label="管理员调整" value="admin_adjust" />
        <ElOption label="手动操作" value="manual" />
        <ElOption label="兑换消费" value="redeem" />
        <ElOption label="SRP 补损" value="srp_payout" />
        <ElOption label="商城购买" value="shop_purchase" />
      </ElSelect>
      <ElButton type="primary" @click="handleSearch">查询</ElButton>
    </template>
  </ArtTableHeader>

  <ArtTable
    :loading="loading"
    :data="data"
    :columns="columns"
    :pagination="pagination"
    @pagination:size-change="handleSizeChange"
    @pagination:current-change="handleCurrentChange"
  />
</template>

<script setup lang="ts">
  import { ElTag, ElButton, ElInput, ElSelect, ElOption } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import { adminListTransactions } from '@/api/sys-wallet'

  defineOptions({ name: 'WalletTransactions' })

  type WalletTransaction = Api.SysWallet.WalletTransaction

  const REF_TYPE_MAP: Record<string, { label: string; tag: string }> = {
    pap_reward: { label: 'PAP 奖励', tag: 'success' },
    admin_adjust: { label: '管理员调整', tag: 'warning' },
    manual: { label: '手动操作', tag: '' },
    redeem: { label: '兑换消费', tag: 'danger' },
    srp_payout: { label: 'SRP 补损', tag: 'primary' },
    shop_purchase: { label: '商城购买', tag: 'info' }
  }
  const getRefTypeLabel = (t: string) => REF_TYPE_MAP[t]?.label ?? t
  const getRefTypeTag = (t: string): any => REF_TYPE_MAP[t]?.tag ?? 'info'

  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)

  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')

  const filterForm = reactive({ user_id: '', ref_type: '' })

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    getData,
    searchParams
  } = useTable({
    core: {
      apiFn: adminListTransactions,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        { prop: 'user_id', label: '用户 ID', width: 90 },
        {
          prop: 'character_name',
          label: '主角色',
          minWidth: 140,
          formatter: (row: WalletTransaction) => h('span', {}, row.character_name || '-')
        },
        {
          label: '金额',
          width: 140,
          formatter: (row: WalletTransaction) =>
            h(
              'span',
              {
                class: `font-medium ${row.amount >= 0 ? 'text-green-600' : 'text-red-500'}`
              },
              `${row.amount >= 0 ? '+' : ''}${formatISK(row.amount)}`
            )
        },
        {
          prop: 'balance_after',
          label: '余额',
          width: 140,
          formatter: (row: WalletTransaction) => h('span', {}, formatISK(row.balance_after))
        },
        {
          prop: 'reason',
          label: '原因',
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'ref_type',
          label: '类型',
          width: 120,
          formatter: (row: WalletTransaction) =>
            h(ElTag, { size: 'small', type: getRefTypeTag(row.ref_type) }, () =>
              getRefTypeLabel(row.ref_type)
            )
        },
        {
          prop: 'operator_id',
          label: '操作人',
          width: 100,
          formatter: (row: WalletTransaction) =>
            h('span', {}, row.operator_id === 0 ? '系统' : `#${row.operator_id}`)
        },
        {
          prop: 'created_at',
          label: '时间',
          width: 200,
          formatter: (row: WalletTransaction) => h('span', {}, formatTime(row.created_at))
        }
      ]
    }
  })

  const handleSearch = () => {
    Object.assign(searchParams, {
      user_id: filterForm.user_id ? Number(filterForm.user_id) : undefined,
      ref_type: filterForm.ref_type || undefined
    })
    getData()
  }

  const filterByUser = (userId: number) => {
    filterForm.user_id = String(userId)
    handleSearch()
  }

  defineExpose({ filterByUser })
</script>
