<!-- 商店管理页面（管理员） -->
<template>
  <div class="shop-admin-page art-full-height">
    <ElTabs v-model="activeTab" type="border-card">
      <!-- 商品管理 -->
      <ElTabPane label="商品管理" name="products">
        <div class="flex items-center gap-4 mb-4 flex-wrap">
          <ElInput v-model="productFilter.name" placeholder="商品名称" clearable style="width: 160px" />
          <ElSelect v-model="productFilter.type" placeholder="商品类型" clearable style="width: 140px">
            <ElOption label="普通商品" value="normal" />
            <ElOption label="兑换码" value="redeem" />
          </ElSelect>
          <ElSelect v-model="productFilter.status" placeholder="状态" clearable style="width: 120px">
            <ElOption label="上架" :value="1" />
            <ElOption label="下架" :value="0" />
          </ElSelect>
          <ElButton type="primary" @click="loadProducts">查询</ElButton>
          <ElButton type="success" @click="showProductDialog()">新增商品</ElButton>
        </div>

        <ElTable v-loading="productLoading" :data="products" stripe border style="width: 100%">
          <ElTableColumn type="index" width="60" label="#" />
          <ElTableColumn prop="name" label="商品名称" min-width="160" />
          <ElTableColumn prop="type" label="类型" width="100" align="center">
            <template #default="{ row }">
              <ElTag :type="row.type === 'redeem' ? 'warning' : 'primary'" size="small" effect="plain">
                {{ row.type === 'redeem' ? '兑换码' : '普通' }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="price" label="价格" width="120" align="right">
            <template #default="{ row }">{{ formatISK(row.price) }}</template>
          </ElTableColumn>
          <ElTableColumn prop="stock" label="库存" width="100" align="center">
            <template #default="{ row }">
              <span v-if="row.stock < 0" class="text-gray-400">无限</span>
              <span v-else :class="row.stock === 0 ? 'text-red-500 font-bold' : ''">{{ row.stock }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="max_per_user" label="限购" width="80" align="center">
            <template #default="{ row }">{{ row.max_per_user > 0 ? row.max_per_user : '不限' }}</template>
          </ElTableColumn>
          <ElTableColumn prop="need_approval" label="需审批" width="80" align="center">
            <template #default="{ row }">
              <ElTag :type="row.need_approval ? 'warning' : 'info'" size="small" effect="plain">
                {{ row.need_approval ? '是' : '否' }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="status" label="状态" width="80" align="center">
            <template #default="{ row }">
              <ElTag :type="row.status === 1 ? 'success' : 'danger'" size="small" effect="plain">
                {{ row.status === 1 ? '上架' : '下架' }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="sort_order" label="排序" width="80" align="center" />
          <ElTableColumn label="操作" width="200" align="center" fixed="right">
            <template #default="{ row }">
              <ElButton size="small" type="primary" @click="showProductDialog(row)">编辑</ElButton>
              <ElButton size="small" type="danger" @click="handleDeleteProduct(row)">删除</ElButton>
            </template>
          </ElTableColumn>
        </ElTable>

        <div v-if="productPagination.total > 0" class="pagination-wrapper">
          <ElPagination
            v-model:current-page="productPagination.current"
            v-model:page-size="productPagination.size"
            :total="productPagination.total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            background
            @size-change="() => { productPagination.current = 1; loadProducts() }"
            @current-change="loadProducts"
          />
        </div>
      </ElTabPane>

      <!-- 订单管理 -->
      <ElTabPane label="订单管理" name="orders">
        <div class="flex items-center gap-4 mb-4 flex-wrap">
          <ElInput v-model="orderFilter.user_id" placeholder="用户 ID" clearable style="width: 140px" />
          <ElSelect v-model="orderFilter.status" placeholder="订单状态" clearable style="width: 140px">
            <ElOption label="待审批" value="pending" />
            <ElOption label="已完成" value="completed" />
            <ElOption label="已拒绝" value="rejected" />
            <ElOption label="余额不足" value="insufficient_funds" />
          </ElSelect>
          <ElButton type="primary" @click="loadOrders">查询</ElButton>
        </div>

        <ElTable v-loading="orderLoading" :data="orderList" stripe border style="width: 100%">
          <ElTableColumn prop="order_no" label="订单号" width="200" />
          <ElTableColumn prop="user_id" label="用户ID" width="100" align="center" />
          <ElTableColumn prop="product_name" label="商品" min-width="140" />
          <ElTableColumn prop="quantity" label="数量" width="70" align="center" />
          <ElTableColumn prop="total_price" label="总价" width="120" align="right">
            <template #default="{ row }">
              <span class="font-medium">{{ formatISK(row.total_price) }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="status" label="状态" width="120" align="center">
            <template #default="{ row }">
              <ElTag :type="orderStatusType(row.status)" size="small" effect="plain">
                {{ orderStatusLabel(row.status) }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="remark" label="用户备注" width="140" show-overflow-tooltip />
          <ElTableColumn prop="review_remark" label="审批备注" width="140" show-overflow-tooltip />
          <ElTableColumn prop="created_at" label="下单时间" width="180">
            <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="200" align="center" fixed="right">
            <template #default="{ row }">
              <template v-if="row.status === 'pending'">
                <ElButton size="small" type="success" @click="handleApprove(row)">通过</ElButton>
                <ElButton size="small" type="danger" @click="handleReject(row)">拒绝</ElButton>
              </template>
              <span v-else class="text-gray-400 text-sm">-</span>
            </template>
          </ElTableColumn>
        </ElTable>

        <div v-if="orderPagination.total > 0" class="pagination-wrapper">
          <ElPagination
            v-model:current-page="orderPagination.current"
            v-model:page-size="orderPagination.size"
            :total="orderPagination.total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            background
            @size-change="() => { orderPagination.current = 1; loadOrders() }"
            @current-change="loadOrders"
          />
        </div>
      </ElTabPane>

      <!-- 兑换码管理 -->
      <ElTabPane label="兑换码管理" name="redeem">
        <div class="flex items-center gap-4 mb-4 flex-wrap">
          <ElInput v-model="redeemFilter.product_id" placeholder="商品 ID" clearable style="width: 140px" />
          <ElSelect v-model="redeemFilter.status" placeholder="状态" clearable style="width: 120px">
            <ElOption label="未使用" value="unused" />
            <ElOption label="已使用" value="used" />
            <ElOption label="已过期" value="expired" />
          </ElSelect>
          <ElButton type="primary" @click="loadRedeemCodes">查询</ElButton>
        </div>

        <ElTable v-loading="redeemLoading" :data="redeemList" stripe border style="width: 100%">
          <ElTableColumn type="index" width="60" label="#" />
          <ElTableColumn prop="product_id" label="商品ID" width="100" align="center" />
          <ElTableColumn prop="user_id" label="用户ID" width="100" align="center" />
          <ElTableColumn prop="order_id" label="订单ID" width="100" align="center" />
          <ElTableColumn prop="code" label="兑换码" width="220">
            <template #default="{ row }">
              <code class="text-sm font-mono">{{ row.code }}</code>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="status" label="状态" width="100" align="center">
            <template #default="{ row }">
              <ElTag :type="row.status === 'unused' ? 'success' : row.status === 'used' ? 'info' : 'danger'" size="small" effect="plain">
                {{ redeemStatusLabel(row.status) }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="created_at" label="创建时间" width="180">
            <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
          </ElTableColumn>
          <ElTableColumn prop="expires_at" label="过期时间" width="180">
            <template #default="{ row }">{{ row.expires_at ? formatTime(row.expires_at) : '-' }}</template>
          </ElTableColumn>
        </ElTable>

        <div v-if="redeemPagination.total > 0" class="pagination-wrapper">
          <ElPagination
            v-model:current-page="redeemPagination.current"
            v-model:page-size="redeemPagination.size"
            :total="redeemPagination.total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            background
            @size-change="() => { redeemPagination.current = 1; loadRedeemCodes() }"
            @current-change="loadRedeemCodes"
          />
        </div>
      </ElTabPane>
    </ElTabs>

    <!-- 商品编辑对话框 -->
    <ElDialog v-model="productDialogVisible" :title="editingProduct ? '编辑商品' : '新增商品'" width="560px" destroy-on-close>
      <ElForm ref="productFormRef" :model="productForm" :rules="productRules" label-width="100px">
        <ElFormItem label="商品名称" prop="name">
          <ElInput v-model="productForm.name" placeholder="请输入商品名称" />
        </ElFormItem>
        <ElFormItem label="描述">
          <ElInput v-model="productForm.description" type="textarea" :rows="3" placeholder="商品描述（可选）" />
        </ElFormItem>
        <ElFormItem label="图片 URL">
          <ElInput v-model="productForm.image" placeholder="商品图片链接（可选）" />
        </ElFormItem>
        <ElFormItem label="价格" prop="price">
          <ElInputNumber v-model="productForm.price" :min="0.01" :precision="2" :step="10" style="width: 200px" />
        </ElFormItem>
        <ElFormItem label="类型" prop="type">
          <ElSelect v-model="productForm.type" style="width: 200px">
            <ElOption label="普通商品" value="normal" />
            <ElOption label="兑换码/服务" value="redeem" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="库存">
          <ElInputNumber v-model="productForm.stock" :min="-1" style="width: 200px" />
          <span class="ml-2 text-xs text-gray-400">-1 = 无限库存</span>
        </ElFormItem>
        <ElFormItem label="限购/人">
          <ElInputNumber v-model="productForm.max_per_user" :min="0" style="width: 200px" />
          <span class="ml-2 text-xs text-gray-400">0 = 不限购</span>
        </ElFormItem>
        <ElFormItem label="需要审批">
          <ElSwitch v-model="productForm.need_approval" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElSelect v-model="productForm.status" style="width: 200px">
            <ElOption label="上架" :value="1" />
            <ElOption label="下架" :value="0" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="排序">
          <ElInputNumber v-model="productForm.sort_order" :min="0" style="width: 200px" />
          <span class="ml-2 text-xs text-gray-400">越大越靠前</span>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="productDialogVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="productSubmitting" @click="submitProduct">确定</ElButton>
      </template>
    </ElDialog>

    <!-- 审批备注对话框 -->
    <ElDialog v-model="reviewDialogVisible" :title="reviewAction === 'approve' ? '审批通过' : '拒绝订单'" width="400px" destroy-on-close>
      <ElForm label-width="80px">
        <ElFormItem label="订单号">
          <span class="font-medium">{{ reviewOrderNo }}</span>
        </ElFormItem>
        <ElFormItem label="审批备注">
          <ElInput v-model="reviewRemark" type="textarea" :rows="3" placeholder="审批备注（可选）" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="reviewDialogVisible = false">取消</ElButton>
        <ElButton :type="reviewAction === 'approve' ? 'success' : 'danger'" :loading="reviewSubmitting" @click="submitReview">
          {{ reviewAction === 'approve' ? '确认通过' : '确认拒绝' }}
        </ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import {
    ElTabs,
    ElTabPane,
    ElTable,
    ElTableColumn,
    ElTag,
    ElButton,
    ElPagination,
    ElInput,
    ElSelect,
    ElOption,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInputNumber,
    ElSwitch,
    ElCard,
    ElMessage,
    ElMessageBox
  } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import {
    adminListProducts,
    adminCreateProduct,
    adminUpdateProduct,
    adminDeleteProduct,
    adminListOrders,
    adminApproveOrder,
    adminRejectOrder,
    adminListRedeemCodes
  } from '@/api/shop'

  defineOptions({ name: 'SystemShop' })

  const activeTab = ref('products')

  // ─── 商品管理 ───
  const products = ref<Api.Shop.Product[]>([])
  const productLoading = ref(false)
  const productFilter = reactive({ name: '', type: '', status: undefined as number | undefined })
  const productPagination = reactive({ current: 1, size: 20, total: 0 })

  async function loadProducts() {
    productLoading.value = true
    try {
      const res = await adminListProducts({
        current: productPagination.current,
        size: productPagination.size,
        name: productFilter.name || undefined,
        type: productFilter.type || undefined,
        status: productFilter.status
      }) as any
      products.value = res.list ?? res.records ?? []
      productPagination.total = res.total ?? 0
    } finally {
      productLoading.value = false
    }
  }

  // 商品编辑对话框
  const productDialogVisible = ref(false)
  const editingProduct = ref<Api.Shop.Product | null>(null)
  const productSubmitting = ref(false)
  const productFormRef = ref<FormInstance>()

  const productForm = reactive({
    name: '',
    description: '',
    image: '',
    price: 0,
    type: 'normal' as 'normal' | 'redeem',
    stock: -1,
    max_per_user: 0,
    need_approval: false,
    status: 1 as number,
    sort_order: 0
  })

  const productRules: FormRules = {
    name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
    price: [{ required: true, message: '请输入价格', trigger: 'blur' }],
    type: [{ required: true, message: '请选择类型', trigger: 'change' }]
  }

  function showProductDialog(product?: Api.Shop.Product) {
    editingProduct.value = product ?? null
    if (product) {
      Object.assign(productForm, {
        name: product.name,
        description: product.description,
        image: product.image,
        price: product.price,
        type: product.type,
        stock: product.stock,
        max_per_user: product.max_per_user,
        need_approval: product.need_approval,
        status: product.status,
        sort_order: product.sort_order
      })
    } else {
      Object.assign(productForm, {
        name: '', description: '', image: '', price: 0,
        type: 'normal', stock: -1, max_per_user: 0,
        need_approval: false, status: 1, sort_order: 0
      })
    }
    productDialogVisible.value = true
  }

  async function submitProduct() {
    try {
      await productFormRef.value?.validate()
    } catch {
      return
    }

    productSubmitting.value = true
    try {
      if (editingProduct.value) {
        await adminUpdateProduct({
          id: editingProduct.value.id,
          ...productForm
        })
        ElMessage.success('更新成功')
      } else {
        await adminCreateProduct(productForm)
        ElMessage.success('创建成功')
      }
      productDialogVisible.value = false
      loadProducts()
    } catch (e: any) {
      ElMessage.error(e?.message || '操作失败')
    } finally {
      productSubmitting.value = false
    }
  }

  async function handleDeleteProduct(product: Api.Shop.Product) {
    try {
      await ElMessageBox.confirm(`确定删除商品「${product.name}」？`, '删除确认', {
        type: 'warning',
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      })
      await adminDeleteProduct(product.id)
      ElMessage.success('删除成功')
      loadProducts()
    } catch { /* cancelled */ }
  }

  // ─── 订单管理 ───
  const orderList = ref<Api.Shop.Order[]>([])
  const orderLoading = ref(false)
  const orderFilter = reactive({ user_id: '', status: '' })
  const orderPagination = reactive({ current: 1, size: 20, total: 0 })

  async function loadOrders() {
    orderLoading.value = true
    try {
      const params: Api.Shop.OrderSearchParams = {
        current: orderPagination.current,
        size: orderPagination.size,
        status: orderFilter.status || undefined
      }
      if (orderFilter.user_id) params.user_id = Number(orderFilter.user_id)
      const res = await adminListOrders(params) as any
      orderList.value = res.list ?? res.records ?? []
      orderPagination.total = res.total ?? 0
    } finally {
      orderLoading.value = false
    }
  }

  // 审批对话框
  const reviewDialogVisible = ref(false)
  const reviewAction = ref<'approve' | 'reject'>('approve')
  const reviewOrderId = ref(0)
  const reviewOrderNo = ref('')
  const reviewRemark = ref('')
  const reviewSubmitting = ref(false)

  function handleApprove(order: Api.Shop.Order) {
    reviewAction.value = 'approve'
    reviewOrderId.value = order.id
    reviewOrderNo.value = order.order_no
    reviewRemark.value = ''
    reviewDialogVisible.value = true
  }

  function handleReject(order: Api.Shop.Order) {
    reviewAction.value = 'reject'
    reviewOrderId.value = order.id
    reviewOrderNo.value = order.order_no
    reviewRemark.value = ''
    reviewDialogVisible.value = true
  }

  async function submitReview() {
    reviewSubmitting.value = true
    try {
      const params: Api.Shop.OrderReviewParams = {
        order_id: reviewOrderId.value,
        remark: reviewRemark.value
      }
      if (reviewAction.value === 'approve') {
        await adminApproveOrder(params)
        ElMessage.success('审批通过')
      } else {
        await adminRejectOrder(params)
        ElMessage.success('已拒绝')
      }
      reviewDialogVisible.value = false
      loadOrders()
    } catch (e: any) {
      ElMessage.error(e?.message || '操作失败')
    } finally {
      reviewSubmitting.value = false
    }
  }

  // ─── 兑换码管理 ───
  const redeemList = ref<Api.Shop.RedeemCode[]>([])
  const redeemLoading = ref(false)
  const redeemFilter = reactive({ product_id: '', status: '' })
  const redeemPagination = reactive({ current: 1, size: 20, total: 0 })

  async function loadRedeemCodes() {
    redeemLoading.value = true
    try {
      const params: Api.Shop.RedeemSearchParams = {
        current: redeemPagination.current,
        size: redeemPagination.size,
        status: redeemFilter.status || undefined
      }
      if (redeemFilter.product_id) params.product_id = Number(redeemFilter.product_id)
      const res = await adminListRedeemCodes(params) as any
      redeemList.value = res.list ?? res.records ?? []
      redeemPagination.total = res.total ?? 0
    } finally {
      redeemLoading.value = false
    }
  }

  // ─── 工具函数 ───
  function formatISK(v: number) {
    return v.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
  }

  function formatTime(t: string) {
    return new Date(t).toLocaleString()
  }

  const orderStatusMap: Record<string, { label: string; type: string }> = {
    pending: { label: '待审批', type: 'warning' },
    paid: { label: '已付款', type: 'success' },
    approved: { label: '已审批', type: 'success' },
    rejected: { label: '已拒绝', type: 'danger' },
    completed: { label: '已完成', type: 'success' },
    cancelled: { label: '已取消', type: 'info' },
    insufficient_funds: { label: '余额不足', type: 'danger' }
  }

  function orderStatusLabel(status: string) {
    return orderStatusMap[status]?.label ?? status
  }

  function orderStatusType(status: string) {
    return (orderStatusMap[status]?.type ?? 'info') as any
  }

  function redeemStatusLabel(status: string) {
    const map: Record<string, string> = { unused: '未使用', used: '已使用', expired: '已过期' }
    return map[status] ?? status
  }

  // ─── 初始化 ───
  onMounted(() => {
    loadProducts()
  })

  watch(activeTab, (tab) => {
    if (tab === 'orders' && orderList.value.length === 0) loadOrders()
    if (tab === 'redeem' && redeemList.value.length === 0) loadRedeemCodes()
  })
</script>

<style scoped>
  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
    padding: 8px 0;
  }
</style>
