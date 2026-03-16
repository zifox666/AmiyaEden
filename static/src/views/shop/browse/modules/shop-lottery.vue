<!-- 抽奖页面（用户端） -->
<template>
  <div class="lottery-page">
    <!-- 加载中 -->
    <div v-if="loading" v-loading="true" class="loading-placeholder" />

    <!-- 无活动 -->
    <ElEmpty
      v-else-if="activities.length === 0"
      :description="t('lottery.noActivities')"
      class="mt-8"
    />

    <!-- 活动卡片网格 -->
    <div v-else class="activity-grid">
      <div
        v-for="act in activities"
        :key="act.id"
        class="activity-card"
        @click="openDrawDialog(act)"
      >
        <div class="activity-cover">
          <img v-if="act.image" :src="act.image" :alt="act.name" />
          <div v-else class="activity-cover-placeholder">
            <el-icon :size="48"><Trophy /></el-icon>
          </div>
          <div class="activity-overlay">
            <ElButton type="warning" round size="large">{{ t('lottery.startDraw') }}</ElButton>
          </div>
        </div>
        <div class="activity-body">
          <h3 class="activity-name">{{ act.name }}</h3>
          <p v-if="act.description" class="activity-desc">{{ act.description }}</p>
          <div class="activity-meta">
            <span v-if="act.cost_per_draw > 0" class="cost-badge">
              {{ act.cost_per_draw }} {{ t('lottery.points') }} / {{ t('lottery.perDraw') }}
            </span>
            <span v-else class="cost-badge free">{{ t('lottery.free') }}</span>
          </div>
          <!-- 奖品预览 -->
          <div class="prize-chips">
            <ElTag
              v-for="prize in act.prizes?.slice(0, 5)"
              :key="prize.id"
              :type="TIER_TAG_TYPE[prize.tier] ?? 'info'"
              size="small"
              effect="plain"
            >
              <img v-if="prize.image" :src="prize.image" class="prize-chip-img" />
              {{ prize.name }}
            </ElTag>
            <span v-if="act.prizes && act.prizes.length > 5" class="text-xs text-gray-400">
              +{{ act.prizes.length - 5 }} {{ t('lottery.more') }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- 我的抽奖记录 -->
    <ElCard v-if="myRecords.length > 0" class="mt-6" shadow="never">
      <template #header>
        <span class="font-semibold">{{ t('lottery.myRecords') }}</span>
      </template>
      <div class="record-list">
        <div v-for="rec in myRecords" :key="rec.id" class="record-item">
          <div class="flex items-center gap-2">
            <img v-if="rec.prize_image" :src="rec.prize_image" class="record-prize-img" />
            <ElTag :type="TIER_TAG_TYPE[rec.prize_tier] ?? 'info'" size="small">
              {{ TIER_LABELS[rec.prize_tier] ?? rec.prize_tier }}
            </ElTag>
            <span class="font-medium">{{ rec.prize_name }}</span>
            <ElTag
              :type="rec.delivery_status === 'delivered' ? 'success' : 'warning'"
              size="small"
              effect="plain"
              class="ml-1"
            >
              {{
                rec.delivery_status === 'delivered'
                  ? t('lottery.delivery.delivered')
                  : t('lottery.delivery.pending')
              }}
            </ElTag>
            <span class="text-xs text-gray-400 ml-auto">{{ formatTime(rec.created_at) }}</span>
          </div>
        </div>
      </div>
    </ElCard>

    <!-- 抽奖对话框 -->
    <ElDialog
      v-model="drawDialogVisible"
      :title="selectedActivity?.name ?? t('lottery.draw')"
      width="520px"
      :close-on-click-modal="drawState === 'idle'"
      :close-on-press-escape="drawState === 'idle'"
      destroy-on-close
      align-center
    >
      <div class="draw-modal">
        <!-- 待机状态：活动信息 + 抽奖按钮 -->
        <template v-if="drawState === 'idle'">
          <div v-if="selectedActivity?.image" class="draw-cover mb-4">
            <img :src="selectedActivity.image" :alt="selectedActivity.name" />
          </div>
          <div class="draw-info mb-4">
            <div v-if="selectedActivity?.description" class="text-sm text-gray-400 mb-3">
              {{ selectedActivity.description }}
            </div>
            <div class="flex items-center gap-2 text-sm">
              <span>{{ t('lottery.costLabel') }}</span>
              <strong
                v-if="selectedActivity && selectedActivity.cost_per_draw > 0"
                class="text-orange-500 text-base"
              >
                {{ selectedActivity.cost_per_draw }} {{ t('lottery.points') }}
              </strong>
              <strong v-else class="text-green-500">{{ t('lottery.free') }}</strong>
              <span class="ml-auto text-gray-400"
                >{{ t('lottery.balance') }}: <strong>{{ walletBalance ?? '-' }}</strong></span
              >
            </div>
          </div>
          <!-- 奖品池 -->
          <div v-if="selectedActivity?.prizes?.length" class="prize-pool mb-4">
            <div class="text-xs text-gray-400 mb-2">{{ t('lottery.prizePool') }}</div>
            <div class="flex flex-wrap gap-2">
              <div
                v-for="prize in selectedActivity.prizes"
                :key="prize.id"
                class="prize-pool-item"
                :class="`tier-${prize.tier}`"
              >
                <img v-if="prize.image" :src="prize.image" class="prize-pool-img" />
                <span>{{ prize.name }}</span>
                <span class="prize-pool-prob"
                  >{{ getProbPercent(prize, selectedActivity.prizes) }}%</span
                >
              </div>
            </div>
          </div>
          <ElButton
            type="warning"
            size="large"
            class="w-full"
            style="font-size: 18px; height: 48px"
            @click="startDraw"
          >
            {{ t('lottery.startDraw') }}
          </ElButton>
        </template>

        <!-- 抽奖动画：奖品滚动 -->
        <template v-else-if="drawState === 'drawing'">
          <div class="draw-animation">
            <div class="slot-machine-wrapper">
              <div class="slot-machine-viewport">
                <div class="slot-machine-strip" :style="stripStyle">
                  <div
                    v-for="(prize, idx) in rollingPrizes"
                    :key="idx"
                    class="slot-machine-item"
                    :class="`tier-${prize.tier}`"
                  >
                    <img v-if="prize.image" :src="prize.image" class="slot-prize-img" />
                    <el-icon v-else :size="40"><Trophy /></el-icon>
                    <span class="slot-prize-name">{{ prize.name }}</span>
                  </div>
                </div>
              </div>
              <div class="slot-indicator" />
            </div>
            <div class="draw-animation-text">{{ t('lottery.drawing') }}</div>
          </div>
        </template>

        <!-- 结果展示 -->
        <template v-else-if="drawState === 'result' && drawResult">
          <div class="draw-result" :class="`result-${drawResult.prize.tier}`">
            <!-- 粒子效果 -->
            <div class="particles" v-if="drawResult.prize.tier !== 'normal'">
              <div v-for="n in 12" :key="n" class="particle" :style="getParticleStyle(n)" />
            </div>
            <!-- 奖品展示 -->
            <div class="result-icon-wrap">
              <div class="result-tier-ring" :class="`ring-${drawResult.prize.tier}`">
                <img
                  v-if="drawResult.prize.image"
                  :src="drawResult.prize.image"
                  class="result-prize-img"
                />
                <el-icon v-else :size="56"><Trophy /></el-icon>
              </div>
            </div>
            <div class="result-tier-label" :class="`label-${drawResult.prize.tier}`">
              {{ TIER_LABELS[drawResult.prize.tier] ?? drawResult.prize.tier }}
            </div>
            <div class="result-name">{{ drawResult.prize.name }}</div>
            <div class="flex gap-2 mt-5">
              <ElButton @click="drawDialogVisible = false">{{ t('lottery.close') }}</ElButton>
              <ElButton type="warning" @click="startDraw">{{ t('lottery.drawAgain') }}</ElButton>
            </div>
          </div>
        </template>
      </div>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { ElTag, ElButton, ElEmpty, ElMessage } from 'element-plus'
  import { Trophy } from '@element-plus/icons-vue'
  import { useI18n } from 'vue-i18n'
  import { fetchLotteryActivities, drawLottery, fetchMyLotteryRecords } from '@/api/shop'
  import { fetchMyWallet } from '@/api/sys-wallet'

  defineOptions({ name: 'ShopLottery' })

  const { t } = useI18n()

  type LotteryActivity = Api.Shop.LotteryActivity
  type LotteryPrize = Api.Shop.LotteryPrize

  const TIER_TAG_TYPE: Record<string, any> = {
    normal: 'info',
    rare: 'warning',
    legendary: 'danger'
  }
  const TIER_LABELS = computed<Record<string, string>>(() => ({
    normal: t('lottery.tier.normal'),
    rare: t('lottery.tier.rare'),
    legendary: t('lottery.tier.legendary')
  }))

  // ─── 活动列表 ───
  const loading = ref(false)
  const activities = ref<LotteryActivity[]>([])

  async function loadActivities() {
    loading.value = true
    try {
      const res = (await fetchLotteryActivities({ current: 1, size: 50 })) as any
      activities.value = res?.list ?? []
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('lottery.loadFailed'))
    } finally {
      loading.value = false
    }
  }

  // ─── 我的记录 ───
  const myRecords = ref<Api.Shop.LotteryRecord[]>([])

  async function loadMyRecords() {
    try {
      const res = (await fetchMyLotteryRecords({ current: 1, size: 10 })) as any
      myRecords.value = res?.list ?? []
    } catch {
      /* ignore */
    }
  }

  function formatTime(t: string) {
    return new Date(t).toLocaleString('zh-CN', { hour12: false })
  }

  // ─── 钱包余额 ───
  const walletBalance = ref<number | null>(null)

  async function loadWallet() {
    try {
      const res = (await fetchMyWallet()) as any
      walletBalance.value = res?.balance ?? null
    } catch {
      /* ignore */
    }
  }

  // ─── 抽奖对话框 ───
  const drawDialogVisible = ref(false)
  const selectedActivity = ref<LotteryActivity | null>(null)
  const drawState = ref<'idle' | 'drawing' | 'result'>('idle')
  const drawResult = ref<Api.Shop.DrawResult | null>(null)

  function openDrawDialog(act: LotteryActivity) {
    selectedActivity.value = act
    drawState.value = 'idle'
    drawResult.value = null
    drawDialogVisible.value = true
    loadWallet()
  }

  // ─── 滚动动画相关 ───
  const ITEM_HEIGHT = 80 // 每个奖品卡片高度 px
  const rollingPrizes = ref<LotteryPrize[]>([])
  const stripStyle = ref<Record<string, string>>({})

  function buildRollingPrizes(prizes: LotteryPrize[], winnerIndex: number): LotteryPrize[] {
    // 构建一条长列表：先随机重复 N 轮，最后一轮以 winner 结尾
    const rounds = 4
    const items: LotteryPrize[] = []
    for (let r = 0; r < rounds; r++) {
      const shuffled = [...prizes].sort(() => Math.random() - 0.5)
      items.push(...shuffled)
    }
    // 最后追加一轮随机顺序，确保 winner 在末尾
    const lastRound = [...prizes]
      .filter((_, i) => i !== winnerIndex)
      .sort(() => Math.random() - 0.5)
    lastRound.push(prizes[winnerIndex])
    items.push(...lastRound)
    return items
  }

  async function startDraw() {
    if (!selectedActivity.value) return
    const prizes = selectedActivity.value.prizes ?? []
    if (prizes.length === 0) return

    drawState.value = 'drawing'
    drawResult.value = null

    // 立刻开始快速滚动（无限循环预览）
    const previewItems = [...prizes, ...prizes, ...prizes].sort(() => Math.random() - 0.5)
    rollingPrizes.value = previewItems
    stripStyle.value = {
      transition: 'none',
      transform: 'translateY(0)'
    }

    // 开始 API 请求
    let result: Api.Shop.DrawResult | null = null
    let error: string | null = null
    try {
      result = (await drawLottery(selectedActivity.value.id)) as any
    } catch (e: any) {
      error = e?.message ?? t('lottery.drawFailed')
    }

    if (error || !result) {
      drawState.value = 'idle'
      ElMessage.error(error ?? t('lottery.drawFailed'))
      return
    }

    // 找到中奖奖品在 prizes 中的下标
    const winnerIdx = prizes.findIndex((p) => p.id === result!.prize.id)
    const safeWinnerIdx = winnerIdx >= 0 ? winnerIdx : 0

    // 构建最终滚动列表
    const fullList = buildRollingPrizes(prizes, safeWinnerIdx)
    rollingPrizes.value = fullList

    // 需要滚动到倒数第1个（winner）位于视窗中央
    const winnerPos = fullList.length - 1
    const viewportCenter = 1 // 视窗显示3个，中间是第2个（index 1）
    const targetOffset = (winnerPos - viewportCenter) * ITEM_HEIGHT

    await nextTick()
    // 先无动画跳到顶部
    stripStyle.value = {
      transition: 'none',
      transform: 'translateY(0)'
    }

    await nextTick()
    // 缓出动画滚到目标位置 (3s, cubic-bezier 减速)
    stripStyle.value = {
      transition: 'transform 3s cubic-bezier(0.15, 0.8, 0.3, 1)',
      transform: `translateY(-${targetOffset}px)`
    }

    // 等动画结束
    await new Promise((resolve) => setTimeout(resolve, 3100))

    drawResult.value = result
    drawState.value = 'result'
    loadWallet()
    loadMyRecords()
  }

  // ─── 概率计算 ───
  function getProbPercent(prize: LotteryPrize, all: LotteryPrize[]) {
    const total = all.reduce((sum, p) => sum + Math.max(p.probability_weight, 1), 0)
    return ((Math.max(prize.probability_weight, 1) / total) * 100).toFixed(1)
  }

  // ─── 粒子动画 ───
  function getParticleStyle(n: number) {
    const angle = (n / 12) * 360
    const delay = Math.random() * 0.5
    return {
      '--angle': `${angle}deg`,
      '--delay': `${delay}s`
    } as any
  }

  // ─── 初始化 ───
  onMounted(() => {
    loadActivities()
    loadMyRecords()
    loadWallet()
  })

  defineExpose({ load: loadActivities })
</script>

<style scoped lang="scss">
  .lottery-page {
    padding: 8px 0;
  }

  .loading-placeholder {
    height: 200px;
  }

  /* ─── 活动网格 ─── */
  .activity-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 16px;
  }

  .activity-card {
    border-radius: 12px;
    overflow: hidden;
    border: 1px solid var(--el-border-color);
    background: var(--el-bg-color);
    cursor: pointer;
    transition:
      transform 0.2s,
      box-shadow 0.2s;

    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);

      .activity-overlay {
        opacity: 1;
      }
    }
  }

  .activity-cover {
    position: relative;
    width: 100%;
    height: 160px;
    overflow: hidden;
    background: var(--el-fill-color-light);

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
  }

  .activity-cover-placeholder {
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--el-text-color-placeholder);
  }

  .activity-overlay {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0;
    transition: opacity 0.2s;
  }

  .activity-body {
    padding: 12px 14px;
  }

  .activity-name {
    font-size: 16px;
    font-weight: 600;
    margin: 0 0 4px;
  }

  .activity-desc {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    margin: 0 0 8px;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .activity-meta {
    margin-bottom: 8px;
  }

  .cost-badge {
    display: inline-block;
    padding: 2px 10px;
    background: var(--el-color-warning-light-9);
    color: var(--el-color-warning);
    border-radius: 20px;
    font-size: 13px;
    font-weight: 600;

    &.free {
      background: var(--el-color-success-light-9);
      color: var(--el-color-success);
    }
  }

  .prize-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    margin-top: 8px;
  }

  /* ─── 抽奖弹窗 ─── */
  .draw-modal {
    text-align: center;
  }

  .draw-cover {
    img {
      width: 100%;
      max-height: 180px;
      object-fit: cover;
      border-radius: 8px;
    }
  }

  .draw-info {
    text-align: left;
    padding: 0 4px;
  }

  .prize-pool {
    text-align: left;
    padding: 12px;
    background: var(--el-fill-color-lighter);
    border-radius: 8px;
  }

  .prize-pool-item {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 12px;
    border: 1px solid currentColor;

    &.tier-normal {
      color: var(--el-color-info);
    }
    &.tier-rare {
      color: var(--el-color-warning);
    }
    &.tier-legendary {
      color: #d4af37;
    }
  }

  .prize-pool-prob {
    font-size: 11px;
    opacity: 0.7;
  }

  /* ─── 动画状态（滚动老虎机） ─── */
  .draw-animation {
    padding: 20px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
  }

  .slot-machine-wrapper {
    position: relative;
    width: 220px;
    height: 240px;
  }

  .slot-machine-viewport {
    width: 220px;
    height: 240px; /* 显示3个奖品 */
    overflow: hidden;
    position: relative;
    border-radius: 12px;
    background: var(--el-fill-color-lighter);
    border: 2px solid var(--el-color-warning-light-5);
  }

  .slot-machine-strip {
    display: flex;
    flex-direction: column;
    will-change: transform;
  }

  .slot-machine-item {
    width: 220px;
    height: 80px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    flex-shrink: 0;
    padding: 0 12px;
    box-sizing: border-box;
    border-bottom: 1px solid var(--el-border-color-lighter);

    &.tier-normal {
      color: var(--el-color-info);
    }
    &.tier-rare {
      color: var(--el-color-warning);
    }
    &.tier-legendary {
      color: #d4af37;
    }
  }

  .slot-prize-img {
    width: 48px;
    height: 48px;
    border-radius: 8px;
    object-fit: cover;
    flex-shrink: 0;
  }

  .slot-prize-name {
    font-size: 14px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .slot-indicator {
    width: 220px;
    height: 80px;
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    border: 2px solid var(--el-color-warning);
    border-radius: 8px;
    pointer-events: none;
    box-shadow: 0 0 12px rgba(230, 162, 60, 0.3);
  }

  .draw-animation-text {
    font-size: 18px;
    font-weight: 600;
    color: var(--el-color-warning);
    animation: pulse 1s ease-in-out infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
    }
    50% {
      opacity: 0.3;
    }
  }

  /* ─── 结果状态 ─── */
  .draw-result {
    padding: 20px 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    position: relative;
    overflow: hidden;

    &.result-normal .result-icon-wrap {
      animation: result-pop-normal 0.5s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
    }

    &.result-rare .result-icon-wrap {
      animation: result-pop-rare 0.6s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
    }

    &.result-legendary .result-icon-wrap {
      animation: result-pop-legendary 0.7s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
    }
  }

  .result-icon-wrap {
    opacity: 0;
  }

  @keyframes result-pop-normal {
    from {
      opacity: 0;
      transform: scale(0.5);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  @keyframes result-pop-rare {
    from {
      opacity: 0;
      transform: scale(0.3) rotate(-15deg);
    }
    to {
      opacity: 1;
      transform: scale(1) rotate(0deg);
    }
  }

  @keyframes result-pop-legendary {
    from {
      opacity: 0;
      transform: scale(0) rotate(-30deg);
    }
    to {
      opacity: 1;
      transform: scale(1) rotate(0deg);
    }
  }

  .result-tier-ring {
    width: 100px;
    height: 100px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 3px solid transparent;

    &.ring-normal {
      border-color: var(--el-color-info);
      color: var(--el-color-info);
      box-shadow: 0 0 16px var(--el-color-info-light-5);
    }

    &.ring-rare {
      border-color: var(--el-color-warning);
      color: var(--el-color-warning);
      box-shadow: 0 0 24px var(--el-color-warning);
      animation: rare-glow 1.5s ease-in-out infinite alternate;
    }

    &.ring-legendary {
      border-color: #d4af37;
      color: #d4af37;
      box-shadow:
        0 0 32px #d4af37,
        0 0 64px rgba(212, 175, 55, 0.3);
      animation: legendary-glow 1s ease-in-out infinite alternate;
    }
  }

  @keyframes rare-glow {
    from {
      box-shadow: 0 0 16px var(--el-color-warning);
    }
    to {
      box-shadow:
        0 0 36px var(--el-color-warning),
        0 0 60px rgba(230, 162, 60, 0.3);
    }
  }

  @keyframes legendary-glow {
    from {
      box-shadow: 0 0 24px #d4af37;
    }
    to {
      box-shadow:
        0 0 48px #d4af37,
        0 0 96px rgba(212, 175, 55, 0.4),
        inset 0 0 20px rgba(212, 175, 55, 0.1);
    }
  }

  .result-tier-label {
    font-size: 13px;
    font-weight: 700;
    letter-spacing: 2px;
    text-transform: uppercase;
    margin-top: 4px;

    &.label-normal {
      color: var(--el-color-info);
    }
    &.label-rare {
      color: var(--el-color-warning);
    }
    &.label-legendary {
      color: #d4af37;
    }
  }

  .result-name {
    font-size: 22px;
    font-weight: 700;
    margin-top: 4px;
  }

  .result-prize-img {
    width: 56px;
    height: 56px;
    border-radius: 8px;
    object-fit: cover;
  }

  .prize-chip-img {
    width: 16px;
    height: 16px;
    border-radius: 3px;
    vertical-align: middle;
    margin-right: 2px;
  }

  .prize-pool-img {
    width: 24px;
    height: 24px;
    border-radius: 4px;
    object-fit: cover;
  }

  .record-prize-img {
    width: 24px;
    height: 24px;
    border-radius: 4px;
    object-fit: cover;
  }

  /* ─── 粒子 ─── */
  .particles {
    position: absolute;
    inset: 0;
    pointer-events: none;
    overflow: hidden;
  }

  .particle {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--el-color-warning);
    animation: particle-fly 1s ease-out var(--delay) forwards;
    opacity: 0;

    .result-legendary & {
      background: #d4af37;
      width: 10px;
      height: 10px;
    }
  }

  @keyframes particle-fly {
    0% {
      opacity: 1;
      transform: rotate(var(--angle)) translateX(0);
    }
    100% {
      opacity: 0;
      transform: rotate(var(--angle)) translateX(100px);
    }
  }

  /* ─── 记录 ─── */
  .record-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .record-item {
    padding: 10px 12px;
    border: 1px solid var(--el-border-color);
    border-radius: 6px;
    background: var(--el-fill-color-lighter);
  }
</style>
