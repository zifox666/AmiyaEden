<!-- 商店页面（用户端） -->
<template>
  <div class="shop-page art-full-height">
    <ElTabs v-model="activeTab" class="art-table-card p-4">
      <!-- 商品列表 -->
      <ElTabPane :label="$t('shop.products')" name="products">
        <ShopProducts ref="productsRef" :balance="walletBalance" @buy="openBuyDialog" />
      </ElTabPane>

      <!-- 抽奖 -->
      <ElTabPane label="抽奖" name="lottery">
        <ShopLottery ref="lotteryRef" />
      </ElTabPane>

      <!-- 我的订单 -->
      <ElTabPane :label="$t('shop.myOrders')" name="orders">
        <ShopOrders ref="ordersRef" />
      </ElTabPane>
    </ElTabs>

    <!-- 购买对话框 -->
    <ElDialog
      v-model="buyDialogVisible"
      :title="$t('shop.buyTitle')"
      width="420px"
      destroy-on-close
    >
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
          <span class="text-red-500 font-bold text-lg">{{
            formatISK(buyProduct.price * buyQuantity)
          }}</span>
        </ElFormItem>
        <ElFormItem :label="$t('shop.remark')">
          <ElInput
            v-model="buyRemark"
            type="textarea"
            :rows="2"
            :placeholder="$t('shop.remarkPlaceholder')"
          />
        </ElFormItem>
        <div v-if="buyProduct.need_approval" class="text-xs text-orange-500 mb-2">
          <el-icon><Warning /></el-icon> {{ $t('shop.approvalNotice') }}
        </div>
      </ElForm>
      <template #footer>
        <ElButton @click="buyDialogVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="buyLoading" @click="confirmBuy">{{
          $t('shop.confirmBuy')
        }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { Warning } from '@element-plus/icons-vue'
  import {
    ElTabs,
    ElTabPane,
    ElDialog,
    ElForm,
    ElFormItem,
    ElInputNumber,
    ElInput,
    ElButton,
    ElMessage
  } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { buyProduct as apiBuyProduct } from '@/api/shop'
  import { fetchMyWallet } from '@/api/sys-wallet'
  import ShopProducts from './modules/shop-products.vue'
  import ShopLottery from './modules/shop-lottery.vue'
  import ShopOrders from './modules/shop-orders.vue'

  defineOptions({ name: 'Shop' })
  const { t } = useI18n()

  // ─── Tab ───
  const activeTab = ref('products')

  // ─── 子面板 refs ───
  const productsRef = ref<InstanceType<typeof ShopProducts>>()
  const lotteryRef = ref<InstanceType<typeof ShopLottery>>()
  const ordersRef = ref<InstanceType<typeof ShopOrders>>()

  // ─── 钱包余额 ───
  const walletBalance = ref<number | null>(null)

  async function loadWallet() {
    try {
      const res = (await fetchMyWallet()) as any
      walletBalance.value = res?.balance ?? null
    } catch {
      /* ignore */
    }
  }

  // ─── 购买对话框 ───
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

  const formatISK = (v: number) =>
    v.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

  async function confirmBuy() {
    if (!buyProduct.value) return
    buyLoading.value = true
    try {
      await apiBuyProduct({
        product_id: buyProduct.value.id,
        quantity: buyQuantity.value,
        remark: buyRemark.value
      })
      ElMessage.success(t('shopBrowse.purchaseSuccess'))
      buyDialogVisible.value = false
      productsRef.value?.refresh()
      loadWallet()
    } catch (e: any) {
      ElMessage.error(e?.message || t('shopBrowse.purchaseFailed'))
    } finally {
      buyLoading.value = false
    }
  }

  // ─── Tab 切换懒加载 ───
  watch(activeTab, (tab) => {
    if (tab === 'lottery') lotteryRef.value?.load()
    if (tab === 'orders') ordersRef.value?.load()
  })

  // ─── 初始化 ───
  onMounted(() => {
    loadWallet()
  })
</script>

<style scoped lang="scss">
  .shop-page {
    // 让 ElTabs 的内容区填满剩余高度，使子面板能继承高度
    :deep(.el-tabs__content) {
      flex: 1;
      overflow: hidden;
    }

    // tab-pane 撑满内容区，成为 flex 列容器
    :deep(.el-tab-pane) {
      height: 100%;
      display: flex;
      flex-direction: column;

      // tab pane 内的卡片无需全局 margin-top
      .art-table-card {
        margin-top: 0;
      }
    }
  }
</style>
