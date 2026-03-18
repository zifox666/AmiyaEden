<!-- 商品管理面板 -->
<template>
  <ElCard class="art-table-card" shadow="never">
    <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
      <template #left>
        <div class="flex items-center gap-2">
          <ElButton type="success" :icon="Plus" @click="openCreateDialog">{{
            t('shop.manage.createProduct')
          }}</ElButton>
          <ElInput
            v-model="nameFilter"
            :placeholder="t('shop.manage.filterName')"
            clearable
            style="width: 160px"
            @keyup.enter="handleSearch"
          />
          <ElSelect
            v-model="typeFilter"
            :placeholder="t('shop.manage.filterType')"
            clearable
            style="width: 140px"
            @change="handleSearch"
          >
            <ElOption :label="t('shop.manage.typeNormal')" value="normal" />
          </ElSelect>
          <ElSelect
            v-model="statusFilter"
            :placeholder="t('shop.manage.filterStatus')"
            clearable
            style="width: 120px"
            @change="handleSearch"
          >
            <ElOption :label="t('shop.manage.statusOnSale')" :value="1" />
            <ElOption :label="t('shop.manage.statusOffSale')" :value="0" />
          </ElSelect>
          <ElButton type="primary" @click="handleSearch">{{ t('shop.manage.search') }}</ElButton>
          <ElButton @click="handleReset">{{ t('shop.manage.reset') }}</ElButton>
        </div>
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

  <!-- 商品编辑对话框 -->
  <ElDialog
    v-model="dialogVisible"
    :title="editingProduct ? t('shop.manage.editProduct') : t('shop.manage.createProduct')"
    width="580px"
    destroy-on-close
  >
    <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="100px">
      <ElFormItem :label="$t('shop.productName')" prop="name">
        <ElInput v-model="formData.name" :placeholder="$t('shop.manage.namePlaceholder')" />
      </ElFormItem>
      <ElFormItem :label="$t('shop.manage.sdeSearch')">
        <SdeSearchSelect
          v-model="sdeTypeId"
          :placeholder="$t('shop.manage.sdeSearchPlaceholder')"
          style="width: 100%"
          @select="onSdeSelect"
        />
      </ElFormItem>
      <ElFormItem :label="t('shop.manage.description')">
        <ElInput
          v-model="formData.description"
          type="textarea"
          :rows="3"
          :placeholder="t('shop.manage.descriptionPlaceholder')"
        />
      </ElFormItem>
      <!-- 图片上传区域 -->
      <ElFormItem :label="t('shop.manage.productImage')">
        <div class="image-upload-area">
          <div v-if="formData.image" class="image-preview">
            <img :src="formData.image" :alt="t('shop.manage.productImage')" />
            <div class="image-actions">
              <ElButton size="small" type="danger" text @click="formData.image = ''">
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
            :http-request="handleImageUpload"
          >
            <div class="upload-placeholder">
              <el-icon v-if="!imageUploading" :size="32"><Plus /></el-icon>
              <el-icon v-else :size="32" class="animate-spin"><Loading /></el-icon>
              <span>{{
                imageUploading ? t('shop.manage.uploading') : t('shop.manage.uploadImage')
              }}</span>
              <span class="upload-hint">{{ t('shop.manage.uploadHint') }}</span>
            </div>
          </ElUpload>
        </div>
      </ElFormItem>
      <ElFormItem :label="t('shop.manage.price')" prop="price">
        <ElInputNumber
          v-model="formData.price"
          :min="0.01"
          :precision="2"
          :step="10"
          style="width: 200px"
        />
      </ElFormItem>
      <ElFormItem :label="t('shop.manage.type')" prop="type">
        <ElSelect v-model="formData.type" style="width: 200px">
          <ElOption :label="t('shop.manage.typeNormal')" value="normal" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem :label="t('shop.manage.stock')">
        <ElInputNumber v-model="formData.stock" :min="-1" style="width: 200px" />
        <span class="ml-2 text-xs text-gray-400">{{ t('shop.manage.stockUnlimitedHint') }}</span>
      </ElFormItem>
      <ElFormItem :label="t('shop.manage.limitPerUser')">
        <ElInputNumber v-model="formData.max_per_user" :min="0" style="width: 200px" />
        <span class="ml-2 text-xs text-gray-400">{{ t('shop.manage.limitPerUserHint') }}</span>
      </ElFormItem>
      <ElFormItem :label="t('shop.manage.limitPeriod')">
        <ElSelect v-model="formData.limit_period" style="width: 200px">
          <ElOption :label="t('shop.manage.periodForever')" value="forever" />
          <ElOption :label="t('shop.manage.periodDaily')" value="daily" />
          <ElOption :label="t('shop.manage.periodWeekly')" value="weekly" />
          <ElOption :label="t('shop.manage.periodMonthly')" value="monthly" />
        </ElSelect>
        <span class="ml-2 text-xs text-gray-400">{{ t('shop.manage.limitPeriodHint') }}</span>
      </ElFormItem>
      <ElFormItem :label="t('shop.manage.needApproval')">
        <ElSwitch v-model="formData.need_approval" />
      </ElFormItem>
      <ElFormItem :label="t('shop.manage.status')">
        <ElSelect v-model="formData.status" style="width: 200px">
          <ElOption :label="t('shop.manage.statusOnSale')" :value="1" />
          <ElOption :label="t('shop.manage.statusOffSale')" :value="0" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem :label="t('shop.manage.sortOrder')">
        <ElInputNumber v-model="formData.sort_order" :min="0" style="width: 200px" />
        <span class="ml-2 text-xs text-gray-400">{{ t('shop.manage.sortHint') }}</span>
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="dialogVisible = false">{{ t('common.cancel') }}</ElButton>
      <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">{{
        t('common.confirm')
      }}</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import {
    ElTag,
    ElButton,
    ElInput,
    ElSelect,
    ElOption,
    ElSwitch,
    ElMessage,
    ElMessageBox,
    ElUpload
  } from 'element-plus'
  import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
  import { Plus, Delete, Loading } from '@element-plus/icons-vue'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import SdeSearchSelect from '@/components/business/SdeSearchSelect.vue'
  import {
    adminListProducts,
    adminCreateProduct,
    adminUpdateProduct,
    adminDeleteProduct,
    uploadShopImage
  } from '@/api/shop'
  import { useTable } from '@/hooks/core/useTable'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'ManageProducts' })
  const { t } = useI18n()

  type Product = Api.Shop.Product

  // ─── 商品类型/状态映射 ───
  const PRODUCT_TYPE_CONFIG = computed(
    () =>
      ({
        normal: { label: t('shop.manage.typeNormalShort'), type: 'primary' }
      }) as unknown as Record<string, { label: string; type: string }>
  )

  const PRODUCT_STATUS_CONFIG = computed(
    () =>
      ({
        1: { label: t('shop.manage.statusOnSale'), type: 'success' },
        0: { label: t('shop.manage.statusOffSale'), type: 'danger' }
      }) as unknown as Record<number, { label: string; type: string }>
  )

  const formatISK = (v: number) =>
    v.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

  // ─── 搜索过滤状态 ───
  const nameFilter = ref('')
  const typeFilter = ref('')
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
      apiFn: adminListProducts,
      apiParams: { current: 1, size: 20 },
      immediate: false,
      columnsFactory: () => [
        { type: 'index', width: 60, label: '#' },
        {
          prop: 'image',
          label: t('shop.manage.colImage'),
          width: 64,
          formatter: (row: Product) => {
            if (!row.image) return null
            return h('img', {
              src: row.image,
              style: 'width:40px;height:40px;object-fit:cover;border-radius:4px'
            })
          }
        },
        {
          prop: 'name',
          label: t('shop.productName'),
          minWidth: 140,
          showOverflowTooltip: true
        },
        {
          prop: 'type',
          label: t('shop.manage.colType'),
          width: 100,
          formatter: (row: Product) => {
            const cfg = PRODUCT_TYPE_CONFIG.value[row.type] ?? { label: row.type, type: 'info' }
            return h(
              ElTag,
              { type: cfg.type as any, size: 'small', effect: 'plain' },
              () => cfg.label
            )
          }
        },
        {
          prop: 'price',
          label: t('shop.manage.price'),
          width: 130,
          formatter: (row: Product) =>
            h('span', { class: 'font-medium text-orange-600' }, formatISK(row.price))
        },
        {
          prop: 'stock',
          label: t('shop.manage.stock'),
          width: 90,
          formatter: (row: Product) => {
            if (row.stock < 0)
              return h('span', { class: 'text-gray-400' }, t('shop.manage.stockUnlimited'))
            return h(
              'span',
              { class: row.stock === 0 ? 'text-red-500 font-bold' : '' },
              String(row.stock)
            )
          }
        },
        {
          prop: 'need_approval',
          label: t('shop.manage.colApproval'),
          width: 90,
          formatter: (row: Product) =>
            h(
              ElTag,
              { type: row.need_approval ? 'warning' : 'info', size: 'small', effect: 'plain' },
              () => (row.need_approval ? t('shop.manage.yes') : t('shop.manage.no'))
            )
        },
        {
          prop: 'status',
          label: t('shop.manage.status'),
          width: 90,
          formatter: (row: Product) => {
            const cfg = PRODUCT_STATUS_CONFIG.value[row.status] ?? {
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
          prop: 'sort_order',
          label: t('shop.manage.colSort'),
          width: 80
        },
        {
          prop: 'actions',
          label: t('common.operation'),
          width: 120,
          fixed: 'right',
          formatter: (row: Product) =>
            h('div', { class: 'flex gap-1' }, [
              h(ArtButtonTable, { type: 'edit', onClick: () => openEditDialog(row) }),
              h(ArtButtonTable, { type: 'delete', onClick: () => handleDelete(row) })
            ])
        }
      ]
    }
  })

  function handleSearch() {
    Object.assign(searchParams, {
      name: nameFilter.value || undefined,
      type: typeFilter.value || undefined,
      status: statusFilter.value,
      current: 1
    })
    getData()
  }

  function handleReset() {
    nameFilter.value = ''
    typeFilter.value = ''
    statusFilter.value = undefined
    resetSearchParams()
  }

  // ─── 对话框状态 ───
  const dialogVisible = ref(false)
  const submitLoading = ref(false)
  const formRef = ref<FormInstance>()
  const editingProduct = ref<Product | null>(null)

  const formData = reactive({
    name: '',
    description: '',
    image: '',
    price: 0,
    type: 'normal' as 'normal' | 'redeem',
    stock: -1,
    max_per_user: 0,
    limit_period: 'forever' as 'forever' | 'daily' | 'weekly' | 'monthly',
    need_approval: false,
    status: 1 as number,
    sort_order: 0
  })

  const formRules = computed<FormRules>(() => ({
    name: [{ required: true, message: t('shop.manage.validName'), trigger: 'blur' }],
    price: [{ required: true, message: t('shop.manage.validPrice'), trigger: 'blur' }],
    type: [{ required: true, message: t('shop.manage.validType'), trigger: 'change' }]
  }))

  function resetForm() {
    Object.assign(formData, {
      name: '',
      description: '',
      image: '',
      price: 0,
      type: 'normal',
      stock: -1,
      max_per_user: 0,
      limit_period: 'forever',
      need_approval: false,
      status: 1,
      sort_order: 0
    })
    editingProduct.value = null
  }

  function openCreateDialog() {
    resetForm()
    dialogVisible.value = true
  }

  function openEditDialog(row: Product) {
    editingProduct.value = row
    Object.assign(formData, {
      name: row.name,
      description: row.description,
      image: row.image,
      price: row.price,
      type: row.type,
      stock: row.stock,
      max_per_user: row.max_per_user,
      limit_period: row.limit_period || 'forever',
      need_approval: row.need_approval,
      status: row.status,
      sort_order: row.sort_order
    })
    dialogVisible.value = true
  }

  async function handleSubmit() {
    if (!formRef.value) return
    await formRef.value.validate()
    submitLoading.value = true
    try {
      if (editingProduct.value) {
        await adminUpdateProduct({ id: editingProduct.value.id, ...formData })
        ElMessage.success(t('shop.manage.updateSuccess'))
      } else {
        await adminCreateProduct({ ...formData })
        ElMessage.success(t('shop.manage.createSuccess'))
      }
      dialogVisible.value = false
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('shop.manage.operationFailed'))
    } finally {
      submitLoading.value = false
    }
  }

  async function handleDelete(row: Product) {
    await ElMessageBox.confirm(
      t('shop.manage.deleteConfirm', { name: row.name }),
      t('shop.manage.deleteTitle'),
      {
        confirmButtonText: t('shop.manage.deleteBtn'),
        cancelButtonText: t('common.cancel'),
        type: 'warning'
      }
    )
    try {
      await adminDeleteProduct(row.id)
      ElMessage.success(t('shop.manage.deleteSuccess'))
      refreshData()
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('shop.manage.deleteFailed'))
    }
  }

  // ─── 图片上传 ───
  const imageUploading = ref(false)

  function handleImageBeforeUpload(file: File) {
    const maxSize = 5 * 1024 * 1024
    if (file.size > maxSize) {
      ElMessage.error(t('shop.manage.imageTooLarge'))
      return false
    }
    return true
  }

  async function handleImageUpload(options: UploadRequestOptions) {
    imageUploading.value = true
    try {
      const res = (await uploadShopImage(options.file as File)) as any
      formData.image = res?.url ?? ''
      ElMessage.success(t('shop.manage.uploadSuccess'))
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('shop.manage.uploadFailed'))
    } finally {
      imageUploading.value = false
    }
  }

  // ─── SDE 物品搜索 ───
  const sdeTypeId = ref<number | null>(null)

  function onSdeSelect(item: Api.Sde.FuzzySearchItem | null) {
    if (item) {
      if (!formData.name) {
        formData.name = item.name
      }
      formData.image = `https://images.evetech.net/types/${item.id}/icon?size=256`
      ElMessage.success(t('shop.manage.sdeSelected', { name: item.name }))
    }
  }

  defineExpose({ load: getData, refresh: refreshData })
</script>

<style scoped>
  .image-upload-area {
    width: 160px;
    height: 160px;
    border: 1px dashed var(--el-border-color);
    border-radius: 6px;
    overflow: hidden;
    position: relative;
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
    top: 4px;
    right: 4px;
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
    font-size: 13px;
    transition: color 0.2s;
  }

  .upload-placeholder:hover {
    color: var(--el-color-primary);
  }

  .upload-hint {
    font-size: 11px;
    color: var(--el-text-color-secondary);
    text-align: center;
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
