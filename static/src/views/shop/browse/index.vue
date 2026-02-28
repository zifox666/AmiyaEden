<!-- 商店页面（用户端） -->
<template>
  <div class="shop-page art-full-height">
    <ElTabs v-model="activeTab">
      <!-- 商品列表 -->
      <ElTabPane :label="$t('shop.products')" name="products">
        <div class="shop-products">
          <!-- 筛选 -->
          <div class="filter-bar mb-4 flex items-center gap-3">
            <ElSelect v-model="productTypeFilter" :placeholder="$t('shop.allTypes')" clearable style="width: 140px" @change="loadProducts">
              <ElOption :label="$t('shop.typeNormal')" value="normal" />
              <ElOption :label="$t('shop.typeRedeem')" value="redeem" />
            </ElSelect>
            <ElButton :loading="productsLoading" @click="loadProducts">
              <el-icon class="mr-1"><Refresh /></el-icon>
              {{ $t('common.refresh') }}
            </ElButton>
            <div class="ml-auto text-sm text-gray-500">
              {{ $t('shop.myBalance') }}:
              <span class="font-bold text-lg" :class="wallet && wallet.balance > 0 ? 'text-green-600' : 'text-red-500'">
                {{ wallet ? formatISK(wallet.balance) : '-' }}
              </span>
            </div>
          </div>

          <!-- 商品卡片网格 -->
          <div v-loading="productsLoading" class="product-grid">
            <ElCard v-for="item in products" :key="item.id" shadow="hover" class="product-card">
              <div class="product-image" v-if="item.image">
                <img :src="item.image" :alt="item.name" />
              </div>
              <div class="product-image placeholder" v-else>
                <el-icon :size="48"><ShoppingBag /></el-icon>
              </div>
              <div class="product-info">
                <h3 class="product-name">{{ item.name }}</h3>
                <p class="product-desc" v-if="item.description">{{ item.description }}</p>
                <div class="product-meta">
                  <div class="price">{{ formatISK(item.price) }}</div>
                  <div class="stock text-xs text-gray-400">
                    <template v-if="item.stock < 0">{{ $t('shop.unlimitedStock') }}</template>
                    <template v-else>{{ $t('shop.stockRemaining', { n: item.stock }) }}</template>
                  </div>
                  <div v-if="item.max_per_user > 0" class="limit text-xs text-gray-400">
                    {{ $t('shop.limitPerUser', { n: item.max_per_user }) }}
                  </div>
                </div>
                <div class="product-tags mt-2">
                  <ElTag v-if="item.type === 'redeem'" size="small" type="warning" effect="plain">{{ $t('shop.typeRedeem') }}</ElTag>
                  <ElTag v-if="item.need_approval" size="small" type="info" effect="plain">{{ $t('shop.needApproval') }}</ElTag>
                </div>
                <ElButton class="mt-3" type="primary" style="width: 100%" :disabled="item.stock === 0" @click="openBuyDialog(item)">
                  {{ item.stock === 0 ? $t('shop.soldOut') : $t('shop.buy') }}
                </ElButton>
              </div>
            </ElCard>
          </div>

          <ElEmpty v-if="!productsLoading && products.length === 0" :description="$t('shop.noProducts')" />

          <div v-if="productPagination.total > 0" class="pagination-wrapper">
            <ElPagination
              v-model:current-page="productPagination.current"
              v-model:page-size="productPagination.size"
              :total="productPagination.total"
              :page-sizes="[12, 24, 48]"
              layout="total, sizes, prev, pager, next"
              background
              @size-change="(s: number) => { productPagination.size = s; loadProducts() }"
              @current-change="(p: number) => { productPagination.current = p; loadProducts() }"
            />
          </div>
        </div>
      </ElTabPane>

      <!-- 我的订单 -->
      <ElTabPane :label="$t('shop.myOrders')" name="orders">
        <div class="filter-bar mb-4 flex items-center gap-3">
          <ElSelect v-model="orderStatusFilter" :placeholder="$t('shop.allStatuses')" clearable style="width: 160px" @change="loadOrders">
            <ElOption label="待审批" value="pending" />
            <ElOption label="已完成" value="completed" />
            <ElOption label="已拒绝" value="rejected" />
            <ElOption label="已取消" value="cancelled" />
          </ElSelect>
          <ElButton :loading="ordersLoading" @click="loadOrders">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
        </div>

        <ElTable v-loading="ordersLoading" :data="orders" stripe border style="width: 100%">
          <ElTableColumn prop="order_no" :label="$t('shop.orderNo')" width="200" />
          <ElTableColumn prop="product_name" :label="$t('shop.productName')" min-width="160" />
          <ElTableColumn prop="quantity" :label="$t('shop.quantity')" width="80" align="center" />
          <ElTableColumn prop="unit_price" :label="$t('shop.unitPrice')" width="120" align="right">
            <template #default="{ row }">{{ formatISK(row.unit_price) }}</template>
          </ElTableColumn>
          <ElTableColumn prop="total_price" :label="$t('shop.totalPrice')" width="120" align="right">
            <template #default="{ row }">
              <span class="font-medium text-red-500">{{ formatISK(row.total_price) }}</span>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="status" :label="$t('shop.status')" width="120" align="center">
            <template #default="{ row }">
              <ElTag :type="orderStatusType(row.status)" size="small" effect="plain">{{ orderStatusLabel(row.status) }}</ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="created_at" :label="$t('shop.orderTime')" width="180">
            <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
          </ElTableColumn>
        </ElTable>

        <ElEmpty v-if="!ordersLoading && orders.length === 0" :description="$t('shop.noOrders')" />

        <div v-if="orderPagination.total > 0" class="pagination-wrapper">
          <ElPagination
            v-model:current-page="orderPagination.current"
            v-model:page-size="orderPagination.size"
            :total="orderPagination.total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            background
            @size-change="(s: number) => { orderPagination.size = s; loadOrders() }"
            @current-change="(p: number) => { orderPagination.current = p; loadOrders() }"
          />
        </div>
      </ElTabPane>

      <!-- 我的兑换码 -->
      <ElTabPane :label="$t('shop.myRedeemCodes')" name="redeem">
        <ElButton class="mb-4" :loading="redeemLoading" @click="loadRedeemCodes">
          <el-icon class="mr-1"><Refresh /></el-icon>
          {{ $t('common.refresh') }}
        </ElButton>

        <ElTable v-loading="redeemLoading" :data="redeemCodes" stripe border style="width: 100%">
          <ElTableColumn type="index" width="60" label="#" />
          <ElTableColumn prop="code" :label="$t('shop.redeemCode')" width="220">
            <template #default="{ row }">
              <code class="text-sm font-mono">{{ row.code }}</code>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="status" :label="$t('shop.status')" width="100" align="center">
            <template #default="{ row }">
              <ElTag :type="row.status === 'unused' ? 'success' : row.status === 'used' ? 'info' : 'danger'" size="small" effect="plain">
                {{ redeemStatusLabel(row.status) }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="created_at" :label="$t('shop.orderTime')" width="180">
            <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
          </ElTableColumn>
          <ElTableColumn prop="expires_at" :label="$t('shop.expiresAt')" width="180">
            <template #default="{ row }">{{ row.expires_at ? formatTime(row.expires_at) : '-' }}</template>
          </ElTableColumn>
        </ElTable>

        <ElEmpty v-if="!redeemLoading && redeemCodes.length === 0" :description="$t('shop.noRedeemCodes')" />

        <div v-if="redeemPagination.total > 0" class="pagination-wrapper">
          <ElPagination
            v-model:current-page="redeemPagination.current"
            v-model:page-size="redeemPagination.size"
            :total="redeemPagination.total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            background
            @size-change="(s: number) => { redeemPagination.size = s; loadRedeemCodes() }"
            @current-change="(p: number) => { redeemPagination.current = p; loadRedeemCodes() }"
          />
        </div>
      </ElTabPane>
    </ElTabs>

    <!-- 购买对话框 -->
    <ElDialog v-model="buyDialogVisible" :title="$t('shop.buyTitle')" width="420px" destroy-on-close>
      <ElForm v-if="buyProduct" label-width="80px">
        <ElFormItem :label="$t('shop.productName')">
          <span class="font-medium">{{ buyProduct.name }}</span>
        </ElFormItem>
        <ElFormItem :label="$t('shop.unitPrice')">
          <span class="text-orange-600 font-medium">{{ formatISK(buyProduct.price) }}</span>
        </ElFormItem>
        <ElFormItem :label="$t('shop.quantity')">
          <ElInputNumber v-model="buyQuantity" :min="1" :max="buyMaxQty" style="width: 160px" />
        </ElFormItem>
        <ElFormItem :label="$t('shop.totalPrice')">
          <span class="text-red-500 font-bold text-lg">{{ formatISK(buyProduct.price * buyQuantity) }}</span>
        </ElFormItem>
        <ElFormItem :label="$t('shop.remark')">
          <ElInput v-model="buyRemark" type="textarea" :rows="2" :placeholder="$t('shop.remarkPlaceholder')" />
        </ElFormItem>
        <div v-if="buyProduct.need_approval" class="text-xs text-orange-500 mb-2">
          <el-icon><Warning /></el-icon> {{ $t('shop.approvalNotice') }}
        </div>
      </ElForm>
      <template #footer>
        <ElButton @click="buyDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="buyLoading" @click="confirmBuy">{{ $t('shop.confirmBuy') }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { Refresh, ShoppingBag, Warning } from '@element-plus/icons-vue'
  import {
    ElTabs,
    ElTabPane,
    ElCard,
    ElTable,
    ElTableColumn,
    ElTag,
    ElButton,
    ElPagination,
    ElEmpty,
    ElSelect,
    ElOption,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInputNumber,
    ElInput,
    ElMessage
  } from 'element-plus'
  import { fetchProducts, buyProduct as apiBuyProduct, fetchMyOrders, fetchMyRedeemCodes } from '@/api/shop'
  import { fetchMyWallet } from '@/api/sys-wallet'

  defineOptions({ name: 'Shop' })

  // ---- Tab ----
  const activeTab = ref('products')

  // ---- 钱包余额 ----
  const wallet = ref<Api.SysWallet.Wallet | null>(null)

  async function loadWallet() {
    try {
      const res = await fetchMyWallet() as any
      wallet.value = res
    } catch { /* ignore */ }
  }

  // ---- 商品列表 ----
  const products = ref<Api.Shop.Product[]>([])
  const productsLoading = ref(false)
  const productTypeFilter = ref('')
  const productPagination = reactive({ current: 1, size: 12, total: 0 })

  async function loadProducts() {
    productsLoading.value = true
    try {
      const res = await fetchProducts({
        current: productPagination.current,
        size: productPagination.size,
        type: productTypeFilter.value || undefined
      }) as any
      products.value = res.list ?? res.records ?? []
      productPagination.total = res.total ?? 0
    } finally {
      productsLoading.value = false
    }
  }

  // ---- 我的订单 ----
  const orders = ref<Api.Shop.Order[]>([])
  const ordersLoading = ref(false)
  const orderStatusFilter = ref('')
  const orderPagination = reactive({ current: 1, size: 20, total: 0 })

  async function loadOrders() {
    ordersLoading.value = true
    try {
      const res = await fetchMyOrders({
        current: orderPagination.current,
        size: orderPagination.size,
        status: orderStatusFilter.value || undefined
      }) as any
      orders.value = res.list ?? res.records ?? []
      orderPagination.total = res.total ?? 0
    } finally {
      ordersLoading.value = false
    }
  }

  // ---- 兑换码 ----
  const redeemCodes = ref<Api.Shop.RedeemCode[]>([])
  const redeemLoading = ref(false)
  const redeemPagination = reactive({ current: 1, size: 20, total: 0 })

  async function loadRedeemCodes() {
    redeemLoading.value = true
    try {
      const res = await fetchMyRedeemCodes({
        current: redeemPagination.current,
        size: redeemPagination.size
      }) as any
      redeemCodes.value = res.list ?? res.records ?? []
      redeemPagination.total = res.total ?? 0
    } finally {
      redeemLoading.value = false
    }
  }

  // ---- 购买对话框 ----
  const buyDialogVisible = ref(false)
  const buyProductRef = ref<Api.Shop.Product | null>(null)
  const buyProduct = computed(() => buyProductRef.value)
  const buyQuantity = ref(1)
  const buyRemark = ref('')
  const buyLoading = ref(false)

  const buyMaxQty = computed(() => {
    if (!buyProduct.value) return 1
    const stock = buyProduct.value.stock
    const limit = buyProduct.value.max_per_user
    if (stock < 0 && limit <= 0) return 999
    if (stock < 0) return limit
    if (limit <= 0) return stock
    return Math.min(stock, limit)
  })

  function openBuyDialog(product: Api.Shop.Product) {
    buyProductRef.value = product
    buyQuantity.value = 1
    buyRemark.value = ''
    buyDialogVisible.value = true
  }

  async function confirmBuy() {
    if (!buyProduct.value) return
    buyLoading.value = true
    try {
      await apiBuyProduct({
        product_id: buyProduct.value.id,
        quantity: buyQuantity.value,
        remark: buyRemark.value
      })
      ElMessage.success('购买成功')
      buyDialogVisible.value = false
      loadProducts()
      loadWallet()
    } catch (e: any) {
      ElMessage.error(e?.message || '购买失败')
    } finally {
      buyLoading.value = false
    }
  }

  // ---- 工具函数 ----
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

  // ---- 初始化 ----
  onMounted(() => {
    loadWallet()
    loadProducts()
  })

  watch(activeTab, (tab) => {
    if (tab === 'orders' && orders.value.length === 0) loadOrders()
    if (tab === 'redeem' && redeemCodes.value.length === 0) loadRedeemCodes()
  })
</script>

<style scoped>
  .product-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 16px;
  }

  .product-card {
    display: flex;
    flex-direction: column;
  }

  .product-card :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    padding: 0;
  }

  .product-image {
    width: 100%;
    height: 160px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border-bottom: 1px solid var(--el-border-color-lighter);
    background: var(--el-fill-color-light);
  }

  .product-image img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .product-image.placeholder {
    color: var(--el-text-color-placeholder);
  }

  .product-info {
    padding: 16px;
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .product-name {
    font-size: 16px;
    font-weight: 600;
    margin: 0 0 6px;
    line-height: 1.4;
  }

  .product-desc {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    margin: 0 0 8px;
    line-clamp: 2;
    -webkit-line-clamp: 2;
    display: -webkit-box;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .product-meta {
    margin-top: auto;
  }

  .product-meta .price {
    font-size: 20px;
    font-weight: 700;
    color: var(--el-color-warning);
    margin-bottom: 4px;
  }

  .pagination-wrapper {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
    padding: 8px 0;
  }
</style>
