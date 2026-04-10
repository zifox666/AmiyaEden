<template>
  <div class="info-ships-page art-full-height">
    <!-- 人物切换器 -->
    <ElCard shadow="never" class="mb-2">
      <div class="flex items-center justify-between flex-wrap gap-4">
        <div class="flex items-center gap-4">
          <span class="text-sm text-gray-500">{{ $t('info.selectCharacter') }}</span>
          <ElSelect
            v-model="selectedCharacterId"
            :placeholder="$t('info.selectCharacterPlaceholder')"
            @change="onCharacterChange"
            style="width: 240px"
          >
            <ElOption
              v-for="char in characters"
              :key="char.character_id"
              :value="char.character_id"
              :label="char.character_name"
            >
              <div class="flex items-center gap-2">
                <ElAvatar :src="buildEveCharacterPortraitUrl(char.character_id, 24)" :size="24" />
                <span>{{ char.character_name }}</span>
              </div>
            </ElOption>
          </ElSelect>
          <ElButton :loading="loading" size="small" @click="loadData">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
        </div>
        <!-- 统计 -->
        <div v-if="shipData" class="flex items-center gap-4 text-sm text-gray-500">
          <span>
            {{ $t('info.flyableShips') }}:
            <strong class="text-green-500">{{ shipData.flyable_ships }}</strong>
            / {{ shipData.total_ships }}
          </span>
        </div>
      </div>
    </ElCard>

    <!-- 主体区域 -->
    <div v-loading="loading" class="ships-main">
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
          :placeholder="$t('info.searchShip')"
          clearable
          style="width: 200px"
          size="small"
          :prefix-icon="Search"
        />

        <!-- 是否只显示可用 -->
        <ElCheckbox v-model="onlyFlyable" :label="$t('info.onlyFlyable')" size="small" />
      </div>

      <!-- 舰船分组展示 -->
      <div v-if="groupedShips.length > 0" class="ships-groups">
        <div v-for="grp in groupedShips" :key="grp.groupName" class="market-group-section">
          <!-- 舰船组标题 -->
          <div class="market-group-header" @click="toggleMarketGroup(grp.groupName)">
            <span class="mg-arrow" :class="{ expanded: !collapsedGroups.has(grp.groupName) }"
              >▶</span
            >
            <span class="mg-title">{{ grp.groupName }}</span>
            <span class="mg-count">{{ grp.flyable }}/{{ grp.total }}</span>
          </div>

          <!-- 舰船网格（所有种族平铺） -->
          <div v-if="!collapsedGroups.has(grp.groupName)" class="ship-grid">
            <ElPopover
              v-for="ship in grp.ships"
              :key="ship.type_id"
              placement="bottom"
              :width="280"
              trigger="hover"
            >
              <template #reference>
                <div
                  class="ship-card"
                  :class="{ flyable: ship.can_fly, 'not-flyable': !ship.can_fly }"
                >
                  <img
                    :src="`https://images.evetech.net/types/${ship.type_id}/icon?size=64`"
                    :alt="ship.type_name"
                    class="ship-icon"
                    loading="lazy"
                  />
                  <span class="ship-label">{{ ship.type_name }}</span>
                  <span v-if="ship.race_name" class="race-badge">{{ ship.race_name }}</span>
                </div>
              </template>
              <!-- Popover 内容：技能需求 -->
              <div class="ship-popover">
                <div class="popover-title">{{ ship.type_name }}</div>
                <div class="popover-subtitle">{{ ship.group_name }}</div>
                <div v-if="ship.skill_reqs.length > 0" class="popover-skills">
                  <div
                    v-for="sr in ship.skill_reqs"
                    :key="sr.skill_id"
                    class="popover-skill-row"
                    :class="{ met: sr.met, unmet: !sr.met }"
                  >
                    <span class="popover-skill-name">{{
                      sr.skill_name || `ID ${sr.skill_id}`
                    }}</span>
                    <span class="popover-skill-level">
                      Lv{{ sr.current_level }} / {{ sr.required_level }}
                    </span>
                    <span v-if="sr.met" class="check">✓</span>
                    <span v-else class="cross">✗</span>
                  </div>
                </div>
                <div v-else class="text-xs text-gray-400">{{ $t('info.noSkillReqs') }}</div>
              </div>
            </ElPopover>
          </div>
        </div>
      </div>

      <ElEmpty v-else-if="!loading" :description="$t('info.noShipData')" :image-size="60" />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { Refresh, Search } from '@element-plus/icons-vue'
  import {
    ElCard,
    ElSelect,
    ElOption,
    ElAvatar,
    ElButton,
    ElEmpty,
    ElInput,
    ElCheckbox,
    ElPopover
  } from 'element-plus'
  import { fetchMyCharacters } from '@/api/auth'
  import { fetchInfoShips } from '@/api/eve-info'
  import { useUserStore } from '@/store/modules/user'
  import { buildEveCharacterPortraitUrl } from '@/utils/eve-image'

  defineOptions({ name: 'EveInfoShips' })

  const userStore = useUserStore()

  // ---- 数据 ----
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const selectedCharacterId = ref<number>()
  const shipData = ref<Api.EveInfo.ShipResponse | null>(null)
  const loading = ref(false)
  const searchKeyword = ref('')
  const selectedRace = ref<number | string>('')
  const selectedGroup = ref('')
  const onlyFlyable = ref(false)
  const collapsedGroups = ref(new Set<string>())

  // ---- 计算属性 ----

  /** 种族选项 */
  const raceOptions = computed(() => {
    if (!shipData.value?.ships) return []
    const map = new Map<number, string>()
    for (const s of shipData.value.ships) {
      if (s.race_id && !map.has(s.race_id)) {
        map.set(s.race_id, s.race_name || `Race ${s.race_id}`)
      }
    }
    return Array.from(map.entries())
      .map(([id, name]) => ({ id, name }))
      .sort((a, b) => a.name.localeCompare(b.name))
  })

  /** 舰船组选项 */
  const groupOptions = computed(() => {
    if (!shipData.value?.ships) return []
    const set = new Map<string, number>()
    for (const s of shipData.value.ships) {
      const name = s.group_name || 'Unknown'
      set.set(name, (set.get(name) ?? 0) + 1)
    }
    return Array.from(set.entries())
      .map(([name, count]) => ({ name, count }))
      .sort((a, b) => a.name.localeCompare(b.name))
  })

  /** 筛选后的舰船 */
  const filteredShips = computed(() => {
    if (!shipData.value?.ships) return []
    let list = shipData.value.ships

    if (selectedRace.value) {
      list = list.filter((s) => s.race_id === selectedRace.value)
    }
    if (selectedGroup.value) {
      list = list.filter((s) => (s.group_name || 'Unknown') === selectedGroup.value)
    }
    if (onlyFlyable.value) {
      list = list.filter((s) => s.can_fly)
    }
    if (searchKeyword.value) {
      const kw = searchKeyword.value.toLowerCase()
      list = list.filter(
        (s) => s.type_name?.toLowerCase().includes(kw) || s.group_name?.toLowerCase().includes(kw)
      )
    }
    return list
  })

  /** 按 group 平铺分组 */
  interface GroupSection {
    groupName: string
    total: number
    flyable: number
    ships: Api.EveInfo.ShipItem[]
  }

  const groupedShips = computed<GroupSection[]>(() => {
    const ships = filteredShips.value
    if (!ships.length) return []

    const grpMap = new Map<string, Api.EveInfo.ShipItem[]>()
    for (const s of ships) {
      const grpName = s.group_name || 'Unknown'
      if (!grpMap.has(grpName)) grpMap.set(grpName, [])
      grpMap.get(grpName)!.push(s)
    }

    const sections: GroupSection[] = []
    for (const [grpName, grpShips] of grpMap) {
      // 组内按种族名排序，使同种族舰船聊在一起
      const sorted = [...grpShips].sort(
        (a, b) =>
          (a.race_name || '').localeCompare(b.race_name || '') ||
          a.type_name.localeCompare(b.type_name)
      )
      sections.push({
        groupName: grpName,
        total: grpShips.length,
        flyable: grpShips.filter((s) => s.can_fly).length,
        ships: sorted
      })
    }

    return sections.sort((a, b) => a.groupName.localeCompare(b.groupName))
  })

  // ---- 交互 ----
  const toggleMarketGroup = (name: string) => {
    if (collapsedGroups.value.has(name)) {
      collapsedGroups.value.delete(name)
    } else {
      collapsedGroups.value.add(name)
    }
    // trigger reactivity
    collapsedGroups.value = new Set(collapsedGroups.value)
  }

  const loadCharacters = async () => {
    try {
      characters.value = await fetchMyCharacters()
      if (characters.value.length > 0 && !selectedCharacterId.value) {
        selectedCharacterId.value = characters.value[0].character_id
        loadData()
      }
    } catch {
      characters.value = []
    }
  }

  const loadData = async () => {
    if (!selectedCharacterId.value) return
    loading.value = true
    try {
      shipData.value = await fetchInfoShips({
        character_id: selectedCharacterId.value,
        language: userStore.language
      })
    } catch {
      shipData.value = null
    } finally {
      loading.value = false
    }
  }

  const onCharacterChange = () => {
    selectedRace.value = ''
    selectedGroup.value = ''
    searchKeyword.value = ''
    onlyFlyable.value = false
    loadData()
  }

  onMounted(() => {
    loadCharacters()
  })
</script>

<style scoped>
  /* ===== 主体 ===== */
  .ships-main {
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
  .ships-groups {
    flex: 1;
    overflow-y: auto;
    scrollbar-width: thin;
    scrollbar-color: transparent transparent;
  }

  .ships-groups:hover {
    scrollbar-color: rgba(144, 147, 153, 0.4) transparent;
  }

  .ships-groups::-webkit-scrollbar {
    width: 4px;
  }

  .ships-groups::-webkit-scrollbar-thumb {
    background: transparent;
    border-radius: 2px;
    transition: background 0.2s;
  }

  .ships-groups:hover::-webkit-scrollbar-thumb {
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

  .race-groups {
    padding-left: 0;
  }

  /* ===== 舰船网格 ===== */
  .ship-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    padding: 6px 0 4px 12px;
  }

  .race-name {
    font-weight: 500;
  }

  .race-count {
    font-size: 12px;
  }

  /* ===== 舰船网格 ===== */
  .ship-grid {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    padding: 6px 0 4px 12px;
  }

  .ship-card {
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
  }

  .ship-card.flyable {
    background: rgba(103, 194, 58, 0.06);
    border-color: rgba(103, 194, 58, 0.2);
  }

  .ship-card.not-flyable {
    background: var(--el-fill-color-lighter);
    opacity: 0.55;
  }

  .ship-card:hover {
    opacity: 1;
    border-color: var(--el-color-primary-light-5);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  }

  .ship-icon {
    width: 48px;
    height: 48px;
    border-radius: 4px;
    margin-bottom: 4px;
  }

  .ship-label {
    font-size: 11px;
    line-height: 1.3;
    word-break: break-word;
    max-height: 2.6em;
    overflow: hidden;
  }

  .race-badge {
    display: inline-block;
    margin-top: 2px;
    font-size: 10px;
    line-height: 1.4;
    padding: 0 4px;
    border-radius: 3px;
    background: var(--el-fill-color);
    color: var(--el-text-color-placeholder);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 80px;
  }

  /* ===== Popover ===== */
  .ship-popover .popover-title {
    font-weight: 600;
    font-size: 14px;
    margin-bottom: 2px;
  }

  .ship-popover .popover-subtitle {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    margin-bottom: 8px;
  }

  .popover-skill-row {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    padding: 2px 0;
  }

  .popover-skill-row.prereq {
    opacity: 0.8;
  }

  .prereq-indent {
    color: var(--el-text-color-placeholder);
    font-size: 11px;
    flex-shrink: 0;
  }

  .popover-skill-row.met {
    color: var(--el-color-success);
  }

  .popover-skill-row.unmet {
    color: var(--el-color-danger);
  }

  .popover-skill-name {
    flex: 1;
  }

  .popover-skill-level {
    font-weight: 500;
    white-space: nowrap;
  }

  .popover-skill-req {
    color: var(--el-text-color-placeholder);
    font-weight: 400;
    margin-left: 2px;
  }

  .check {
    font-weight: bold;
  }

  .cross {
    font-weight: bold;
  }

  /* ===== 响应式 ===== */
  @media (max-width: 768px) {
    .ship-card {
      width: 76px;
    }
  }
</style>
