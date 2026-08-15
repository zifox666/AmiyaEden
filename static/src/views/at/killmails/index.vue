<template>
  <div class="art-card p-5"
    ><div class="flex-b mb-4"
      ><div
        ><h2 class="text-xl">{{ t('at.km.allTitle') }}</h2
        ><p class="text-sm text-g-500 mt-1">{{ t('at.km.allDescription') }}</p></div
      ><ElButton @click="load">{{ t('at.common.refresh') }}</ElButton></div
    ><ElInput
      v-model="keyword"
      clearable
      :placeholder="t('at.km.search')"
      class="mb-3"
      style="max-width: 340px" /><ElTable :data="filtered" v-loading="loading"
      ><ElTableColumn prop="killmail_time" :label="t('at.km.time')" min-width="190" /><ElTableColumn
        prop="victim_name"
        :label="t('at.km.victim')"
        min-width="180" /><ElTableColumn
        prop="victim_ship_name"
        :label="t('at.km.ship')"
        min-width="180" /><ElTableColumn
        prop="resolution_status"
        :label="t('at.common.status')"
        width="150"
        ><template #default="{ row }"
          ><ElTag>{{ row.resolution_status }}</ElTag></template
        ></ElTableColumn
      ><ElTableColumn
        prop="assigned_round_id"
        :label="t('at.km.roundId')"
        width="120" /><ElTableColumn prop="km_id" label="KMID" min-width="300" /></ElTable
  ></div>
</template>
<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { atApi } from '@/api/at'
  const { t } = useI18n(),
    loading = ref(false),
    keyword = ref(''),
    items = ref<Api.AT.Killmail[]>([])
  const filtered = computed(() => {
    const q = keyword.value.toLowerCase()
    return items.value.filter(
      (x) => !q || `${x.victim_name} ${x.victim_ship_name} ${x.km_id}`.toLowerCase().includes(q)
    )
  })
  async function load() {
    loading.value = true
    try {
      items.value = await atApi.listKillmails()
    } finally {
      loading.value = false
    }
  }
  onMounted(load)
</script>
