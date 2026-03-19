<template>
  <div class="info-assets-page art-full-height">
    <!-- 统计栏 -->
    <ElCard shadow="never" class="mb-2">
      <div class="flex items-center justify-between flex-wrap gap-4">
        <div class="flex items-center gap-4">
          <ElButton :loading="loading" size="small" @click="loadData">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
        </div>
        <div v-if="assetsData" class="flex items-center gap-4 text-sm text-gray-500">
          <span>
            {{ $t('info.assetCount') }}:
            <strong class="text-blue-500">{{ assetsData.total_items }}</strong>
          </span>
        </div>
      </div>
    </ElCard>

    <!-- 主体区域 -->
    <div v-loading="loading" class="assets-main">
      <!-- 搜索栏 -->
      <div class="filter-bar">
        <ElInput
          v-model="searchKeyword"
          :placeholder="$t('info.searchAsset')"
          clearable
          style="width: 280px"
          size="small"
          :prefix-icon="Search"
        />
      </div>

      <!-- 按位置分组 -->
      <div v-if="filteredLocations.length > 0" class="assets-groups">
        <div v-for="loc in filteredLocations" :key="loc.location_id" class="location-section">
          <!-- 位置标题 -->
          <div class="location-header" @click="toggleLocation(loc.location_id)">
            <span class="mg-arrow" :class="{ expanded: !collapsedLocations.has(loc.location_id) }"
              >▶</span
            >
            <span class="mg-title">{{ loc.location_name }}</span>
            <span class="mg-count">{{ countItems(loc.items) }}</span>
          </div>

          <!-- 物品列表 -->
          <div v-if="!collapsedLocations.has(loc.location_id)" class="asset-items">
            <template v-for="item in loc.items" :key="item.item_id">
              <div
                class="asset-item"
                @click="item.children?.length ? toggleItem(item.item_id) : undefined"
              >
                <span
                  v-if="item.children?.length"
                  class="mg-arrow child-toggle"
                  :class="{ expanded: expandedItems.has(item.item_id) }"
                  >▶</span
                >
                <img
                  :src="getItemIcon(item)"
                  :alt="item.type_name"
                  class="asset-icon"
                  loading="lazy"
                />
                <div class="asset-info">
                  <span class="asset-type-name">{{ item.type_name }}</span>
                  <span v-if="item.asset_name" class="asset-name-tag">{{ item.asset_name }}</span>
                </div>
                <span class="asset-group">{{ item.group_name }}</span>
                <span class="asset-qty">
                  {{ item.quantity > 1 ? `x${item.quantity}` : '' }}
                </span>
                <span class="asset-owner">{{ item.character_name }}</span>
              </div>
              <!-- 子物品 -->
              <div
                v-if="item.children?.length && expandedItems.has(item.item_id)"
                class="child-items"
              >
                <div v-for="child in item.children" :key="child.item_id" class="asset-item child">
                  <img
                    :src="getItemIcon(child)"
                    :alt="child.type_name"
                    class="asset-icon"
                    loading="lazy"
                  />
                  <div class="asset-info">
                    <span class="asset-type-name">{{ child.type_name }}</span>
                    <span v-if="child.asset_name" class="asset-name-tag">{{
                      child.asset_name
                    }}</span>
                  </div>
                  <span class="asset-group">{{ child.group_name }}</span>
                  <span class="asset-qty">
                    {{ child.quantity > 1 ? `x${child.quantity}` : '' }}
                  </span>
                  <span class="asset-owner">{{ child.character_name }}</span>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>

      <ElEmpty v-else-if="!loading" :description="$t('info.noAssetData')" :image-size="60" />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { Refresh, Search } from '@element-plus/icons-vue'
  import { ElCard, ElButton, ElEmpty, ElInput } from 'element-plus'
  import { fetchInfoAssets } from '@/api/eve-info'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'EveInfoAssets' })

  const userStore = useUserStore()

  const assetsData = ref<Api.EveInfo.AssetsResponse | null>(null)
  const loading = ref(false)
  const searchKeyword = ref('')
  const collapsedLocations = ref(new Set<number>())
  const expandedItems = ref(new Set<number>())

  /** 蓝图拷贝 categoryID=9 */
  const CATEGORY_BLUEPRINT = 9

  /** 获取物品图标 URL */
  const getItemIcon = (item: Api.EveInfo.AssetItemNode) => {
    if (item.category_id === CATEGORY_BLUEPRINT) {
      const suffix = item.is_blueprint_copy ? 'bpc' : 'bp'
      return `https://images.evetech.net/types/${item.type_id}/${suffix}?size=32`
    }
    return `https://images.evetech.net/types/${item.type_id}/icon?size=32`
  }

  /** 递归统计物品数（含子物品） */
  const countItems = (items: Api.EveInfo.AssetItemNode[]) => {
    let count = items.length
    for (const item of items) {
      if (item.children) count += item.children.length
    }
    return count
  }

  /** 递归匹配搜索 */
  const matchSearch = (item: Api.EveInfo.AssetItemNode, kw: string): boolean => {
    if (item.type_name?.toLowerCase().includes(kw)) return true
    if (item.group_name?.toLowerCase().includes(kw)) return true
    if (item.asset_name?.toLowerCase().includes(kw)) return true
    if (item.children?.some((c) => matchSearch(c, kw))) return true
    return false
  }

  /** 搜索过滤后的位置列表 */
  const filteredLocations = computed(() => {
    if (!assetsData.value?.locations) return []
    const kw = searchKeyword.value.toLowerCase().trim()
    if (!kw) return assetsData.value.locations

    const result: Api.EveInfo.AssetLocationNode[] = []
    for (const loc of assetsData.value.locations) {
      // 位置名匹配则显示全部
      if (loc.location_name?.toLowerCase().includes(kw)) {
        result.push(loc)
        continue
      }
      // 按物品筛选
      const matchedItems = loc.items.filter((item) => matchSearch(item, kw))
      if (matchedItems.length > 0) {
        result.push({ ...loc, items: matchedItems })
      }
    }
    return result
  })

  const toggleLocation = (id: number) => {
    if (collapsedLocations.value.has(id)) {
      collapsedLocations.value.delete(id)
    } else {
      collapsedLocations.value.add(id)
    }
    collapsedLocations.value = new Set(collapsedLocations.value)
  }

  const toggleItem = (id: number) => {
    if (expandedItems.value.has(id)) {
      expandedItems.value.delete(id)
    } else {
      expandedItems.value.add(id)
    }
    expandedItems.value = new Set(expandedItems.value)
  }

  const loadData = async () => {
    loading.value = true
    try {
      assetsData.value = await fetchInfoAssets({
        language: userStore.language
      })
    } catch {
      assetsData.value = null
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    loadData()
  })
</script>

<style scoped>
  /* ===== 主体 ===== */
  .assets-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    padding: 16px;
    overflow: hidden;
  }

  /* ===== 筛选栏 ===== */
  .filter-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
    flex-wrap: wrap;
  }

  /* ===== 分组 ===== */
  .assets-groups {
    flex: 1;
    overflow-y: auto;
    scrollbar-width: thin;
    scrollbar-color: transparent transparent;
  }

  .assets-groups:hover {
    scrollbar-color: rgba(144, 147, 153, 0.4) transparent;
  }

  .assets-groups::-webkit-scrollbar {
    width: 4px;
  }

  .assets-groups::-webkit-scrollbar-thumb {
    background: transparent;
    border-radius: 2px;
    transition: background 0.2s;
  }

  .assets-groups:hover::-webkit-scrollbar-thumb {
    background: rgba(144, 147, 153, 0.4);
  }

  .location-section {
    margin-bottom: 8px;
  }

  .location-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    background: var(--el-fill-color-light);
    border-radius: 4px;
    cursor: pointer;
    user-select: none;
    font-weight: 600;
    font-size: 14px;
  }

  .location-header:hover {
    background: var(--el-fill-color);
  }

  .mg-arrow {
    font-size: 10px;
    transition: transform 0.15s;
    color: var(--el-text-color-secondary);
  }

  .mg-arrow.expanded {
    transform: rotate(90deg);
  }

  .mg-title {
    flex: 1;
  }

  .mg-count {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    font-weight: 400;
  }

  /* ===== 物品列表 ===== */
  .asset-items {
    padding: 4px 0;
  }

  .asset-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 12px;
    border-radius: 4px;
    transition: background 0.15s;
    cursor: default;
  }

  .asset-item:hover {
    background: var(--el-fill-color-light);
  }

  .asset-icon {
    width: 28px;
    height: 28px;
    border-radius: 3px;
    border: 1px solid var(--el-border-color-lighter);
    flex-shrink: 0;
  }

  .asset-info {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .asset-type-name {
    font-size: 13px;
    color: var(--el-text-color-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .asset-name-tag {
    font-size: 11px;
    color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
    padding: 0 5px;
    border-radius: 3px;
    white-space: nowrap;
    flex-shrink: 0;
  }

  .asset-group {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    white-space: nowrap;
    width: 120px;
    text-align: right;
    flex-shrink: 0;
  }

  .asset-qty {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    font-weight: 500;
    width: 50px;
    text-align: right;
    flex-shrink: 0;
  }

  .asset-owner {
    font-size: 12px;
    color: var(--el-text-color-regular);
    width: 100px;
    text-align: right;
    flex-shrink: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .child-toggle {
    cursor: pointer;
    padding: 2px 4px;
  }

  /* ===== 子物品 ===== */
  .child-items {
    padding-left: 28px;
    border-left: 2px solid var(--el-border-color-lighter);
    margin-left: 24px;
  }

  .asset-item.child {
    padding-left: 8px;
  }
</style>
