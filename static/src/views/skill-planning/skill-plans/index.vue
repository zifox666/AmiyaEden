<template>
  <div class="skill-plan-page art-full-height">
    <div class="skill-plan-layout">
      <ElCard class="skill-plan-list-card" shadow="never">
        <div class="skill-plan-list-toolbar">
          <ElInput
            v-model="keyword"
            :placeholder="$t('skillPlan.searchPlaceholder')"
            clearable
            @keyup.enter="handleSearch"
            @clear="handleSearch"
          />
          <ElButton v-if="canManage" type="primary" :icon="Plus" @click="openCreateDialog">
            {{ $t('skillPlan.create') }}
          </ElButton>
        </div>

        <div class="skill-plan-list-header">
          <span>{{ $t('skillPlan.listTitle') }}</span>
          <span>{{ pagination.total }}</span>
        </div>

        <div v-loading="listLoading" class="skill-plan-list">
          <template v-if="plans.length">
            <button
              v-for="plan in plans"
              :key="plan.id"
              type="button"
              class="skill-plan-item"
              :class="{ active: plan.id === selectedPlanId }"
              @click="selectPlan(plan.id)"
            >
              <div class="skill-plan-item__heading">
                <img
                  v-if="plan.ship_type_id"
                  :src="getShipIconUrl(plan.ship_type_id, 64)"
                  class="skill-plan-item__icon"
                  alt=""
                  loading="lazy"
                />
                <span class="skill-plan-item__title">{{ plan.title }}</span>
              </div>
            </button>
          </template>

          <ElEmpty v-else :description="$t('skillPlan.emptyList')" :image-size="72" />
        </div>

        <div class="skill-plan-pagination">
          <ElPagination
            small
            background
            layout="prev, pager, next"
            :current-page="pagination.current"
            :page-size="pagination.size"
            :total="pagination.total"
            @current-change="handlePageChange"
          />
        </div>
      </ElCard>

      <ElCard class="skill-plan-detail-card" shadow="never">
        <div v-loading="detailLoading" class="skill-plan-detail">
          <template v-if="selectedPlan">
            <div class="skill-plan-detail__header">
              <div class="skill-plan-detail__identity">
                <img
                  v-if="selectedPlan.ship_type_id"
                  :src="getShipIconUrl(selectedPlan.ship_type_id, 64)"
                  class="skill-plan-detail__icon"
                  alt=""
                  loading="lazy"
                />
                <div class="skill-plan-detail__text">
                  <div class="skill-plan-detail__eyebrow">{{ $t('skillPlan.detailTitle') }}</div>
                  <h2 class="skill-plan-detail__title">{{ selectedPlan.title }}</h2>
                  <div v-if="selectedPlan.ship_name" class="skill-plan-detail__ship-name">
                    {{ selectedPlan.ship_name }}
                  </div>
                </div>
              </div>

              <div v-if="canManage" class="skill-plan-detail__actions">
                <ElButton @click="openEditDialog">{{ $t('common.edit') }}</ElButton>
                <ElButton type="danger" @click="handleDelete">{{ $t('common.delete') }}</ElButton>
              </div>
            </div>

            <div class="skill-plan-summary">
              <div class="skill-plan-summary__item">
                <span class="label">{{ $t('skillPlan.skillCount') }}</span>
                <strong>{{ selectedPlan.skill_count }}</strong>
              </div>
              <div class="skill-plan-summary__item">
                <span class="label">{{ $t('common.updatedAt') }}</span>
                <strong>{{ formatTime(selectedPlan.updated_at) }}</strong>
              </div>
              <div class="skill-plan-summary__item">
                <span class="label">{{ $t('common.createdAt') }}</span>
                <strong>{{ formatTime(selectedPlan.created_at) }}</strong>
              </div>
            </div>

            <div class="skill-plan-description">
              <div class="skill-plan-description__label">
                {{ $t('skillPlan.fields.description') }}
              </div>
              <p>{{ selectedPlan.description || $t('skillPlan.descriptionEmpty') }}</p>
            </div>

            <div class="skill-plan-table-header">
              <span>{{ $t('skillPlan.skillListTitle') }}</span>
            </div>

            <ElTable :data="selectedPlan.skills" stripe class="skill-plan-table">
              <ElTableColumn :label="$t('skillPlan.table.skill')" min-width="280">
                <template #default="{ row }">
                  <div class="skill-cell">
                    <span class="skill-cell__name">{{ row.skill_name }}</span>
                  </div>
                </template>
              </ElTableColumn>
              <ElTableColumn
                prop="group_name"
                :label="$t('skillPlan.table.group')"
                min-width="180"
              />
              <ElTableColumn :label="$t('skillPlan.table.requiredLevel')" width="140">
                <template #default="{ row }">
                  <ElTag type="primary" effect="light">
                    {{ $t('skillPlan.level', { level: row.required_level }) }}
                  </ElTag>
                </template>
              </ElTableColumn>
            </ElTable>
          </template>

          <ElEmpty v-else :description="$t('skillPlan.emptyDetail')" :image-size="84" />
        </div>
      </ElCard>
    </div>

    <SkillPlanDialog
      v-model:visible="dialogVisible"
      :editing="editingPlan"
      @success="handleDialogSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { Plus } from '@element-plus/icons-vue'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { formatTime } from '@utils/common'
  import { deleteSkillPlan, fetchSkillPlanDetail, fetchSkillPlanList } from '@/api/skill-plan'
  import SkillPlanDialog from './modules/skill-plan-dialog.vue'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'SkillPlans' })

  type SkillPlanListItem = Api.SkillPlan.SkillPlanListItem
  type SkillPlanDetail = Api.SkillPlan.SkillPlanDetail
  type SkillPlanDisplayState = Pick<
    Api.SkillPlan.SkillPlanDetail,
    'id' | 'ship_type_id' | 'ship_name'
  >

  const { t } = useI18n()
  const userStore = useUserStore()

  const canManage = computed(() => {
    const roles = userStore.getUserInfo?.roles ?? []
    return roles.some((role) => ['super_admin', 'admin', 'fc'].includes(role))
  })

  const currentLang = computed(() => userStore.language || 'zh')

  const keyword = ref('')
  const plans = ref<SkillPlanListItem[]>([])
  const selectedPlanId = ref<number | null>(null)
  const selectedPlan = ref<SkillPlanDetail | null>(null)
  const editingPlan = ref<SkillPlanDetail | null>(null)
  const dialogVisible = ref(false)
  const listLoading = ref(false)
  const detailLoading = ref(false)
  const planDisplayState = reactive<Record<number, SkillPlanDisplayState>>({})

  const pagination = reactive({
    current: 1,
    size: 10,
    total: 0
  })

  const getShipIconUrl = (shipTypeId: number, size = 64) =>
    `https://images.evetech.net/types/${shipTypeId}/icon?size=${size}`

  function mergePlanDisplayState(plan: SkillPlanDisplayState) {
    planDisplayState[plan.id] = {
      id: plan.id,
      ship_type_id: plan.ship_type_id,
      ship_name: plan.ship_name
    }

    const index = plans.value.findIndex((item) => item.id === plan.id)
    if (index >= 0) {
      plans.value[index] = {
        ...plans.value[index],
        ship_type_id: plan.ship_type_id
      }
    }
  }

  async function hydratePlanListShipIcons(planIds: number[]) {
    const uniqueIds = [...new Set(planIds.filter(Boolean))]
    if (!uniqueIds.length) return

    const results = await Promise.allSettled(
      uniqueIds.map((id) => fetchSkillPlanDetail(id, currentLang.value))
    )

    for (const result of results) {
      if (result.status !== 'fulfilled') continue
      mergePlanDisplayState(result.value)
    }
  }

  async function loadPlans(preferredPlanId?: number | null) {
    listLoading.value = true
    try {
      const response = await fetchSkillPlanList({
        current: pagination.current,
        size: pagination.size,
        keyword: keyword.value.trim() || undefined
      })

      plans.value = response.list ?? []
      pagination.total = response.total ?? 0

      if (!plans.value.length) {
        selectedPlanId.value = null
        selectedPlan.value = null
        return
      }

      const targetId = preferredPlanId ?? selectedPlanId.value
      const nextPlanId =
        targetId && plans.value.some((plan) => plan.id === targetId) ? targetId : plans.value[0].id

      await loadPlanDetail(nextPlanId)
      void hydratePlanListShipIcons(plans.value.map((plan) => plan.id))
    } catch {
      plans.value = []
      pagination.total = 0
      selectedPlanId.value = null
      selectedPlan.value = null
    } finally {
      listLoading.value = false
    }
  }

  async function loadPlanDetail(planId: number) {
    if (!planId) return
    detailLoading.value = true
    try {
      selectedPlan.value = await fetchSkillPlanDetail(planId, currentLang.value)
      selectedPlanId.value = planId
      mergePlanDisplayState(selectedPlan.value)
    } catch (error: any) {
      selectedPlan.value = null
      ElMessage.error(error?.message ?? t('httpMsg.requestFailed'))
    } finally {
      detailLoading.value = false
    }
  }

  async function selectPlan(planId: number) {
    if (planId === selectedPlanId.value && selectedPlan.value) return
    await loadPlanDetail(planId)
  }

  function handleSearch() {
    pagination.current = 1
    loadPlans()
  }

  function handlePageChange(page: number) {
    pagination.current = page
    loadPlans()
  }

  function openCreateDialog() {
    editingPlan.value = null
    dialogVisible.value = true
  }

  function openEditDialog() {
    if (!selectedPlan.value) return
    editingPlan.value = selectedPlan.value
    dialogVisible.value = true
  }

  async function handleDelete() {
    if (!selectedPlan.value) return

    try {
      await ElMessageBox.confirm(
        t('skillPlan.deleteConfirm', { title: selectedPlan.value.title }),
        t('skillPlan.delete'),
        {
          type: 'warning',
          confirmButtonText: t('common.confirm'),
          cancelButtonText: t('common.cancel')
        }
      )

      await deleteSkillPlan(selectedPlan.value.id)
      ElMessage.success(t('skillPlan.deleteSuccess'))

      const deletedId = selectedPlan.value.id
      if (plans.value.length === 1 && pagination.current > 1) {
        pagination.current -= 1
      }

      selectedPlanId.value = null
      selectedPlan.value = null
      await loadPlans(plans.value.find((plan) => plan.id !== deletedId)?.id ?? null)
    } catch (error: any) {
      if (error === 'cancel') return
      ElMessage.error(error?.message ?? t('httpMsg.requestFailed'))
    }
  }

  async function handleDialogSuccess(plan: SkillPlanDetail) {
    dialogVisible.value = false
    editingPlan.value = null
    selectedPlanId.value = plan.id
    selectedPlan.value = plan
    mergePlanDisplayState(plan)
    await loadPlans(plan.id)
  }

  watch(currentLang, (lang, previousLang) => {
    if (lang !== previousLang && selectedPlanId.value) {
      loadPlanDetail(selectedPlanId.value)
    }
  })

  watch(dialogVisible, (visible) => {
    if (!visible) {
      editingPlan.value = null
    }
  })

  onMounted(() => {
    loadPlans()
  })
</script>

<style scoped lang="scss">
  .skill-plan-page {
    height: 100%;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .skill-plan-layout {
    flex: 1 1 auto;
    min-height: 0;
    display: flex;
    align-items: flex-start;
    gap: 16px;
    overflow: hidden;
  }

  .skill-plan-list-card,
  .skill-plan-detail-card {
    height: 100%;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .skill-plan-list-card {
    flex: 0 0 320px;
    width: 320px;
  }

  .skill-plan-detail-card {
    flex: 1 1 0;
    min-width: 0;
  }

  .skill-plan-list-card :deep(.el-card__body),
  .skill-plan-detail-card :deep(.el-card__body) {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .skill-plan-list-toolbar {
    display: flex;
    gap: 12px;
    margin-bottom: 12px;
  }

  .skill-plan-list-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    color: var(--el-text-color-secondary);
    font-size: 13px;
    margin-bottom: 12px;
  }

  .skill-plan-list {
    flex: 1;
    overflow: auto;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .skill-plan-item {
    width: 100%;
    cursor: pointer;
    appearance: none;
    border: 1px solid var(--el-border-color-light);
    border-radius: 14px;
    padding: 14px;
    text-align: left;
    background: var(--el-bg-color);
    transition:
      border-color 0.2s ease,
      box-shadow 0.2s ease,
      transform 0.2s ease;
  }

  .skill-plan-item:hover {
    border-color: var(--el-color-primary-light-5);
    transform: translateY(-1px);
  }

  .skill-plan-item.active {
    border-color: var(--el-color-primary);
    box-shadow: 0 10px 24px rgb(64 158 255 / 12%);
    background: linear-gradient(180deg, rgb(236 245 255 / 100%) 0%, var(--el-bg-color) 100%);
  }

  .skill-plan-item__title {
    display: block;
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .skill-plan-item__heading {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .skill-plan-item__icon {
    width: 32px;
    height: 32px;
    border-radius: 8px;
    flex-shrink: 0;
  }

  .skill-plan-pagination {
    display: flex;
    justify-content: center;
    margin-top: 12px;
  }

  .skill-plan-detail {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .skill-plan-detail__header {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    align-items: flex-start;
    margin-bottom: 18px;
  }

  .skill-plan-detail__identity {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .skill-plan-detail__icon {
    width: 56px;
    height: 56px;
    border-radius: 14px;
  }

  .skill-plan-detail__text {
    min-width: 0;
  }

  .skill-plan-detail__eyebrow {
    color: var(--el-color-primary);
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 0.08em;
    text-transform: uppercase;
  }

  .skill-plan-detail__title {
    margin: 4px 0 0;
    font-size: 28px;
    line-height: 1.2;
    color: var(--el-text-color-primary);
  }

  .skill-plan-detail__ship-name {
    color: var(--el-text-color-secondary);
    font-size: 13px;
    line-height: 1.4;
  }

  .skill-plan-detail__actions {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .skill-plan-summary {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
    margin-bottom: 16px;
  }

  .skill-plan-summary__item {
    border-radius: 14px;
    padding: 14px 16px;
    background: var(--el-fill-color-extra-light);
    border: 1px solid var(--el-border-color-lighter);
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .skill-plan-summary__item .label {
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }

  .skill-plan-description {
    padding: 16px;
    border-radius: 16px;
    background: linear-gradient(
      180deg,
      var(--el-fill-color-extra-light) 0%,
      var(--el-bg-color) 100%
    );
    border: 1px solid var(--el-border-color-lighter);
    margin-bottom: 18px;
  }

  .skill-plan-description__label,
  .skill-plan-table-header {
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-secondary);
    margin-bottom: 10px;
  }

  .skill-plan-description p {
    margin: 0;
    line-height: 1.7;
    color: var(--el-text-color-primary);
    white-space: pre-wrap;
  }

  .skill-plan-table {
    width: 100%;
  }

  .skill-cell {
    display: inline-flex;
    align-items: center;
  }

  .skill-cell__name {
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  @media (max-width: 1100px) {
    .skill-plan-layout {
      flex-direction: column;
    }

    .skill-plan-list-card {
      flex: 0 0 auto;
      width: 100%;
    }

    .skill-plan-detail-card {
      flex: 1 1 auto;
      width: 100%;
    }

    .skill-plan-summary {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .skill-plan-list-toolbar,
    .skill-plan-detail__header {
      flex-direction: column;
    }

    .skill-plan-detail__title {
      font-size: 22px;
    }
  }

  @media (max-width: 768px) {
    .skill-plan-detail__identity {
      align-items: flex-start;
    }

    .skill-plan-detail__icon {
      width: 48px;
      height: 48px;
    }
  }
</style>
