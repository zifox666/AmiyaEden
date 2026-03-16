<!-- 抽奖活动管理面板 -->
<template>
  <ElCard class="art-table-card" shadow="never">
    <!-- 标签页：活动管理 / 抽奖记录 -->
    <ElTabs v-model="activeSubTab">
      <ElTabPane :label="t('lottery.manage.activitiesTab')" name="activities">
        <!-- 工具栏 -->
        <div class="flex items-center gap-2 mb-3">
          <ElButton type="success" :icon="Plus" @click="openCreateActivity">{{
            t('lottery.manage.createActivity')
          }}</ElButton>
          <ElButton :loading="activityLoading" @click="loadActivities">
            <el-icon class="mr-1"><Refresh /></el-icon>{{ t('common.refresh') }}
          </ElButton>
        </div>

        <!-- 活动列表 -->
        <div v-loading="activityLoading" class="activity-list">
          <ElEmpty
            v-if="!activityLoading && activities.length === 0"
            :description="t('lottery.manage.noActivities')"
          />
          <ElCard v-for="act in activities" :key="act.id" class="activity-card mb-3" shadow="hover">
            <div class="activity-header">
              <div class="activity-image" v-if="act.image">
                <img :src="act.image" :alt="act.name" />
              </div>
              <div class="activity-info">
                <div class="flex items-center gap-2">
                  <span class="activity-name">{{ act.name }}</span>
                  <ElTag
                    :type="act.status === 1 ? 'success' : 'danger'"
                    size="small"
                    effect="plain"
                  >
                    {{ act.status === 1 ? t('lottery.status.active') : t('lottery.status.closed') }}
                  </ElTag>
                </div>
                <div class="activity-meta text-sm text-gray-400 mt-1">
                  <span
                    >{{ t('lottery.manage.costPerDraw') }}:
                    <strong class="text-orange-500"
                      >{{ act.cost_per_draw }} {{ t('lottery.points') }}</strong
                    ></span
                  >
                  <span v-if="act.start_at" class="ml-3"
                    >{{ t('lottery.manage.startAt') }}: {{ formatTime(act.start_at) }}</span
                  >
                  <span v-if="act.end_at" class="ml-3"
                    >{{ t('lottery.manage.endAt') }}: {{ formatTime(act.end_at) }}</span
                  >
                </div>
                <div v-if="act.description" class="text-xs text-gray-400 mt-1">{{
                  act.description
                }}</div>
              </div>
              <div class="activity-actions flex gap-2 ml-auto flex-shrink-0">
                <ElButton size="small" @click="openPrizeManager(act)">{{
                  t('lottery.manage.managePrizes')
                }}</ElButton>
                <ElButton size="small" type="primary" @click="openEditActivity(act)">{{
                  t('common.edit')
                }}</ElButton>
                <ElButton size="small" type="danger" @click="handleDeleteActivity(act)">{{
                  t('common.delete')
                }}</ElButton>
              </div>
            </div>

            <!-- 奖品预览 -->
            <div
              v-if="act.prizes && act.prizes.length > 0"
              class="prizes-preview mt-3 pt-3 border-t border-gray-700"
            >
              <div class="text-xs text-gray-400 mb-2">{{
                t('lottery.manage.prizeCount', { n: act.prizes.length })
              }}</div>
              <div class="flex flex-wrap gap-2">
                <ElTag
                  v-for="prize in act.prizes"
                  :key="prize.id"
                  :type="TIER_TAG_TYPE[prize.tier] ?? 'info'"
                  size="small"
                >
                  <img v-if="prize.image" :src="prize.image" class="prize-tag-img" />
                  {{ prize.name }}
                  <span v-if="prize.total_stock > 0"
                    >({{ prize.drawn_count }}/{{ prize.total_stock }})</span
                  >
                </ElTag>
              </div>
            </div>
            <div v-else class="mt-2 text-xs text-gray-400">{{ t('lottery.manage.noPrizes') }}</div>
          </ElCard>
        </div>
      </ElTabPane>

      <ElTabPane :label="t('lottery.manage.recordsTab')" name="records">
        <div class="flex items-center gap-2 mb-3">
          <ElSelect
            v-model="recordActivityFilter"
            :placeholder="t('lottery.manage.filterActivity')"
            clearable
            style="width: 200px"
            @change="loadRecords"
          >
            <ElOption
              v-for="act in allActivities"
              :key="act.id"
              :label="act.name"
              :value="act.id"
            />
          </ElSelect>
          <ElButton :loading="recordLoading" @click="loadRecords">
            <el-icon class="mr-1"><Refresh /></el-icon>{{ t('common.refresh') }}
          </ElButton>
        </div>
        <ArtTable
          :loading="recordLoading"
          :data="records"
          :columns="recordColumns"
          :pagination="recordPagination"
          @pagination:size-change="handleRecordSizeChange"
          @pagination:current-change="handleRecordCurrentChange"
        />
      </ElTabPane>
    </ElTabs>
  </ElCard>

  <!-- 创建/编辑活动对话框 -->
  <ElDialog
    v-model="activityDialogVisible"
    :title="editingActivity ? t('lottery.manage.editActivity') : t('lottery.manage.createActivity')"
    width="560px"
    destroy-on-close
  >
    <ElForm ref="activityFormRef" :model="activityForm" :rules="activityRules" label-width="100px">
      <ElFormItem :label="t('lottery.manage.fields.name')" prop="name">
        <ElInput
          v-model="activityForm.name"
          :placeholder="t('lottery.manage.fields.namePlaceholder')"
        />
      </ElFormItem>
      <ElFormItem :label="t('lottery.manage.fields.description')">
        <ElInput
          v-model="activityForm.description"
          type="textarea"
          :rows="2"
          :placeholder="t('lottery.manage.fields.descriptionPlaceholder')"
        />
      </ElFormItem>
      <ElFormItem :label="t('lottery.manage.fields.coverImage')">
        <div class="image-upload-area">
          <div v-if="activityForm.image" class="image-preview">
            <img :src="activityForm.image" alt="封面" />
            <div class="image-actions">
              <ElButton size="small" type="danger" text @click="activityForm.image = ''">
                <el-icon><Delete /></el-icon>
              </ElButton>
            </div>
          </div>
          <ElUpload
            v-else
            class="image-uploader"
            :show-file-list="false"
            accept="image/jpeg,image/png,image/gif,image/webp"
            :before-upload="handleImageBeforeUpload"
            :http-request="(opts) => handleImageUpload(opts, 'activity')"
          >
            <div class="upload-placeholder">
              <el-icon v-if="!imageUploading" :size="28"><Plus /></el-icon>
              <el-icon v-else :size="28" class="animate-spin"><Loading /></el-icon>
              <span class="text-xs">{{
                imageUploading ? t('lottery.manage.uploading') : t('lottery.manage.uploadCover')
              }}</span>
            </div>
          </ElUpload>
        </div>
      </ElFormItem>
      <ElFormItem :label="t('lottery.manage.fields.costPerDraw')" prop="cost_per_draw">
        <ElInputNumber
          v-model="activityForm.cost_per_draw"
          :min="0"
          :precision="2"
          style="width: 200px"
        />
        <span class="ml-2 text-xs text-gray-400">{{ t('lottery.manage.fields.costHint') }}</span>
      </ElFormItem>
      <ElFormItem :label="t('lottery.manage.fields.status')">
        <ElSelect v-model="activityForm.status" style="width: 140px">
          <ElOption :label="t('lottery.status.active')" :value="1" />
          <ElOption :label="t('lottery.status.closed')" :value="0" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem :label="t('lottery.manage.fields.startAt')">
        <ElDatePicker
          v-model="activityForm.start_at"
          type="datetime"
          :placeholder="t('lottery.manage.fields.noLimit')"
          clearable
          style="width: 240px"
        />
      </ElFormItem>
      <ElFormItem :label="t('lottery.manage.fields.endAt')">
        <ElDatePicker
          v-model="activityForm.end_at"
          type="datetime"
          :placeholder="t('lottery.manage.fields.noLimit')"
          clearable
          style="width: 240px"
        />
      </ElFormItem>
      <ElFormItem :label="t('lottery.manage.fields.sortOrder')">
        <ElInputNumber v-model="activityForm.sort_order" :min="0" style="width: 140px" />
        <span class="ml-2 text-xs text-gray-400">{{ t('lottery.manage.fields.sortHint') }}</span>
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="activityDialogVisible = false">{{ t('common.cancel') }}</ElButton>
      <ElButton type="primary" :loading="activitySubmitting" @click="handleActivitySubmit">{{
        t('common.confirm')
      }}</ElButton>
    </template>
  </ElDialog>

  <!-- 奖品管理抽屉 -->
  <ElDrawer
    v-model="prizeDrawerVisible"
    :title="t('lottery.manage.managePrizesTitle', { name: prizeActivity?.name ?? '' })"
    size="560px"
    destroy-on-close
  >
    <div class="flex items-center gap-2 mb-4">
      <ElButton type="success" size="small" :icon="Plus" @click="openCreatePrize">{{
        t('lottery.manage.addPrize')
      }}</ElButton>
    </div>

    <div v-if="currentPrizes.length === 0" class="text-center text-gray-400 py-8">{{
      t('lottery.manage.noPrizesHint')
    }}</div>

    <!-- 奖品列表 -->
    <div class="prize-list">
      <div v-for="prize in currentPrizes" :key="prize.id" class="prize-item">
        <div class="prize-tier-bar" :class="`tier-${prize.tier}`"></div>
        <img v-if="prize.image" :src="prize.image" class="prize-list-img" />
        <div class="prize-content">
          <div class="flex items-center gap-2">
            <ElTag :type="TIER_TAG_TYPE[prize.tier] ?? 'info'" size="small">
              {{ TIER_LABELS[prize.tier] ?? prize.tier }}
            </ElTag>
            <span class="font-medium">{{ prize.name }}</span>
          </div>
          <div class="text-xs text-gray-400 mt-1">
            {{ t('lottery.manage.weight') }}: {{ prize.probability_weight }}
            <template v-if="prize.total_stock > 0">
              &nbsp;|&nbsp; {{ t('lottery.manage.stock') }}: {{ prize.drawn_count }}/{{
                prize.total_stock
              }}
            </template>
            <template v-else> &nbsp;|&nbsp; {{ t('lottery.manage.stockUnlimited') }} </template>
          </div>
        </div>
        <div class="prize-actions flex gap-1 ml-auto">
          <ElButton size="small" @click="openEditPrize(prize)">{{ t('common.edit') }}</ElButton>
          <ElButton size="small" type="danger" @click="handleDeletePrize(prize)">{{
            t('common.delete')
          }}</ElButton>
        </div>
      </div>
    </div>

    <!-- 总权重提示 -->
    <div v-if="currentPrizes.length > 0" class="mt-4 p-3 rounded bg-gray-800 text-xs text-gray-400">
      {{ t('lottery.manage.totalWeight') }}: {{ totalWeight }} &nbsp;|&nbsp;
      <span v-for="prize in currentPrizes" :key="prize.id">
        {{ prize.name }}: {{ ((prize.probability_weight / totalWeight) * 100).toFixed(1) }}%
      </span>
    </div>
  </ElDrawer>

  <!-- 创建/编辑奖品对话框 -->
  <ElDialog
    v-model="prizeDialogVisible"
    :title="editingPrize ? t('lottery.manage.editPrize') : t('lottery.manage.addPrize')"
    width="440px"
    destroy-on-close
  >
    <ElForm ref="prizeFormRef" :model="prizeForm" :rules="prizeRules" label-width="100px">
      <ElFormItem :label="t('lottery.manage.prizeFields.name')" prop="name">
        <ElInput
          v-model="prizeForm.name"
          :placeholder="t('lottery.manage.prizeFields.namePlaceholder')"
        />
      </ElFormItem>
      <ElFormItem :label="t('lottery.manage.prizeFields.image')">
        <div class="image-upload-area small">
          <div v-if="prizeForm.image" class="image-preview">
            <img :src="prizeForm.image" :alt="t('lottery.manage.prizeFields.image')" />
            <div class="image-actions">
              <ElButton size="small" type="danger" text @click="prizeForm.image = ''">
                <el-icon><Delete /></el-icon>
              </ElButton>
            </div>
          </div>
          <ElUpload
            v-else
            class="image-uploader"
            :show-file-list="false"
            accept="image/jpeg,image/png,image/gif,image/webp"
            :before-upload="handleImageBeforeUpload"
            :http-request="(opts) => handlePrizeImageUpload(opts)"
          >
            <div class="upload-placeholder">
              <el-icon v-if="!prizeImageUploading" :size="22"><Plus /></el-icon>
              <el-icon v-else :size="22" class="animate-spin"><Loading /></el-icon>
              <span class="text-xs">{{
                prizeImageUploading
                  ? t('lottery.manage.uploading')
                  : t('lottery.manage.uploadImage')
              }}</span>
            </div>
          </ElUpload>
        </div>
        <SdeSearchSelect
          v-model="prizeSdeTypeId"
          :placeholder="t('lottery.manage.prizeFields.sdeSearchPlaceholder')"
          style="width: 200px; margin-left: 12px"
          @select="onPrizeSdeSelect"
        />
      </ElFormItem>
      <ElFormItem :label="t('lottery.manage.prizeFields.tier')" prop="tier">
        <ElSelect v-model="prizeForm.tier" style="width: 160px">
          <ElOption :label="t('lottery.tier.normal')" value="normal" />
          <ElOption :label="t('lottery.tier.rare')" value="rare" />
          <ElOption :label="t('lottery.tier.legendary')" value="legendary" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem :label="t('lottery.manage.prizeFields.weight')">
        <ElInputNumber v-model="prizeForm.probability_weight" :min="1" style="width: 140px" />
        <span class="ml-2 text-xs text-gray-400">{{
          t('lottery.manage.prizeFields.weightHint')
        }}</span>
      </ElFormItem>
      <ElFormItem :label="t('lottery.manage.prizeFields.totalStock')">
        <ElInputNumber v-model="prizeForm.total_stock" :min="0" style="width: 140px" />
        <span class="ml-2 text-xs text-gray-400">{{
          t('lottery.manage.prizeFields.stockHint')
        }}</span>
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="prizeDialogVisible = false">{{ t('common.cancel') }}</ElButton>
      <ElButton type="primary" :loading="prizeSubmitting" @click="handlePrizeSubmit">{{
        t('common.confirm')
      }}</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import {
    ElTag,
    ElButton,
    ElSelect,
    ElOption,
    ElInput,
    ElMessage,
    ElMessageBox,
    ElUpload,
    ElTabs,
    ElTabPane,
    ElDrawer,
    ElDatePicker,
    ElEmpty
  } from 'element-plus'
  import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
  import { Plus, Refresh, Delete, Loading } from '@element-plus/icons-vue'
  import SdeSearchSelect from '@/components/business/SdeSearchSelect.vue'
  import {
    adminListLotteryActivities,
    adminCreateLotteryActivity,
    adminUpdateLotteryActivity,
    adminDeleteLotteryActivity,
    adminCreateLotteryPrize,
    adminUpdateLotteryPrize,
    adminDeleteLotteryPrize,
    adminListLotteryRecords,
    adminUpdateLotteryRecordDelivery,
    uploadShopImage
  } from '@/api/shop'

  defineOptions({ name: 'ManageLottery' })

  const { t } = useI18n()

  type LotteryActivity = Api.Shop.LotteryActivity
  type LotteryPrize = Api.Shop.LotteryPrize
  type LotteryRecord = Api.Shop.LotteryRecord

  const TIER_TAG_TYPE: Record<string, any> = {
    normal: 'info',
    rare: 'warning',
    legendary: 'danger'
  }
  const TIER_LABELS = computed(() => ({
    normal: t('lottery.tier.normal'),
    rare: t('lottery.tier.rare'),
    legendary: t('lottery.tier.legendary')
  }))

  // ─── 活动列表 ───
  const activeSubTab = ref('activities')
  const activityLoading = ref(false)
  const activities = ref<LotteryActivity[]>([])
  const allActivities = ref<LotteryActivity[]>([]) // 用于 records 筛选

  async function loadActivities() {
    activityLoading.value = true
    try {
      const res = (await adminListLotteryActivities({ current: 1, size: 100 })) as any
      activities.value = res?.list ?? []
      allActivities.value = activities.value
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('lottery.manage.loadFailed'))
    } finally {
      activityLoading.value = false
    }
  }

  function formatTime(t: string | null) {
    if (!t) return '-'
    return new Date(t).toLocaleString('zh-CN', { hour12: false })
  }

  // ─── 创建/编辑活动 ───
  const activityDialogVisible = ref(false)
  const activitySubmitting = ref(false)
  const activityFormRef = ref<FormInstance>()
  const editingActivity = ref<LotteryActivity | null>(null)
  const imageUploading = ref(false)

  const activityForm = reactive({
    name: '',
    description: '',
    image: '',
    cost_per_draw: 0,
    status: 1 as number,
    start_at: null as Date | null,
    end_at: null as Date | null,
    sort_order: 0
  })

  const activityRules = computed<FormRules>(() => ({
    name: [{ required: true, message: t('lottery.manage.validName'), trigger: 'blur' }],
    cost_per_draw: [{ required: true, message: t('lottery.manage.validCost'), trigger: 'blur' }]
  }))

  function resetActivityForm() {
    Object.assign(activityForm, {
      name: '',
      description: '',
      image: '',
      cost_per_draw: 0,
      status: 1,
      start_at: null,
      end_at: null,
      sort_order: 0
    })
    editingActivity.value = null
  }

  function openCreateActivity() {
    resetActivityForm()
    activityDialogVisible.value = true
  }

  function openEditActivity(act: LotteryActivity) {
    editingActivity.value = act
    Object.assign(activityForm, {
      name: act.name,
      description: act.description,
      image: act.image,
      cost_per_draw: act.cost_per_draw,
      status: act.status,
      start_at: act.start_at ? new Date(act.start_at) : null,
      end_at: act.end_at ? new Date(act.end_at) : null,
      sort_order: act.sort_order
    })
    activityDialogVisible.value = true
  }

  async function handleActivitySubmit() {
    if (!activityFormRef.value) return
    await activityFormRef.value.validate()
    activitySubmitting.value = true
    try {
      const payload = {
        ...activityForm,
        start_at: activityForm.start_at ? activityForm.start_at.toISOString() : null,
        end_at: activityForm.end_at ? activityForm.end_at.toISOString() : null
      }
      if (editingActivity.value) {
        await adminUpdateLotteryActivity({ id: editingActivity.value.id, ...payload })
        ElMessage.success(t('lottery.manage.updateSuccess'))
      } else {
        await adminCreateLotteryActivity(payload)
        ElMessage.success(t('lottery.manage.createSuccess'))
      }
      activityDialogVisible.value = false
      loadActivities()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('lottery.manage.operationFailed'))
    } finally {
      activitySubmitting.value = false
    }
  }

  async function handleDeleteActivity(act: LotteryActivity) {
    await ElMessageBox.confirm(
      t('lottery.manage.deleteActivityConfirm', { name: act.name }),
      t('common.confirm'),
      {
        confirmButtonText: t('common.delete'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    try {
      await adminDeleteLotteryActivity(act.id)
      ElMessage.success(t('lottery.manage.deleteSuccess'))
      loadActivities()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('lottery.manage.deleteFailed'))
    }
  }

  // ─── 图片上传 ───
  function handleImageBeforeUpload(file: File) {
    if (file.size > 5 * 1024 * 1024) {
      ElMessage.error(t('lottery.manage.imageTooLarge'))
      return false
    }
    return true
  }

  async function handleImageUpload(options: UploadRequestOptions, target: 'activity') {
    imageUploading.value = true
    try {
      const res = (await uploadShopImage(options.file as File)) as any
      if (target === 'activity') activityForm.image = res?.url ?? ''
      ElMessage.success(t('lottery.manage.uploadSuccess'))
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('lottery.manage.uploadFailed'))
    } finally {
      imageUploading.value = false
    }
  }

  // ─── 奖品图片上传 ───
  const prizeImageUploading = ref(false)

  async function handlePrizeImageUpload(options: UploadRequestOptions) {
    prizeImageUploading.value = true
    try {
      const res = (await uploadShopImage(options.file as File)) as any
      prizeForm.image = res?.url ?? ''
      ElMessage.success(t('lottery.manage.uploadSuccess'))
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('lottery.manage.uploadFailed'))
    } finally {
      prizeImageUploading.value = false
    }
  }

  // ─── SDE 搜索选择奖品图片 ───
  const prizeSdeTypeId = ref<number | null>(null)

  function onPrizeSdeSelect(item: Api.Sde.FuzzySearchItem | null) {
    if (item) {
      prizeForm.image = `https://images.evetech.net/types/${item.id}/icon?size=64`
      if (!prizeForm.name) {
        prizeForm.name = item.name
      }
    }
  }

  // ─── 奖品管理 ───
  const prizeDrawerVisible = ref(false)
  const prizeActivity = ref<LotteryActivity | null>(null)
  const currentPrizes = computed(() => prizeActivity.value?.prizes ?? [])
  const totalWeight = computed(() =>
    currentPrizes.value.reduce((sum, p) => sum + Math.max(p.probability_weight, 1), 0)
  )

  function openPrizeManager(act: LotteryActivity) {
    prizeActivity.value = act
    prizeDrawerVisible.value = true
  }

  const prizeDialogVisible = ref(false)
  const prizeSubmitting = ref(false)
  const prizeFormRef = ref<FormInstance>()
  const editingPrize = ref<LotteryPrize | null>(null)

  const prizeForm = reactive({
    name: '',
    image: '',
    tier: 'normal' as Api.Shop.LotteryPrizeTier,
    probability_weight: 10,
    total_stock: 0
  })

  const prizeRules = computed<FormRules>(() => ({
    name: [{ required: true, message: t('lottery.manage.validPrizeName'), trigger: 'blur' }],
    tier: [{ required: true, message: t('lottery.manage.validTier'), trigger: 'change' }]
  }))

  function openCreatePrize() {
    editingPrize.value = null
    Object.assign(prizeForm, {
      name: '',
      image: '',
      tier: 'normal',
      probability_weight: 10,
      total_stock: 0
    })
    prizeSdeTypeId.value = null
    prizeDialogVisible.value = true
  }

  function openEditPrize(prize: LotteryPrize) {
    editingPrize.value = prize
    Object.assign(prizeForm, {
      name: prize.name,
      image: prize.image,
      tier: prize.tier,
      probability_weight: prize.probability_weight,
      total_stock: prize.total_stock
    })
    prizeSdeTypeId.value = null
    prizeDialogVisible.value = true
  }

  async function handlePrizeSubmit() {
    if (!prizeFormRef.value) return
    await prizeFormRef.value.validate()
    prizeSubmitting.value = true
    try {
      if (editingPrize.value) {
        await adminUpdateLotteryPrize({ id: editingPrize.value.id, ...prizeForm })
        ElMessage.success(t('lottery.manage.updateSuccess'))
      } else {
        await adminCreateLotteryPrize({ activity_id: prizeActivity.value!.id, ...prizeForm })
        ElMessage.success(t('lottery.manage.addSuccess'))
      }
      prizeDialogVisible.value = false
      // 重新加载活动列表以刷新奖品
      await loadActivities()
      // 重定向当前 prizeActivity 到刷新后的数据
      if (prizeActivity.value) {
        const updated = activities.value.find((a) => a.id === prizeActivity.value!.id)
        if (updated) prizeActivity.value = updated
      }
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('lottery.manage.operationFailed'))
    } finally {
      prizeSubmitting.value = false
    }
  }

  async function handleDeletePrize(prize: LotteryPrize) {
    await ElMessageBox.confirm(
      t('lottery.manage.deletePrizeConfirm', { name: prize.name }),
      t('common.confirm'),
      {
        confirmButtonText: t('common.delete'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    try {
      await adminDeleteLotteryPrize(prize.id)
      ElMessage.success(t('lottery.manage.deleteSuccess'))
      await loadActivities()
      if (prizeActivity.value) {
        const updated = activities.value.find((a) => a.id === prizeActivity.value!.id)
        if (updated) prizeActivity.value = updated
      }
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('lottery.manage.deleteFailed'))
    }
  }

  // ─── 抽奖记录 ───
  const recordLoading = ref(false)
  const records = ref<LotteryRecord[]>([])
  const recordActivityFilter = ref<number | undefined>(undefined)
  const recordPagination = reactive({ current: 1, size: 20, total: 0 })

  const recordColumns = computed(
    () =>
      [
        { type: 'index', width: 60, label: '#' },
        { prop: 'user_id', label: t('lottery.manage.recordColumns.userId'), width: 90 },
        {
          prop: 'activity_name',
          label: t('lottery.manage.recordColumns.activity'),
          minWidth: 120,
          showOverflowTooltip: true
        },
        {
          prop: 'prize_name',
          label: t('lottery.manage.recordColumns.prize'),
          minWidth: 160,
          formatter: (row: LotteryRecord) =>
            h('div', { class: 'flex items-center gap-2' }, [
              row.prize_image
                ? h('img', { src: row.prize_image, class: 'prize-record-img' })
                : null,
              h(
                ElTag,
                { type: TIER_TAG_TYPE[row.prize_tier], size: 'small', effect: 'plain' },
                () => TIER_LABELS.value[row.prize_tier] ?? row.prize_tier
              ),
              h('span', {}, row.prize_name)
            ])
        },
        {
          prop: 'cost',
          label: t('lottery.manage.recordColumns.cost'),
          width: 90,
          formatter: (row: LotteryRecord) =>
            h('span', { class: 'text-orange-500' }, String(row.cost))
        },
        {
          prop: 'delivery_status',
          label: t('lottery.manage.recordColumns.deliveryStatus'),
          width: 110,
          formatter: (row: LotteryRecord) => {
            const isDelivered = row.delivery_status === 'delivered'
            return h(
              ElTag,
              { type: isDelivered ? 'success' : 'warning', size: 'small', effect: 'plain' },
              () => (isDelivered ? t('lottery.delivery.delivered') : t('lottery.delivery.pending'))
            )
          }
        },
        {
          prop: 'created_at',
          label: t('lottery.manage.recordColumns.time'),
          width: 170,
          formatter: (row: LotteryRecord) =>
            new Date(row.created_at).toLocaleString('zh-CN', { hour12: false })
        },
        {
          prop: 'operation',
          label: t('common.operation'),
          width: 120,
          formatter: (row: LotteryRecord) => {
            if (row.delivery_status === 'delivered') return null
            return h(
              ElButton,
              {
                size: 'small',
                type: 'success',
                onClick: () => handleMarkDelivered(row)
              },
              () => t('lottery.delivery.markDelivered')
            )
          }
        }
      ] as any[]
  )

  async function loadRecords() {
    recordLoading.value = true
    try {
      const res = (await adminListLotteryRecords({
        current: recordPagination.current,
        size: recordPagination.size,
        activity_id: recordActivityFilter.value
      })) as any
      records.value = res?.list ?? []
      recordPagination.total = res?.total ?? 0
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('lottery.manage.loadFailed'))
    } finally {
      recordLoading.value = false
    }
  }

  async function handleMarkDelivered(row: LotteryRecord) {
    try {
      await adminUpdateLotteryRecordDelivery(row.id, 'delivered')
      row.delivery_status = 'delivered'
      ElMessage.success(t('lottery.delivery.markSuccess'))
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('lottery.manage.operationFailed'))
    }
  }

  function handleRecordSizeChange(size: number) {
    recordPagination.size = size
    recordPagination.current = 1
    loadRecords()
  }

  function handleRecordCurrentChange(page: number) {
    recordPagination.current = page
    loadRecords()
  }

  // ─── 初始化 ───
  onMounted(() => {
    loadActivities()
  })

  watch(activeSubTab, (tab) => {
    if (tab === 'records' && records.value.length === 0) {
      loadRecords()
    }
  })

  defineExpose({ load: loadActivities, refresh: loadActivities })
</script>

<style scoped>
  .activity-card {
    transition: border-color 0.2s;
  }

  .activity-header {
    display: flex;
    align-items: flex-start;
    gap: 12px;
  }

  .activity-image {
    width: 64px;
    height: 64px;
    flex-shrink: 0;
    border-radius: 6px;
    overflow: hidden;
    background: var(--el-fill-color);
  }

  .activity-image img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .activity-info {
    flex: 1;
  }

  .activity-name {
    font-size: 16px;
    font-weight: 600;
  }

  .prize-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .prize-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    border: 1px solid var(--el-border-color);
    border-radius: 6px;
    background: var(--el-fill-color-lighter);
  }

  .prize-tier-bar {
    width: 4px;
    height: 36px;
    border-radius: 2px;
    flex-shrink: 0;
  }

  .tier-normal {
    background: var(--el-color-info);
  }
  .tier-rare {
    background: var(--el-color-warning);
  }
  .tier-legendary {
    background: var(--el-color-danger);
  }

  /* 奖品图片 */
  .prize-tag-img {
    width: 16px;
    height: 16px;
    border-radius: 2px;
    vertical-align: middle;
    margin-right: 4px;
  }

  .prize-list-img {
    width: 40px;
    height: 40px;
    border-radius: 4px;
    object-fit: cover;
    flex-shrink: 0;
  }

  .prize-record-img {
    width: 24px;
    height: 24px;
    border-radius: 3px;
    object-fit: cover;
  }

  /* 图片上传 */
  .image-upload-area {
    width: 100px;
    height: 100px;
    border: 1px dashed var(--el-border-color);
    border-radius: 6px;
    overflow: hidden;
    background: var(--el-fill-color-lighter);
  }

  .image-preview {
    width: 100%;
    height: 100%;
    position: relative;
  }

  .image-preview img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .image-actions {
    position: absolute;
    top: 2px;
    right: 2px;
    background: rgba(0, 0, 0, 0.5);
    border-radius: 4px;
  }

  .image-uploader {
    width: 100%;
    height: 100%;
  }

  .image-uploader :deep(.el-upload) {
    width: 100%;
    height: 100%;
    display: block;
  }

  .upload-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 4px;
    cursor: pointer;
    color: var(--el-text-color-placeholder);
    font-size: 12px;
  }

  .upload-placeholder:hover {
    color: var(--el-color-primary);
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }
    to {
      transform: rotate(360deg);
    }
  }

  .animate-spin {
    animation: spin 1s linear infinite;
  }
</style>
