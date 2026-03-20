<template>
  <div class="corp-npc-kills-page art-full-height">
    <!-- 日期范围筛选 -->
    <ElCard class="art-card" shadow="never">
      <div class="flex items-center gap-4 flex-wrap">
        <ElDatePicker
          v-model="dateRange"
          type="daterange"
          :start-placeholder="$t('npcKill.startDate')"
          :end-placeholder="$t('npcKill.endDate')"
          format="YYYY-MM-DD"
          value-format="YYYY-MM-DD"
          style="width: 280px"
        />

        <ElButton type="primary" :loading="loading" @click="handleSearch">
          {{ $t('npcKill.search') }}
        </ElButton>
        <ElButton @click="handleReset">{{ $t('npcKill.reset') }}</ElButton>
      </div>
    </ElCard>

    <!-- 总览卡片 -->
    <div v-if="reportData" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4 my-4">
      <ElCard shadow="never" class="text-center">
        <p class="text-sm text-gray-500">{{ $t('npcKill.totalBounty') }}</p>
        <p class="text-xl font-bold text-green-600 mt-1">{{
          formatISK(reportData.summary.total_bounty)
        }}</p>
      </ElCard>
      <ElCard shadow="never" class="text-center">
        <p class="text-sm text-gray-500">{{ $t('npcKill.totalESS') }}</p>
        <p class="text-xl font-bold text-blue-600 mt-1">{{
          formatISK(reportData.summary.total_ess)
        }}</p>
      </ElCard>
      <ElCard shadow="never" class="text-center">
        <p class="text-sm text-gray-500">{{ $t('npcKill.totalTax') }}</p>
        <p class="text-xl font-bold text-red-500 mt-1">{{
          formatISK(reportData.summary.total_tax)
        }}</p>
      </ElCard>
      <ElCard shadow="never" class="text-center">
        <p class="text-sm text-gray-500">{{ $t('npcKill.actualIncome') }}</p>
        <p class="text-xl font-bold text-green-600 mt-1">{{
          formatISK(reportData.summary.actual_income)
        }}</p>
      </ElCard>
      <ElCard shadow="never" class="text-center">
        <p class="text-sm text-gray-500">{{ $t('npcKill.totalRecords') }}</p>
        <p class="text-xl font-bold mt-1">{{ reportData.summary.total_records }}</p>
      </ElCard>
      <ElCard shadow="never" class="text-center">
        <p class="text-sm text-gray-500">{{ $t('npcKill.estimatedHours') }}</p>
        <p class="text-xl font-bold mt-1">{{ reportData.summary.estimated_hours }}</p>
      </ElCard>
    </div>

    <!-- 成员统计表格 -->
    <ElCard v-if="reportData" shadow="never" class="art-table-card mb-4">
      <template #header>
        <span class="font-medium">{{ $t('npcKill.members') }}</span>
      </template>
      <ElTable
        :data="reportData.members"
        stripe
        border
        max-height="500"
        :default-sort="{ prop: 'actual_income', order: 'descending' }"
      >
        <ElTableColumn type="index" width="55" label="#" align="center" />
        <ElTableColumn
          prop="character_name"
          :label="$t('npcKill.characterName')"
          min-width="140"
          show-overflow-tooltip
        />
        <ElTableColumn
          prop="total_bounty"
          :label="$t('npcKill.totalBounty')"
          width="160"
          align="right"
          sortable
        >
          <template #default="{ row }">
            <span class="text-green-600 font-medium">{{ formatISK(row.total_bounty) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn
          prop="total_ess"
          :label="$t('npcKill.totalESS')"
          width="160"
          align="right"
          sortable
        >
          <template #default="{ row }">
            <span class="text-blue-600 font-medium">{{ formatISK(row.total_ess) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn
          prop="total_tax"
          :label="$t('npcKill.totalTax')"
          width="140"
          align="right"
          sortable
        >
          <template #default="{ row }">
            <span class="text-red-500 font-medium">{{ formatISK(row.total_tax) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn
          prop="actual_income"
          :label="$t('npcKill.actualIncome')"
          width="160"
          align="right"
          sortable
        >
          <template #default="{ row }">
            <span class="text-green-600 font-bold">{{ formatISK(row.actual_income) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn
          prop="record_count"
          :label="$t('npcKill.recordCount')"
          width="100"
          align="right"
          sortable
        />
      </ElTable>
    </ElCard>

    <div v-if="reportData" class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-4">
      <!-- 按地点分类 -->
      <ElCard shadow="never" class="art-table-card">
        <template #header>
          <span class="font-medium">{{ $t('npcKill.bySystem') }}</span>
        </template>
        <ElTable :data="reportData.by_system" stripe border max-height="400">
          <ElTableColumn type="index" width="55" label="#" align="center" />
          <ElTableColumn
            prop="solar_system_name"
            :label="$t('npcKill.solarSystem')"
            min-width="160"
            show-overflow-tooltip
          />
          <ElTableColumn
            prop="count"
            :label="$t('npcKill.systemCount')"
            width="100"
            align="right"
            sortable
          />
          <ElTableColumn
            prop="amount"
            :label="$t('npcKill.systemAmount')"
            width="160"
            align="right"
            sortable
          >
            <template #default="{ row }">
              <span class="text-green-600 font-medium">{{ formatISK(row.amount) }}</span>
            </template>
          </ElTableColumn>
        </ElTable>
      </ElCard>

      <!-- 时间趋势 -->
      <ElCard shadow="never" class="art-table-card">
        <template #header>
          <span class="font-medium">{{ $t('npcKill.trend') }}</span>
        </template>
        <ElTable :data="reportData.trend" stripe border max-height="400">
          <ElTableColumn prop="date" :label="$t('npcKill.trendDate')" width="140" />
          <ElTableColumn
            prop="amount"
            :label="$t('npcKill.trendAmount')"
            min-width="160"
            align="right"
            sortable
          >
            <template #default="{ row }">
              <span class="text-green-600 font-medium">{{ formatISK(row.amount) }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn
            prop="count"
            :label="$t('npcKill.trendCount')"
            width="100"
            align="right"
            sortable
          />
        </ElTable>
      </ElCard>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ElDatePicker } from 'element-plus'
  import { fetchCorpNpcKills } from '@/api/npc-kill'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'CorpNpcKillReport' })

  const { t } = useI18n()

  // ─── 状态 ───
  const dateRange = ref<[string, string] | null>(null)
  const reportData = ref<Api.NpcKill.NpcKillCorpResponse | null>(null)
  const loading = ref(false)

  // ─── ISK 格式化 ───
  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v)

  // ─── 加载数据 ───
  const loadData = async () => {
    loading.value = true
    try {
      const params: Api.NpcKill.NpcKillCorpRequest = {}
      if (dateRange.value) {
        params.start_date = dateRange.value[0]
        params.end_date = dateRange.value[1]
      }
      reportData.value = (await fetchCorpNpcKills(params)) ?? null
    } catch {
      reportData.value = null
    } finally {
      loading.value = false
    }
  }

  const handleSearch = () => {
    loadData()
  }

  const handleReset = () => {
    dateRange.value = null
    loadData()
  }

  // ─── 初始化 ───
  onMounted(() => {
    loadData()
  })
</script>
