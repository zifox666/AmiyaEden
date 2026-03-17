<!-- 舰队管理页面 -->
<template>
  <div class="fleet-page art-full-height">
    <!-- 搜索栏 -->
    <FleetSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams" />

    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElButton type="primary" :icon="Plus" @click="openCreateDialog">
            {{ $t('fleet.create') }}
          </ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <!-- 创建/编辑弹窗 -->
    <ElDialog
      v-model="dialogVisible"
      :title="editingFleet ? $t('fleet.edit') : $t('fleet.create')"
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
          <ElSelect
            v-model="formData.character_id"
            :placeholder="$t('fleet.fields.fcPlaceholder')"
            style="width: 100%"
          >
            <ElOption
              v-for="c in characters"
              :key="c.character_id"
              :label="c.character_name"
              :value="c.character_id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="$t('fleet.fields.fleetConfig')">
          <ElSelect
            v-model="formData.fleet_config_id"
            :placeholder="$t('fleet.fields.fleetConfigPlaceholder')"
            style="width: 100%"
            clearable
          >
            <ElOption v-for="fc in fleetConfigs" :key="fc.id" :label="fc.name" :value="fc.id" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="$t('fleet.fields.autoSrpMode')">
          <ElSelect v-model="formData.auto_srp_mode" style="width: 100%">
            <ElOption :label="$t('fleet.autoSrp.disabled')" value="disabled" />
            <ElOption :label="$t('fleet.autoSrp.submitOnly')" value="submit_only" />
            <ElOption :label="$t('fleet.autoSrp.autoApprove')" value="auto_approve" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem :label="$t('fleet.fields.timeRange')" prop="time_range">
          <ElDatePicker
            v-model="formData.time_range"
            type="datetimerange"
            range-separator="~"
            :start-placeholder="$t('fleet.fields.startAt')"
            :end-placeholder="$t('fleet.fields.endAt')"
            value-format="YYYY-MM-DDTHH:mm:ssZ"
            style="width: 100%"
            @change="formRef?.validateField('time_range')"
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
        <ElFormItem v-if="!editingFleet" :label="$t('fleet.fields.sendPing')">
          <ElSwitch v-model="formData.send_ping" />
          <span class="ml-2 text-xs text-gray-400">{{ $t('fleet.fields.sendPingHint') }}</span>
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
  import { useTable } from '@/hooks/core/useTable'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import FleetSearch from './modules/fleet-search.vue'
  import { fetchFleetList, createFleet, updateFleet, deleteFleet } from '@/api/fleet'
  import { fetchFleetConfigList } from '@/api/fleet-config'
  import { fetchMyCharacters } from '@/api/auth'
  import {
    ElTag,
    ElButton,
    ElMessageBox,
    ElSwitch,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import { Plus } from '@element-plus/icons-vue'
  import { useRouter } from 'vue-router'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'Fleets' })

  type FleetItem = Api.Fleet.FleetItem

  const { t } = useI18n()
  const router = useRouter()

  // ─── 重要度颜色映射 ───
  const IMPORTANCE_MAP: Record<string, string> = {
    strat_op: 'danger',
    cta: 'warning',
    other: 'info'
  }

  // ─── 时间格式化 ───
  const formatTime = (v: string) => {
    if (!v) return '-'
    return new Date(v).toLocaleString()
  }

  // 生成 YYYY-MM-DDTHH:mm:ssZ 格式的本地时间字符串（DatePicker value-format）
  function fmtLocalISO(d: Date): string {
    const pad = (n: number) => String(n).padStart(2, '0')
    const tz = -d.getTimezoneOffset()
    const sign = tz >= 0 ? '+' : '-'
    const tzH = Math.floor(Math.abs(tz) / 60)
    const tzM = Math.abs(tz) % 60
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}${sign}${pad(tzH)}:${pad(tzM)}`
  }
  function defaultTimeRange(): [string, string] {
    const now = Date.now()
    return [fmtLocalISO(new Date(now - 5_400_000)), fmtLocalISO(new Date(now + 5_400_000))]
  }

  // ─── 表格 ───
  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchFleetList,
      apiParams: { current: 1, size: 20 },
      columnsFactory: () => [
        { type: 'index', width: 60, label: '序号' },
        {
          prop: 'title',
          label: t('fleet.fields.title'),
          minWidth: 180,
          showOverflowTooltip: true,
          formatter: (row: FleetItem) =>
            h(
              ElButton,
              { type: 'primary', link: true, onClick: () => goDetail(row) },
              () => row.title
            )
        },
        {
          prop: 'importance',
          label: t('fleet.fields.importance'),
          width: 120,
          formatter: (row: FleetItem) =>
            h(
              ElTag,
              {
                type: (IMPORTANCE_MAP[row.importance] ?? 'info') as any,
                effect: 'dark',
                size: 'small'
              },
              () => t(`fleet.importance.${row.importance}`)
            )
        },
        {
          prop: 'fc_character_name',
          label: t('fleet.fields.fc'),
          width: 160,
          showOverflowTooltip: true
        },
        {
          prop: 'pap_count',
          label: t('fleet.fields.papCount'),
          width: 100
        },
        {
          prop: 'start_at',
          label: t('fleet.fields.timeRange'),
          width: 320,
          formatter: (row: FleetItem) =>
            h(
              'span',
              { class: 'text-xs text-gray-500' },
              `${formatTime(row.start_at)} ~ ${formatTime(row.end_at)}`
            )
        },
        {
          prop: 'created_at',
          label: '创建时间',
          width: 180,
          formatter: (row: FleetItem) => h('span', {}, formatTime(row.created_at))
        },
        {
          prop: 'actions',
          label: t('common.operation'),
          width: 200,
          fixed: 'right',
          formatter: (row: FleetItem) =>
            h('div', { class: 'flex gap-1' }, [
              h(ArtButtonTable, { type: 'view', onClick: () => goDetail(row) }),
              h(ArtButtonTable, { type: 'edit', onClick: () => openEditDialog(row) }),
              h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) })
            ])
        }
      ]
    }
  })

  // ─── 搜索 ───
  const searchForm = ref<{ importance: string | undefined }>({ importance: undefined })

  const handleSearch = (params: Record<string, any>) => {
    Object.assign(searchParams, params)
    getData()
  }

  // ─── FC 候选角色列表 ───
  const characters = ref<Api.Auth.EveCharacter[]>([])

  const loadCharacters = async () => {
    try {
      const res = await fetchMyCharacters()
      characters.value = res ?? []
    } catch {
      characters.value = []
    }
  }

  // ─── 导航 ───
  const goDetail = (row: FleetItem) => {
    router.push({ name: 'FleetDetail', params: { id: row.id } })
  }

  // ─── 创建 / 编辑 ───
  const dialogVisible = ref(false)
  const submitLoading = ref(false)
  const formRef = ref<FormInstance>()
  const editingFleet = ref<FleetItem | null>(null)

  const formData = reactive({
    title: '',
    description: '',
    importance: 'other' as 'strat_op' | 'cta' | 'other',
    pap_count: 1,
    character_id: undefined as number | undefined,
    time_range: null as [string, string] | null,
    send_ping: true,
    fleet_config_id: undefined as number | undefined,
    auto_srp_mode: 'disabled' as 'disabled' | 'submit_only' | 'auto_approve'
  })

  const formRules: FormRules = {
    title: [{ required: true, message: t('fleet.fields.titlePlaceholder'), trigger: 'blur' }],
    importance: [{ required: true, message: t('fleet.fields.importance'), trigger: 'change' }],
    pap_count: [{ required: true, message: t('fleet.fields.papCount'), trigger: 'blur' }],
    character_id: [{ required: true, message: t('fleet.fields.fcPlaceholder'), trigger: 'change' }],
    time_range: [{ required: true, message: t('fleet.fields.timeRange'), trigger: 'change' }]
  }

  function resetForm() {
    formData.title = ''
    formData.description = ''
    formData.importance = 'other'
    formData.pap_count = 1
    formData.character_id = undefined
    formData.time_range = null
    formData.send_ping = true
    formData.fleet_config_id = undefined
    formData.auto_srp_mode = 'disabled'
    editingFleet.value = null
  }

  function openCreateDialog() {
    resetForm()
    formData.time_range = defaultTimeRange()
    dialogVisible.value = true
  }

  function openEditDialog(row: FleetItem) {
    editingFleet.value = row
    formData.title = row.title
    formData.description = row.description
    formData.importance = row.importance
    formData.pap_count = row.pap_count
    formData.character_id = row.fc_character_id
    formData.time_range = [row.start_at, row.end_at]
    formData.fleet_config_id = row.fleet_config_id ?? undefined
    formData.auto_srp_mode = row.auto_srp_mode ?? 'disabled'
    dialogVisible.value = true
  }

  async function handleSubmit() {
    if (!formRef.value) return
    await formRef.value.validate()
    submitLoading.value = true
    try {
      const [start_at, end_at] = formData.time_range || ['', '']
      if (editingFleet.value) {
        await updateFleet(editingFleet.value.id, {
          title: formData.title,
          description: formData.description,
          importance: formData.importance,
          pap_count: formData.pap_count,
          character_id: formData.character_id,
          start_at,
          end_at,
          fleet_config_id: formData.fleet_config_id ?? null,
          auto_srp_mode: formData.auto_srp_mode
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
          end_at,
          send_ping: formData.send_ping,
          fleet_config_id: formData.fleet_config_id ?? null,
          auto_srp_mode: formData.auto_srp_mode
        })
        ElMessage.success(t('fleet.createSuccess'))
      }
      dialogVisible.value = false
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('common.error'))
    } finally {
      submitLoading.value = false
    }
  }

  // ─── 删除 ───
  async function handleDelete(row: FleetItem) {
    await ElMessageBox.confirm(t('fleet.deleteConfirm', { name: row.title }), t('fleet.delete'), {
      confirmButtonText: t('common.confirm'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    try {
      await deleteFleet(row.id)
      ElMessage.success(t('fleet.deleteSuccess'))
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('common.error'))
    }
  }

  // ─── 初始化 ───
  // ─── 舰队配置列表 ───
  const fleetConfigs = ref<Api.FleetConfig.FleetConfigItem[]>([])

  async function loadFleetConfigs() {
    try {
      const res = await fetchFleetConfigList({ current: 1, size: 100 })
      fleetConfigs.value = res?.list ?? []
    } catch {
      fleetConfigs.value = []
    }
  }

  onMounted(() => {
    loadCharacters()
    loadFleetConfigs()
  })
</script>
