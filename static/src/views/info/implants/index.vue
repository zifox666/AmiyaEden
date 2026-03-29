<template>
  <div class="info-implants-page art-full-height">
    <!-- 顶部：人物选择器 -->
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
        <div v-if="implantsData" class="flex items-center gap-4 text-sm text-gray-500">
          <span
            >{{ $t('info.jumpCloneCount') }}:
            <strong class="text-blue-400">{{ implantsData.jump_clones.length }}</strong></span
          >
        </div>
      </div>
    </ElCard>

    <!-- 主体区域 -->
    <div v-loading="loading" class="implants-main">
      <div class="implants-panel">
        <template v-if="implantsData">
          <!-- 跳跃疲劳信息 -->
          <div class="fatigue-bar">
            <!-- 跳跃就绪状态标签 -->
            <div class="fatigue-item">
              <span class="fatigue-label">{{ $t('info.jumpFatigueExpire') }}:</span>
              <el-tag v-if="isFatigueExpired" type="success" size="small" round>
                ✓ {{ $t('info.jumpReady') }}
              </el-tag>
              <el-tag v-else type="warning" size="small" round>
                ⏳ {{ $t('info.jumpFatigueRemaining') }}{{ fatigueRemaining }}
              </el-tag>
            </div>
            <div class="fatigue-item" v-if="implantsData.last_jump_date">
              <span class="fatigue-label">{{ $t('info.lastJumpDate') }}:</span>
              <span class="fatigue-value">{{ formatTime(implantsData.last_jump_date) }}</span>
            </div>
            <div class="fatigue-item" v-if="implantsData.last_clone_jump_date">
              <span class="fatigue-label">{{ $t('info.lastCloneJump') }}:</span>
              <span class="fatigue-value">{{ formatTime(implantsData.last_clone_jump_date) }}</span>
            </div>
          </div>

          <!-- 基地空间站信息 -->
          <div class="home-station" v-if="implantsData.home_location">
            <div class="section-header">
              <el-icon><HomeFilled /></el-icon>
              <span>{{ $t('info.homeStation') }}</span>
            </div>
            <div class="location-text">
              {{
                implantsData.home_location.location_name ||
                `${implantsData.home_location.location_type}-${implantsData.home_location.location_id}`
              }}
            </div>
          </div>

          <!-- 当前活跃植入体 -->
          <div class="clone-section">
            <div class="section-header">
              <el-icon><Cpu /></el-icon>
              <span>{{ $t('info.activeImplants') }}</span>
              <span class="implant-count">({{ implantsData.active_implants.length }})</span>
            </div>
            <div class="implant-list" v-if="implantsData.active_implants.length > 0">
              <div
                v-for="implant in implantsData.active_implants"
                :key="implant.implant_id"
                class="implant-item"
              >
                <img
                  :src="`https://images.evetech.net/types/${implant.implant_id}/icon?size=32`"
                  class="implant-icon"
                  loading="lazy"
                />
                <span class="implant-name">{{
                  implant.implant_name || `Type ${implant.implant_id}`
                }}</span>
              </div>
            </div>
            <div v-else class="no-implants">{{ $t('info.noImplants') }}</div>
          </div>

          <!-- 跳跃克隆体列表 -->
          <div class="section-header clone-section-title">
            <el-icon><Connection /></el-icon>
            <span>{{ $t('info.jumpClones') }}</span>
            <span class="implant-count">({{ implantsData.jump_clones.length }})</span>
          </div>

          <div v-if="implantsData.jump_clones.length > 0" class="jump-clones-list">
            <div
              v-for="clone in implantsData.jump_clones"
              :key="clone.jump_clone_id"
              class="clone-card"
            >
              <div class="clone-card-header">
                <div class="clone-location">
                  <el-icon class="location-icon"><Location /></el-icon>
                  <span class="location-name">
                    {{
                      clone.location.location_name ||
                      `${clone.location.location_type}-${clone.location.location_id}`
                    }}
                  </span>
                </div>
                <span class="clone-id">#{{ clone.jump_clone_id }}</span>
              </div>
              <div class="clone-implants" v-if="clone.implants.length > 0">
                <div
                  v-for="implant in clone.implants"
                  :key="implant.implant_id"
                  class="implant-item"
                >
                  <img
                    :src="`https://images.evetech.net/types/${implant.implant_id}/icon?size=32`"
                    class="implant-icon"
                    loading="lazy"
                  />
                  <span class="implant-name">{{
                    implant.implant_name || `Type ${implant.implant_id}`
                  }}</span>
                </div>
              </div>
              <div v-else class="no-implants">{{ $t('info.noImplants') }}</div>
            </div>
          </div>

          <ElEmpty v-else :description="$t('info.noJumpClones')" />
        </template>

        <ElEmpty v-else-if="!loading" :description="$t('info.noImplantData')" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted } from 'vue'
  import { Refresh, HomeFilled, Location, Connection } from '@element-plus/icons-vue'
  import { ElCard, ElSelect, ElOption, ElAvatar, ElButton, ElEmpty, ElTag } from 'element-plus'
  import { fetchMyCharacters } from '@/api/auth'
  import { fetchInfoImplants } from '@/api/eve-info'
  import { formatTime } from '@utils/common'
  import { useUserStore } from '@/store/modules/user'

  // Cpu icon (Element Plus doesn't have Cpu, use Monitor as substitute)
  const Cpu = {
    name: 'Cpu',
    render() {
      return h('svg', { viewBox: '0 0 1024 1024', xmlns: 'http://www.w3.org/2000/svg' }, [
        h('path', {
          fill: 'currentColor',
          d: 'M320 256a64 64 0 0 0-64 64v384a64 64 0 0 0 64 64h384a64 64 0 0 0 64-64V320a64 64 0 0 0-64-64H320zm0-64h384a128 128 0 0 1 128 128v384a128 128 0 0 1-128 128H320a128 128 0 0 1-128-128V320a128 128 0 0 1 128-128z'
        }),
        h('path', {
          fill: 'currentColor',
          d: 'M512 64a32 32 0 0 1 32 32v128a32 32 0 0 1-64 0V96a32 32 0 0 1 32-32zm160 0a32 32 0 0 1 32 32v128a32 32 0 0 1-64 0V96a32 32 0 0 1 32-32zM352 64a32 32 0 0 1 32 32v128a32 32 0 0 1-64 0V96a32 32 0 0 1 32-32zm160 736a32 32 0 0 1 32 32v96a32 32 0 0 1-64 0v-96a32 32 0 0 1 32-32zm160 0a32 32 0 0 1 32 32v96a32 32 0 0 1-64 0v-96a32 32 0 0 1 32-32zM352 800a32 32 0 0 1 32 32v96a32 32 0 0 1-64 0v-96a32 32 0 0 1 32-32zM64 512a32 32 0 0 1 32-32h128a32 32 0 0 1 0 64H96a32 32 0 0 1-32-32zm0 160a32 32 0 0 1 32-32h128a32 32 0 0 1 0 64H96a32 32 0 0 1-32-32zM64 352a32 32 0 0 1 32-32h128a32 32 0 0 1 0 64H96a32 32 0 0 1-32-32zm736 160a32 32 0 0 1 32-32h96a32 32 0 0 1 0 64h-96a32 32 0 0 1-32-32zm0 160a32 32 0 0 1 32-32h96a32 32 0 0 1 0 64h-96a32 32 0 0 1-32-32zm0-320a32 32 0 0 1 32-32h96a32 32 0 0 1 0 64h-96a32 32 0 0 1-32-32zM384 384h256v256H384V384z'
        })
      ])
    }
  }

  import { h } from 'vue'

  defineOptions({ name: 'EveInfoImplants' })

  const userStore = useUserStore()

  // ---- 数据 ----
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const selectedCharacterId = ref<number>()
  const implantsData = ref<Api.EveInfo.ImplantsResponse | null>(null)
  const loading = ref(false)

  // ---- 计算属性 ----
  const isFatigueExpired = computed(() => {
    if (!implantsData.value?.jump_fatigue_expire) return true
    return new Date(implantsData.value.jump_fatigue_expire) <= new Date()
  })

  const fatigueRemaining = computed(() => {
    if (!implantsData.value?.jump_fatigue_expire) return ''
    const expire = new Date(implantsData.value.jump_fatigue_expire)
    const now = new Date()
    const diff = expire.getTime() - now.getTime()
    if (diff <= 0) return ''
    const hours = Math.floor(diff / (1000 * 60 * 60))
    const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
    if (hours > 0) return `${hours}h ${minutes}m`
    return `${minutes}m`
  })

  // ---- 工具方法 ----
  // ---- 交互 ----
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
      implantsData.value = await fetchInfoImplants({
        character_id: selectedCharacterId.value,
        language: userStore.language
      })
    } catch {
      implantsData.value = null
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

<style scoped lang="scss">
  .info-implants-page {
    display: flex;
    flex-direction: column;
    height: 100%;
    box-sizing: border-box;
  }

  .implants-main {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .implants-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-light);
    border-radius: 6px;
    padding: 16px;
    overflow-y: auto;
    scrollbar-width: thin;
    scrollbar-color: transparent transparent;
  }

  .implants-panel:hover {
    scrollbar-color: rgba(144, 147, 153, 0.4) transparent;
  }

  .implants-panel::-webkit-scrollbar {
    width: 4px;
  }

  .implants-panel::-webkit-scrollbar-thumb {
    background: transparent;
    border-radius: 2px;
  }

  .implants-panel:hover::-webkit-scrollbar-thumb {
    background: rgba(144, 147, 153, 0.4);
  }

  /* ===== 疲劳信息 ===== */
  .fatigue-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 16px;
    padding: 10px 14px;
    margin-bottom: 12px;
    background: var(--el-fill-color-lighter);
    border-radius: 6px;
    font-size: 13px;
  }

  .fatigue-item {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .fatigue-label {
    color: var(--el-text-color-secondary);
  }

  .fatigue-value {
    color: var(--el-text-color-primary);
    font-weight: 500;
  }

  .fatigue-countdown {
    font-size: 12px;
    margin-left: 4px;
  }

  /* ===== 基地空间站信息 ===== */
  .home-station {
    padding: 10px 14px;
    margin-bottom: 12px;
    background: var(--el-fill-color-lighter);
    border-radius: 6px;
  }

  .location-text {
    font-size: 13px;
    color: var(--el-text-color-primary);
    margin-top: 6px;
    padding-left: 22px;
  }

  /* ===== 通用标题 ===== */
  .section-header {
    display: flex;
    align-items: center;
    gap: 6px;
    font-weight: 600;
    font-size: 14px;
    color: var(--el-text-color-primary);
    margin-bottom: 8px;
  }

  .clone-section-title {
    margin-top: 16px;
    margin-bottom: 12px;
  }

  .implant-count {
    font-weight: 400;
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }

  /* ===== 克隆区块 ===== */
  .clone-section {
    padding: 10px 14px;
    margin-bottom: 12px;
    background: var(--el-fill-color-lighter);
    border-radius: 6px;
  }

  /* ===== 植入体列表 ===== */
  .implant-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .implant-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 8px;
    border-radius: 4px;
    transition: background 0.15s;
  }

  .implant-item:hover {
    background: var(--el-fill-color);
  }

  .implant-icon {
    width: 24px;
    height: 24px;
    border-radius: 3px;
    flex-shrink: 0;
  }

  .implant-name {
    font-size: 13px;
    color: var(--el-text-color-primary);
    line-height: 1.4;
  }

  .no-implants {
    font-size: 12px;
    color: var(--el-text-color-placeholder);
    padding: 4px 8px;
  }

  /* ===== 跳跃克隆体卡片 ===== */
  .jump-clones-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .clone-card {
    background: var(--el-fill-color-lighter);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    padding: 12px 14px;
    transition: border-color 0.15s;
  }

  .clone-card:hover {
    border-color: var(--el-border-color);
  }

  .clone-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
  }

  .clone-location {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
  }

  .location-icon {
    color: var(--el-color-primary);
    font-size: 16px;
  }

  .location-name {
    word-break: break-word;
  }

  .clone-id {
    font-size: 12px;
    color: var(--el-text-color-placeholder);
    flex-shrink: 0;
    margin-left: 12px;
  }

  .clone-implants {
    padding-left: 4px;
  }

  /* ===== 响应式 ===== */
  @media (max-width: 768px) {
    .fatigue-bar {
      flex-direction: column;
      gap: 6px;
    }
  }
</style>
