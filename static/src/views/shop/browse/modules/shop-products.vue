<!-- 商品列表面板 -->
<template>
  <div class="shop-products">
    <!-- 筛选栏 -->
    <div class="filter-bar mb-4 flex items-center gap-3">
      <ElSelect
        v-model="typeFilter"
        :placeholder="$t('shop.allTypes')"
        clearable
        style="width: 140px"
        @change="handleTypeChange"
      >
        <ElOption :label="$t('shop.typeNormal')" value="normal" />
        <ElOption :label="$t('shop.typeRedeem')" value="redeem" />
      </ElSelect>
      <ElButton :loading="loading" @click="refreshData">
        <el-icon class="mr-1"><Refresh /></el-icon>
        {{ $t('common.refresh') }}
      </ElButton>
      <div class="ml-auto text-sm text-gray-500">
        {{ $t('shop.myBalance') }}:
        <span
          class="font-bold text-lg"
          :class="balance !== null && balance > 0 ? 'text-green-600' : 'text-red-500'"
        >
          {{ balance !== null ? formatISK(balance) : '-' }}
        </span>
      </div>
    </div>

    <!-- 商品卡片网格 -->
    <div v-loading="loading" class="product-grid">
      <ElCard v-for="item in data" :key="item.id" shadow="hover" class="product-card">
        <div v-if="item.image" class="product-image">
          <img :src="item.image" :alt="item.name" @error="onImgError" />
        </div>
        <div v-else class="product-image placeholder">
          <el-icon :size="48"><ShoppingBag /></el-icon>
        </div>
        <div class="product-info">
          <h3 class="product-name">{{ item.name }}</h3>
          <p v-if="item.description" class="product-desc">{{ item.description }}</p>
          <div class="product-meta">
            <div class="price">{{ formatISK(item.price) }}</div>
            <div class="stock text-xs text-gray-400">
              <template v-if="item.stock < 0">{{ $t('shop.unlimitedStock') }}</template>
              <template v-else>{{ $t('shop.stockRemaining', { n: item.stock }) }}</template>
            </div>
            <div v-if="item.max_per_user > 0" class="limit text-xs text-gray-400">
              {{ $t('shop.limitPerUser', { n: item.max_per_user }) }}
              <span v-if="item.limit_period && item.limit_period !== 'forever'">
                ({{ $t('shop.period.' + item.limit_period) }})
              </span>
            </div>
          </div>
          <div class="product-tags mt-2">
            <ElTag v-if="item.type === 'redeem'" size="small" type="warning" effect="plain">
              {{ $t('shop.typeRedeem') }}
            </ElTag>
            <ElTag v-if="item.need_approval" size="small" type="info" effect="plain">
              {{ $t('shop.needApproval') }}
            </ElTag>
          </div>
          <ElButton
            class="mt-3"
            type="primary"
            style="width: 100%"
            :disabled="item.stock === 0"
            @click="emit('buy', item)"
          >
            {{ item.stock === 0 ? $t('shop.soldOut') : $t('shop.buy') }}
          </ElButton>
        </div>
      </ElCard>
    </div>

    <ElEmpty v-if="!loading && data.length === 0" :description="$t('shop.noProducts')" />

    <div v-if="pagination.total > 0" class="pagination-wrapper">
      <ElPagination
        :current-page="pagination.current"
        :page-size="pagination.size"
        :total="pagination.total"
        :page-sizes="[12, 24, 36, 48]"
        layout="total, sizes, prev, pager, next"
        background
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { Refresh, ShoppingBag } from '@element-plus/icons-vue'
  import { ElCard, ElTag, ElButton, ElSelect, ElOption, ElPagination, ElEmpty } from 'element-plus'
  import { fetchProducts } from '@/api/shop'
  import { useTable } from '@/hooks/core/useTable'

  defineOptions({ name: 'ShopProducts' })

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const props = defineProps<{ balance: number | null }>()
  const emit = defineEmits<{
    (e: 'buy', product: Api.Shop.Product): void
    (e: 'refreshed'): void
  }>()

  const typeFilter = ref<string | undefined>(undefined)

  const formatISK = (v: number) =>
    v.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })

  /** 图片加载失败时尝试回退：render → icon */
  function onImgError(e: Event) {
    const img = e.target as HTMLImageElement
    if (img.src.includes('/render')) {
      img.src = img.src.replace('/render', '/icon')
    }
  }

  const {
    data,
    loading,
    pagination,
    searchParams,
    getData,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchProducts,
      apiParams: { current: 1, size: 12, type: undefined as string | undefined }
    }
  })

  function handleTypeChange() {
    searchParams.type = typeFilter.value || undefined
    getData()
  }

  // 供父页面调用（购买成功后刷新）
  defineExpose({ refresh: refreshData })
</script>

<style scoped>
  .product-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
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
    aspect-ratio: 1;
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
    object-fit: contain;
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
