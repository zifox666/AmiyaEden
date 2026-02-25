<!-- SRP 舰船补损价格表管理 -->
<template>
  <div class="srp-prices-page art-full-height">
    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="flex items-center justify-between">
          <h2 class="text-base font-medium">{{ $t('srp.prices.title') }}</h2>
          <div class="flex gap-2">
            <ElInput v-model="keyword" :placeholder="$t('srp.prices.searchPlaceholder')" clearable style="width: 200px"
              @keyup.enter="loadPrices" @clear="loadPrices" />
            <ElButton :loading="loading" @click="loadPrices">
              <el-icon class="mr-1"><Refresh /></el-icon>
              {{ $t('srp.prices.refresh') }}
            </ElButton>
            <ElButton type="primary" @click="openAddDialog">
              <el-icon class="mr-1"><Plus /></el-icon>
              {{ $t('srp.prices.addPrice') }}
            </ElButton>
          </div>
        </div>
      </template>

      <ElTable v-loading="loading" :data="prices" stripe border style="width: 100%">
        <ElTableColumn prop="ship_type_id" :label="$t('srp.prices.columns.typeId')" width="110" align="center" />
        <ElTableColumn prop="ship_name" :label="$t('srp.prices.columns.name')" min-width="200" />
        <ElTableColumn prop="amount" :label="$t('srp.prices.columns.amount')" width="200" align="right">
          <template #default="{ row }">
            <span class="font-medium text-blue-600">{{ formatISK(row.amount) }} ISK</span>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="updated_at" :label="$t('srp.prices.columns.updatedAt')" width="200">
          <template #default="{ row }">{{ formatTime(row.updated_at) }}</template>
        </ElTableColumn>
        <ElTableColumn :label="$t('srp.prices.columns.action')" width="150" align="center" fixed="right">
          <template #default="{ row }">
            <ElButton size="small" type="primary" link @click="openEditDialog(row)">{{ $t('srp.prices.editBtn') }}</ElButton>
            <ElPopconfirm :title="$t('srp.prices.deleteConfirm')" @confirm="handleDelete(row.id)">
              <template #reference>
                <ElButton size="small" type="danger" link>{{ $t('srp.prices.deleteBtn') }}</ElButton>
              </template>
            </ElPopconfirm>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <ElDialog v-model="dialogVisible" :title="editTarget ? $t('srp.prices.editDialog') : $t('srp.prices.addDialog')" width="460px">
      <ElForm ref="formRef" :model="form" :rules="rules" label-width="130px">
        <ElFormItem :label="$t('srp.prices.fields.typeId')" prop="ship_type_id">
          <ElInputNumber v-model="form.ship_type_id" :min="1" style="width: 100%" />
        </ElFormItem>
        <ElFormItem :label="$t('srp.prices.fields.name')" prop="ship_name">
          <ElInput v-model="form.ship_name" :placeholder="$t('srp.prices.fields.namePlaceholder')" />
        </ElFormItem>
        <ElFormItem :label="$t('srp.prices.fields.amount')" prop="amount">
          <ElInputNumber v-model="form.amount" :min="0" :precision="2" :step="10000000" style="width: 100%" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">{{ $t('srp.prices.cancelBtn') }}</ElButton>
        <ElButton type="primary" :loading="saving" @click="handleSave">{{ $t('srp.prices.saveBtn') }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { Plus, Refresh } from '@element-plus/icons-vue'
  import {
    ElCard, ElTable, ElTableColumn, ElButton, ElInput, ElInputNumber,
    ElDialog, ElForm, ElFormItem, ElPopconfirm, ElMessage, type FormInstance, type FormRules
  } from 'element-plus'
  import { fetchShipPrices, upsertShipPrice, deleteShipPrice } from '@/api/srp'

  defineOptions({ name: 'SrpPrices' })

  const { t } = useI18n()

  const prices = ref<Api.Srp.ShipPrice[]>([])
  const loading = ref(false)
  const keyword = ref('')

  const loadPrices = async () => {
    loading.value = true
    try {
      const list = await fetchShipPrices(keyword.value || undefined)
      prices.value = list ?? []
    } catch { prices.value = [] }
    finally { loading.value = false }
  }

  const dialogVisible = ref(false)
  const saving = ref(false)
  const formRef = ref<FormInstance>()
  const editTarget = ref<Api.Srp.ShipPrice | null>(null)

  const form = reactive({ id: 0, ship_type_id: 0, ship_name: '', amount: 0 })

  const rules: FormRules = {
    ship_type_id: [{ required: true, validator: (_r, v, cb) => v > 0 ? cb() : cb(new Error(t('srp.prices.validTypeId'))), trigger: 'change' }],
    ship_name: [{ required: true, message: t('srp.prices.validName'), trigger: 'blur' }],
    amount: [{ required: true, validator: (_r, v, cb) => v >= 0 ? cb() : cb(new Error(t('srp.prices.validAmount'))), trigger: 'change' }],
  }

  const openAddDialog = () => {
    editTarget.value = null
    form.id = 0; form.ship_type_id = 0; form.ship_name = ''; form.amount = 0
    dialogVisible.value = true
  }

  const openEditDialog = (row: Api.Srp.ShipPrice) => {
    editTarget.value = row
    form.id = row.id; form.ship_type_id = row.ship_type_id; form.ship_name = row.ship_name; form.amount = row.amount
    dialogVisible.value = true
  }

  const handleSave = async () => {
    await formRef.value?.validate()
    saving.value = true
    try {
      await upsertShipPrice({ ...form })
      ElMessage.success(editTarget.value ? t('srp.prices.updateSuccess') : t('srp.prices.addSuccess'))
      dialogVisible.value = false
      loadPrices()
    } catch { /* handled */ }
    finally { saving.value = false }
  }

  const handleDelete = async (id: number) => {
    try {
      await deleteShipPrice(id)
      ElMessage.success(t('srp.prices.deleteSuccess'))
      loadPrices()
    } catch { /* handled */ }
  }

  const formatTime = (v: string) => v ? new Date(v).toLocaleString() : '-'
  const formatISK = (v: number) =>
    new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(v ?? 0)

  onMounted(loadPrices)
</script>