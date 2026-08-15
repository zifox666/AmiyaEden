<template>
  <div class="art-card p-5">
    <div class="flex-b mb-4"
      ><div
        ><h2 class="text-xl">{{ t('at.km.reviewTitle') }}</h2
        ><p class="text-sm text-g-500 mt-1">{{ t('at.km.reviewDescription') }}</p></div
      ><ElButton @click="load">{{ t('at.common.refresh') }}</ElButton></div
    >
    <ArtTable
      :data="pagedItems"
      :loading="loading"
      :columns="columns"
      :pagination="pagination"
      @pagination:size-change="changePageSize"
      @pagination:current-change="pagination.current = $event"
    >
      <template #resolution_status="{ row }"
        ><ElTag>{{ row.resolution_status }}</ElTag></template
      >
      <template #assign="{ row }">
        <div class="flex gap-2">
          <ElInputNumber
            v-model="assignments[row.km_id].match"
            :placeholder="t('at.km.matchId')"
            :min="1"
            controls-position="right"
          />
          <ElInputNumber
            v-model="assignments[row.km_id].round"
            :placeholder="t('at.km.roundId')"
            :min="1"
            controls-position="right"
          />
          <ElButton type="primary" @click="assign(row as Api.AT.Killmail)">{{
            t('at.km.assign')
          }}</ElButton>
        </div>
      </template>
    </ArtTable>
  </div>
</template>
<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { ElMessage } from 'element-plus'
  import type { ColumnOption } from '@/types'
  import { atApi } from '@/api/at'
  const { t } = useI18n(),
    items = ref<Api.AT.Killmail[]>([]),
    loading = ref(false),
    assignments = reactive<Record<string, { match: number; round: number }>>({})
  const pagination = reactive({ current: 1, size: 10, total: 0 })
  const pagedItems = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return items.value.slice(start, start + pagination.size)
  })
  watch(
    () => items.value.length,
    (total) => (pagination.total = total),
    { immediate: true }
  )
  const columns = computed<ColumnOption[]>(() => [
    { prop: 'killmail_time', label: t('at.km.time'), minWidth: 190 },
    { prop: 'victim_name', label: t('at.km.victim'), minWidth: 180 },
    { prop: 'victim_ship_name', label: t('at.km.ship'), minWidth: 180 },
    { prop: 'resolution_status', label: t('at.common.status'), width: 140, useSlot: true },
    { prop: 'assign', label: t('at.km.assign'), width: 430, fixed: 'right', useSlot: true }
  ])
  function changePageSize(size: number) {
    pagination.size = size
    pagination.current = 1
  }
  async function load() {
    loading.value = true
    try {
      items.value = await atApi.listPendingKillmails()
      for (const item of items.value) assignments[item.km_id] ??= { match: 0, round: 0 }
    } finally {
      loading.value = false
    }
  }
  async function assign(row: Api.AT.Killmail) {
    const value = assignments[row.km_id]
    const result = await atApi.assignKillmail(row.km_id, value.match, value.round)
    ElMessage.success(`${t('at.km.assigned')}: ${result.status}`)
    await load()
  }
  onMounted(load)
</script>
