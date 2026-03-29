import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { SYSTEM_IDENTITY } from '@/constants/system-identity'

export const useSysConfigStore = defineStore(
  'sysConfig',
  () => {
    const config = ref<Api.SysConfig.BasicConfig>({
      corp_id: SYSTEM_IDENTITY.corporationId,
      site_title: SYSTEM_IDENTITY.displayName
    })

    const loading = ref(false)
    const loaded = ref(false)

    const logoUrl = computed(
      () => `https://images.evetech.net/corporations/${config.value.corp_id}/logo?size=128`
    )

    const siteTitle = computed(() => config.value.site_title)
    const hasCanonicalIdentity = computed(
      () =>
        config.value.corp_id === SYSTEM_IDENTITY.corporationId &&
        config.value.site_title === SYSTEM_IDENTITY.displayName
    )

    async function loadConfig() {
      loading.value = true
      try {
        config.value = {
          corp_id: SYSTEM_IDENTITY.corporationId,
          site_title: SYSTEM_IDENTITY.displayName
        }
      } finally {
        loaded.value = true
        loading.value = false
      }
    }

    async function ensureLoaded() {
      if (loading.value) return
      if (loaded.value && hasCanonicalIdentity.value) return
      await loadConfig()
    }

    return {
      config,
      logoUrl,
      siteTitle,
      loading,
      loaded,
      loadConfig,
      ensureLoaded
    }
  },
  {
    persist: {
      key: 'sysConfig',
      storage: localStorage
    }
  }
)
