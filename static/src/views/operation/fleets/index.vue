<!-- 舰队管理页面 -->
<template>
  <div class="fleet-page art-full-height">
    <!-- 搜索栏 -->
    <ElCard class="art-search-card" shadow="never">
      <div class="flex items-center gap-3 flex-wrap">
        <ElSelect
          v-model="searchForm.importance"
          :placeholder="$t('fleet.fields.importance')"
          clearable
          style="width: 140px"
          @change="handleSearch"
        >
          <ElOption label="Strat Op" value="strat_op" />
          <ElOption label="CTA" value="cta" />
          <ElOption label="Other" value="other" />
        </ElSelect>
        <ElButton @click="resetSearch">
          {{ $t('table.searchBar.reset') }}
        </ElButton>
      </div>
    </ElCard>

    <ElCard class="art-table-card" shadow="never">
      <!-- 操作栏 -->
      <div class="flex items-center justify-between mb-4">
        <h3 class="text-base font-medium">{{ $t('fleet.title') }}</h3>
        <div class="flex items-center gap-2">
          <ElButton :loading="loading" @click="loadFleets">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
          <ElButton type="primary" @click="showCreateDialog">
            <el-icon class="mr-1"><Plus /></el-icon>
            {{ $t('fleet.create') }}
          </ElButton>
        </div>
      </div>

      <!-- 舰队表格 -->
      <ElTable v-loading="loading" :data="fleets" stripe border style="width: 100%">
        <ElTableColumn prop="title" :label="$t('fleet.fields.title')" min-width="180">
          <template #default="{ row }">
            <ElButton type="primary" link @click="goDetail(row)">{{ row.title }}</ElButton>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="importance" :label="$t('fleet.fields.importance')" width="120" align="center">
          <template #default="{ row }">
            <ElTag :type="importanceType(row.importance)" size="small" effect="dark">
              {{ $t(`fleet.importance.${row.importance}`) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="fc_character_name" :label="$t('fleet.fields.fc')" width="160" />
        <ElTableColumn prop="pap_count" :label="$t('fleet.fields.papCount')" width="100" align="center" />
        <ElTableColumn :label="$t('fleet.fields.timeRange')" width="320">
          <template #default="{ row }">
            <span class="text-xs text-gray-500">{{ formatTime(row.start_at) }} ~ {{ formatTime(row.end_at) }}</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('common.operation')" width="160" fixed="right" align="center">
          <template #default="{ row }">
            <ElButton type="primary" link size="small" @click="goDetail(row)">
              {{ $t('common.detail') }}
            </ElButton>
            <ElButton type="warning" link size="small" @click="showEditDialog(row)">
              {{ $t('common.edit') }}
            </ElButton>
            <ElButton type="danger" link size="small" @click="handleDelete(row)">
              {{ $t('common.delete') }}
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <ElPagination
          v-model:current-page="pagination.current"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next, jumper"
          background
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </ElCard>

    <!-- 创建/编辑弹窗 -->
    <ElDialog
      v-model="dialogVisible"
      :title="isEdit ? $t('fleet.edit') : $t('fleet.create')"
      width="520px"
      destroy-on-close
    >
      <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="100px">
        <ElFormItem :label="$t('fleet.fields.title')" prop="title">
          <ElInput v-model="formData.title" :placeholder="$t('fleet.fields.titlePlaceholder')" />
        </ElFormItem>
        <ElFormItem :label="$t('fleet.fields.importance')" prop="importance">
          <ElSelect v-model="formData.importance" style="width: 100%">
            <ElOption label="Strat Op" value="strat_op" />
            <ElOption label="CTA" value="cta" />
            <ElOption label="Other" value="other" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="$t('fleet.fields.papCount')" prop="pap_count">
          <ElInputNumber v-model="formData.pap_count" :min="0" :max="100" style="width: 100%" />
        </ElFormItem>
        <ElFormItem :label="$t('fleet.fields.fc')" prop="character_id">
          <ElSelect v-model="formData.character_id" :placeholder="$t('fleet.fields.fcPlaceholder')" style="width: 100%">
            <ElOption
              v-for="c in characters"
              :key="c.character_id"
              :label="c.character_name"
              :value="c.character_id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="$t('fleet.fields.timeRange')" prop="time_range">
          <ElDatePicker
            v-model="timeRange"
            type="datetimerange"
            range-separator="~"
            :start-placeholder="$t('fleet.fields.startAt')"
            :end-placeholder="$t('fleet.fields.endAt')"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            style="width: 100%"
          />
        </ElFormItem>
        <ElFormItem :label="$t('fleet.fields.description')">
          <ElInput
            v-model="formData.description"
            type="textarea"
            :rows="3"
            :placeholder="$t('fleet.fields.descriptionPlaceholder')"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ $t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { Refresh, Plus } from '@element-plus/icons-vue'
  import {
    ElCard,
    ElTable,
    ElTableColumn,
    ElTag,
    ElButton,
    ElPagination,
    ElSelect,
    ElOption,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInput,
    ElDatePicker,
    ElMessageBox,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useRouter } from 'vue-router'
  import { fetchFleetList, createFleet, updateFleet, deleteFleet } from '@/api/fleet'
  import { fetchMyCharacters } from '@/api/auth'

  defineOptions({ name: 'Fleets' })

  const { t } = useI18n()
  const router = useRouter()

  // ---- 数据 ----
  const fleets = ref<Api.Fleet.FleetItem[]>([])
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const loading = ref(false)
  const submitLoading = ref(false)

  // ---- 分页 ----
  const pagination = reactive({ current: 1, size: 20, total: 0 })

  // ---- 搜索 ----
  const searchForm = reactive<{ importance: string | undefined }>({ importance: undefined })

  // ---- 弹窗 ----
  const dialogVisible = ref(false)
  const isEdit = ref(false)
  const editId = ref('')
  const formRef = ref<FormInstance>()
  const timeRange = ref<[string, string] | null>(null)

  const formData = reactive({
    title: '',
    description: '',
    importance: 'other' as 'strat_op' | 'cta' | 'other',
    pap_count: 1,
    character_id: undefined as number | undefined
  })

  const formRules: FormRules = {
    title: [{ required: true, message: t('fleet.fields.titlePlaceholder'), trigger: 'blur' }],
    importance: [{ required: true, message: t('fleet.fields.importance'), trigger: 'change' }],
    pap_count: [{ required: true, message: t('fleet.fields.papCount'), trigger: 'blur' }],
    character_id: [{ required: true, message: t('fleet.fields.fcPlaceholder'), trigger: 'change' }],
    start_at: [{ required: true, message: t('fleet.fields.timeRange'), trigger: 'change' }]
  }

  // ---- 等级样式 ----
  const IMPORTANCE_MAP: Record<string, string> = {
    strat_op: 'danger',
    cta: 'warning',
    other: 'info'
  }
  const importanceType = (v: string) => (IMPORTANCE_MAP[v] || 'info') as any

  // ---- 时间格式化 ----
  const formatTime = (v: string) => {
    if (!v) return '-'
    return new Date(v).toLocaleString()
  }

  // ---- 加载数据 ----
  const loadFleets = async () => {
    loading.value = true
    try {
      const params: Api.Fleet.FleetSearchParams = {
        current: pagination.current,
        size: pagination.size
      }
      if (searchForm.importance) params.importance = searchForm.importance

      const res = await fetchFleetList(params)
      if (res) {
        fleets.value = res.records ?? []
        pagination.total = res.total ?? 0
        pagination.current = res.current ?? 1
        pagination.size = res.size ?? 20
      } else {
        fleets.value = []
        pagination.total = 0
      }
    } catch {
      fleets.value = []
      pagination.total = 0
    } finally {
      loading.value = false
    }
  }

  const loadCharacters = async () => {
    try {
      const res = await fetchMyCharacters()
      characters.value = res ?? []
    } catch {
      characters.value = []
    }
  }

  // ---- 搜索 ----
  const handleSearch = () => {
    pagination.current = 1
    loadFleets()
  }

  const resetSearch = () => {
    searchForm.importance = undefined
    pagination.current = 1
    loadFleets()
  }

  // ---- 分页 ----
  const handleSizeChange = () => {
    pagination.current = 1
    loadFleets()
  }
  const handleCurrentChange = () => {
    loadFleets()
  }

  // ---- 导航 ----
  const goDetail = (row: Api.Fleet.FleetItem) => {
    router.push({ name: 'FleetDetail', params: { id: row.id } })
  }

  // ---- 弹窗操作 ----
  const resetForm = () => {
    formData.title = ''
    formData.description = ''
    formData.importance = 'other'
    formData.pap_count = 1
    formData.character_id = undefined
    timeRange.value = null
  }

  const showCreateDialog = () => {
    resetForm()
    isEdit.value = false
    editId.value = ''
    dialogVisible.value = true
  }

  const showEditDialog = (row: Api.Fleet.FleetItem) => {
    isEdit.value = true
    editId.value = row.id
    formData.title = row.title
    formData.description = row.description
    formData.importance = row.importance
    formData.pap_count = row.pap_count
    formData.character_id = row.fc_character_id
    timeRange.value = [row.start_at, row.end_at]
    dialogVisible.value = true
  }

  const handleSubmit = async () => {
    if (!formRef.value) return
    try {
      await formRef.value.validate()
    } catch {
      return
    }

    submitLoading.value = true
    try {
      const [start_at, end_at] = timeRange.value || ['', '']
      if (isEdit.value) {
        await updateFleet(editId.value, {
          title: formData.title,
          description: formData.description,
          importance: formData.importance,
          pap_count: formData.pap_count,
          character_id: formData.character_id,
          start_at,
          end_at
        })
        ElMessage.success(t('fleet.updateSuccess'))
      } else {
        await createFleet({
          title: formData.title,
          description: formData.description,
          importance: formData.importance,
          pap_count: formData.pap_count,
          character_id: formData.character_id!,
          start_at,
          end_at
        })
        ElMessage.success(t('fleet.createSuccess'))
      }
      dialogVisible.value = false
      loadFleets()
    } catch (e) {
      console.error('Submit fleet error:', e)
    } finally {
      submitLoading.value = false
    }
  }

  // ---- 删除 ----
  const handleDelete = (row: Api.Fleet.FleetItem) => {
    ElMessageBox.confirm(t('fleet.deleteConfirm', { name: row.title }), t('fleet.delete'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'error'
    })
      .then(async () => {
        await deleteFleet(row.id)
        ElMessage.success(t('fleet.deleteSuccess'))
        loadFleets()
      })
      .catch(() => {})
  }

  // ---- 初始化 ----
  onMounted(() => {
    loadFleets()
    loadCharacters()
  })
</script>

<style scoped>
  .art-search-card {
    margin-bottom: 16px;
  }
  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
  }
</style>
