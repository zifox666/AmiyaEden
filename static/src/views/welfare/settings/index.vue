<!-- 福利设置页面 -->
<template>
  <div class="welfare-settings-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <div class="flex items-center gap-2">
            <ElButton v-if="canManage" type="success" :icon="Plus" @click="openCreateDialog">{{
              t('welfareSettings.create')
            }}</ElButton>
            <ElTag v-if="reorderSaving" type="info" size="small">{{
              t('welfareSettings.reorderSaving')
            }}</ElTag>
            <ElInput
              v-model="nameFilter"
              :placeholder="t('welfareSettings.filterName')"
              clearable
              style="width: 160px"
              @keyup.enter="handleSearch"
            />
            <ElSelect
              v-model="statusFilter"
              :placeholder="t('welfareSettings.filterStatus')"
              clearable
              style="width: 120px"
              @change="handleSearch"
            >
              <ElOption :label="t('welfareSettings.statusActive')" :value="1" />
              <ElOption :label="t('welfareSettings.statusDisabled')" :value="0" />
            </ElSelect>
            <ElButton type="primary" @click="handleSearch">{{ t('common.search') }}</ElButton>
            <ElButton @click="handleReset">{{ t('common.reset') }}</ElButton>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        ref="tableRef"
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        row-key="id"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <!-- 导入历史记录对话框 -->
    <ElDialog
      v-model="importDialogVisible"
      :title="t('welfareSettings.importTitle')"
      width="560px"
      destroy-on-close
    >
      <p class="mb-2 text-sm text-gray-500">{{ t('welfareSettings.importHint') }}</p>
      <ElInput
        v-model="importCSV"
        type="textarea"
        :rows="10"
        :placeholder="t('welfareSettings.importPlaceholder')"
      />
      <template #footer>
        <ElButton @click="importDialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="importLoading" @click="handleImport">{{
          t('common.confirm')
        }}</ElButton>
      </template>
    </ElDialog>

    <!-- 福利编辑对话框 -->
    <ElDialog
      v-model="dialogVisible"
      :title="editingItem ? t('welfareSettings.edit') : t('welfareSettings.create')"
      width="560px"
      destroy-on-close
    >
      <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="120px">
        <ElFormItem :label="t('welfareSettings.name')" prop="name">
          <ElInput v-model="formData.name" :placeholder="t('welfareSettings.namePlaceholder')" />
        </ElFormItem>
        <ElFormItem :label="t('welfareSettings.description')">
          <ElInput
            v-model="formData.description"
            type="textarea"
            :rows="3"
            :placeholder="t('welfareSettings.descriptionPlaceholder')"
          />
        </ElFormItem>
        <ElFormItem :label="t('welfareSettings.distMode')" prop="dist_mode">
          <ElRadioGroup v-model="formData.dist_mode">
            <ElRadio value="per_user">{{ t('welfareSettings.distModePerUser') }}</ElRadio>
            <ElRadio value="per_character">{{ t('welfareSettings.distModePerCharacter') }}</ElRadio>
          </ElRadioGroup>
          <div v-if="hasMaxCharAge" class="text-xs text-gray-400 mt-1">
            {{ t('welfareSettings.distModeLockedHint') }}
          </div>
        </ElFormItem>
        <ElFormItem :label="t('welfareSettings.maxCharAgeMonths')">
          <ElInputNumber
            v-model="formData.max_char_age_months"
            :min="0"
            :step="1"
            :precision="0"
            :placeholder="t('welfareSettings.maxCharAgePlaceholder')"
            controls-position="right"
            style="width: 200px"
          />
          <div class="text-xs text-gray-400 mt-1">
            {{ t('welfareSettings.maxCharAgeHint') }}
          </div>
        </ElFormItem>
        <ElFormItem :label="t('welfareSettings.minimumPap')">
          <ElInputNumber
            v-model="formData.minimum_pap"
            :min="0"
            :step="1"
            :precision="0"
            :placeholder="t('welfareSettings.minimumPapPlaceholder')"
            controls-position="right"
            style="width: 200px"
          />
          <div class="text-xs text-gray-400 mt-1">
            {{ t('welfareSettings.minimumPapHint') }}
          </div>
        </ElFormItem>
        <ElFormItem :label="t('welfareSettings.requireSkillPlan')">
          <ElSwitch v-model="formData.require_skill_plan" />
        </ElFormItem>
        <ElFormItem :label="t('welfareSettings.requireEvidence')">
          <ElSwitch v-model="formData.require_evidence" />
        </ElFormItem>
        <ElFormItem v-if="formData.require_evidence" :label="t('welfareSettings.exampleEvidence')">
          <div class="flex flex-col gap-2" style="width: 100%">
            <ElUpload
              :show-file-list="false"
              accept="image/*"
              :before-upload="handleExampleEvidenceUpload"
            >
              <ElButton size="small" :loading="exampleEvidenceUploading">
                {{ t('welfareSettings.uploadExampleEvidence') }}
              </ElButton>
            </ElUpload>
            <img
              v-if="formData.example_evidence"
              :src="formData.example_evidence"
              class="mt-1 rounded border"
              style="max-height: 120px; max-width: 100%; object-fit: contain"
            />
          </div>
        </ElFormItem>
        <ElFormItem
          v-if="formData.require_skill_plan"
          :label="t('welfareSettings.skillPlan')"
          prop="skill_plan_ids"
        >
          <ElSelect
            v-model="formData.skill_plan_ids"
            :placeholder="t('welfareSettings.skillPlanPlaceholder')"
            :loading="skillPlansLoading"
            filterable
            multiple
            collapse-tags
            collapse-tags-tooltip
            style="width: 100%"
          >
            <ElOption
              v-for="plan in skillPlans"
              :key="plan.id"
              :label="plan.title"
              :value="plan.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('welfareSettings.status')">
          <ElSelect v-model="formData.status" style="width: 200px">
            <ElOption :label="t('welfareSettings.statusActive')" :value="1" />
            <ElOption :label="t('welfareSettings.statusDisabled')" :value="0" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="t('welfareSettings.sortOrder')">
          <ElInputNumber
            v-model="formData.sort_order"
            :min="0"
            :step="1"
            :precision="0"
            controls-position="right"
            style="width: 200px"
          />
          <div class="text-xs text-gray-400 mt-1">{{ t('welfareSettings.sortOrderHint') }}</div>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">{{ t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">{{
          t('common.confirm')
        }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import {
    ElTag,
    ElButton,
    ElInput,
    ElInputNumber,
    ElSelect,
    ElOption,
    ElSwitch,
    ElRadioGroup,
    ElRadio,
    ElUpload,
    ElMessage,
    ElMessageBox
  } from 'element-plus'
  import { useUserStore } from '@/store/modules/user'
  import type { FormInstance, FormRules, UploadRawFile } from 'element-plus'
  import { Plus } from '@element-plus/icons-vue'
  import { useDraggable } from 'vue-draggable-plus'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import {
    adminListWelfares,
    adminCreateWelfare,
    adminUpdateWelfare,
    adminDeleteWelfare,
    adminImportWelfareRecords,
    adminReorderWelfares,
    uploadWelfareEvidence
  } from '@/api/welfare'
  import { fetchSkillPlanList } from '@/api/skill-plan'
  import { useTable } from '@/hooks/core/useTable'
  import { useI18n } from 'vue-i18n'
  import { formatTime } from '@utils/common'

  defineOptions({ name: 'WelfareSettings' })
  const { t } = useI18n()

  const userStore = useUserStore()
  const canManage = computed(() => {
    const roles = userStore.getUserInfo?.roles ?? []
    return roles.some((role) => ['super_admin', 'admin'].includes(role))
  })

  type WelfareItem = Api.Welfare.WelfareItem

  // ─── 状态/模式映射 ───
  const STATUS_CONFIG = computed(
    () =>
      ({
        1: { label: t('welfareSettings.statusActive'), type: 'success' },
        0: { label: t('welfareSettings.statusDisabled'), type: 'danger' }
      }) as Record<number, { label: string; type: string }>
  )

  const DIST_MODE_CONFIG = computed(
    () =>
      ({
        per_user: { label: t('welfareSettings.distModePerUser'), type: 'primary' },
        per_character: { label: t('welfareSettings.distModePerCharacter'), type: 'warning' }
      }) as Record<string, { label: string; type: string }>
  )

  // ─── 搜索过滤状态 ───
  const nameFilter = ref('')
  const statusFilter = ref<number | undefined>(undefined)

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    searchParams,
    getData,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: adminListWelfares,
      apiParams: { current: 1, size: 50 },
      columnsFactory: () => [
        ...(canManage.value
          ? [
              {
                prop: 'drag',
                label: '',
                width: 40,
                formatter: () =>
                  h(
                    'span',
                    {
                      class: 'drag-handle cursor-grab text-gray-400 hover:text-gray-600 select-none',
                      title: t('welfareSettings.dragHint')
                    },
                    '⠿'
                  )
              }
            ]
          : []),
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'name',
          label: t('welfareSettings.name'),
          minWidth: 160,
          showOverflowTooltip: true
        },
        {
          prop: 'description',
          label: t('welfareSettings.description'),
          minWidth: 200,
          showOverflowTooltip: true,
          formatter: (row: WelfareItem) => row.description || '-'
        },
        {
          prop: 'dist_mode',
          label: t('welfareSettings.distMode'),
          width: 120,
          formatter: (row: WelfareItem) => {
            const cfg = DIST_MODE_CONFIG.value[row.dist_mode] ?? {
              label: row.dist_mode,
              type: 'info'
            }
            return h(
              ElTag,
              { type: cfg.type as any, size: 'small', effect: 'plain' },
              () => cfg.label
            )
          }
        },
        {
          prop: 'require_skill_plan',
          label: t('welfareSettings.requireSkillPlan'),
          width: 120,
          formatter: (row: WelfareItem) =>
            h(
              ElTag,
              {
                type: row.require_skill_plan ? 'warning' : 'info',
                size: 'small',
                effect: 'plain'
              },
              () => (row.require_skill_plan ? t('welfareSettings.yes') : t('welfareSettings.no'))
            )
        },
        {
          prop: 'status',
          label: t('welfareSettings.status'),
          width: 90,
          formatter: (row: WelfareItem) => {
            const cfg = STATUS_CONFIG.value[row.status] ?? {
              label: String(row.status),
              type: 'info'
            }
            return h(
              ElTag,
              { type: cfg.type as any, size: 'small', effect: 'plain' },
              () => cfg.label
            )
          }
        },
        {
          prop: 'created_at',
          label: t('welfareSettings.createdAt'),
          width: 180,
          formatter: (row: WelfareItem) => formatTime(row.created_at)
        },
        ...(canManage.value
          ? [
              {
                prop: 'actions',
                label: t('common.operation'),
                width: 180,
                fixed: 'right',
                formatter: (row: WelfareItem) =>
                  h('div', { class: 'flex gap-1' }, [
                    h(ArtButtonTable, { type: 'edit', onClick: () => openEditDialog(row) }),
                    h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) }),
                    h(ArtButtonTable, {
                      icon: 'ri:upload-2-line',
                      elType: 'warning',
                      label: t('welfareSettings.importBtn'),
                      onClick: () => openImportDialog(row)
                    })
                  ])
              }
            ]
          : [])
      ]
    }
  })

  function handleSearch() {
    Object.assign(searchParams, {
      name: nameFilter.value || undefined,
      status: statusFilter.value,
      current: 1
    })
    getData()
  }

  function handleReset() {
    nameFilter.value = ''
    statusFilter.value = undefined
    resetSearchParams()
  }

  // ─── 技能计划列表 ───
  const skillPlans = ref<Api.SkillPlan.SkillPlanListItem[]>([])
  const skillPlansLoading = ref(false)

  onMounted(() => loadSkillPlans())

  async function loadSkillPlans() {
    if (skillPlans.value.length > 0) return
    skillPlansLoading.value = true
    try {
      const res = await fetchSkillPlanList({ current: 1, size: 200 })
      skillPlans.value = res?.list ?? []
    } catch {
      skillPlans.value = []
    } finally {
      skillPlansLoading.value = false
    }
  }

  // ─── 对话框状态 ───
  const dialogVisible = ref(false)
  const submitLoading = ref(false)
  const formRef = ref<FormInstance>()
  const editingItem = ref<WelfareItem | null>(null)

  const formData = reactive({
    name: '',
    description: '',
    dist_mode: 'per_user' as 'per_user' | 'per_character',
    require_skill_plan: false,
    skill_plan_ids: [] as number[],
    max_char_age_months: undefined as number | undefined,
    minimum_pap: undefined as number | undefined,
    require_evidence: false,
    example_evidence: '',
    status: 1 as number,
    sort_order: 0
  })

  const exampleEvidenceUploading = ref(false)

  async function handleExampleEvidenceUpload(file: UploadRawFile) {
    exampleEvidenceUploading.value = true
    try {
      const res = await uploadWelfareEvidence(file)
      formData.example_evidence = res.url
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('welfareSettings.uploadFailed'))
    } finally {
      exampleEvidenceUploading.value = false
    }
    return false // prevent default upload behavior
  }

  const hasMaxCharAge = computed(
    () => formData.max_char_age_months != null && formData.max_char_age_months > 0
  )

  const formRules = computed<FormRules>(() => ({
    name: [{ required: true, message: t('welfareSettings.validName'), trigger: 'blur' }],
    dist_mode: [{ required: true, message: t('welfareSettings.validDistMode'), trigger: 'change' }],
    skill_plan_ids: [
      {
        validator: (_rule: any, value: any, callback: any) => {
          if (formData.require_skill_plan && (!value || value.length === 0)) {
            callback(new Error(t('welfareSettings.validSkillPlan')))
          } else {
            callback()
          }
        },
        trigger: 'change'
      }
    ]
  }))

  function resetForm() {
    Object.assign(formData, {
      name: '',
      description: '',
      dist_mode: 'per_user',
      require_skill_plan: false,
      skill_plan_ids: [],
      max_char_age_months: undefined,
      minimum_pap: undefined,
      require_evidence: false,
      example_evidence: '',
      status: 1,
      sort_order: 0
    })
    editingItem.value = null
  }

  function openCreateDialog() {
    resetForm()
    loadSkillPlans()
    dialogVisible.value = true
  }

  function openEditDialog(row: WelfareItem) {
    editingItem.value = row
    Object.assign(formData, {
      name: row.name,
      description: row.description,
      dist_mode: row.dist_mode,
      require_skill_plan: row.require_skill_plan,
      skill_plan_ids: row.skill_plan_ids ?? [],
      max_char_age_months: row.max_char_age_months ?? undefined,
      minimum_pap: row.minimum_pap ?? undefined,
      require_evidence: row.require_evidence ?? false,
      example_evidence: row.example_evidence ?? '',
      status: row.status,
      sort_order: row.sort_order ?? 0
    })
    loadSkillPlans()
    dialogVisible.value = true
  }

  async function handleSubmit() {
    if (!formRef.value) return
    await formRef.value.validate()
    submitLoading.value = true
    try {
      const payload = {
        ...formData,
        skill_plan_ids: formData.require_skill_plan ? formData.skill_plan_ids : [],
        max_char_age_months: formData.max_char_age_months || null,
        minimum_pap: formData.minimum_pap || null,
        example_evidence: formData.require_evidence ? formData.example_evidence : ''
      }
      if (editingItem.value) {
        await adminUpdateWelfare({ id: editingItem.value.id, ...payload })
        ElMessage.success(t('welfareSettings.updateSuccess'))
      } else {
        await adminCreateWelfare(payload)
        ElMessage.success(t('welfareSettings.createSuccess'))
      }
      dialogVisible.value = false
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('welfareSettings.operationFailed'))
    } finally {
      submitLoading.value = false
    }
  }

  // ─── 导入历史记录 ───
  const importDialogVisible = ref(false)
  const importLoading = ref(false)
  const importCSV = ref('')
  const importWelfareID = ref<number>(0)

  function openImportDialog(row: WelfareItem) {
    importWelfareID.value = row.id
    importCSV.value = ''
    importDialogVisible.value = true
  }

  async function handleImport() {
    if (!importCSV.value.trim()) return
    importLoading.value = true
    try {
      const res = await adminImportWelfareRecords({
        welfare_id: importWelfareID.value,
        csv: importCSV.value
      })
      ElMessage.success(t('welfareSettings.importSuccess', { count: res.count }))
      importDialogVisible.value = false
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('welfareSettings.operationFailed'))
    } finally {
      importLoading.value = false
    }
  }

  async function handleDelete(row: WelfareItem) {
    await ElMessageBox.confirm(
      t('welfareSettings.deleteConfirm', { name: row.name }),
      t('welfareSettings.deleteTitle'),
      {
        confirmButtonText: t('welfareSettings.deleteBtn'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    try {
      await adminDeleteWelfare(row.id)
      ElMessage.success(t('welfareSettings.deleteSuccess'))
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('welfareSettings.deleteFailed'))
    }
  }

  // ─── Drag-and-drop reordering ───
  const tableRef = ref<any>()
  const reorderSaving = ref(false)

  const { start: startDraggable } = useDraggable<WelfareItem>(null, data, {
    handle: '.drag-handle',
    animation: 150,
    onEnd: async () => {
      reorderSaving.value = true
      try {
        await adminReorderWelfares(data.value.map((row) => row.id))
      } catch (e: any) {
        ElMessage.error(e?.message ?? t('welfareSettings.reorderFailed'))
      } finally {
        reorderSaving.value = false
      }
    }
  })

  onMounted(() => {
    if (!canManage.value) return
    nextTick(() => {
      const tbody = tableRef.value?.$el?.querySelector(
        '.el-table__body tbody'
      ) as HTMLElement | null
      if (tbody) startDraggable(tbody)
    })
  })
</script>
