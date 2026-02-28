<!-- 通知组件 -->
<template>
  <div
    class="art-notification-panel art-card-sm !shadow-xl"
    :style="{
      transform: show ? 'scaleY(1)' : 'scaleY(0.9)',
      opacity: show ? 1 : 0
    }"
    v-show="visible"
    @click.stop
  >
    <div class="flex-cb px-3.5 mt-3.5">
      <span class="text-base font-medium text-g-800">{{ $t('notice.title') }}</span>
      <span
        class="text-xs text-g-800 px-1.5 py-1 c-p select-none rounded hover:bg-g-200"
        @click="handleMarkAllRead"
      >
        {{ $t('notice.btnRead') }}
      </span>
    </div>

    <div class="box-border flex items-center w-full h-12.5 px-3.5 border-b-d">
      <span class="text-[13px] text-g-700">
        {{ $t('notice.bar[0]') }} ({{ unreadCount }})
        <span v-if="unreadCount > 0" class="ml-1 text-xs text-danger"
          >{{ unreadCount }} {{ $t('notice.text[1]') || '未读' }}</span
        >
      </span>
    </div>

    <div class="w-full h-[calc(100%-95px)]">
      <div class="h-[calc(100%-60px)] overflow-y-scroll scrollbar-thin">
        <!-- 加载状态 -->
        <div v-if="loading" class="relative top-25 h-full text-g-500 text-center !bg-transparent">
          <ArtSvgIcon icon="ri:loader-4-line" class="text-3xl animate-spin" />
          <p class="mt-3.5 text-xs !bg-transparent">{{ $t('notice.text[2]') || '加载中...' }}</p>
        </div>

        <!-- 通知列表 -->
        <ul v-else-if="notificationList.length > 0">
          <li
            v-for="item in notificationList"
            :key="item.id"
            class="box-border flex-c px-3.5 py-3.5 c-p last:border-b-0 hover:bg-g-200/60"
            :class="{ 'opacity-50': item.is_read }"
            @click="handleMarkRead(item)"
          >
            <div
              class="size-9 leading-9 text-center rounded-lg flex-cc"
              :class="[getNoticeStyle(item.type).iconClass]"
            >
              <ArtSvgIcon class="text-lg !bg-transparent" :icon="getNoticeStyle(item.type).icon" />
            </div>
            <div class="w-[calc(100%-45px)] ml-3.5">
              <h4 class="text-sm font-normal leading-5.5 text-g-900">
                <span class="font-medium">{{ item.type }}</span>
                <span
                  v-if="!item.is_read"
                  class="ml-1.5 inline-block size-1.5 bg-danger rounded-full align-middle"
                ></span>
              </h4>
              <p v-if="item.text" class="mt-1 text-xs text-g-600 line-clamp-2">{{ item.text }}</p>
              <p class="mt-1.5 text-xs text-g-500">{{ formatTime(item.timestamp) }}</p>
            </div>
          </li>
        </ul>

        <!-- 空状态 -->
        <div v-else class="relative top-25 h-full text-g-500 text-center !bg-transparent">
          <ArtSvgIcon icon="system-uicons:inbox" class="text-5xl" />
          <p class="mt-3.5 text-xs !bg-transparent"
            >{{ $t('notice.text[0]') }}{{ $t('notice.bar[0]') }}</p
          >
        </div>
      </div>

      <div class="relative box-border w-full px-3.5 flex gap-2">
        <ElButton v-if="hasMore" class="flex-1 mt-3" @click="loadMore" v-ripple>
          {{ $t('notice.viewAll') || '加载更多' }}
        </ElButton>
      </div>
    </div>

    <div class="h-25"></div>
  </div>
</template>

<script setup lang="ts">
  import { ref, watch, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import { fetchNotifications, markAsRead, markAllAsRead } from '@/api/notification'

  defineOptions({ name: 'ArtNotification' })

  interface NoticeStyle {
    icon: string
    iconClass: string
  }

  const { t } = useI18n()

  const props = defineProps<{
    value: boolean
  }>()

  const emit = defineEmits<{
    'update:value': [value: boolean]
    'unread-change': [count: number]
  }>()

  const show = ref(false)
  const visible = ref(false)
  const loading = ref(false)

  // 通知数据
  const notificationList = ref<Api.Notification.NotificationItem[]>([])
  const totalCount = ref(0)
  const unreadCount = ref(0)
  const currentPage = ref(1)
  const pageSize = 20
  const hasMore = ref(false)

  // ─── 通知类型样式映射 ───

  const getNoticeStyle = (type: string): NoticeStyle => {
    // EVE 通知类型前缀映射
    if (type.startsWith('War') || type.startsWith('AllWar')) {
      return { icon: 'ri:sword-line', iconClass: 'bg-danger/12 text-danger' }
    }
    if (type.startsWith('Corp') || type.startsWith('AllAnchoringMsg')) {
      return { icon: 'ri:building-line', iconClass: 'bg-info/12 text-info' }
    }
    if (type.startsWith('Sov') || type.startsWith('Sovereignty')) {
      return { icon: 'ri:flag-line', iconClass: 'bg-warning/12 text-warning' }
    }
    if (type.startsWith('Structure') || type.startsWith('Tower') || type.startsWith('Orbital')) {
      return { icon: 'ri:building-2-line', iconClass: 'bg-theme/12 text-theme' }
    }
    if (type.startsWith('Kill') || type.includes('Kill')) {
      return { icon: 'ri:skull-line', iconClass: 'bg-danger/12 text-danger' }
    }
    if (type.startsWith('Contact') || type.startsWith('Buddy')) {
      return { icon: 'ri:user-line', iconClass: 'bg-success/12 text-success' }
    }
    if (type.startsWith('Insurance')) {
      return { icon: 'ri:shield-check-line', iconClass: 'bg-success/12 text-success' }
    }
    if (type.startsWith('Bill') || type.includes('Payment') || type.includes('Tax')) {
      return { icon: 'ri:money-dollar-circle-line', iconClass: 'bg-warning/12 text-warning' }
    }
    return { icon: 'ri:notification-3-line', iconClass: 'bg-theme/12 text-theme' }
  }

  // ─── 时间格式化 ───

  const formatTime = (timestamp: string): string => {
    const date = new Date(timestamp)
    const now = new Date()
    const diff = now.getTime() - date.getTime()
    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(diff / 3600000)
    const days = Math.floor(diff / 86400000)

    if (minutes < 1) return '刚刚'
    if (minutes < 60) return `${minutes} 分钟前`
    if (hours < 24) return `${hours} 小时前`
    if (days < 7) return `${days} 天前`
    return date.toLocaleString()
  }

  // ─── 数据加载 ───

  const loadNotifications = async (page = 1) => {
    loading.value = true
    try {
      const data = await fetchNotifications({ page, page_size: pageSize })
      if (page === 1) {
        notificationList.value = data.list || []
      } else {
        notificationList.value.push(...(data.list || []))
      }
      totalCount.value = data.total
      unreadCount.value = data.unread_count
      currentPage.value = page
      hasMore.value = notificationList.value.length < data.total
      emit('unread-change', data.unread_count)
    } catch {
      // 静默失败
    } finally {
      loading.value = false
    }
  }

  const loadMore = () => {
    loadNotifications(currentPage.value + 1)
  }

  // ─── 标记已读 ───

  const handleMarkRead = async (item: Api.Notification.NotificationItem) => {
    if (item.is_read) return
    try {
      await markAsRead({ notification_ids: [item.id] })
      item.is_read = true
      unreadCount.value = Math.max(0, unreadCount.value - 1)
      emit('unread-change', unreadCount.value)
    } catch {
      // 静默失败
    }
  }

  const handleMarkAllRead = async () => {
    if (unreadCount.value === 0) return
    try {
      await markAllAsRead()
      notificationList.value.forEach((n) => (n.is_read = true))
      unreadCount.value = 0
      emit('unread-change', 0)
    } catch {
      // 静默失败
    }
  }

  // ─── 动画 ───

  const showNotice = (open: boolean) => {
    if (open) {
      visible.value = true
      loadNotifications(1) // 每次打开刷新
      setTimeout(() => {
        show.value = true
      }, 5)
    } else {
      show.value = false
      setTimeout(() => {
        visible.value = false
      }, 350)
    }
  }

  // ─── 公开方法 ───

  defineExpose({
    refreshUnread: () => loadNotifications(1)
  })

  // ─── 生命周期 ───

  onMounted(() => {
    // 初始加载未读数
    loadNotifications(1)
  })

  watch(
    () => props.value,
    (newValue) => {
      showNotice(newValue)
    }
  )
</script>

<style scoped>
  @reference '@styles/core/tailwind.css';

  .art-notification-panel {
    @apply absolute 
    top-14.5 
    right-5 
    w-90 
    h-125 
    overflow-hidden 
    transition-all 
    duration-300
    origin-top 
    will-change-[top,left] 
    max-[640px]:top-[65px]
    max-[640px]:right-0
    max-[640px]:w-full 
    max-[640px]:h-[80vh];
  }

  .bar-active {
    color: var(--theme-color) !important;
    border-bottom: 2px solid var(--theme-color);
  }

  .scrollbar-thin::-webkit-scrollbar {
    width: 5px !important;
  }

  .dark .scrollbar-thin::-webkit-scrollbar-track {
    background-color: var(--default-box-color);
  }

  .dark .scrollbar-thin::-webkit-scrollbar-thumb {
    background-color: #222 !important;
  }
</style>
