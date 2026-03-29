<template>
  <div class="info-skill-page art-full-height">
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
                <ElAvatar :src="char.portrait_url" :size="24" />
                <span>{{ char.character_name }}</span>
              </div>
            </ElOption>
          </ElSelect>
          <ElButton :loading="loading" size="small" @click="loadData">
            <el-icon class="mr-1"><Refresh /></el-icon>
            {{ $t('common.refresh') }}
          </ElButton>
        </div>
      </div>
    </ElCard>

    <!-- 主体区域：左侧技能 + 右侧队列 -->
    <div v-loading="loading" class="skill-main">
      <!-- ========== 左侧：技能面板 ========== -->
      <div class="skill-panel">
        <!-- 技能标题 + 总 SP -->
        <div class="panel-header">
          <span class="panel-title">{{ $t('info.skillList') }}</span>
          <span class="total-sp">
            {{ formatNumber(skillData?.total_sp ?? 0) }} {{ $t('info.totalSPLabel') }}
          </span>
        </div>

        <!-- 筛选栏 -->
        <div class="filter-bar">
          <ElSelect
            v-model="selectedGroup"
            :placeholder="$t('info.allSkills')"
            clearable
            style="width: 160px"
            size="small"
          >
            <ElOption :label="$t('info.allSkills')" :value="''" />
            <ElOption
              v-for="group in skillGroups"
              :key="group.groupName"
              :label="group.groupName"
              :value="group.groupName"
            />
          </ElSelect>
          <ElInput
            v-model="searchKeyword"
            :placeholder="$t('info.searchSkill')"
            clearable
            style="width: 180px"
            size="small"
            :prefix-icon="Search"
          />
        </div>

        <!-- 技能分类网格 -->
        <div class="category-grid">
          <div
            v-for="group in skillGroups"
            :key="group.groupName"
            class="category-cell"
            :class="{ active: selectedGroup === group.groupName }"
            :style="{ '--progress': group.progress + '%' }"
            @click="toggleGroup(group.groupName)"
          >
            <span class="category-name">{{ group.groupName }}</span>
            <span class="category-count">{{ group.count }}</span>
          </div>
        </div>

        <!-- 技能列表 -->
        <div v-if="filteredSkills.length > 0" class="skill-list">
          <div
            v-for="skill in filteredSkills"
            :key="skill.skill_id"
            class="skill-row"
            :class="{ 'in-queue': isInQueue(skill.skill_id), unlearned: !skill.learned }"
          >
            <!-- 未吸收：书本图标 -->
            <div v-if="!skill.learned" class="skill-book" :title="$t('info.skillNotLearned')">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 16 16"
                fill="currentColor"
                aria-hidden="true"
              >
                <path
                  d="M1 2.828c.885-.37 2.154-.769 3.388-.893 1.33-.134 2.458.063 3.112.752v9.746c-.935-.53-2.12-.603-3.213-.493-1.18.12-2.37.461-3.287.811V2.828zm7.5-.141c.654-.689 1.782-.886 3.112-.752 1.234.124 2.503.523 3.388.893v9.923c-.918-.35-2.107-.692-3.287-.81-1.094-.111-2.278-.039-3.213.492V2.687zM8 1.783C7.015.936 5.587.81 4.287.94c-1.514.153-3.042.672-3.994 1.105A.5.5 0 0 0 0 2.5v11a.5.5 0 0 0 .707.455c.882-.4 2.303-.881 3.68-1.02 1.409-.142 2.59.087 3.223.877a.5.5 0 0 0 .78 0c.633-.79 1.814-1.019 3.222-.877 1.378.139 2.8.62 3.681 1.02A.5.5 0 0 0 16 13.5v-11a.5.5 0 0 0-.293-.455c-.952-.433-2.48-.952-3.994-1.105C10.413.809 8.985.936 8 1.783z"
                />
              </svg>
            </div>
            <!-- 已吸收：等级进度条 -->
            <div v-else class="level-bars">
              <span
                v-for="i in 5"
                :key="i"
                class="level-bar"
                :class="{
                  trained: i <= skill.active_level,
                  'partially-trained':
                    i === skill.active_level + 1 && skill.trained_level > skill.active_level
                }"
              />
            </div>
            <span class="skill-name">{{ skill.skill_name || `Type ${skill.skill_id}` }}</span>
            <span class="skill-status">
              <span v-if="getQueueRemainingTime(skill.skill_id)" class="training-time">
                {{ getQueueRemainingTime(skill.skill_id) }}
              </span>
              <span v-else-if="skill.active_level >= 5" class="trained-check">✓</span>
            </span>
          </div>
        </div>
        <ElEmpty v-else-if="!loading" :description="$t('info.noSkillData')" :image-size="60" />
      </div>

      <!-- ========== 右侧：技能队列面板 ========== -->
      <div class="queue-panel">
        <!-- 队列标题 -->
        <div class="panel-header">
          <span class="panel-title">{{ $t('info.skillQueue') }}</span>
          <span class="queue-capacity">
            {{ skillData?.skill_queue?.length ?? 0 }}<span class="queue-max">/150</span>
          </span>
        </div>

        <!-- 当前训练进度（第一个） -->
        <div v-if="currentTraining" class="current-training">
          <div class="training-chevrons">
            <span v-for="i in 8" :key="i" class="chevron">›</span>
          </div>
          <div class="training-info">
            <span class="training-name">
              {{ currentTraining.skill_name || `Type ${currentTraining.skill_id}` }}
              {{ romanLevel(currentTraining.finished_level) }}
            </span>
            <span class="training-countdown">
              {{ formatRemainingTime(currentTraining.finish_date) }}
            </span>
          </div>
          <ElProgress
            :percentage="calcTimeProgress(currentTraining)"
            :stroke-width="4"
            :show-text="false"
            color="#5b9bd5"
            class="mt-1"
          />
        </div>

        <!-- 队列列表 -->
        <div class="queue-list">
          <div v-for="item in queueWithoutFirst" :key="item.queue_position" class="queue-item">
            <div class="level-bars small">
              <span
                v-for="i in 5"
                :key="i"
                class="level-bar"
                :class="{ trained: i < item.finished_level }"
              />
            </div>
            <span class="queue-skill-name">
              {{ item.skill_name || `Type ${item.skill_id}` }}
              {{ romanLevel(item.finished_level) }}
            </span>
            <span class="queue-time">{{ formatRemainingTime(item.finish_date) }}</span>
          </div>
          <ElEmpty
            v-if="!loading && (!skillData?.skill_queue || skillData.skill_queue.length === 0)"
            :description="$t('info.noSkillQueue')"
            :image-size="60"
          />
        </div>

        <!-- 底部统计 -->
        <div v-if="skillData" class="queue-footer">
          <div class="unallocated-sp">
            <span class="sp-value">{{ formatNumber(skillData.unallocated_sp) }}</span>
            {{ $t('info.unallocatedSPSuffix') }}
          </div>
          <div class="total-training-time">
            <span class="footer-label">{{ $t('info.totalTrainingTime') }}</span>
            <span class="time-value">{{ totalQueueTime }}</span>
          </div>
          <div class="queued-sp">
            {{ formatNumber(totalQueueSP) }}{{ $t('info.queuedSPSuffix') }}
          </div>
        </div>
      </div>
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
    ElProgress,
    ElInput
  } from 'element-plus'
  import { fetchMyCharacters } from '@/api/auth'
  import { fetchInfoSkills } from '@/api/eve-info'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'EveInfoSkill' })

  const userStore = useUserStore()

  // ---- 数据 ----
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const selectedCharacterId = ref<number>()
  const skillData = ref<Api.EveInfo.SkillResponse | null>(null)
  const loading = ref(false)
  const searchKeyword = ref('')
  const selectedGroup = ref('')

  // ---- 格式化工具 ----
  const formatNumber = (v: number) => new Intl.NumberFormat('en-US').format(v)

  const romanLevel = (level: number): string => {
    const numerals = ['', 'I', 'II', 'III', 'IV', 'V']
    return numerals[level] || String(level)
  }

  const formatRemainingTime = (finishDate: number): string => {
    if (!finishDate) return ''
    const now = Math.floor(Date.now() / 1000)
    let remaining = finishDate - now
    if (remaining <= 0) return '已完成'
    const days = Math.floor(remaining / 86400)
    remaining %= 86400
    const hours = Math.floor(remaining / 3600)
    remaining %= 3600
    const minutes = Math.floor(remaining / 60)
    const seconds = remaining % 60
    let result = ''
    if (days > 0) result += `${days}天 `
    if (hours > 0 || days > 0) result += `${hours}小时`
    if (days === 0) {
      if (minutes > 0) result += ` ${minutes}分`
      if (hours === 0 && minutes < 10) result += ` ${seconds}秒`
    }
    return result.trim()
  }

  // ---- 计算属性 ----

  /** 技能分组 */
  interface SkillGroup {
    groupName: string
    count: number
    skills: Api.EveInfo.SkillItem[]
    progress: number
  }

  const skillGroups = computed<SkillGroup[]>(() => {
    if (!skillData.value?.skills) return []
    const map = new Map<string, SkillGroup>()
    for (const s of skillData.value.skills) {
      const key = s.group_name || 'Unknown'
      if (!map.has(key)) {
        map.set(key, { groupName: key, count: 0, skills: [], progress: 0 })
      }
      const g = map.get(key)!
      g.count++
      g.skills.push(s)
    }
    // 计算每个分类的训练完成度：已有等级之和 / (技能数 * 5)
    for (const g of map.values()) {
      const totalLevels = g.count * 5
      const trainedLevels = g.skills.reduce((sum, s) => sum + (s.active_level ?? 0), 0)
      g.progress = totalLevels > 0 ? Math.round((trainedLevels / totalLevels) * 100) : 0
    }
    return Array.from(map.values()).sort((a, b) => a.groupName.localeCompare(b.groupName))
  })

  /** 筛选后的技能列表 */
  const filteredSkills = computed(() => {
    if (!skillData.value?.skills) return []
    let list = skillData.value.skills
    if (selectedGroup.value) {
      list = list.filter((s) => s.group_name === selectedGroup.value)
    }
    if (searchKeyword.value) {
      const kw = searchKeyword.value.toLowerCase()
      list = list.filter(
        (s) => s.skill_name?.toLowerCase().includes(kw) || s.group_name?.toLowerCase().includes(kw)
      )
    }
    // 按组名 -> 技能名排序
    return [...list].sort(
      (a, b) => a.group_name.localeCompare(b.group_name) || a.skill_name.localeCompare(b.skill_name)
    )
  })

  /** 当前正在训练的技能（根据时间计算，跳过已完成的） */
  const currentTraining = computed(() => {
    if (!skillData.value?.skill_queue?.length) return null
    const now = Math.floor(Date.now() / 1000)
    return skillData.value.skill_queue.find((q) => q.finish_date > now) ?? null
  })

  /** 队列中当前训练之后的技能 */
  const queueWithoutFirst = computed(() => {
    if (!skillData.value?.skill_queue?.length) return []
    const current = currentTraining.value
    if (!current) return []
    return skillData.value.skill_queue.filter((q) => q.queue_position > current.queue_position)
  })

  /** 队列中 skill_id 集合（快速查找） */
  const queueSkillMap = computed(() => {
    const map = new Map<number, Api.EveInfo.SkillQueueItem>()
    for (const q of skillData.value?.skill_queue ?? []) {
      // 只保留第一个匹配（最近的）
      if (!map.has(q.skill_id)) map.set(q.skill_id, q)
    }
    return map
  })

  /** 判断技能是否在队列中 */
  const isInQueue = (skillId: number) => queueSkillMap.value.has(skillId)

  /** 获取队列中该技能的剩余时间 */
  const getQueueRemainingTime = (skillId: number): string => {
    const q = queueSkillMap.value.get(skillId)
    if (!q) return ''
    return formatRemainingTime(q.finish_date)
  }

  /** 计算当前训练的时间进度百分比 */
  const calcTimeProgress = (item: Api.EveInfo.SkillQueueItem): number => {
    if (!item.start_date || !item.finish_date) return 0
    const now = Math.floor(Date.now() / 1000)
    const total = item.finish_date - item.start_date
    if (total <= 0) return 100
    const elapsed = now - item.start_date
    return Math.min(100, Math.max(0, Math.round((elapsed / total) * 100)))
  }

  /** 队列总训练时间 */
  const totalQueueTime = computed(() => {
    if (!skillData.value?.skill_queue?.length) return '-'
    const lastItem = skillData.value.skill_queue[skillData.value.skill_queue.length - 1]
    if (!lastItem.finish_date) return '-'
    return formatRemainingTime(lastItem.finish_date)
  })

  /** 队列总技能点 */
  const totalQueueSP = computed(() => {
    if (!skillData.value?.skill_queue?.length) return 0
    return skillData.value.skill_queue.reduce((sum, item) => {
      return sum + Math.max(0, item.level_end_sp - item.training_start_sp)
    }, 0)
  })

  // ---- 交互 ----
  const toggleGroup = (groupName: string) => {
    selectedGroup.value = selectedGroup.value === groupName ? '' : groupName
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
      skillData.value = await fetchInfoSkills({
        character_id: selectedCharacterId.value,
        language: userStore.language
      })
    } catch {
      skillData.value = null
    } finally {
      loading.value = false
    }
  }

  const onCharacterChange = () => {
    selectedGroup.value = ''
    searchKeyword.value = ''
    loadData()
  }

  onMounted(() => {
    loadCharacters()
  })
</script>

<style scoped>
  /* ===== 主体布局 ===== */
  .skill-main {
    display: flex;
    gap: 12px;
    min-height: 0;
    flex: 1;
  }

  .skill-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    padding: 16px;
    overflow: hidden;
  }

  .queue-panel {
    width: 420px;
    min-width: 360px;
    display: flex;
    flex-direction: column;
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    padding: 16px;
    overflow: hidden;
  }

  /* ===== 面板头 ===== */
  .panel-header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .panel-title {
    font-size: 16px;
    font-weight: 600;
  }

  .total-sp {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .queue-capacity {
    font-size: 18px;
    font-weight: 600;
  }

  .queue-max {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    font-weight: 400;
  }

  /* ===== 筛选栏 ===== */
  .filter-bar {
    display: flex;
    gap: 8px;
    margin-bottom: 12px;
  }

  /* ===== 分类网格 ===== */
  .category-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 4px;
    margin-bottom: 12px;
  }

  .category-cell {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 10px;
    border-radius: 4px;
    background: var(--el-fill-color-light);
    cursor: pointer;
    font-size: 13px;
    transition: all 0.15s;
    user-select: none;
    overflow: hidden;
  }

  .category-cell::before {
    content: '';
    position: absolute;
    inset: 0;
    width: var(--progress, 0%);
    background: rgba(91, 164, 207, 0.18);
    transition: width 0.4s ease;
    pointer-events: none;
  }

  .category-cell:hover {
    background: var(--el-fill-color);
  }

  .category-cell.active {
    background: var(--el-color-primary-light-8);
    color: var(--el-color-primary);
    font-weight: 500;
  }

  .category-cell.active::before {
    background: rgba(var(--el-color-primary-rgb), 0.15);
  }

  .category-name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .category-count {
    margin-left: 6px;
    font-weight: 600;
    font-size: 14px;
    flex-shrink: 0;
    scrollbar-width: none;
  }

  /* ===== 技能列表 ===== */
  .skill-list {
    flex: 1;
    overflow-y: auto;
    scrollbar-width: none;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    align-content: start;
    gap: 2px;
  }

  .skill-row {
    display: flex;
    align-items: center;
    padding: 5px 8px;
    border-radius: 3px;
    font-size: 13px;
    gap: 8px;
    min-width: 0;
  }

  .skill-row:hover {
    background: var(--el-fill-color-light);
  }

  .skill-row.in-queue {
    background: var(--el-color-primary-light-9);
  }

  .skill-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .skill-status {
    flex-shrink: 0;
    font-size: 12px;
  }

  .training-time {
    color: var(--el-color-warning);
  }

  .trained-check {
    color: var(--el-color-success);
    font-weight: bold;
  }

  /* ===== 等级指示条 ===== */
  .level-bars {
    display: flex;
    gap: 2px;
    flex-shrink: 0;
  }

  .level-bar {
    width: 12px;
    height: 10px;
    border-radius: 1px;
    background: var(--el-border-color-lighter);
    display: inline-block;
  }

  .level-bar.trained {
    background: #5ba4cf;
  }

  .level-bar.partially-trained {
    background: linear-gradient(90deg, #5ba4cf 50%, var(--el-border-color-lighter) 50%);
  }

  .level-bars.small .level-bar {
    width: 10px;
    height: 8px;
  }

  /* ===== 未吸收技能（书本图标） ===== */
  .skill-row.unlearned {
    opacity: 0.45;
  }

  .skill-row.unlearned:hover {
    opacity: 0.8;
  }

  .skill-book {
    display: flex;
    align-items: center;
    justify-content: flex-start;
    width: 68px; /* 与 5 个等级条总宽对齐 */
    flex-shrink: 0;
    color: var(--el-text-color-placeholder);
  }

  .skill-book svg {
    width: 14px;
    height: 14px;
  }

  /* ===== 当前训练 ===== */
  .current-training {
    background: var(--el-fill-color-light);
    border-radius: 6px;
    padding: 10px 12px;
    margin-bottom: 8px;
  }

  .training-chevrons {
    display: flex;
    gap: 1px;
    font-size: 12px;
    color: #5ba4cf;
    line-height: 1;
    margin-bottom: 4px;
    letter-spacing: -2px;
    font-weight: bold;
  }

  .training-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .training-name {
    font-weight: 600;
    font-size: 14px;
  }

  .training-countdown {
    font-size: 16px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  /* ===== 队列列表 ===== */
  .queue-list {
    flex: 1;
    overflow-y: auto;
    scrollbar-width: none;
  }

  .queue-item {
    display: flex;
    align-items: center;
    padding: 5px 6px;
    border-radius: 3px;
    font-size: 13px;
    gap: 8px;
  }

  .queue-item:hover {
    background: var(--el-fill-color-light);
  }

  .queue-skill-name {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .queue-time {
    flex-shrink: 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    text-align: right;
    min-width: 90px;
  }

  /* ===== 队列底部统计 ===== */
  .queue-footer {
    margin-top: auto;
    padding-top: 12px;
    border-top: 1px solid var(--el-border-color-lighter);
    font-size: 13px;
  }

  .unallocated-sp {
    text-align: right;
    margin-bottom: 8px;
    color: var(--el-text-color-secondary);
  }

  .unallocated-sp .sp-value {
    color: var(--el-color-success);
    font-weight: 600;
  }

  .total-training-time {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    margin-bottom: 4px;
  }

  .footer-label {
    color: var(--el-text-color-secondary);
  }

  .time-value {
    font-size: 20px;
    font-weight: 600;
  }

  .queued-sp {
    text-align: right;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  /* ===== 响应式 ===== */
  @media (max-width: 900px) {
    .skill-main {
      flex-direction: column;
    }

    .queue-panel {
      width: 100%;
      min-width: 0;
    }
  }
</style>
