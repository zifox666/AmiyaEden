<template>
  <div class="art-card p-5"
    ><div class="flex-b mb-4"
      ><div
        ><h2 class="text-xl">{{ t('at.br.allTitle') }}</h2
        ><p class="text-sm text-g-500 mt-1">{{ t('at.br.allDescription') }}</p></div
      ><ElButton @click="load">{{ t('at.common.refresh') }}</ElButton></div
    ><ElTable :data="items" v-loading="loading"
      ><ElTableColumn :label="t('at.matches.teams')" min-width="240"
        ><template #default="{ row }"
          >{{ row.team_a_name }} vs {{ row.team_b_name }}</template
        ></ElTableColumn
      ><ElTableColumn prop="round_no" :label="t('at.matches.round')" width="100" /><ElTableColumn
        prop="version"
        :label="t('at.scoreRules.version')"
        width="100"
      /><ElTableColumn prop="status" :label="t('at.common.status')" width="130" /><ElTableColumn
        prop="created_at"
        :label="t('at.scoreRules.publishedAt')"
        min-width="190"
      /><ElTableColumn :label="t('at.common.actions')" width="120"
        ><template #default="{ row }"
          ><ElButton link type="primary" @click="router.push(`/at/br/${row.round_id}`)">{{
            t('at.common.view')
          }}</ElButton></template
        ></ElTableColumn
      ></ElTable
    ></div
  >
</template>
<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { atApi } from '@/api/at'
  const { t } = useI18n(),
    router = useRouter(),
    loading = ref(false),
    items = ref<Api.AT.BRListItem[]>([])
  async function load() {
    loading.value = true
    try {
      items.value = await atApi.listBRs()
    } finally {
      loading.value = false
    }
  }
  onMounted(load)
</script>
