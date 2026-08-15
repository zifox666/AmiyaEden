<template>
  <div class="art-card p-5">
    <div class="flex-b mb-4">
      <div
        ><h2 class="text-xl">{{ t('at.km.listenerTitle') }}</h2
        ><p class="text-sm text-g-500 mt-1">{{ t('at.km.listenerDescription') }}</p></div
      >
      <div class="flex gap-2"
        ><ElButton @click="paused = !paused">{{
          paused ? t('at.km.resume') : t('at.km.pause')
        }}</ElButton
        ><ElButton @click="clearSucceeded">{{ t('at.km.clearSucceeded') }}</ElButton
        ><ElButton v-if="canReview" type="primary" @click="router.push('/at/km-review')">{{
          t('at.km.openReview')
        }}</ElButton></div
      >
    </div>
    <ElAlert
      :type="listenerState.type"
      :title="listenerState.text"
      :closable="false"
      class="mb-4"
    />
    <ElInput
      v-model="manualText"
      type="textarea"
      :rows="10"
      :placeholder="t('at.km.pastePlaceholder')"
    />
    <div class="mt-3"
      ><ElButton type="primary" @click="enqueue(manualText)">{{
        t('at.km.addQueue')
      }}</ElButton></div
    >
    <ArtTable
      class="mt-5"
      :data="pagedQueue"
      :columns="columns"
      :pagination="pagination"
      @pagination:size-change="changePageSize"
      @pagination:current-change="pagination.current = $event"
    >
      <template #status="{ row }"
        ><ElTag
          v-if="row.status"
          :type="row.status === 'success' ? 'success' : row.status === 'error' ? 'danger' : 'info'"
          >{{ t(`at.km.queueStatus.${row.status}`) }}</ElTag
        ><span v-else>-</span></template
      >
      <template #operation="{ row }"
        ><ElButton
          v-if="row.status === 'error'"
          link
          type="primary"
          @click="submit(row as QueueItem)"
          >{{ t('at.km.retry') }}</ElButton
        ></template
      >
    </ArtTable>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { useRouter } from 'vue-router'
  import { atApi } from '@/api/at'
  import type { ColumnOption } from '@/types'
  import { useUserStore } from '@/store/modules/user'
  const { t } = useI18n(),
    router = useRouter()
  const userStore = useUserStore()
  const canReview = computed(() =>
    (userStore.info.roles ?? []).some((role) => ['fc', 'admin', 'super_admin'].includes(role))
  )
  type QueueItem = {
    fingerprint: string
    text: string
    preview: string
    createdAt: string
    status: 'queued' | 'submitting' | 'success' | 'error'
    message: string
  }
  type ClipboardPermission = 'granted' | 'denied' | 'prompt'
  const queue = ref<QueueItem[]>([]),
    paused = ref(false),
    focused = ref(document.hasFocus()),
    visible = ref(!document.hidden),
    permission = ref<ClipboardPermission | 'unsupported'>('prompt'),
    manualText = ref('')
  const pagination = reactive({ current: 1, size: 10, total: 0 })
  const pagedQueue = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return queue.value.slice(start, start + pagination.size)
  })
  watch(
    () => queue.value.length,
    (total) => (pagination.total = total),
    { immediate: true }
  )
  const columns = computed<ColumnOption[]>(() => [
    { prop: 'createdAt', label: t('at.km.detectedAt'), minWidth: 190 },
    { prop: 'preview', label: t('at.km.preview'), minWidth: 280, showOverflowTooltip: true },
    { prop: 'status', label: t('at.common.status'), width: 130, useSlot: true },
    { prop: 'message', label: t('at.km.result'), minWidth: 220 },
    { prop: 'operation', label: t('at.common.actions'), width: 100, fixed: 'right', useSlot: true }
  ])
  function changePageSize(size: number) {
    pagination.size = size
    pagination.current = 1
  }
  let timer: number | undefined
  const listenerState = computed(() => {
    if (!window.isSecureContext || !navigator.clipboard)
      return { type: 'error' as const, text: t('at.km.unsupported') }
    if (permission.value === 'denied')
      return { type: 'error' as const, text: t('at.km.permissionDenied') }
    if (paused.value) return { type: 'warning' as const, text: t('at.km.paused') }
    if (!focused.value || !visible.value)
      return { type: 'warning' as const, text: t('at.km.backgroundPaused') }
    return { type: 'success' as const, text: t('at.km.listening') }
  })
  async function fingerprint(text: string) {
    const bytes = new TextEncoder().encode(text.replace(/\r\n?/g, '\n').trim())
    const hash = await crypto.subtle.digest('SHA-256', bytes)
    return [...new Uint8Array(hash)].map((b) => b.toString(16).padStart(2, '0')).join('')
  }
  async function enqueue(text: string) {
    if (!text.trim()) return
    const id = await fingerprint(text)
    if (queue.value.some((item) => item.fingerprint === id)) return
    const item = reactive<QueueItem>({
      fingerprint: id,
      text,
      preview: text.trim().slice(0, 160),
      createdAt: new Date().toLocaleString(),
      status: 'queued',
      message: ''
    })
    queue.value.unshift(item)
    manualText.value = ''
    await submit(item)
  }
  async function submit(item: QueueItem) {
    item.status = 'submitting'
    try {
      const result = await atApi.submitKillmails(item.text)
      item.status = 'success'
      item.message = result.items
        .map((entry) => `${entry.km_id}${entry.duplicate ? ` (${t('at.km.duplicate')})` : ''}`)
        .join(', ')
    } catch (error) {
      item.status = 'error'
      item.message = error instanceof Error ? error.message : String(error)
    }
  }
  async function poll() {
    if (
      paused.value ||
      !focused.value ||
      !visible.value ||
      permission.value === 'denied' ||
      !window.isSecureContext
    )
      return
    try {
      await enqueue(await navigator.clipboard.readText())
    } catch {
      /* Browser exposes permission changes inconsistently; the status remains explicit. */
    }
  }
  const clearSucceeded = () => {
    queue.value = queue.value.filter((item) => item.status !== 'success')
  }
  const handleFocus = () => (focused.value = true)
  const handleBlur = () => (focused.value = false)
  const handleVisibility = () => (visible.value = !document.hidden)
  onMounted(async () => {
    window.addEventListener('focus', handleFocus)
    window.addEventListener('blur', handleBlur)
    document.addEventListener('visibilitychange', handleVisibility)
    try {
      const queryPermission = navigator.permissions.query as unknown as (descriptor: {
        name: string
      }) => Promise<{ state: ClipboardPermission; onchange: (() => void) | null }>
      const result = await queryPermission({ name: 'clipboard-read' })
      permission.value = result.state
      result.onchange = () => (permission.value = result.state)
    } catch {
      permission.value = 'prompt'
    }
    timer = window.setInterval(poll, 1500)
  })
  onUnmounted(() => {
    if (timer) window.clearInterval(timer)
    window.removeEventListener('focus', handleFocus)
    window.removeEventListener('blur', handleBlur)
    document.removeEventListener('visibilitychange', handleVisibility)
  })
</script>
