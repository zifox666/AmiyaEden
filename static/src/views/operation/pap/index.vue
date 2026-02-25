<!-- 我的 PAP 记录页面 -->
<template>
  <div class="pap-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <div>
            <h2 class="text-lg font-medium">{{ $t('fleet.pap.myTitle') }}</h2>
          </div>
          <ElButton :loading="loading" @click="loadPapLogs">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
        </div>
      </template>

      <!-- 汇总统计 -->
      <div class="flex items-center gap-6 mb-4 px-2">
        <div class="text-center">
          <p class="text-2xl font-bold text-primary">{{ totalPap }}</p>
          <p class="text-xs text-gray-500 mt-1">总 PAP</p>
        </div>
        <div class="text-center">
          <p class="text-2xl font-bold text-green-600">{{ papLogs.length }}</p>
          <p class="text-xs text-gray-500 mt-1">参与次数</p>
        </div>
      </div>

      <!-- PAP 表格 -->
      <ElTable v-loading="loading" :data="papLogs" stripe border style="width: 100%">
        <ElTableColumn type="index" width="60" label="#" />
        <ElTableColumn prop="fleet_id" label="舰队 ID" min-width="260">
          <template #default="{ row }">
            <code class="text-xs">{{ row.fleet_id }}</code>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="character_id" label="角色 ID" width="120" align="center" />
        <ElTableColumn prop="pap_count" :label="$t('fleet.pap.count')" width="120" align="center">
          <template #default="{ row }">
            <ElTag type="success" size="small">+{{ row.pap_count }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="issued_by" :label="$t('fleet.pap.issuedBy')" width="120" align="center" />
        <ElTableColumn prop="created_at" :label="$t('fleet.pap.issuedAt')" width="200">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </ElTableColumn>
      </ElTable>

      <ElEmpty v-if="!loading && papLogs.length === 0" :description="$t('fleet.pap.empty')" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { Refresh } from '@element-plus/icons-vue'
  import { ElCard, ElTable, ElTableColumn, ElTag, ElButton, ElEmpty } from 'element-plus'
  import { fetchMyPapLogs } from '@/api/fleet'

  defineOptions({ name: 'MyPap' })

  const papLogs = ref<Api.Fleet.PapLog[]>([])
  const loading = ref(false)

  const totalPap = computed(() => papLogs.value.reduce((sum, p) => sum + p.pap_count, 0))

  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')

  const loadPapLogs = async () => {
    loading.value = true
    try {
      papLogs.value = (await fetchMyPapLogs()) ?? []
    } catch {
      papLogs.value = []
    } finally {
      loading.value = false
    }
  }

  onMounted(loadPapLogs)
</script>
