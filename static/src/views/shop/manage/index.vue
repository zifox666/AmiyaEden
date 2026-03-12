<!-- 商店管理页面（管理员） -->
<template>
  <div class="shop-admin-page art-full-height">
    <ElTabs v-model="activeTab" class="art-table-card p-4">
      <!-- 商品管理 -->
      <ElTabPane :label="$t('shopAdmin.tabs.products')" name="products">
        <ManageProducts ref="productsRef" />
      </ElTabPane>

      <!-- 抽奖管理 -->
      <ElTabPane label="抽奖管理" name="lottery">
        <ManageLottery ref="lotteryRef" />
      </ElTabPane>

      <!-- 订单管理 -->
      <ElTabPane :label="$t('shopAdmin.tabs.orders')" name="orders">
        <ManageOrders ref="ordersRef" />
      </ElTabPane>
    </ElTabs>
  </div>
</template>

<script setup lang="ts">
  import ManageProducts from './modules/manage-products.vue'
  import ManageLottery from './modules/manage-lottery.vue'
  import ManageOrders from './modules/manage-orders.vue'

  defineOptions({ name: 'SystemShop' })

  const activeTab = ref('products')

  const productsRef = ref<InstanceType<typeof ManageProducts>>()
  const lotteryRef = ref<InstanceType<typeof ManageLottery>>()
  const ordersRef = ref<InstanceType<typeof ManageOrders>>()

  // Tab 切换懒加载
  watch(activeTab, (tab) => {
    if (tab === 'lottery') lotteryRef.value?.load()
    if (tab === 'orders') ordersRef.value?.load()
  })

  // 初始化加载商品列表
  onMounted(() => {
    productsRef.value?.load()
  })
</script>

<style scoped lang="scss">
  .shop-admin-page {
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
