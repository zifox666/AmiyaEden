<!-- 通过邀请链接加入舰队 -->
<template>
  <div class="join-page art-full-height flex items-center justify-center">
    <ElCard style="width: 420px" shadow="never">
      <template #header>
        <div class="text-center">
          <h2 class="text-lg font-medium">{{ $t('fleet.join.title') }}</h2>
        </div>
      </template>

      <!-- 邀请码信息 -->
      <div v-if="inviteCode" class="mb-4 p-3 rounded bg-gray-50 text-sm">
        <span class="text-gray-500">{{ $t('fleet.invite.code') }}：</span>
        <code class="text-xs break-all">{{ inviteCode }}</code>
      </div>

      <!-- 无邀请码 -->
      <ElEmpty v-if="!inviteCode" description="无效的邀请链接，缺少邀请码" />

      <template v-else>
        <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="80px">
          <ElFormItem :label="$t('fleet.join.selectCharacter')" prop="character_id">
            <ElSelect
              v-model="formData.character_id"
              :placeholder="$t('fleet.join.selectCharacterPlaceholder')"
              style="width: 100%"
              :loading="charLoading"
            >
              <ElOption
                v-for="c in characters"
                :key="c.character_id"
                :label="c.character_name"
                :value="c.character_id"
              >
                <div class="flex items-center gap-2">
                  <img
                    :src="c.portrait_url"
                    :alt="c.character_name"
                    class="w-6 h-6 rounded-full object-cover"
                  />
                  <span>{{ c.character_name }}</span>
                </div>
              </ElOption>
            </ElSelect>
          </ElFormItem>
        </ElForm>

        <div class="flex justify-end gap-2 mt-4">
          <ElButton @click="goBack">{{ $t('common.cancel') }}</ElButton>
          <ElButton
            type="primary"
            :loading="submitLoading"
            :disabled="!formData.character_id"
            @click="handleJoin"
          >
            {{ $t('common.confirm') }}
          </ElButton>
        </div>
      </template>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ElCard, ElForm, ElFormItem, ElSelect, ElOption, ElButton, ElEmpty, type FormInstance, type FormRules } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useRoute, useRouter } from 'vue-router'
  import { fetchMyCharacters } from '@/api/auth'
  import { joinFleet } from '@/api/fleet'

  defineOptions({ name: 'JoinFleet' })

  const { t } = useI18n()
  const route = useRoute()
  const router = useRouter()

  const inviteCode = computed(() => route.query.code as string | undefined)

  // ---- 数据 ----
  const characters = ref<Api.Auth.EveCharacter[]>([])
  const charLoading = ref(false)
  const submitLoading = ref(false)
  const formRef = ref<FormInstance>()

  const formData = reactive({ character_id: undefined as number | undefined })

  const formRules: FormRules = {
    character_id: [{ required: true, message: t('fleet.join.selectCharacterPlaceholder'), trigger: 'change' }]
  }

  // ---- 加载角色列表 ----
  const loadCharacters = async () => {
    charLoading.value = true
    try {
      const res = await fetchMyCharacters()
      characters.value = res ?? []
      // 默认选中第一个
      if (characters.value.length === 1) {
        formData.character_id = characters.value[0].character_id
      }
    } catch {
      characters.value = []
    } finally {
      charLoading.value = false
    }
  }

  // ---- 提交 ----
  const handleJoin = async () => {
    if (!formRef.value) return
    try {
      await formRef.value.validate()
    } catch {
      return
    }

    submitLoading.value = true
    try {
      await joinFleet({
        code: inviteCode.value!,
        character_id: formData.character_id!
      })
      ElMessage.success(t('fleet.join.joinSuccess'))
      router.push({ name: 'MyPap' })
    } catch (e: any) {
      const msg = e?.message || t('fleet.join.invalidCode')
      ElMessage.error(msg)
    } finally {
      submitLoading.value = false
    }
  }

  const goBack = () => {
    router.back()
  }

  onMounted(() => {
    if (inviteCode.value) {
      loadCharacters()
    }
  })
</script>
