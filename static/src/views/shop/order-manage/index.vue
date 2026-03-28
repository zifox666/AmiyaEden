<!-- 商店订单管理页面（管理员 / 福利官） -->
<template>
  <div class="shop-order-manage-page art-full-height">
    <ElTabs v-model="activeTab" class="art-table-card p-4">
      <ElTabPane :label="$t('shopAdmin.tabs.orders')" name="orders">
        <ElAlert type="success" :closable="false" class="mb-4" show-icon>
          <p>{{ $t('shopAdmin.tabs.ordersSubtitle') }}</p>
        </ElAlert>
        <ManageOrders ref="ordersRef" />
      </ElTabPane>

      <ElTabPane :label="$t('shopAdmin.tabs.orderHistory')" name="orderHistory">
        <ManageOrderHistory ref="orderHistoryRef" />
      </ElTabPane>
    </ElTabs>
  </div>
</template>

<script setup lang="ts">
  import ManageOrders from '../manage/modules/manage-orders.vue'
  import ManageOrderHistory from '../manage/modules/manage-order-history.vue'

  defineOptions({ name: 'ShopOrderManage' })

  const activeTab = ref('orders')

  const ordersRef = ref<InstanceType<typeof ManageOrders>>()
  const orderHistoryRef = ref<InstanceType<typeof ManageOrderHistory>>()

  watch(activeTab, (tab) => {
    if (tab === 'orders') ordersRef.value?.load()
    if (tab === 'orderHistory') orderHistoryRef.value?.load()
  })

  onMounted(() => {
    ordersRef.value?.load()
  })
</script>

<style scoped lang="scss">
  .shop-order-manage-page {
    :deep(.el-tabs__content) {
      flex: 1;
      overflow: hidden;
    }

    :deep(.el-tab-pane) {
      height: 100%;
      display: flex;
      flex-direction: column;

      .art-table-card {
        margin-top: 0;
      }
    }
  }
</style>
