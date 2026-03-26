<!-- SRP 舰船补损价格表管理 -->
<template>
  <div class="srp-prices-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="loadPrices">
        <template #left>
          <ElInput
            v-model="keyword"
            :placeholder="$t('srp.prices.searchPlaceholder')"
            clearable
            style="width: 200px"
            @keyup.enter="loadPrices"
            @clear="loadPrices"
          />
          <ElButton type="primary" :icon="Plus" @click="openAddDialog">
            {{ $t('srp.prices.addPrice') }}
          </ElButton>
          <ArtExcelExport
            :data="exportPricesData"
            :headers="pricesExportHeaders"
            :filename="`srp-prices_${new Date().toLocaleDateString()}`"
            sheet-name="SRP价格表"
            :button-text="$t('srp.prices.exportBtn')"
            type="success"
          />
          <ArtExcelImport @import-success="handleImport">
            {{ $t('srp.prices.importBtn') }}
          </ArtExcelImport>
        </template>
      </ArtTableHeader>

      <ArtTable :loading="loading" :data="prices" :columns="columns" />
    </ElCard>

    <ElDialog
      v-model="dialogVisible"
      :title="editTarget ? $t('srp.prices.editDialog') : $t('srp.prices.addDialog')"
      width="460px"
      destroy-on-close
    >
      <ElForm ref="formRef" :model="form" :rules="rules" label-width="130px">
        <ElFormItem :label="$t('srp.prices.fields.ship')" prop="ship_type_id">
          <SdeSearchSelect
            v-model="form.ship_type_id"
            :category-ids="[6]"
            :initial-options="dialogInitialOptions"
            :placeholder="$t('srp.prices.fields.namePlaceholder')"
            style="width: 100%"
            @select="onShipSelect"
          />
        </ElFormItem>
        <ElFormItem :label="$t('srp.prices.fields.amount')" prop="amount">
          <div class="million-isk-input" style="width: 100%">
            <ElInputNumber
              :model-value="toMillionISKInput(form.amount)"
              :min="0"
              :precision="2"
              :step="1"
              class="million-isk-input__control"
              @update:model-value="(v) => (form.amount = fromMillionISKInput(v))"
            />
            <span class="million-isk-input__suffix">{{ $t('common.millionIsk') }}</span>
          </div>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">{{ $t('srp.prices.cancelBtn') }}</ElButton>
        <ElButton type="primary" :loading="saving" @click="handleSave">{{
          $t('srp.prices.saveBtn')
        }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { Plus } from '@element-plus/icons-vue'
  import {
    ElCard,
    ElButton,
    ElInput,
    ElInputNumber,
    ElDialog,
    ElForm,
    ElFormItem,
    ElMessage,
    ElMessageBox,
    type FormInstance,
    type FormRules
  } from 'element-plus'
  import { useTableColumns } from '@/hooks/core/useTableColumns'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ArtExcelExport from '@/components/core/forms/art-excel-export/index.vue'
  import ArtExcelImport from '@/components/core/forms/art-excel-import/index.vue'
  import { fetchShipPrices, upsertShipPrice, deleteShipPrice } from '@/api/srp'
  import SdeSearchSelect from '@/components/business/SdeSearchSelect.vue'
  import { fromMillionISKInput, toMillionISKInput } from '@/utils/iskUnits'

  defineOptions({ name: 'SrpPrices' })

  const { t } = useI18n()

  type ShipPrice = Api.Srp.ShipPrice

  // ─── 格式化 ───
  const formatTime = (v: string) => (v ? new Date(v).toLocaleString() : '-')
  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(
      (v ?? 0) / 1_000_000
    )

  // ─── 列配置 ───
  const { columns, columnChecks } = useTableColumns<ShipPrice>(() => [
    { type: 'index', width: 60, label: '#' },
    {
      prop: 'ship_type_id',
      label: t('srp.prices.columns.typeId'),
      width: 110
    },
    {
      prop: 'ship_name',
      label: t('srp.prices.columns.name'),
      minWidth: 200,
      showOverflowTooltip: true
    },
    {
      prop: 'amount',
      label: t('srp.prices.columns.amount'),
      width: 200,
      formatter: (row: ShipPrice) =>
        h('span', { class: 'font-medium text-blue-600' }, `${formatISK(row.amount)} M ISK`)
    },
    {
      prop: 'updated_at',
      label: t('srp.prices.columns.updatedAt'),
      width: 180,
      formatter: (row: ShipPrice) => h('span', {}, formatTime(row.updated_at))
    },
    {
      prop: 'actions',
      label: t('srp.prices.columns.action'),
      width: 120,
      fixed: 'right',
      formatter: (row: ShipPrice) =>
        h('div', { class: 'flex gap-1' }, [
          h(ArtButtonTable, { type: 'edit', onClick: () => openEditDialog(row) }),
          h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) })
        ])
    }
  ])

  // ─── 导出 ───
  const pricesExportHeaders = {
    ship_type_id: 'TypeID',
    ship_name: '舰船名称',
    amount: '标准补损金额',
    updated_at: '最后更新'
  }
  const exportPricesData = computed(() =>
    prices.value.map((p) => ({
      ship_type_id: p.ship_type_id,
      ship_name: p.ship_name,
      amount: p.amount,
      updated_at: p.updated_at ? new Date(p.updated_at).toLocaleString() : '-'
    }))
  )

  // ─── 导入 ───
  const handleImport = async (rows: Record<string, unknown>[]) => {
    const items = rows
      .map((row) => ({
        ship_type_id: Number(row['TypeID'] ?? row['ship_type_id'] ?? 0),
        ship_name: String(row['舰船名称'] ?? row['ship_name'] ?? ''),
        amount: Number(row['标准补损金额'] ?? row['amount'] ?? 0)
      }))
      .filter((item) => item.ship_type_id > 0 && item.ship_name)
    if (!items.length) {
      ElMessage.warning(t('srp.prices.importNoData'))
      return
    }
    loading.value = true
    let success = 0
    try {
      for (const item of items) {
        await upsertShipPrice(item)
        success++
      }
      ElMessage.success(t('srp.prices.importSuccess', { count: success }))
      loadPrices()
    } catch {
      ElMessage.error(t('srp.prices.importFailed'))
    } finally {
      loading.value = false
    }
  }

  // ─── 数据 ───
  const prices = ref<ShipPrice[]>([])
  const loading = ref(false)
  const keyword = ref('')

  const loadPrices = async () => {
    loading.value = true
    try {
      const list = await fetchShipPrices(keyword.value || undefined)
      prices.value = list ?? []
    } catch {
      prices.value = []
    } finally {
      loading.value = false
    }
  }

  // ─── 新增 / 编辑 ───
  const dialogVisible = ref(false)
  const saving = ref(false)
  const formRef = ref<FormInstance>()
  const editTarget = ref<ShipPrice | null>(null)
  const dialogInitialOptions = ref<Api.Sde.FuzzySearchItem[]>([])

  const form = reactive({ id: 0, ship_type_id: 0 as number | null, ship_name: '', amount: 0 })

  function onShipSelect(item: Api.Sde.FuzzySearchItem | null) {
    if (item) {
      form.ship_type_id = item.id
      form.ship_name = item.name
    } else {
      form.ship_type_id = null
      form.ship_name = ''
    }
  }

  const rules: FormRules = {
    ship_type_id: [
      {
        required: true,
        validator: (_r, v, cb) => (v && v > 0 ? cb() : cb(new Error(t('srp.prices.validTypeId')))),
        trigger: 'change'
      }
    ],
    ship_name: [{ required: true, message: t('srp.prices.validName'), trigger: 'blur' }],
    amount: [
      {
        required: true,
        validator: (_r, v, cb) => (v >= 0 ? cb() : cb(new Error(t('srp.prices.validAmount')))),
        trigger: 'change'
      }
    ]
  }

  const openAddDialog = () => {
    editTarget.value = null
    form.id = 0
    form.ship_type_id = null
    form.ship_name = ''
    form.amount = 0
    dialogInitialOptions.value = []
    dialogVisible.value = true
  }

  const openEditDialog = (row: ShipPrice) => {
    editTarget.value = row
    form.id = row.id
    form.ship_type_id = row.ship_type_id
    form.ship_name = row.ship_name
    form.amount = row.amount
    dialogInitialOptions.value = [
      { id: row.ship_type_id, name: row.ship_name, group_id: 0, group_name: '', category: 'type' }
    ]
    dialogVisible.value = true
  }

  const handleSave = async () => {
    await formRef.value?.validate()
    saving.value = true
    try {
      await upsertShipPrice({ ...form, ship_type_id: form.ship_type_id! })
      ElMessage.success(
        editTarget.value ? t('srp.prices.updateSuccess') : t('srp.prices.addSuccess')
      )
      dialogVisible.value = false
      loadPrices()
    } catch {
      /* handled */
    } finally {
      saving.value = false
    }
  }

  // ─── 删除 ───
  const handleDelete = async (row: ShipPrice) => {
    await ElMessageBox.confirm(t('srp.prices.deleteConfirm'), t('common.confirm'), {
      confirmButtonText: t('srp.prices.deleteBtn'),
      cancelButtonText: t('common.cancel'),
      type: 'warning'
    })
    try {
      await deleteShipPrice(row.id)
      ElMessage.success(t('srp.prices.deleteSuccess'))
      loadPrices()
    } catch {
      /* handled */
    }
  }

  onMounted(loadPrices)
</script>

<style scoped>
  .million-isk-input {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
    gap: 8px;
  }

  .million-isk-input__control {
    width: 100%;
  }

  .million-isk-input__suffix {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    white-space: nowrap;
  }
</style>
