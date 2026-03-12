<template>
  <div class="art-card h-128 p-5 mb-5 max-sm:mb-4">
    <div class="art-card-header">
      <div class="title">
        <h4>{{ $t('dashboardConsole.fleetList.title') }}</h4>
        <p>{{ $t('dashboardConsole.fleetList.records', { count: fleets.length }) }}</p>
      </div>
    </div>
    <div class="h-[calc(100%-40px)] overflow-auto mt-2">
      <ElScrollbar>
        <div v-if="fleets.length === 0" class="flex-cc h-full text-g-500 text-sm">
          {{ $t('dashboardConsole.fleetList.empty') }}
        </div>
        <div
          v-for="(item, index) in fleets"
          :key="`${item.source}-${item.id}-${index}`"
          class="flex-cb py-3 border-b border-g-300 text-sm last:border-b-0"
        >
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <ElTag
                :type="item.source === 'alliance' ? 'warning' : 'primary'"
                size="small"
                effect="plain"
              >
                {{
                  item.source === 'alliance'
                    ? $t('dashboardConsole.fleetList.source.alliance')
                    : $t('dashboardConsole.fleetList.source.internal')
                }}
              </ElTag>
              <span class="text-g-800 font-medium truncate">{{ item.title }}</span>
            </div>
            <div class="flex items-center gap-3 mt-1 text-xs text-g-500">
              <span>{{ formatTime(item.start_at) }}</span>
              <span v-if="item.character_name">{{ item.character_name }}</span>
              <span v-if="item.ship_type_name">{{ item.ship_type_name }}</span>
              <span v-if="item.importance" class="capitalize">{{
                importanceLabel(item.importance)
              }}</span>
            </div>
          </div>
          <div class="text-right ml-3 shrink-0">
            <span class="text-theme font-medium">{{ item.pap_count }}</span>
            <span class="text-xs text-g-500 ml-0.5">PAP</span>
          </div>
        </div>
      </ElScrollbar>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'

  const { t } = useI18n()
  defineProps<{
    fleets: Api.Dashboard.FleetItem[]
  }>()

  const importanceLabel = (importance: string): string => {
    return t(`fleet.importance.${importance}`)
  }

  const formatTime = (time: string): string => {
    if (!time) return ''
    const d = new Date(time)
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  }
</script>
