<template>
  <div class="info-fittings-page art-full-height">
    <!-- 统计栏 -->
    <ElCard shadow="never" class="mb-2">
      <div class="flex items-center justify-between flex-wrap gap-4">
        <div class="flex items-center gap-4">
          <ElButton :loading="loading" size="small" @click="loadData">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
        </div>
        <div v-if="fittingsData" class="flex items-center gap-4 text-sm text-gray-500">
          <span>
            {{ $t('info.fittingCount') }}:
            <strong class="text-blue-500">{{ fittingsData.total }}</strong>
          </span>
        </div>
      </div>
    </ElCard>

    <!-- 主体区域 -->
    <div v-loading="loading" class="fittings-main">
      <!-- 筛选栏 -->
      <div class="filter-bar">
        <!-- 种族筛选 -->
        <ElSelect
          v-model="selectedRace"
          :placeholder="$t('info.allRaces')"
          clearable
          style="width: 140px"
          size="small"
        >
          <ElOption :label="$t('info.allRaces')" :value="''" />
          <ElOption
            v-for="race in raceOptions"
            :key="race.id"
            :label="race.name"
            :value="race.id"
          />
        </ElSelect>

        <!-- 舰船组筛选 -->
        <ElSelect
          v-model="selectedGroup"
          :placeholder="$t('info.allGroups')"
          clearable
          style="width: 200px"
          size="small"
        >
          <ElOption :label="$t('info.allGroups')" :value="''" />
          <ElOption
            v-for="grp in groupOptions"
            :key="grp.name"
            :label="grp.name"
            :value="grp.name"
          />
        </ElSelect>

        <!-- 搜索 -->
        <ElInput
          v-model="searchKeyword"
          :placeholder="$t('info.searchFitting')"
          clearable
          style="width: 200px"
          size="small"
          :prefix-icon="Search"
        />
      </div>

      <!-- 分组展示 -->
      <div v-if="groupedFittings.length > 0" class="fittings-groups">
        <div v-for="grp in groupedFittings" :key="grp.groupName" class="market-group-section">
          <!-- 舰船组标题 -->
          <div class="market-group-header" @click="toggleGroup(grp.groupName)">
            <span class="mg-arrow" :class="{ expanded: !collapsedGroups.has(grp.groupName) }"
              >▶</span
            >
            <span class="mg-title">{{ grp.groupName }}</span>
            <span class="mg-count">{{ grp.fittings.length }}</span>
          </div>

          <!-- 装配网格 -->
          <div v-if="!collapsedGroups.has(grp.groupName)" class="fitting-grid">
            <div
              v-for="fit in grp.fittings"
              :key="`${fit.fitting_id}-${fit.character_id}`"
              class="fitting-card"
              @click="openDetail(fit)"
            >
              <img
                :src="`https://images.evetech.net/types/${fit.ship_type_id}/icon?size=64`"
                :alt="fit.ship_name"
                class="fitting-icon"
                loading="lazy"
              />
              <span class="fitting-label">{{ fit.name }}</span>
              <span v-if="fit.race_name" class="race-badge">{{ fit.race_name }}</span>
            </div>
          </div>
        </div>
      </div>

      <ElEmpty v-else-if="!loading" :description="$t('info.noFittingData')" :image-size="60" />
    </div>

    <!-- 装配详情弹窗 -->
    <ElDialog
      v-model="detailVisible"
      :title="$t('info.fittingDetail')"
      width="680px"
      :close-on-click-modal="true"
      destroy-on-close
    >
      <div v-if="selectedFitting" class="km-detail">
        <!-- 头部 -->
        <div class="km-header">
          <img
            :src="`https://images.evetech.net/types/${selectedFitting.ship_type_id}/icon?size=64`"
            class="km-ship-icon"
            alt="ship"
          />
          <div class="km-header-info">
            <h3 class="km-ship-name">{{ selectedFitting.name }}</h3>
            <p class="km-meta">{{ selectedFitting.ship_name }}</p>
            <p v-if="selectedFitting.description" class="km-meta">{{
              selectedFitting.description
            }}</p>
          </div>
        </div>

        <!-- 槽位列表 -->
        <div class="km-slots">
          <div v-for="slot in selectedFitting.slots" :key="slot.flag_name" class="km-slot-group">
            <div class="km-slot-header">
              <span>{{ slot.flag_text || slot.flag_name }}</span>
            </div>
            <div class="km-slot-items">
              <div
                v-for="(item, idx) in slot.items"
                :key="`${item.type_id}-${item.flag}-${idx}`"
                class="km-item"
              >
                <img
                  :src="`https://images.evetech.net/types/${item.type_id}/icon?size=32`"
                  class="km-item-icon"
                  alt=""
                />
                <span class="km-item-name">{{ item.type_name }}</span>
                <span v-if="item.quantity > 1" class="km-item-qty">x{{ item.quantity }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <template #footer>
        <ElButton @click="detailVisible = false">{{ $t('common.close') }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { Refresh, Search } from '@element-plus/icons-vue'
  import { ElCard, ElSelect, ElOption, ElButton, ElEmpty, ElInput, ElDialog } from 'element-plus'
  import { fetchInfoFittings } from '@/api/eve-info'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'EveInfoFittings' })

  const userStore = useUserStore()

  // ---- 数据 ----
  const fittingsData = ref<Api.EveInfo.FittingsListResponse | null>(null)
  const loading = ref(false)
  const searchKeyword = ref('')
  const selectedRace = ref<number | string>('')
  const selectedGroup = ref('')
  const collapsedGroups = ref(new Set<string>())
  const detailVisible = ref(false)
  const selectedFitting = ref<Api.EveInfo.FittingResponse | null>(null)

  // ---- 计算属性 ----

  /** 种族选项 */
  const raceOptions = computed(() => {
    if (!fittingsData.value?.fittings) return []
    const map = new Map<number, string>()
    for (const f of fittingsData.value.fittings) {
      if (f.race_id && !map.has(f.race_id)) {
        map.set(f.race_id, f.race_name || `Race ${f.race_id}`)
      }
    }
    return Array.from(map.entries())
      .map(([id, name]) => ({ id, name }))
      .sort((a, b) => a.name.localeCompare(b.name))
  })

  /** 舰船组选项 */
  const groupOptions = computed(() => {
    if (!fittingsData.value?.fittings) return []
    const set = new Map<string, number>()
    for (const f of fittingsData.value.fittings) {
      const name = f.group_name || 'Unknown'
      set.set(name, (set.get(name) ?? 0) + 1)
    }
    return Array.from(set.entries())
      .map(([name, count]) => ({ name, count }))
      .sort((a, b) => a.name.localeCompare(b.name))
  })

  /** 筛选后的装配 */
  const filteredFittings = computed(() => {
    if (!fittingsData.value?.fittings) return []
    let list = fittingsData.value.fittings

    if (selectedRace.value) {
      list = list.filter((f) => f.race_id === selectedRace.value)
    }
    if (selectedGroup.value) {
      list = list.filter((f) => (f.group_name || 'Unknown') === selectedGroup.value)
    }
    if (searchKeyword.value) {
      const kw = searchKeyword.value.toLowerCase()
      list = list.filter(
        (f) =>
          f.name?.toLowerCase().includes(kw) ||
          f.ship_name?.toLowerCase().includes(kw) ||
          f.group_name?.toLowerCase().includes(kw)
      )
    }
    return list
  })

  /** 按 Group -> Race -> Ship 分组 */
  interface FittingGroupSection {
    groupName: string
    fittings: Api.EveInfo.FittingResponse[]
  }

  const groupedFittings = computed<FittingGroupSection[]>(() => {
    const fittings = filteredFittings.value
    if (!fittings.length) return []

    const grpMap = new Map<string, Api.EveInfo.FittingResponse[]>()
    for (const f of fittings) {
      const grpName = f.group_name || 'Unknown'
      if (!grpMap.has(grpName)) grpMap.set(grpName, [])
      grpMap.get(grpName)!.push(f)
    }

    const sections: FittingGroupSection[] = []
    for (const [grpName, grpFittings] of grpMap) {
      const sorted = [...grpFittings].sort(
        (a, b) =>
          (a.race_name || '').localeCompare(b.race_name || '') ||
          (a.ship_name || '').localeCompare(b.ship_name || '') ||
          a.name.localeCompare(b.name)
      )
      sections.push({ groupName: grpName, fittings: sorted })
    }

    return sections.sort((a, b) => a.groupName.localeCompare(b.groupName))
  })

  // ---- 交互 ----
  const toggleGroup = (name: string) => {
    if (collapsedGroups.value.has(name)) {
      collapsedGroups.value.delete(name)
    } else {
      collapsedGroups.value.add(name)
    }
    collapsedGroups.value = new Set(collapsedGroups.value)
  }

  const openDetail = (fitting: Api.EveInfo.FittingResponse) => {
    selectedFitting.value = fitting
    detailVisible.value = true
  }

  const loadData = async () => {
    loading.value = true
    try {
      fittingsData.value = await fetchInfoFittings({
        language: userStore.language
      })
    } catch {
      fittingsData.value = null
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
  .fittings-main {
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
  .fittings-groups {
    flex: 1;
    overflow-y: auto;
    scrollbar-width: thin;
    scrollbar-color: transparent transparent;
  }

  .fittings-groups:hover {
    scrollbar-color: rgba(144, 147, 153, 0.4) transparent;
  }

  .fittings-groups::-webkit-scrollbar {
    width: 4px;
  }

  .fittings-groups::-webkit-scrollbar-thumb {
    background: transparent;
    border-radius: 2px;
    transition: background 0.2s;
  }

  .fittings-groups:hover::-webkit-scrollbar-thumb {
    background: rgba(144, 147, 153, 0.4);
  }

  .market-group-section {
    margin-bottom: 8px;
  }

  .market-group-header {
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

  .market-group-header:hover {
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

  /* ===== 装配网格 ===== */
  .fitting-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    padding: 6px 0 4px 12px;
  }

  .fitting-card {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 84px;
    flex-shrink: 0;
    padding: 6px 4px 4px;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.15s;
    border: 1px solid transparent;
    text-align: center;
    background: var(--el-fill-color-lighter);
  }

  .fitting-card:hover {
    border-color: var(--el-color-primary-light-5);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  .fitting-icon {
    width: 48px;
    height: 48px;
    border-radius: 4px;
    margin-bottom: 4px;
  }

  .fitting-label {
    font-size: 11px;
    line-height: 1.3;
    word-break: break-word;
    max-height: 2.6em;
    overflow: hidden;
  }

  .race-badge {
    display: inline-block;
    font-size: 10px;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color);
    padding: 0 4px;
    border-radius: 3px;
    margin-top: 2px;
  }

  /* ===== 详情弹窗（复用 KM 样式） ===== */
  .km-detail {
    min-height: 200px;
  }

  .km-header {
    display: flex;
    align-items: flex-start;
    gap: 16px;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .km-ship-icon {
    width: 64px;
    height: 64px;
    border-radius: 6px;
    border: 1px solid var(--el-border-color);
    flex-shrink: 0;
  }

  .km-header-info {
    flex: 1;
    min-width: 0;
  }

  .km-ship-name {
    font-size: 18px;
    font-weight: 600;
    margin: 0 0 4px;
    color: var(--el-text-color-primary);
  }

  .km-meta {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    margin: 2px 0;
  }

  .km-slots {
    margin-top: 16px;
  }

  .km-slot-group {
    margin-bottom: 12px;
  }

  .km-slot-header {
    padding: 6px 12px;
    background: var(--el-color-primary-light-9);
    border-left: 3px solid var(--el-color-primary);
    font-size: 13px;
    font-weight: 600;
    color: var(--el-color-primary);
    border-radius: 0 4px 4px 0;
  }

  .km-slot-items {
    padding: 4px 0;
  }

  .km-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 12px;
    border-radius: 4px;
    transition: background 0.15s;
  }

  .km-item:hover {
    background: var(--el-fill-color-light);
  }

  .km-item-icon {
    width: 28px;
    height: 28px;
    border-radius: 3px;
    border: 1px solid var(--el-border-color-lighter);
    flex-shrink: 0;
  }

  .km-item-name {
    flex: 1;
    font-size: 13px;
    color: var(--el-text-color-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .km-item-qty {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    font-weight: 500;
    flex-shrink: 0;
  }
</style>
