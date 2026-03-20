<!-- 系统管理 - 联盟 PAP 所有成员视图 -->
<template>
  <div class="alliance-pap-page art-full-height">
    <!-- 搜索栏（含拉取按钮） -->
    <PapSearch
      v-model="searchForm"
      :fetching="fetching"
      @search="handleSearch"
      @reset="resetSearchParams"
      @fetch="triggerFetch"
      @import="handleImport"
    />

    <ElCard class="art-table-card" shadow="never">
      <!-- 工具栏 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElButton type="warning" :icon="Setting" @click="settleDialogVisible = true">
            {{ t('alliancePap.settle.openBtn') }}
          </ElButton>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <!-- 兑换配置 & 月度结算弹窗 -->
    <PapSettle v-model="settleDialogVisible" />
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { ElTag, ElMessage, ElButton } from 'element-plus'
  import { Setting } from '@element-plus/icons-vue'
  import { useI18n } from 'vue-i18n'
  import {
    fetchAllAlliancePAP,
    triggerAlliancePAPFetch,
    importAlliancePAP,
    type AlliancePAPSummary,
    type PAPImportInfo
  } from '@/api/alliance-pap'
  import PapSearch from './modules/pap-search.vue'
  import PapSettle from './modules/pap-settle.vue'

  defineOptions({ name: 'AlliancePAP' })

  const { t } = useI18n()

  const now = new Date()
  const currentMonth = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`

  // ─── 搜索 ───
  const searchForm = ref<Record<string, any>>({ month: currentMonth })

  function parseMonth() {
    const month = searchForm.value.month || currentMonth
    const [yearStr, monthStr] = month.split('-')
    return { year: Number(yearStr), month: Number(monthStr) }
  }

  // ─── 表格 ───
  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    searchParams,
    resetSearchParams,
    getData,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchAllAlliancePAP,
      apiParams: { current: 1, size: 20, ...parseMonth() },
      columnsFactory: () => [
        { type: 'index', width: 60, label: t('alliancePap.columns.rank') },
        {
          prop: 'main_character',
          label: t('alliancePap.columns.mainCharacter'),
          minWidth: 140
        },
        {
          prop: 'total_pap',
          label: t('alliancePap.columns.monthlyPap'),
          width: 100,
          sortable: true,
          formatter: (row: AlliancePAPSummary) =>
            h(ElTag, { type: 'success', size: 'small' }, () => row.total_pap)
        },
        {
          prop: 'yearly_total_pap',
          label: t('alliancePap.columns.yearlyPap'),
          width: 110,
          sortable: true,
          formatter: (row: AlliancePAPSummary) =>
            h(ElTag, { type: 'info', size: 'small' }, () => row.yearly_total_pap)
        },
        {
          prop: 'monthly_rank',
          label: t('alliancePap.columns.corpMonthlyRank'),
          width: 110,
          sortable: true,
          formatter: (row: AlliancePAPSummary) =>
            h('span', [
              h('span', { class: 'text-green-600 font-medium' }, `#${row.monthly_rank}`),
              h('span', { class: 'text-xs text-gray-400' }, ` / ${row.total_in_corp}`)
            ])
        },
        {
          prop: 'global_monthly_rank',
          label: t('alliancePap.columns.allianceMonthlyRank'),
          width: 120,
          sortable: true,
          formatter: (row: AlliancePAPSummary) =>
            h('span', [
              h('span', { class: 'text-yellow-500 font-medium' }, `#${row.global_monthly_rank}`),
              h('span', { class: 'text-xs text-gray-400' }, ` / ${row.total_global}`)
            ])
        },
        {
          prop: 'yearly_rank',
          label: t('alliancePap.columns.corpYearlyRank'),
          width: 110,
          sortable: true,
          formatter: (row: AlliancePAPSummary) =>
            h('span', { class: 'text-purple-500 font-medium' }, `#${row.yearly_rank}`)
        },
        {
          prop: 'global_yearly_rank',
          label: t('alliancePap.columns.allianceYearlyRank'),
          width: 110,
          sortable: true,
          formatter: (row: AlliancePAPSummary) =>
            h('span', { class: 'text-blue-500 font-medium' }, `#${row.global_yearly_rank}`)
        },
        {
          prop: 'calculated_at',
          label: t('alliancePap.columns.calculatedAt'),
          width: 170,
          formatter: (row: AlliancePAPSummary) =>
            row.calculated_at ? new Date(row.calculated_at).toLocaleString() : '-'
        },
        {
          prop: 'is_archived',
          label: t('alliancePap.columns.status'),
          width: 90,
          formatter: (row: AlliancePAPSummary) =>
            h(ElTag, { type: row.is_archived ? 'info' : 'success', size: 'small' }, () =>
              row.is_archived ? t('alliancePap.status.archived') : t('alliancePap.status.current')
            )
        }
      ]
    }
  })

  // ─── 搜索 / 重置 ───
  function handleSearch() {
    const { year, month } = parseMonth()
    Object.assign(searchParams, { year, month })
    getData()
  }

  // ─── 拉取最新数据 ───
  const fetching = ref(false)
  const settleDialogVisible = ref(false)

  async function triggerFetch() {
    fetching.value = true
    try {
      await triggerAlliancePAPFetch(parseMonth())
      ElMessage.success(t('alliancePap.fetchTriggered'))
    } catch {
      ElMessage.error(t('alliancePap.fetchFailed'))
    } finally {
      fetching.value = false
    }
  }

  // ─── 从表格/SEAT导入 PAP ───
  const handleImport = async (rows: Record<string, unknown>[]) => {
    const { year, month } = parseMonth()
    const items = rows
      .map((row) => ({
        primary_character_name: String(row['主角色'] ?? row['primary_character_name'] ?? ''),
        monthly_pap: Number(row['月 PAP'] ?? row['monthly_pap'] ?? 0),
        calculated_at: String(row['数据时间'] ?? row['calculated_at'] ?? 0)
      }))
      .filter((item) => item.primary_character_name && item.calculated_at != '')
    if (!items.length) {
      ElMessage.warning(t('alliancePap.importNoData'))
      return
    }
    fetching.value = true
    let success = 0
    try {
      for (const item of items) {
        const { primary_character_name, monthly_pap, calculated_at } = item
        try {
          await importAlliancePAP({
            year,
            month,
            data: { primary_character_name, monthly_pap, calculated_at }
          })
        } catch (err: any) {
          if (err.message == '主角色不存在' || err.message == '未设置主角色') {
            continue
          }
          throw err
        }
        success++
      }
      ElMessage.success(t('alliancePap.importSuccess', { count: success }))
      handleSearch()
    } catch {
      ElMessage.error(t('alliancePap.importFailed'))
    } finally {
      fetching.value = false
    }
  }
</script>
