<template>
  <div class="score-rules-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <div class="flex-b mb-4">
        <div>
          <h2 class="text-xl">{{ t('at.scoreRules.title') }}</h2>
          <p class="text-sm text-g-500 mt-1">{{ t('at.scoreRules.description') }}</p>
        </div>
        <ElUpload
          v-if="canManage"
          :show-file-list="false"
          :auto-upload="false"
          accept=".xlsx"
          :on-change="upload"
        >
          <ElButton type="primary" :loading="uploading">{{ t('at.scoreRules.import') }}</ElButton>
        </ElUpload>
      </div>
      <ElAlert type="info" :closable="false" class="mb-4">
        <template #title>
          <a :href="scoreSource" target="_blank" rel="noopener">{{ t('at.scoreRules.source') }}</a>
        </template>
      </ElAlert>
      <div class="flex-b mb-3"
        ><div
          ><h3 class="text-lg">{{ t('at.scoreRules.currentShips') }}</h3
          ><span class="text-sm text-g-500">{{ currentDetail?.rule.version || '-' }}</span></div
        ><ElInput
          v-model="shipKeyword"
          clearable
          :placeholder="t('at.scoreRules.searchShip')"
          style="max-width: 320px"
      /></div>
      <ElTable :data="filteredCurrentShips" v-loading="loading" max-height="560" class="mb-4">
        <ElTableColumn prop="type_name" :label="t('at.scoreRules.shipName')" min-width="240"
          ><template #default="{ row }"
            ><div class="flex items-center gap-2"
              ><img
                :src="`https://images.evetech.net/types/${row.type_id}/icon?size=32`"
                class="ship-icon"
              /><span>{{ row.type_name }}</span></div
            ></template
          ></ElTableColumn
        >
        <ElTableColumn prop="base_points" :label="t('at.scoreRules.points')" width="120" sortable />
        <ElTableColumn prop="size_code" :label="t('at.scoreRules.hullClass')" width="140" />
        <ElTableColumn prop="type_id" label="TypeID" width="130" />
      </ElTable>
      <ElCollapse>
        <ElCollapseItem name="versions" :title="t('at.scoreRules.versionManagement')">
          <ArtTable
            :data="pagedRules"
            :loading="loading"
            :columns="columns"
            :pagination="pagination"
            @pagination:size-change="handleSizeChange"
            @pagination:current-change="handleCurrentChange"
          >
            <template #status="{ row }">
              <ElTag :type="row.status === 'published' ? 'success' : 'info'">
                {{ row.status === 'published' ? t('at.common.published') : t('at.common.draft') }}
              </ElTag>
            </template>
            <template #operation="{ row }">
              <div class="flex items-center gap-2">
                <ArtButtonTable
                  icon="ri:eye-line"
                  icon-class="bg-theme/12 text-theme"
                  :title="t('at.common.details')"
                  @click="showDetail(row.id)"
                />
                <ArtButtonTable
                  v-if="canManage && row.status === 'draft'"
                  icon="ri:send-plane-line"
                  icon-class="bg-success/12 text-success"
                  :title="t('at.scoreRules.publish')"
                  @click="publish(row.id)"
                />
              </div>
            </template>
          </ArtTable>
        </ElCollapseItem>
      </ElCollapse>
    </ElCard>

    <ElDialog v-model="detailVisible" :title="detail?.rule.version ?? ''" width="860px">
      <ElDescriptions v-if="detail" :column="2" border class="mb-4">
        <ElDescriptionsItem :label="t('at.scoreRules.version')">{{
          detail.rule.version
        }}</ElDescriptionsItem>
        <ElDescriptionsItem :label="t('at.common.status')">{{
          detail.rule.status
        }}</ElDescriptionsItem>
        <ElDescriptionsItem :label="t('at.scoreRules.shipCount')">{{
          detail.ships.length
        }}</ElDescriptionsItem>
        <ElDescriptionsItem :label="t('at.scoreRules.bannedCount')">{{
          detail.banned.length
        }}</ElDescriptionsItem>
        <ElDescriptionsItem :label="t('at.scoreRules.hash')" :span="2">{{
          detail.rule.file_hash
        }}</ElDescriptionsItem>
      </ElDescriptions>
      <ElTable v-if="detail" :data="detail.ships" max-height="480">
        <ElTableColumn prop="type_name" :label="t('at.scoreRules.shipName')" min-width="240" />
        <ElTableColumn prop="base_points" :label="t('at.scoreRules.points')" width="120" sortable />
        <ElTableColumn prop="size_code" :label="t('at.scoreRules.hullClass')" width="140" />
        <ElTableColumn prop="type_id" label="TypeID" width="130" />
      </ElTable>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import type { UploadFile } from 'element-plus'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import type { ColumnOption } from '@/types'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { atApi } from '@/api/at'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'ATScoreRules' })

  const scoreSource =
    'https://docs.google.com/spreadsheets/d/1AVYlWlvuMKnA3yuqqDCcAkia8pvhpb9OBcM29WFw5rM'
  const { t } = useI18n()
  const userStore = useUserStore()
  const canManage = computed(() =>
    (userStore.info.roles ?? []).some((role) => ['admin', 'super_admin'].includes(role))
  )
  const rules = ref<Api.AT.ScoreRule[]>([])
  const detail = ref<Api.AT.ScoreRuleDetail>()
  const detailVisible = ref(false)
  const currentDetail = ref<Api.AT.ScoreRuleDetail>()
  const shipKeyword = ref('')
  const loading = ref(false)
  const uploading = ref(false)
  const pagination = reactive({ current: 1, size: 10, total: 0 })
  const pagedRules = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return rules.value.slice(start, start + pagination.size)
  })
  const filteredCurrentShips = computed(() => {
    const keyword = shipKeyword.value.trim().toLowerCase()
    return (currentDetail.value?.ships ?? []).filter(
      (ship) =>
        !keyword ||
        ship.type_name.toLowerCase().includes(keyword) ||
        String(ship.type_id).includes(keyword)
    )
  })
  watch(
    () => rules.value.length,
    (total) => (pagination.total = total),
    { immediate: true }
  )
  const columns = computed<ColumnOption[]>(() => [
    { type: 'globalIndex', width: 70, label: '#' },
    { prop: 'version', label: t('at.scoreRules.version'), minWidth: 140 },
    { prop: 'file_hash', label: t('at.scoreRules.hash'), minWidth: 280, showOverflowTooltip: true },
    { prop: 'status', label: t('at.common.status'), width: 130, useSlot: true },
    { prop: 'published_at', label: t('at.scoreRules.publishedAt'), minWidth: 190 },
    { prop: 'operation', label: t('at.common.actions'), width: 140, fixed: 'right', useSlot: true }
  ])

  async function load() {
    loading.value = true
    try {
      rules.value = await atApi.listScoreRules()
      const current = rules.value.find((rule) => rule.status === 'published')
      currentDetail.value = current ? await atApi.getScoreRule(current.id) : undefined
    } finally {
      loading.value = false
    }
  }

  async function showDetail(id: number) {
    detail.value = await atApi.getScoreRule(id)
    detailVisible.value = true
  }

  async function upload(file: UploadFile) {
    if (!file.raw) return
    uploading.value = true
    try {
      await atApi.importScoreRule(file.raw)
      ElMessage.success(t('at.scoreRules.imported'))
      await load()
    } finally {
      uploading.value = false
    }
  }

  async function publish(id: number) {
    await ElMessageBox.confirm(t('at.scoreRules.publishConfirm'), t('at.common.confirm'))
    await atApi.publishScoreRule(id)
    ElMessage.success(t('at.scoreRules.published'))
    await load()
  }

  function handleSizeChange(size: number) {
    pagination.size = size
    pagination.current = 1
  }

  function handleCurrentChange(current: number) {
    pagination.current = current
  }

  onMounted(load)
</script>

<style scoped>
  .ship-icon {
    width: 32px;
    height: 32px;
    border-radius: 4px;
  }
</style>
