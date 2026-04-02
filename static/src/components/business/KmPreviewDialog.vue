<!-- KM 装配详情预览组件（共享） -->
<template>
  <ElDialog
    v-model="visible"
    :title="$t('srp.kmPreview.title')"
    width="680px"
    :close-on-click-modal="true"
    destroy-on-close
    @open="loadDetail"
  >
    <div v-loading="loading" class="km-detail">
      <template v-if="detail">
        <!-- 头部：舰船图标 + 信息 -->
        <div class="km-header">
          <img
            :src="`https://images.evetech.net/types/${detail.ship_type_id}/icon?size=64`"
            class="km-ship-icon"
            alt="ship"
          />
          <div class="km-header-info">
            <h3 class="km-ship-name">{{ detail.ship_name }}</h3>
            <p class="km-meta">{{ detail.character_name }}</p>
            <p class="km-meta">{{ detail.system_name }} · {{ formatTime(detail.killmail_time) }}</p>
            <p v-if="detail.janice_amount" class="km-meta">
              {{ $t('srp.kmPreview.estimatedValue') }}:
              {{ formatIskSmart(detail.janice_amount) }} ISK
            </p>
          </div>
        </div>

        <!-- 槽位列表 -->
        <div class="km-slots">
          <div v-for="slot in detail.slots" :key="slot.flag_id" class="km-slot-group">
            <div class="km-slot-header">
              <span>{{ slot.flag_text || slot.flag_name }}</span>
            </div>
            <div class="km-slot-items">
              <div
                v-for="(item, idx) in slot.items"
                :key="`${item.item_id}-${item.dropped}-${idx}`"
                class="km-item"
              >
                <span class="km-item-bar" :class="item.dropped ? 'bar-dropped' : 'bar-destroyed'" />
                <img
                  :src="`https://images.evetech.net/types/${item.item_id}/icon?size=32`"
                  class="km-item-icon"
                  alt=""
                />
                <span class="km-item-name">{{ item.item_name }}</span>
                <span v-if="item.quantity > 1" class="km-item-qty">x{{ item.quantity }}</span>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- 空状态 -->
      <ElEmpty v-if="!loading && !detail" :description="$t('srp.kmPreview.noData')" />
    </div>

    <template #footer>
      <ElButton @click="visible = false">{{ $t('srp.apply.cancelBtn') }}</ElButton>
      <ElButton v-if="killmailId" type="primary" @click="openZkillboard">
        {{ $t('srp.apply.openZkillboard') }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { ElDialog, ElButton, ElEmpty } from 'element-plus'
  import { fetchKillmailDetail } from '@/api/srp'
  import { formatIskSmart, formatTime } from '@utils/common'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'KmPreviewDialog' })

  const props = defineProps<{
    killmailId: number
  }>()

  const visible = defineModel<boolean>({ default: false })

  useI18n()
  const userStore = useUserStore()

  const loading = ref(false)
  const detail = ref<Api.Srp.KillmailDetailResponse | null>(null)

  const loadDetail = async () => {
    if (!props.killmailId) return
    loading.value = true
    detail.value = null
    try {
      detail.value = await fetchKillmailDetail({
        killmail_id: props.killmailId,
        language: userStore.language
      })
    } catch {
      detail.value = null
    } finally {
      loading.value = false
    }
  }

  const openZkillboard = () => {
    window.open(`https://zkillboard.com/kill/${props.killmailId}/`, '_blank')
  }

  watch(
    () => props.killmailId,
    () => {
      if (visible.value && props.killmailId) loadDetail()
    }
  )
</script>

<style scoped>
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

  /* 槽位 */
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

  .km-item-bar {
    width: 3px;
    align-self: stretch;
    border-radius: 2px;
    flex-shrink: 0;
  }

  .bar-dropped {
    background: var(--el-color-success);
  }

  .bar-destroyed {
    background: var(--el-color-danger);
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
