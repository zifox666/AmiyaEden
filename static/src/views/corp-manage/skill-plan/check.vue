<!-- 技能规划检查页面（管理员/FC） -->
<template>
  <div class="skill-plan-check art-full-height">
    <!-- 顶部操作栏 -->
    <ElCard shadow="never" class="mb-2">
      <div class="flex items-center justify-between flex-wrap gap-4">
        <div class="flex items-center gap-3">
          <ElSelect
            v-model="selectedPlanId"
            :placeholder="$t('skillPlan.selectPlan')"
            style="width: 260px"
            @change="onPlanChange"
          >
            <ElOption v-for="plan in planList" :key="plan.id" :value="plan.id" :label="plan.name" />
          </ElSelect>
          <ElButton type="primary" :loading="checking" :disabled="!selectedPlanId" @click="doCheck">
            <el-icon class="mr-1"><Search /></el-icon>
            {{ $t('skillPlan.checkAll') }}
          </ElButton>
        </div>
        <ElButton :disabled="!checkResult" @click="exportResults">
          <el-icon class="mr-1"><Download /></el-icon>
          {{ $t('skillPlan.exportResults') }}
        </ElButton>
      </div>
    </ElCard>

    <!-- 统计卡片 -->
    <ElCard v-if="checkResult" shadow="never" class="mb-2">
      <div class="summary-grid">
        <div class="summary-item">
          <span class="summary-label">{{ $t('skillPlan.summary.planName') }}</span>
          <span class="summary-value plan-name">{{ checkResult.plan_name }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">{{ $t('skillPlan.summary.totalCharacters') }}</span>
          <span class="summary-value">{{ checkResult.total_characters }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">{{ $t('skillPlan.summary.satisfiedCount') }}</span>
          <span class="summary-value satisfied">{{ checkResult.satisfied_count }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">{{ $t('skillPlan.summary.unsatisfiedCount') }}</span>
          <span class="summary-value unsatisfied">{{ checkResult.unsatisfied_count }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">{{ $t('skillPlan.summary.satisfiedRate') }}</span>
          <span class="summary-value">{{ checkResult.satisfied_rate.toFixed(1) }}%</span>
        </div>
      </div>
    </ElCard>

    <!-- 过滤和表格 -->
    <ElCard v-if="checkResult" shadow="never" class="flex-1">
      <div class="flex items-center justify-between mb-3">
        <div class="flex gap-2">
          <ElButton
            type="primary"
            :plain="activeFilter !== 'all'"
            size="small"
            round
            @click="activeFilter = 'all'"
          >
            {{ $t('skillPlan.filter.all') }} ({{ checkResult.characters.length }})
          </ElButton>
          <ElButton
            type="success"
            :plain="activeFilter !== 'satisfied'"
            size="small"
            round
            @click="activeFilter = 'satisfied'"
          >
            {{ $t('skillPlan.filter.satisfied') }} ({{ checkResult.satisfied_count }})
          </ElButton>
          <ElButton
            type="warning"
            :plain="activeFilter !== 'unsatisfied'"
            size="small"
            round
            @click="activeFilter = 'unsatisfied'"
          >
            {{ $t('skillPlan.filter.unsatisfied') }} ({{ checkResult.unsatisfied_count }})
          </ElButton>
        </div>
        <ElInput
          v-model="searchKeyword"
          :placeholder="$t('skillPlan.searchPlaceholder')"
          clearable
          style="width: 240px"
          size="small"
          :prefix-icon="Search"
        />
      </div>

      <ElTable
        :data="filteredCharacters"
        stripe
        style="width: 100%"
        max-height="calc(100vh - 380px)"
      >
        <ElTableColumn prop="user_name" :label="$t('skillPlan.columns.userName')" width="140" />
        <ElTableColumn
          prop="character_name"
          :label="$t('skillPlan.columns.characterName')"
          width="180"
        />
        <ElTableColumn :label="$t('skillPlan.columns.progress')" width="140">
          <template #default="{ row }">
            <span :class="row.status === 'satisfied' ? 'text-success' : 'text-warning'">
              {{ row.satisfied }} / {{ row.total }}
            </span>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('skillPlan.columns.status')" width="100">
          <template #default="{ row }">
            <ElTag :type="row.status === 'satisfied' ? 'success' : 'warning'" size="small">
              {{
                row.status === 'satisfied'
                  ? $t('skillPlan.status.satisfied')
                  : $t('skillPlan.status.unsatisfied')
              }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn :label="$t('skillPlan.columns.missingSkills')" min-width="400">
          <template #default="{ row }">
            <span v-if="row.missing_skills && row.missing_skills.length > 0" class="missing-text">
              <span v-for="(ms, idx) in row.missing_skills" :key="idx">
                {{ ms.skill_name }} Lv{{ ms.required_level }}({{ $t('skillPlan.currentLevel') }}:{{
                  ms.current_level
                }}){{ (idx as number) < row.missing_skills.length - 1 ? ', ' : '' }}
              </span>
            </span>
            <span v-else class="text-success">-</span>
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted } from 'vue'
  import { fetchAllSkillPlans, checkAllCharacters } from '@/api/skill-plan'
  import {
    ElCard,
    ElSelect,
    ElOption,
    ElButton,
    ElInput,
    ElTable,
    ElTableColumn,
    ElTag,
    ElMessage
  } from 'element-plus'
  import { Search, Download } from '@element-plus/icons-vue'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'SkillPlanCheck' })

  const { t, locale } = useI18n()

  const planList = ref<Api.SkillPlan.SkillPlanDTO[]>([])
  const selectedPlanId = ref<number | ''>('')
  const checking = ref(false)
  const checkResult = ref<Api.SkillPlan.SkillCheckSummary | null>(null)
  const activeFilter = ref<'all' | 'satisfied' | 'unsatisfied'>('all')
  const searchKeyword = ref('')

  onMounted(async () => {
    try {
      planList.value = (await fetchAllSkillPlans()) ?? []
    } catch {
      /* empty */
    }
  })

  function onPlanChange() {
    checkResult.value = null
  }

  async function doCheck() {
    if (!selectedPlanId.value) return
    checking.value = true
    try {
      const lang = locale.value === 'en' ? 'en' : 'zh'
      checkResult.value = await checkAllCharacters(selectedPlanId.value as number, lang)
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e)
      ElMessage.error(msg)
    } finally {
      checking.value = false
    }
  }

  const filteredCharacters = computed(() => {
    if (!checkResult.value) return []
    let list = checkResult.value.characters
    if (activeFilter.value === 'satisfied') {
      list = list.filter((c) => c.status === 'satisfied')
    } else if (activeFilter.value === 'unsatisfied') {
      list = list.filter((c) => c.status === 'unsatisfied')
    }
    if (searchKeyword.value) {
      const kw = searchKeyword.value.toLowerCase()
      list = list.filter(
        (c) => c.character_name.toLowerCase().includes(kw) || c.user_name.toLowerCase().includes(kw)
      )
    }
    return list
  })

  function exportResults() {
    if (!checkResult.value) return
    const rows = checkResult.value.characters.map((c) => ({
      [t('skillPlan.columns.userName')]: c.user_name,
      [t('skillPlan.columns.characterName')]: c.character_name,
      [t('skillPlan.columns.progress')]: `${c.satisfied}/${c.total}`,
      [t('skillPlan.columns.status')]:
        c.status === 'satisfied'
          ? t('skillPlan.status.satisfied')
          : t('skillPlan.status.unsatisfied'),
      [t('skillPlan.columns.missingSkills')]: (c.missing_skills ?? [])
        .map(
          (ms) =>
            `${ms.skill_name} Lv${ms.required_level}(${t('skillPlan.currentLevel')}:${ms.current_level})`
        )
        .join(', ')
    }))

    const headers = Object.keys(rows[0] ?? {})
    const csv = [
      headers.join(','),
      ...rows.map((r) =>
        headers
          .map((h) => `"${String((r as Record<string, string>)[h]).replace(/"/g, '""')}"`)
          .join(',')
      )
    ].join('\n')

    const BOM = '\uFEFF'
    const blob = new Blob([BOM + csv], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${checkResult.value.plan_name}_check_result.csv`
    a.click()
    URL.revokeObjectURL(url)
  }
</script>

<style scoped lang="scss">
  .summary-grid {
    display: flex;
    gap: 32px;
    flex-wrap: wrap;
  }
  .summary-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .summary-label {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }
  .summary-value {
    font-size: 24px;
    font-weight: 700;
    color: var(--el-text-color-primary);
    &.plan-name {
      font-size: 18px;
    }
    &.satisfied {
      color: var(--el-color-success);
    }
    &.unsatisfied {
      color: var(--el-color-warning);
    }
  }
  .missing-text {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 1.6;
  }
  .text-success {
    color: var(--el-color-success);
  }
  .text-warning {
    color: var(--el-color-warning);
  }
</style>
