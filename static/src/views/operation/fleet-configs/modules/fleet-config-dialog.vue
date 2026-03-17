<!-- 舰队配置 创建/编辑/查看 弹窗 -->
<template>
  <ElDialog
    :model-value="visible"
    :title="
      readonly && editing
        ? $t('fleetConfig.view')
        : editing
          ? $t('fleetConfig.edit')
          : $t('fleetConfig.create')
    "
    width="760px"
    destroy-on-close
    @update:model-value="emit('update:visible', $event)"
  >
    <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="100px">
      <ElFormItem :label="$t('fleetConfig.fields.name')" prop="name">
        <ElInput
          v-model="formData.name"
          :placeholder="$t('fleetConfig.fields.namePlaceholder')"
          :disabled="readonly"
        />
      </ElFormItem>
      <ElFormItem :label="$t('fleetConfig.fields.description')">
        <ElInput
          v-model="formData.description"
          type="textarea"
          :rows="2"
          :placeholder="$t('fleetConfig.fields.descriptionPlaceholder')"
          :disabled="readonly"
        />
      </ElFormItem>

      <!-- 装配列表 -->
      <ElDivider content-position="left">{{ $t('fleetConfig.fields.fittings') }}</ElDivider>

      <div v-for="(fit, idx) in formData.fittings" :key="fit.id || idx" class="fitting-entry">
        <div class="fitting-header">
          <strong>{{ fit.fitting_name || $t('fleetConfig.newFitting') }}</strong>
          <ElButton
            v-if="!readonly"
            type="danger"
            link
            size="small"
            class="ml-auto"
            @click="removeFitting(idx)"
          >
            {{ $t('common.delete') }}
          </ElButton>
        </div>

        <div class="fitting-fields">
          <ElFormItem
            :label="$t('fleetConfig.fields.fittingName')"
            :prop="`fittings.${idx}.fitting_name`"
            :rules="[
              {
                required: true,
                message: $t('fleetConfig.fields.fittingNamePlaceholder'),
                trigger: 'blur'
              }
            ]"
          >
            <ElInput
              v-model="fit.fitting_name"
              :placeholder="$t('fleetConfig.fields.fittingNamePlaceholder')"
              :disabled="readonly"
            />
          </ElFormItem>

          <ElFormItem :label="$t('fleetConfig.fields.srpAmount')">
            <ElRow style="width: 100%">
              <ElInputNumber
                v-model="fit.srp_amount"
                :min="0"
                :precision="0"
                :disabled="readonly"
                style="width: 200px"
              />
            </ElRow>
          </ElFormItem>

          <ElFormItem
            :label="$t('fleetConfig.fields.eft')"
            :prop="!readonly ? `fittings.${idx}.eft` : undefined"
            :rules="
              !readonly
                ? [
                    {
                      required: true,
                      message: $t('fleetConfig.fields.eftPlaceholder'),
                      trigger: 'blur'
                    }
                  ]
                : []
            "
          >
            <!-- 查看模式：显示本地化 EFT -->
            <template v-if="readonly">
              <ElInput
                :model-value="eftMap[fit.id ?? 0] ?? ''"
                type="textarea"
                :rows="6"
                readonly
                class="eft-textarea"
              />
            </template>
            <!-- 编辑/创建模式：可输入英文 EFT -->
            <template v-else>
              <ElInput
                v-model="fit.eft"
                type="textarea"
                :rows="6"
                :placeholder="$t('fleetConfig.fields.eftPlaceholder')"
                class="eft-textarea"
                @change="onEFTChange(idx)"
              />
            </template>
          </ElFormItem>

          <div class="fitting-actions">
            <!-- 查看模式：复制 + 保存到游戏 -->
            <template v-if="readonly">
              <ElButton size="small" @click="copyEFT(eftMap[fit.id ?? 0] ?? '')">
                {{ $t('fleetConfig.copyEFT') }}
              </ElButton>
              <ElButton size="small" type="primary" @click="openSaveToGame(fit.id ?? 0)">
                {{ $t('fleetConfig.saveToGame') }}
              </ElButton>
            </template>
            <!-- 编辑模式：仅复制 -->
            <template v-else>
              <ElButton size="small" @click="copyEFT(fit.eft)">
                {{ $t('fleetConfig.copyEFT') }}
              </ElButton>
            </template>
            <!-- 装备设置：所有用户均可查看，管理员可编辑 -->
            <ElButton v-if="fit.id" size="small" type="warning" @click="openItemSettings(fit.id)">
              {{ $t('fleetConfig.itemSettings') }}
            </ElButton>
          </div>
        </div>
      </div>

      <div v-if="!readonly" class="fitting-add-actions">
        <ElButton type="primary" plain @click="addEmptyFitting">
          {{ $t('fleetConfig.addFitting') }}
        </ElButton>
        <ElButton plain @click="openImportDialog">
          {{ $t('fleetConfig.importFromFitting') }}
        </ElButton>
      </div>
    </ElForm>

    <template #footer>
      <ElButton @click="emit('update:visible', false)">{{ $t('common.cancel') }}</ElButton>
      <ElButton v-if="!readonly" type="primary" :loading="submitLoading" @click="handleSubmit">
        {{ $t('common.confirm') }}
      </ElButton>
    </template>

    <!-- 从用户装配导入-->
    <ElDialog
      v-model="showImportDialog"
      :title="$t('fleetConfig.importFromFitting')"
      width="400px"
      append-to-body
      destroy-on-close
    >
      <ElForm label-width="80px">
        <ElFormItem :label="$t('fleetConfig.fields.fitting')">
          <ElSelect
            v-model="importForm.fitting_id"
            :placeholder="$t('fleetConfig.fields.fittingPlaceholder')"
            style="width: 100%"
            filterable
          >
            <ElOption
              v-for="f in userFittings"
              :key="`${f.character_id}_${f.fitting_id}`"
              :label="`${f.ship_name} - ${f.name}`"
              :value="f.fitting_id"
            />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="showImportDialog = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="importLoading" @click="handleImport">
          {{ $t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>

    <!-- 保存到游戏弹窗 -->
    <ElDialog
      v-model="saveToGameVisible"
      :title="$t('fleetConfig.saveToGameTitle')"
      width="380px"
      append-to-body
      destroy-on-close
    >
      <ElForm label-width="80px">
        <ElFormItem :label="$t('fleetConfig.fields.character')">
          <ElSelect
            v-model="saveToGameCharacterID"
            :placeholder="$t('fleet.fields.fcPlaceholder')"
            style="width: 100%"
          >
            <ElOption
              v-for="c in characters"
              :key="c.character_id"
              :label="c.character_name"
              :value="c.character_id"
            />
          </ElSelect>
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="saveToGameVisible = false">{{ $t('common.cancel') }}</ElButton>
        <ElButton type="primary" :loading="saveToGameLoading" @click="handleSaveToGame">
          {{ $t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>

    <!-- 装备设置弹窗 -->
    <ElDialog
      v-model="itemSettingsVisible"
      :title="$t('fleetConfig.itemSettingsTitle')"
      width="780px"
      append-to-body
      destroy-on-close
    >
      <div v-loading="itemSettingsLoading">
        <template v-for="(items, group) in groupedItems" :key="group">
          <ElDivider content-position="left">{{
            $t(`fleetConfig.slotGroup.${group}`, group)
          }}</ElDivider>
          <div v-for="item in items" :key="item.id" class="item-setting-row">
            <div class="item-setting-info">
              <img
                :src="`https://images.evetech.net/types/${item.type_id}/icon?size=32`"
                class="item-setting-icon"
                loading="lazy"
              />
              <span class="item-setting-name">{{ item.type_name }} ×{{ item.quantity }}</span>
            </div>
            <div class="item-setting-controls">
              <ElSelect
                v-model="itemEdits[item.id].importance"
                size="small"
                style="width: 100px"
                :disabled="!canManage"
                @change="onImportanceChange(item.id)"
              >
                <ElOption value="required" :label="$t('fleetConfig.importance.required')" />
                <ElOption value="optional" :label="$t('fleetConfig.importance.optional')" />
                <ElOption value="replaceable" :label="$t('fleetConfig.importance.replaceable')" />
              </ElSelect>
              <ElSelect
                v-model="itemEdits[item.id].penalty"
                size="small"
                style="width: 100px"
                :disabled="!canManage"
              >
                <ElOption value="half" :label="$t('fleetConfig.penalty.half')" />
                <ElOption value="none" :label="$t('fleetConfig.penalty.none')" />
              </ElSelect>
              <ElSelect
                v-if="itemEdits[item.id].importance === 'replaceable'"
                v-model="itemEdits[item.id].replacement_penalty"
                size="small"
                style="width: 120px"
                :disabled="!canManage"
              >
                <ElOption value="half" :label="$t('fleetConfig.replacementPenalty.half')" />
                <ElOption value="none" :label="$t('fleetConfig.replacementPenalty.none')" />
              </ElSelect>
            </div>
            <!-- 替代品列表 -->
            <div v-if="itemEdits[item.id].importance === 'replaceable'" class="item-replacements">
              <ElTag
                v-for="repId in itemEdits[item.id].replacements"
                :key="repId"
                :closable="canManage"
                size="small"
                class="replacement-tag"
                @close="removeReplacement(item.id, repId)"
              >
                {{ replacementNameCache[repId] ?? '' }} #{{ repId }}
              </ElTag>
              <SdeSearchSelect
                v-if="canManage"
                v-model="replacementSearchId"
                :placeholder="$t('fleetConfig.replacementSearch')"
                style="width: 200px"
                @select="onReplacementSelect(item.id, $event)"
              />
            </div>
          </div>
        </template>
      </div>
      <template #footer>
        <ElButton @click="itemSettingsVisible = false">{{ $t('common.close') }}</ElButton>
        <ElButton
          v-if="canManage"
          type="primary"
          :loading="itemSettingsSaving"
          @click="handleSaveItemSettings"
        >
          {{ $t('common.confirm') }}
        </ElButton>
      </template>
    </ElDialog>
  </ElDialog>
</template>

<script setup lang="ts">
  import {
    createFleetConfig,
    updateFleetConfig,
    importFittingFromUser,
    exportFittingToESI,
    fetchFleetConfigEFT,
    fetchFittingItems,
    updateFittingItemsSettings
  } from '@/api/fleet-config'
  import { fetchMyCharacters } from '@/api/auth'
  import { fetchInfoFittings } from '@/api/eve-info'
  import type { FormInstance, FormRules } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import { useClipboard } from '@vueuse/core'
  import { useUserStore } from '@/store/modules/user'
  import SdeSearchSelect from '@/components/business/SdeSearchSelect.vue'

  const props = defineProps<{
    visible: boolean
    editing: Api.FleetConfig.FleetConfigItem | null
    readonly?: boolean
  }>()

  const emit = defineEmits<{
    (e: 'update:visible', val: boolean): void
    (e: 'success'): void
    (e: 'created', config: Api.FleetConfig.FleetConfigItem): void
  }>()

  const { t, locale } = useI18n()
  const { copy } = useClipboard()

  const userStore = useUserStore()
  const canManage = computed(() => {
    const roles = userStore.getUserInfo?.roles ?? []
    return roles.some((r) => ['super_admin', 'admin', 'fc', 'srp'].includes(r))
  })

  const formRef = ref<FormInstance>()
  const submitLoading = ref(false)

  // 内部使用的装配类型（视图模式下附带 id）
  type InternalFitting = Api.FleetConfig.FittingReq & { id?: number }

  const formData = reactive({
    name: '',
    description: '',
    fittings: [] as InternalFitting[]
  })

  const formRules: FormRules = {
    name: [{ required: true, message: t('fleetConfig.fields.namePlaceholder'), trigger: 'blur' }]
  }

  // ─── 本地 EFT 缓存（查看模式）───
  const eftMap = ref<Record<number, string>>({})

  async function loadEFTs(lang: string) {
    if (!props.editing) return
    try {
      const res = await fetchFleetConfigEFT(props.editing.id, lang)
      const map: Record<number, string> = {}
      for (const item of res?.fittings ?? []) {
        map[item.id] = item.eft
      }
      eftMap.value = map
    } catch {
      eftMap.value = {}
    }
  }

  // ─── 初始化（编辑/查看模式填充）───
  watch(
    () => props.visible,
    async (val) => {
      if (!val) return
      eftMap.value = {}
      if (props.editing) {
        formData.name = props.editing.name
        formData.description = props.editing.description
        if (props.readonly) {
          // 查看模式：展示本地化 EFT，formData.fittings 仅用于渲染字段
          formData.fittings = (props.editing.fittings ?? []).map<InternalFitting>((f) => ({
            fitting_name: f.fitting_name,
            eft: '',
            srp_amount: f.srp_amount,
            id: f.id
          }))
          await loadEFTs(locale.value)
        } else {
          // 编辑模式：先填充基础字段，再用英文 EFT 填充 textarea
          formData.fittings = (props.editing.fittings ?? []).map<InternalFitting>((f) => ({
            fitting_name: f.fitting_name,
            eft: '',
            srp_amount: f.srp_amount,
            id: f.id
          }))
          // 异步加载英文 EFT 填入编辑模式
          const eftRes = await fetchFleetConfigEFT(props.editing.id, 'en').catch(() => null)
          if (eftRes) {
            const eftById: Record<number, string> = {}
            for (const item of eftRes.fittings) {
              eftById[item.id] = item.eft
            }
            const idsInOrder = (props.editing.fittings ?? []).map((f) => f.id)
            idsInOrder.forEach((id, idx) => {
              if (formData.fittings[idx] && eftById[id]) {
                formData.fittings[idx].eft = eftById[id]
              }
            })
          }
        }
      } else {
        formData.name = ''
        formData.description = ''
        formData.fittings = []
      }
      loadCharacters()
    }
  )

  // ─── 装配操作 ───
  function addEmptyFitting() {
    formData.fittings.push({
      fitting_name: '',
      eft: '',
      srp_amount: 0
    })
  }

  function removeFitting(idx: number) {
    formData.fittings.splice(idx, 1)
  }

  /** EFT 头行自动填充装配 **/
  function onEFTChange(idx: number) {
    const eft = formData.fittings[idx].eft
    const match = eft.match(/^\[(.+?),\s*(.+?)\]\s*$/m)
    if (match && !formData.fittings[idx].fitting_name) {
      formData.fittings[idx].fitting_name = match[2].trim()
    }
  }

  function copyEFT(eft: string) {
    copy(eft)
    ElMessage.success(t('fleetConfig.eftCopied'))
  }

  // ─── 提交 ───
  async function handleSubmit() {
    if (!formRef.value) return
    await formRef.value.validate()
    if (formData.fittings.length === 0) {
      ElMessage.warning(t('fleetConfig.noFittings'))
      return
    }
    submitLoading.value = true
    try {
      if (props.editing) {
        await updateFleetConfig(props.editing.id, {
          name: formData.name,
          description: formData.description,
          fittings: formData.fittings
        })
        ElMessage.success(t('fleetConfig.updateSuccess'))
      } else {
        const newConfig = await createFleetConfig({
          name: formData.name,
          description: formData.description,
          fittings: formData.fittings
        })
        ElMessage.success(t('fleetConfig.createSuccess'))
        emit('created', newConfig)
        return
      }
      emit('success')
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('common.error'))
    } finally {
      submitLoading.value = false
    }
  }

  // ─── 角色列表 ───
  const characters = ref<Api.Auth.EveCharacter[]>([])

  async function loadCharacters() {
    try {
      const res = await fetchMyCharacters()
      characters.value = res ?? []
    } catch {
      characters.value = []
    }
  }

  // ─── 用户装配导入 ───
  const showImportDialog = ref(false)
  const importLoading = ref(false)
  const importForm = reactive({
    fitting_id: undefined as number | undefined
  })
  const userFittings = ref<Api.EveInfo.FittingResponse[]>([])

  function openImportDialog() {
    showImportDialog.value = true
    loadUserFittings()
  }

  async function loadUserFittings() {
    try {
      const res = await fetchInfoFittings({ language: locale.value })
      userFittings.value = res?.fittings ?? []
    } catch {
      userFittings.value = []
    }
  }

  async function handleImport() {
    if (!importForm.fitting_id) {
      ElMessage.warning(t('fleetConfig.importSelectRequired'))
      return
    }
    const selected = userFittings.value.find((f) => f.fitting_id === importForm.fitting_id)
    if (!selected) {
      ElMessage.warning(t('fleetConfig.importSelectRequired'))
      return
    }
    importLoading.value = true
    try {
      const result = await importFittingFromUser({
        character_id: selected.character_id,
        fitting_id: importForm.fitting_id
      })
      if (result) {
        formData.fittings.push({
          fitting_name: result.fitting_name ?? '',
          eft: result.eft ?? '',
          srp_amount: result.srp_amount ?? 0
        })
      }
      showImportDialog.value = false
      importForm.fitting_id = undefined
      ElMessage.success(t('fleetConfig.importSuccess'))
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('common.error'))
    } finally {
      importLoading.value = false
    }
  }

  // ─── 装备设置 ───
  const itemSettingsVisible = ref(false)
  const itemSettingsLoading = ref(false)
  const itemSettingsSaving = ref(false)
  const itemSettingsFittingId = ref(0)
  const itemSettingsData = ref<Api.FleetConfig.FittingItemsResponse | null>(null)
  const itemEdits = ref<Record<number, Api.FleetConfig.ItemSettingUpdate>>({})
  const replacementSearchId = ref<number | null>(null)
  /** 本地缓存新增替代品的名称 type_id -> type_name */
  const replacementNameCache = ref<Record<number, string>>({})

  async function openItemSettings(fittingId: number) {
    if (!props.editing) return
    itemSettingsFittingId.value = fittingId
    itemSettingsLoading.value = true
    itemSettingsVisible.value = true
    replacementNameCache.value = {}
    try {
      const res = await fetchFittingItems(props.editing.id, fittingId, locale.value)
      itemSettingsData.value = res
      // 初始化编辑状态
      const edits: Record<number, Api.FleetConfig.ItemSettingUpdate> = {}
      for (const item of res?.items ?? []) {
        edits[item.id] = {
          id: item.id,
          importance: item.importance,
          penalty: item.penalty,
          replacement_penalty: item.replacement_penalty,
          replacements: item.replacements?.map((r) => r.type_id) ?? []
        }
        // 缓存已有替代品名称
        for (const r of item.replacements ?? []) {
          replacementNameCache.value[r.type_id] = r.type_name
        }
      }
      itemEdits.value = edits
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('common.error'))
      itemSettingsVisible.value = false
    } finally {
      itemSettingsLoading.value = false
    }
  }

  function onImportanceChange(itemId: number) {
    const edit = itemEdits.value[itemId]
    if (!edit) return
    if (edit.importance !== 'replaceable') {
      edit.replacements = []
    }
  }

  function removeReplacement(itemId: number, typeId: number) {
    const edit = itemEdits.value[itemId]
    if (!edit) return
    edit.replacements = edit.replacements?.filter((id) => id !== typeId) ?? []
  }

  function onReplacementSelect(itemId: number, item: Api.Sde.FuzzySearchItem | null) {
    if (!item) return
    const edit = itemEdits.value[itemId]
    if (!edit) return
    if (!edit.replacements) edit.replacements = []
    if (!edit.replacements.includes(item.id)) {
      edit.replacements.push(item.id)
      replacementNameCache.value[item.id] = item.name
    }
    replacementSearchId.value = null
  }

  /** 按 flag_group 分组 */
  const groupedItems = computed(() => {
    const items = itemSettingsData.value?.items ?? []
    const groups: Record<string, Api.FleetConfig.FittingItemDetail[]> = {}
    for (const item of items) {
      const g = item.flag_group || 'Other'
      if (!groups[g]) groups[g] = []
      groups[g].push(item)
    }
    return groups
  })

  async function handleSaveItemSettings() {
    if (!props.editing) return
    itemSettingsSaving.value = true
    try {
      const items = Object.values(itemEdits.value)
      await updateFittingItemsSettings(props.editing.id, itemSettingsFittingId.value, {
        items
      })
      ElMessage.success(t('fleetConfig.itemSettingsSaved'))
      itemSettingsVisible.value = false
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('common.error'))
    } finally {
      itemSettingsSaving.value = false
    }
  }

  // ─── 保存到游戏───
  const saveToGameVisible = ref(false)
  const saveToGameLoading = ref(false)
  const saveToGameFittingID = ref<number>(0)
  const saveToGameCharacterID = ref<number | undefined>(undefined)

  function openSaveToGame(fittingID: number) {
    saveToGameFittingID.value = fittingID
    saveToGameCharacterID.value = characters.value[0]?.character_id
    saveToGameVisible.value = true
  }

  async function handleSaveToGame() {
    if (!saveToGameCharacterID.value) {
      ElMessage.warning(t('fleetConfig.saveToGameCharacterRequired'))
      return
    }
    if (!props.editing) return
    saveToGameLoading.value = true
    try {
      await exportFittingToESI({
        character_id: saveToGameCharacterID.value,
        fleet_config_id: props.editing.id,
        fitting_item_id: saveToGameFittingID.value
      })
      saveToGameVisible.value = false
      ElMessage.success(t('fleetConfig.saveToGameSuccess'))
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('common.error'))
    } finally {
      saveToGameLoading.value = false
    }
  }
</script>

<style scoped>
  .fitting-entry {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    padding: 12px;
    margin-bottom: 12px;
  }

  .fitting-header {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
    padding-bottom: 8px;
    border-bottom: 1px solid var(--el-border-color-extra-light);
  }

  .fitting-fields {
    padding-top: 4px;
  }

  .fitting-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
    margin-top: 4px;
  }

  .fitting-add-actions {
    display: flex;
    gap: 8px;
    margin-top: 16px;
  }

  .eft-textarea :deep(.el-textarea__inner) {
    font-family: 'Courier New', Courier, monospace;
    font-size: 12px;
  }

  .item-setting-row {
    padding: 8px 0;
    border-bottom: 1px solid var(--el-border-color-extra-light);
  }

  .item-setting-info {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
  }

  .item-setting-icon {
    width: 32px;
    height: 32px;
    border-radius: 4px;
  }

  .item-setting-name {
    font-size: 14px;
  }

  .item-setting-controls {
    display: flex;
    gap: 8px;
    margin-bottom: 6px;
  }

  .item-replacements {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    padding-left: 40px;
  }

  .replacement-tag {
    margin: 0;
  }
</style>
