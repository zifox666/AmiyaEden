<template>
  <div class="info-skill-page art-full-height">
    <!-- 角色切换器 + 技能概览 -->
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
        </div>
        <div v-if="skillData" class="flex gap-6">
          <div class="text-center">
            <p class="text-sm text-gray-500">{{ $t('info.totalSP') }}</p>
            <p class="text-xl font-bold text-blue-600">{{ formatNumber(skillData.total_sp) }}</p>
          </div>
          <div class="text-center">
            <p class="text-sm text-gray-500">{{ $t('info.unallocatedSP') }}</p>
            <p class="text-xl font-bold text-orange-500">{{
              formatNumber(skillData.unallocated_sp)
            }}</p>
          </div>
          <div class="text-center">
            <p class="text-sm text-gray-500">{{ $t('info.skillCount') }}</p>
            <p class="text-xl font-bold">{{ skillData.skill_count }}</p>
          </div>
        </div>
      </div>
    </ElCard>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <!-- 技能队列 -->
      <ElCard class="lg:col-span-1" shadow="never">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="card-title">{{ $t('info.skillQueue') }}</span>
            <ElTag v-if="skillData?.skill_queue?.length" size="small" type="success">
              {{ skillData.skill_queue.length }} {{ $t('info.inQueue') }}
            </ElTag>
          </div>
        </template>

        <div v-if="skillData?.skill_queue?.length" class="skill-queue-list">
          <div
            v-for="item in skillData.skill_queue"
            :key="item.queue_position"
            class="skill-queue-item"
          >
            <div class="flex items-center justify-between">
              <div>
                <span class="font-medium">{{ item.skill_name || `Type ${item.skill_id}` }}</span>
                <ElTag size="small" class="ml-2" type="info">Lv {{ item.finished_level }}</ElTag>
                <ElTag size="small" class="ml-1" effect="plain" v-if="item.queue_position === 1">
                  {{ $t('info.inTraining') }}
                </ElTag>
              </div>
              <span class="text-xs text-gray-400">#{{ item.queue_position }}</span>
            </div>
            <div class="mt-1">
              <ElProgress
                :percentage="calcQueueProgress(item)"
                :stroke-width="6"
                :show-text="false"
                :status="item.queue_position === 0 ? undefined : 'warning'"
              />
              <div class="flex justify-between text-xs text-gray-400 mt-1">
                <span
                  >{{ formatSP(item.training_start_sp) }} /
                  {{ formatSP(item.level_end_sp) }} SP</span
                >
                <span v-if="item.finish_date">{{ formatTimestamp(item.finish_date) }}</span>
              </div>
            </div>
          </div>
        </div>
        <ElEmpty v-else :description="$t('info.noSkillQueue')" :image-size="80" />
      </ElCard>

      <!-- 技能列表 -->
      <ElCard class="lg:col-span-2" shadow="never">
        <template #header>
          <div class="flex items-center justify-between">
            <span class="card-title">{{ $t('info.skillList') }}</span>
            <div class="flex items-center gap-2">
              <ElInput
                v-model="searchKeyword"
                :placeholder="$t('info.searchSkill')"
                clearable
                style="width: 200px"
                size="small"
              />
              <ElButton :loading="loading" size="small" @click="loadData">
                <el-icon class="mr-1"><Refresh /></el-icon>
                {{ $t('common.refresh') }}
              </ElButton>
            </div>
          </div>
        </template>

        <ElTable
          v-loading="loading"
          :data="filteredSkills"
          stripe
          border
          style="width: 100%"
          max-height="600"
          :default-sort="{ prop: 'group_name', order: 'ascending' }"
        >
          <ElTableColumn prop="group_name" :label="$t('info.skillGroup')" width="180" sortable />
          <ElTableColumn prop="skill_name" :label="$t('info.skillName')" min-width="200" sortable>
            <template #default="{ row }">
              {{ row.skill_name || `Type ${row.skill_id}` }}
            </template>
          </ElTableColumn>
          <ElTableColumn
            prop="active_level"
            :label="$t('info.activeLevel')"
            width="100"
            align="center"
            sortable
          >
            <template #default="{ row }">
              <div class="flex items-center justify-center gap-0.5">
                <span
                  v-for="i in 5"
                  :key="i"
                  class="level-dot"
                  :class="i <= row.active_level ? 'active' : ''"
                />
              </div>
            </template>
          </ElTableColumn>
          <ElTableColumn
            prop="trained_level"
            :label="$t('info.trainedLevel')"
            width="100"
            align="center"
            sortable
          >
            <template #default="{ row }"> Lv {{ row.trained_level }} </template>
          </ElTableColumn>
          <ElTableColumn
            prop="skillpoints_in_skill"
            :label="$t('info.skillSP')"
            width="140"
            align="right"
            sortable
          >
            <template #default="{ row }">{{ formatNumber(row.skillpoints_in_skill) }}</template>
          </ElTableColumn>
        </ElTable>

        <ElEmpty
          v-if="!loading && filteredSkills.length === 0"
          :description="$t('info.noSkillData')"
        />
      </ElCard>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { Refresh } from '@element-plus/icons-vue'
  import {
    ElCard,
    ElSelect,
    ElOption,
    ElAvatar,
    ElTable,
    ElTableColumn,
    ElTag,
    ElButton,
    ElEmpty,
    ElProgress,
    ElInput
  } from 'element-plus'
  import { fetchMyCharacters } from '@/api/auth'
  import { fetchInfoSkills } from '@/api/eve-info'

  defineOptions({ name: 'EveInfoSkill' })

  const characters = ref<Api.Auth.EveCharacter[]>([])
  const selectedCharacterId = ref<number>()
  const skillData = ref<Api.EveInfo.SkillResponse | null>(null)
  const loading = ref(false)
  const searchKeyword = ref('')

  const formatNumber = (v: number) => new Intl.NumberFormat('en-US').format(v)

  const formatSP = (v: number) =>
    v >= 1_000_000
      ? `${(v / 1_000_000).toFixed(1)}M`
      : v >= 1_000
        ? `${(v / 1_000).toFixed(0)}K`
        : String(v)

  const formatTimestamp = (ts: number) => {
    if (!ts) return ''
    // ESI 使用 Unix 秒时间戳
    const d = new Date(ts * 1000)
    return d.toLocaleString()
  }

  const calcQueueProgress = (item: Api.EveInfo.SkillQueueItem) => {
    const total = item.level_end_sp - item.level_start_sp
    if (total <= 0) return 100
    const current = item.training_start_sp - item.level_start_sp
    return Math.min(100, Math.max(0, Math.round((current / total) * 100)))
  }

  const filteredSkills = computed(() => {
    if (!skillData.value?.skills) return []
    if (!searchKeyword.value) return skillData.value.skills
    const kw = searchKeyword.value.toLowerCase()
    return skillData.value.skills.filter(
      (s) => s.skill_name?.toLowerCase().includes(kw) || s.group_name?.toLowerCase().includes(kw)
    )
  })

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
        character_id: selectedCharacterId.value
      })
    } catch {
      skillData.value = null
    } finally {
      loading.value = false
    }
  }

  const onCharacterChange = () => {
    loadData()
  }

  onMounted(() => {
    loadCharacters()
  })
</script>

<style scoped>
  .card-title {
    font-size: 15px;
    font-weight: 500;
  }
  .skill-queue-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
    overflow-y: auto;
    max-height: 600px;
    scrollbar-width: none;
  }
  .skill-queue-item {
    padding: 8px 12px;
    border-radius: 6px;
    background: var(--el-fill-color-light);
  }
  .level-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--el-border-color-lighter);
    display: inline-block;
  }
  .level-dot.active {
    background: var(--el-color-primary);
  }
</style>
