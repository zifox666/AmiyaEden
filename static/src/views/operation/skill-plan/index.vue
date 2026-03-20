<!-- 用户技能规划查看页面 -->
<template>
  <div class="skill-plan-user art-full-height">
    <!-- 顶部选择 -->
    <ElCard shadow="never" class="mb-2">
      <div class="flex items-center gap-3 flex-wrap">
        <ElSelect
          v-model="selectedPlanId"
          :placeholder="$t('skillPlan.selectPlan')"
          style="width: 260px"
          @change="onPlanChange"
        >
          <ElOption v-for="p in planList" :key="p.id" :value="p.id" :label="p.name" />
        </ElSelect>
        <ElButton type="primary" :loading="checking" :disabled="!selectedPlanId" @click="doCheck">
          <el-icon class="mr-1"><Search /></el-icon>
          {{ $t('skillPlan.checkMyCharacters') }}
        </ElButton>
      </div>
    </ElCard>

    <!-- 规划说明 -->
    <ElCard v-if="selectedPlan" shadow="never" class="mb-2">
      <div
        class="text-sm"
        :style="{
          color: selectedPlan.description
            ? 'var(--el-text-color-secondary)'
            : 'var(--el-text-color-placeholder)'
        }"
      >
        {{ selectedPlan.description || $t('skillPlan.noDescription') }}
      </div>
    </ElCard>

    <!-- 角色技能结果 -->
    <template v-if="checkResult && checkResult.characters.length > 0">
      <ElCard
        v-for="char in checkResult.characters"
        :key="char.character_id"
        shadow="never"
        class="mb-2 char-card"
      >
        <template #header>
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <ElAvatar
                :src="`https://images.evetech.net/characters/${char.character_id}/portrait?size=64`"
                :size="40"
              />
              <div>
                <div class="font-medium">{{ char.character_name }}</div>
                <div class="text-xs" style="color: var(--el-text-color-secondary)">
                  {{ $t('skillPlan.columns.progress') }}: {{ char.satisfied }} / {{ char.total }}
                </div>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <ElButton
                v-if="char.missing_skills && char.missing_skills.length > 0"
                size="small"
                :icon="CopyDocument"
                @click="copyMissingSkills(char)"
              >
                {{ $t('skillPlan.copyMissing') }}
              </ElButton>
              <ElTag :type="char.status === 'satisfied' ? 'success' : 'warning'" size="default">
                {{
                  char.status === 'satisfied'
                    ? $t('skillPlan.status.satisfied')
                    : $t('skillPlan.status.unsatisfied')
                }}
              </ElTag>
            </div>
          </div>
        </template>

        <!-- 进度条 -->
        <ElProgress
          :percentage="char.total > 0 ? Math.round((char.satisfied / char.total) * 100) : 0"
          :color="char.status === 'satisfied' ? '#67c23a' : '#e6a23c'"
          :stroke-width="12"
          class="mb-4"
        />

        <!-- 缺少的技能明细 -->
        <template v-if="char.missing_skills && char.missing_skills.length > 0">
          <div class="missing-title mb-2">{{ $t('skillPlan.missingSkillsTitle') }}</div>
          <div class="missing-list">
            <div v-for="(ms, idx) in char.missing_skills" :key="idx" class="missing-item">
              <span class="skill-name">{{ ms.skill_name }}</span>
              <span class="skill-level-info">
                Lv{{ ms.current_level }}
                <el-icon class="mx-1"><Right /></el-icon>
                Lv{{ ms.required_level }}
              </span>
            </div>
          </div>
        </template>
        <div v-else class="text-center py-2" style="color: var(--el-color-success)">
          {{ $t('skillPlan.allSatisfied') }}
        </div>
      </ElCard>
    </template>

    <!-- 空状态 -->
    <ElCard v-if="!checking && !checkResult" shadow="never" class="flex-1">
      <ElEmpty :description="$t('skillPlan.selectAndCheck')" />
    </ElCard>
    <ElCard v-if="checkResult && checkResult.characters.length === 0" shadow="never" class="flex-1">
      <ElEmpty :description="$t('skillPlan.noCharacters')" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted } from 'vue'
  import { fetchAllSkillPlans, checkMyCharacters } from '@/api/skill-plan'
  import {
    ElCard,
    ElSelect,
    ElOption,
    ElButton,
    ElAvatar,
    ElTag,
    ElProgress,
    ElEmpty,
    ElMessage
  } from 'element-plus'
  import { Search, Right, CopyDocument } from '@element-plus/icons-vue'
  import { useI18n } from 'vue-i18n'

  defineOptions({ name: 'UserSkillPlan' })

  const { t, locale } = useI18n()

  const planList = ref<Api.SkillPlan.SkillPlanDTO[]>([])
  const selectedPlanId = ref<number | ''>('')
  const checking = ref(false)
  const checkResult = ref<Api.SkillPlan.SkillCheckSummary | null>(null)

  const selectedPlan = computed(() => planList.value.find((p) => p.id === selectedPlanId.value))

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

  function copyMissingSkills(char: Api.SkillPlan.SkillCheckCharacterResult) {
    if (!char.missing_skills || char.missing_skills.length === 0) {
      ElMessage.info(t('skillPlan.copyEmpty'))
      return
    }
    const text = char.missing_skills.map((ms) => `${ms.skill_name} ${ms.required_level}`).join('\n')
    navigator.clipboard.writeText(text).then(() => {
      ElMessage.success(t('skillPlan.copySuccess'))
    })
  }

  async function doCheck() {
    if (!selectedPlanId.value) return
    checking.value = true
    try {
      const lang = locale.value === 'en' ? 'en' : 'zh'
      checkResult.value = await checkMyCharacters(selectedPlanId.value as number, lang)
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : String(e)
      ElMessage.error(msg)
    } finally {
      checking.value = false
    }
  }
</script>

<style scoped lang="scss">
  .char-card {
    :deep(.el-card__header) {
      padding: 12px 20px;
    }
  }
  .missing-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-regular);
  }
  .missing-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 8px;
  }
  .missing-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 12px;
    border-radius: 6px;
    background: var(--el-fill-color-light);
    font-size: 13px;
  }
  .skill-name {
    font-weight: 500;
    color: var(--el-text-color-primary);
  }
  .skill-level-info {
    display: flex;
    align-items: center;
    color: var(--el-color-warning);
    font-weight: 600;
    font-size: 12px;
  }
</style>
