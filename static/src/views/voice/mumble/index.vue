<template>
  <div class="mumble-page">
    <ElCard shadow="never">
      <template #header>
        <div class="card-header">
          <div>
            <h2 class="section-title">{{ $t('mumble.title') }}</h2>
            <p class="section-desc">{{ $t('mumble.description') }}</p>
          </div>
          <ElTag :type="profile?.config.enabled ? 'success' : 'warning'">
            {{ profile?.config.enabled ? $t('mumble.enabled') : $t('mumble.disabled') }}
          </ElTag>
        </div>
      </template>

      <ElDescriptions :column="1" border v-loading="loading">
        <ElDescriptionsItem :label="$t('mumble.username')">
          <ElSpace>
            <span class="mono">{{ profile?.account.username ?? '-' }}</span>
            <ElButton link type="primary" @click="copyText(profile?.account.username)">
              {{ $t('mumble.copy') }}
            </ElButton>
          </ElSpace>
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('mumble.password')">
          <ElSpace>
            <span class="mono">{{ profile?.account.password ?? '-' }}</span>
            <ElButton link type="primary" @click="copyText(profile?.account.password)">
              {{ $t('mumble.copy') }}
            </ElButton>
            <ElButton link type="danger" :loading="resetting" @click="handleResetPassword">
              {{ $t('mumble.resetPassword') }}
            </ElButton>
          </ElSpace>
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('mumble.displayName')">
          {{ profile?.account.display_name ?? '-' }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('mumble.currentGroups')">
          <ElSpace v-if="profile?.account.groups?.length" wrap>
            <ElTag v-for="group in profile.account.groups" :key="group" type="info">
              {{ group }}
            </ElTag>
          </ElSpace>
          <span v-else>-</span>
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('mumble.server')">
          {{ serverText }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('mumble.quickUrl')">
          <ElInput :model-value="profile?.account.quick_url ?? ''" readonly>
            <template #append>
              <ElButton @click="copyText(profile?.account.quick_url)">
                {{ $t('mumble.copy') }}
              </ElButton>
            </template>
          </ElInput>
        </ElDescriptionsItem>
      </ElDescriptions>

      <ElAlert
        v-if="!profile?.account.quick_url"
        class="mt-4"
        type="warning"
        show-icon
        :closable="false"
        :title="$t('mumble.notConfigured')"
      />

      <ElSpace class="mt-4">
        <ElButton type="primary" :disabled="!profile?.account.quick_url" @click="openMumble">
          {{ $t('mumble.openClient') }}
        </ElButton>
        <ElButton @click="loadProfile">{{ $t('common.refresh') }}</ElButton>
      </ElSpace>
    </ElCard>

    <ElCard v-if="isAdmin" shadow="never" class="mt-4">
      <template #header>
        <div class="card-header">
          <div>
            <h2 class="section-title">{{ $t('mumble.roleGroups.title') }}</h2>
            <p class="section-desc">{{ $t('mumble.roleGroups.description') }}</p>
          </div>
          <ElButton type="primary" :loading="savingRoleGroups" @click="handleSaveRoleGroups">
            {{ $t('common.save') }}
          </ElButton>
        </div>
      </template>

      <ElTable :data="roleGroups" border v-loading="loadingRoleGroups">
        <ElTableColumn prop="role_name" :label="$t('mumble.roleGroups.role')" min-width="150">
          <template #default="{ row }">
            <ElSpace>
              <span>{{ row.role_name }}</span>
              <ElTag size="small">{{ row.role_code }}</ElTag>
            </ElSpace>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="group_name" :label="$t('mumble.roleGroups.group')" min-width="220">
          <template #default="{ row }">
            <ElInput
              v-model="row.group_name"
              clearable
              :disabled="!row.enabled"
              :placeholder="$t('mumble.roleGroups.groupPlaceholder')"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn prop="enabled" :label="$t('mumble.roleGroups.enabled')" width="120">
          <template #default="{ row }">
            <ElSwitch v-model="row.enabled" />
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>

    <ElCard v-if="isAdmin" shadow="never" class="mt-4">
      <template #header>
        <h2 class="section-title">{{ $t('mumble.config.title') }}</h2>
      </template>

      <ElForm :model="configForm" label-width="130px" style="max-width: 720px">
        <ElFormItem :label="$t('mumble.config.enabled')">
          <ElSwitch v-model="configForm.enabled" />
        </ElFormItem>
        <ElFormItem :label="$t('mumble.config.url')">
          <ElInput
            v-model="configForm.url"
            clearable
            :placeholder="$t('mumble.config.urlPlaceholder')"
          />
        </ElFormItem>
        <ElFormItem :label="$t('mumble.config.port')">
          <ElInputNumber v-model="configForm.port" :min="1" :max="65535" :controls="false" />
        </ElFormItem>
        <ElFormItem :label="$t('mumble.config.serverName')">
          <ElInput
            v-model="configForm.server_name"
            clearable
            :placeholder="$t('mumble.config.serverNamePlaceholder')"
          />
        </ElFormItem>
        <ElFormItem :label="$t('mumble.config.authSecret')">
          <ElInput
            v-model="configForm.auth_secret"
            clearable
            type="password"
            show-password
            :placeholder="
              profile?.config.auth_secret_set
                ? $t('mumble.config.authSecretSetPlaceholder')
                : $t('mumble.config.authSecretPlaceholder')
            "
          />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" :loading="savingConfig" @click="handleSaveConfig">
            {{ $t('common.save') }}
          </ElButton>
        </ElFormItem>
      </ElForm>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { computed, reactive, ref, onMounted } from 'vue'
  import { useI18n } from 'vue-i18n'
  import {
    ElAlert,
    ElButton,
    ElCard,
    ElDescriptions,
    ElDescriptionsItem,
    ElForm,
    ElFormItem,
    ElInput,
    ElInputNumber,
    ElMessage,
    ElMessageBox,
    ElSpace,
    ElSwitch,
    ElTable,
    ElTableColumn,
    ElTag
  } from 'element-plus'
  import {
    fetchMumbleProfile,
    fetchMumbleRoleGroups,
    resetMumblePassword,
    updateMumbleConfig,
    updateMumbleRoleGroups
  } from '@/api/mumble'
  import { useUserStore } from '@/store/modules/user'

  defineOptions({ name: 'MumbleCenter' })

  const { t } = useI18n()
  const userStore = useUserStore()

  const loading = ref(false)
  const loadingRoleGroups = ref(false)
  const resetting = ref(false)
  const savingConfig = ref(false)
  const savingRoleGroups = ref(false)
  const profile = ref<Api.Mumble.Profile | null>(null)
  const roleGroups = ref<Api.Mumble.RoleGroupMapping[]>([])

  const configForm = reactive<Api.Mumble.Config>({
    enabled: false,
    url: '',
    port: 64738,
    server_name: '',
    auth_secret_set: false,
    auth_secret: ''
  })

  const isAdmin = computed(() => {
    const roles = userStore.info.roles ?? []
    return roles.includes('admin') || roles.includes('super_admin')
  })

  const serverText = computed(() => {
    if (!profile.value?.config.url) return '-'
    return `${profile.value.config.url}:${profile.value.config.port}`
  })

  const syncConfigForm = (config: Api.Mumble.Config) => {
    configForm.enabled = config.enabled
    configForm.url = config.url
    configForm.port = config.port
    configForm.server_name = config.server_name
    configForm.auth_secret_set = config.auth_secret_set
    configForm.auth_secret = ''
  }

  const loadProfile = async () => {
    loading.value = true
    try {
      const data = await fetchMumbleProfile()
      profile.value = data
      syncConfigForm(data.config)
    } catch {
      /* empty */
    } finally {
      loading.value = false
    }
  }

  const loadRoleGroups = async () => {
    if (!isAdmin.value) return
    loadingRoleGroups.value = true
    try {
      roleGroups.value = await fetchMumbleRoleGroups()
    } catch {
      /* empty */
    } finally {
      loadingRoleGroups.value = false
    }
  }

  const copyText = async (value?: string) => {
    if (!value) return
    await navigator.clipboard.writeText(value)
    ElMessage.success(t('mumble.copySuccess'))
  }

  const handleResetPassword = async () => {
    try {
      await ElMessageBox.confirm(t('mumble.resetConfirm'), t('common.tips'), {
        type: 'warning',
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel')
      })
    } catch {
      return
    }
    resetting.value = true
    try {
      const account = await resetMumblePassword()
      if (profile.value) {
        profile.value.account = account
      }
      ElMessage.success(t('mumble.resetSuccess'))
    } catch {
      /* empty */
    } finally {
      resetting.value = false
    }
  }

  const openMumble = () => {
    if (!profile.value?.account.quick_url) return
    window.location.href = profile.value.account.quick_url
  }

  const handleSaveConfig = async () => {
    savingConfig.value = true
    try {
      await updateMumbleConfig({
        enabled: configForm.enabled,
        url: configForm.url,
        port: configForm.port,
        server_name: configForm.server_name,
        auth_secret: configForm.auth_secret || undefined
      })
      ElMessage.success(t('common.updateSuccess'))
      await loadProfile()
    } catch {
      /* empty */
    } finally {
      savingConfig.value = false
    }
  }

  const handleSaveRoleGroups = async () => {
    savingRoleGroups.value = true
    try {
      await updateMumbleRoleGroups({
        mappings: roleGroups.value.map((item) => ({
          role_code: item.role_code,
          group_name: item.group_name,
          enabled: item.enabled
        }))
      })
      ElMessage.success(t('mumble.roleGroups.saveSuccess'))
      await loadRoleGroups()
      await loadProfile()
    } catch {
      /* empty */
    } finally {
      savingRoleGroups.value = false
    }
  }

  onMounted(() => {
    loadProfile()
    loadRoleGroups()
  })
</script>

<style scoped>
  .mumble-page {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .card-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
  }

  .section-title {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }

  .section-desc {
    margin: 6px 0 0;
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }

  .mono {
    font-family:
      ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace;
  }
</style>
