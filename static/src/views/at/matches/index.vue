<template>
  <div class="art-card p-5">
    <div class="flex-b mb-4"
      ><div
        ><h2 class="text-xl">{{ t('at.matches.title') }}</h2
        ><p class="text-sm text-g-500 mt-1">{{ t('at.matches.workflowDescription') }}</p></div
      ><ElButton v-if="canManage" type="primary" @click="openCreate">{{
        t('at.matches.create')
      }}</ElButton></div
    >
    <ArtTable :data="matches" :loading="loading" :columns="columns">
      <template #uuid="{ row }"
        ><code>{{ row.uuid.slice(0, 8) }}</code></template
      >
      <template #teams="{ row }">{{ row.team_a_name }} vs {{ row.team_b_name }}</template>
      <template #score="{ row }">{{ row.team_a_wins }} : {{ row.team_b_wins }}</template>
      <template #status="{ row }"
        ><ElTag>{{ t(`at.status.${row.status}`) }}</ElTag></template
      >
      <template #operation="{ row }"
        ><ElButton link type="primary" @click="router.push(`/at/matches/${row.uuid}`)">{{
          t('at.common.details')
        }}</ElButton></template
      >
    </ArtTable>
    <ElDialog v-model="dialogVisible" :title="t('at.matches.create')" width="680px">
      <ElForm label-position="top"
        ><ElRow :gutter="14"
          ><ElCol :span="12"
            ><ElFormItem :label="t('at.matches.season')"
              ><ElSelect v-model="form.season_id" class="w-full"
                ><ElOption
                  v-for="x in seasons"
                  :key="x.id"
                  :label="x.name"
                  :value="x.id" /></ElSelect></ElFormItem></ElCol
          ><ElCol :span="12"
            ><ElFormItem :label="t('at.matches.matchType')"
              ><ElSelect v-model="form.match_type_id" class="w-full"
                ><ElOption
                  v-for="x in activeMatchTypes"
                  :key="x.id"
                  :label="x.name"
                  :value="x.id" /></ElSelect></ElFormItem></ElCol
        ></ElRow>
        <ElRow :gutter="14"
          ><ElCol :span="12"
            ><ElFormItem :label="t('at.common.format')"
              ><ElSelect v-model="form.format" class="w-full"
                ><ElOption label="10v10" value="10v10" /><ElOption
                  label="7v7"
                  value="7v7" /></ElSelect></ElFormItem></ElCol
          ><ElCol :span="12"
            ><ElFormItem :label="t('at.common.series')"
              ><ElSelect v-model="form.series" class="w-full"
                ><ElOption label="BO3" value="BO3" /><ElOption
                  label="BO5"
                  value="BO5" /></ElSelect></ElFormItem></ElCol
        ></ElRow>
        <ElFormItem :label="t('at.matches.set')"
          ><ElSelect v-model="form.set_id" class="w-full"
            ><ElOption
              v-for="x in compatibleSets"
              :key="x.id"
              :label="x.name"
              :value="x.id" /></ElSelect
        ></ElFormItem>
        <ElRow :gutter="14"
          ><ElCol :span="12"
            ><ElFormItem :label="t('at.matches.teamA')"
              ><ElInput v-model="form.team_a_name" /></ElFormItem></ElCol
          ><ElCol :span="12"
            ><ElFormItem :label="t('at.matches.teamB')"
              ><ElInput v-model="form.team_b_name" /></ElFormItem></ElCol
        ></ElRow>
        <ElFormItem
          ><ElCheckbox v-model="form.allow_self_assign">{{
            t('at.matches.allowSelfAssign')
          }}</ElCheckbox></ElFormItem
        >
      </ElForm>
      <template #footer
        ><ElButton @click="dialogVisible = false">{{ t('at.common.cancel') }}</ElButton
        ><ElButton type="primary" :loading="saving" @click="create">{{
          t('at.common.create')
        }}</ElButton></template
      >
    </ElDialog>
  </div>
</template>
<script setup lang="ts">
  import type { ColumnOption } from '@/types'
  import { useI18n } from 'vue-i18n'
  import { atApi } from '@/api/at'
  import { useUserStore } from '@/store/modules/user'
  const { t } = useI18n(),
    router = useRouter(),
    user = useUserStore()
  const canManage = computed(() =>
    (user.info.roles || []).some((r) => ['fc', 'admin', 'super_admin'].includes(r))
  )
  const loading = ref(false),
    saving = ref(false),
    dialogVisible = ref(false),
    matches = ref<Api.AT.Match[]>([]),
    sets = ref<Api.AT.Set[]>([]),
    seasons = ref<Api.AT.Season[]>([]),
    matchTypes = ref<Api.AT.MatchType[]>([])
  const form = reactive<Api.AT.MatchCreate>({
    season_id: 0,
    match_type_id: 0,
    format: '10v10',
    series: 'BO3',
    team_a_name: '',
    team_b_name: '',
    set_id: 0,
    allow_self_assign: true
  })
  const activeMatchTypes = computed(() => matchTypes.value.filter((x) => x.active)),
    compatibleSets = computed(() => sets.value.filter((x) => x.format === form.format))
  const columns = computed<ColumnOption[]>(() => [
    { prop: 'uuid', label: 'UUID', width: 110, useSlot: true },
    { prop: 'teams', label: t('at.matches.teams'), minWidth: 240, useSlot: true },
    { prop: 'format', label: t('at.common.format'), width: 100 },
    { prop: 'series', label: t('at.common.series'), width: 90 },
    { prop: 'score', label: t('at.matches.score'), width: 100, useSlot: true },
    { prop: 'status', label: t('at.common.status'), width: 130, useSlot: true },
    { prop: 'operation', label: t('at.common.actions'), width: 110, useSlot: true }
  ])
  async function load() {
    loading.value = true
    try {
      ;[matches.value, sets.value, seasons.value, matchTypes.value] = await Promise.all([
        atApi.listMatches(),
        atApi.listSets(),
        atApi.listSeasons(),
        atApi.listMatchTypes()
      ])
    } finally {
      loading.value = false
    }
  }
  function openCreate() {
    form.season_id = seasons.value.find((x) => x.is_current)?.id || seasons.value[0]?.id || 0
    form.match_type_id =
      matchTypes.value.find((x) => x.active && x.name.toLowerCase() === 'training')?.id ||
      activeMatchTypes.value[0]?.id ||
      0
    form.format = '10v10'
    form.series = 'BO3'
    form.allow_self_assign = true
    form.set_id = compatibleSets.value[0]?.id || 0
    dialogVisible.value = true
  }
  async function create() {
    saving.value = true
    try {
      const created = await atApi.createMatch(form)
      dialogVisible.value = false
      await router.push(`/at/matches/${created.uuid}`)
    } finally {
      saving.value = false
    }
  }
  onMounted(load)
</script>
