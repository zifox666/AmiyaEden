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
            <ElRow class="million-isk-input">
              <ElInputNumber
                :model-value="iskToMillionInput(fit.srp_amount)"
                :min="0"
                :precision="2"
                :step="1"
                :disabled="readonly"
                class="million-isk-input__control"
                @update:model-value="updateFittingSrpAmount(fit, $event)"
              />
              <span class="million-isk-input__suffix">{{ $t('common.millionIsk') }}</span>
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
            <!-- 查看模式：复制 -->
            <template v-if="readonly">
              <ElButton size="small" @click="copyEFT(eftMap[fit.id ?? 0] ?? '')">
                {{ $t('fleetConfig.copyEFT') }}
              </ElButton>
            </template>
            <!-- 编辑模式：仅复制 -->
            <template v-else>
              <ElButton size="small" @click="copyEFT(fit.eft)">
                {{ $t('fleetConfig.copyEFT') }}
              </ElButton>
            </template>
            <ElButton v-if="fit.id" size="small" type="primary" @click="openSaveToGame(fit.id)">
              {{ $t('fleetConfig.saveToGame') }}
            </ElButton>
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
            v-model="importForm.selection"
            :placeholder="$t('fleetConfig.fields.fittingPlaceholder')"
            style="width: 100%"
            filterable
          >
            <ElOption
              v-for="f in userFittings"
              :key="`${f.character_id}_${f.fitting_id}`"
              :label="`${f.ship_name} - ${f.name}`"
              :value="buildImportSelectionKey(f.character_id, f.fitting_id)"
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
      width="min(980px, calc(100vw - 32px))"
      append-to-body
      destroy-on-close
    >
      <div v-loading="itemSettingsLoading" class="item-settings-dialog">
        <ElTabs
          v-if="groupedItemEntries.length"
          v-model="itemSettingsActiveGroup"
          class="item-settings-tabs"
        >
          <ElTabPane
            v-for="entry in groupedItemEntries"
            :key="entry.group"
            :label="formatGroupTabLabel(entry.group, entry.items.length)"
            :name="entry.group"
          >
            <div class="item-settings-list">
              <div v-for="item in entry.items" :key="item.id" class="item-setting-row">
                <div class="item-setting-main">
                  <img
                    :src="`https://images.evetech.net/types/${item.type_id}/icon?size=32`"
                    class="item-setting-icon"
                    loading="lazy"
                  />
                  <div class="item-setting-body">
                    <span class="item-setting-name">{{ item.type_name }}</span>
                    <span class="item-setting-meta"
                      >{{ item.flag }} · {{ $t('common.quantity') }} ×{{ item.quantity }}</span
                    >
                  </div>
                </div>

                <div class="item-setting-controls">
                  <div class="item-setting-field">
                    <span class="item-setting-field__label">{{
                      $t('fleetConfig.itemColumns.importance')
                    }}</span>
                    <template v-if="canManage">
                      <ElSelect
                        v-model="itemEdits[item.id].importance"
                        size="small"
                        class="item-settings-inline-select"
                        @change="onImportanceChange(item.id)"
                      >
                        <ElOption value="required" :label="$t('fleetConfig.importance.required')" />
                        <ElOption value="optional" :label="$t('fleetConfig.importance.optional')" />
                        <ElOption
                          value="replaceable"
                          :label="$t('fleetConfig.importance.replaceable')"
                        />
                      </ElSelect>
                    </template>
                    <ElTag v-else size="small" effect="plain">
                      {{ importanceLabel(itemEdits[item.id].importance) }}
                    </ElTag>
                  </div>

                  <div class="item-setting-field">
                    <span class="item-setting-field__label">{{
                      $t('fleetConfig.itemColumns.penalty')
                    }}</span>
                    <template v-if="canManage">
                      <ElSelect
                        v-model="itemEdits[item.id].penalty"
                        size="small"
                        class="item-settings-inline-select"
                      >
                        <ElOption value="half" :label="$t('fleetConfig.penalty.half')" />
                        <ElOption value="none" :label="$t('fleetConfig.penalty.none')" />
                      </ElSelect>
                    </template>
                    <ElTag v-else size="small" effect="plain">
                      {{ penaltyLabel(itemEdits[item.id].penalty) }}
                    </ElTag>
                  </div>

                  <div class="item-setting-field">
                    <span class="item-setting-field__label">{{
                      $t('fleetConfig.itemColumns.replacementPenalty')
                    }}</span>
                    <template v-if="itemEdits[item.id].importance === 'replaceable'">
                      <template v-if="canManage">
                        <ElSelect
                          v-model="itemEdits[item.id].replacement_penalty"
                          size="small"
                          class="item-settings-inline-select"
                        >
                          <ElOption
                            value="half"
                            :label="$t('fleetConfig.replacementPenalty.half')"
                          />
                          <ElOption
                            value="none"
                            :label="$t('fleetConfig.replacementPenalty.none')"
                          />
                        </ElSelect>
                      </template>
                      <ElTag v-else size="small" effect="plain">
                        {{ replacementPenaltyLabel(itemEdits[item.id].replacement_penalty) }}
                      </ElTag>
                    </template>
                    <span v-else class="item-settings-empty">-</span>
                  </div>

                  <div class="item-setting-field item-setting-field--wide">
                    <span class="item-setting-field__label">{{
                      $t('fleetConfig.itemColumns.replacements')
                    }}</span>
                    <template v-if="itemEdits[item.id].importance === 'replaceable'">
                      <div class="item-setting-replacements">
                        <span
                          class="item-setting-replacements__summary"
                          :title="getReplacementTooltip(item.id)"
                        >
                          {{ getReplacementSummary(item.id) }}
                        </span>
                        <ElButton
                          size="small"
                          link
                          type="primary"
                          @click="openReplacementEditor(item.id)"
                        >
                          {{
                            canManage
                              ? $t('fleetConfig.manageReplacements')
                              : $t('fleetConfig.viewReplacements')
                          }}
                        </ElButton>
                      </div>
                    </template>
                    <span v-else class="item-settings-empty">-</span>
                  </div>
                </div>
              </div>
            </div>
          </ElTabPane>
        </ElTabs>
        <div v-else class="item-settings-empty-state">{{ $t('fleetConfig.noItems') }}</div>
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

    <ElDialog
      v-model="replacementEditorVisible"
      :title="replacementEditorTitle"
      width="min(560px, calc(100vw - 32px))"
      append-to-body
      destroy-on-close
      @closed="closeReplacementEditor"
    >
      <template v-if="replacementEditorItem && replacementEditorEdit">
        <div class="replacement-editor-item">
          <img
            :src="`https://images.evetech.net/types/${replacementEditorItem.type_id}/icon?size=32`"
            class="item-setting-icon"
            loading="lazy"
          />
          <div class="replacement-editor-item__body">
            <span class="replacement-editor-item__name">{{ replacementEditorItem.type_name }}</span>
            <span class="replacement-editor-item__meta"
              >{{ replacementEditorItem.flag }} · {{ $t('common.quantity') }} ×{{
                replacementEditorItem.quantity
              }}</span
            >
          </div>
        </div>

        <div class="replacement-editor-tags">
          <ElTag
            v-for="repId in replacementEditorEdit.replacements"
            :key="repId"
            :closable="canManage"
            size="small"
            class="replacement-tag"
            @close="removeReplacement(replacementEditorItem.id, repId)"
          >
            {{ replacementNameCache[repId] ?? '' }} #{{ repId }}
          </ElTag>
          <span v-if="!replacementEditorEdit.replacements?.length" class="item-settings-empty">
            {{ $t('fleetConfig.noReplacements') }}
          </span>
        </div>

        <SdeSearchSelect
          v-if="canManage"
          v-model="replacementSearchId"
          :placeholder="$t('fleetConfig.replacementSearch')"
          class="replacement-editor-search"
          @select="onReplacementSelect(replacementEditorItem.id, $event)"
        />
      </template>

      <template #footer>
        <ElButton @click="closeReplacementEditor">{{ $t('common.close') }}</ElButton>
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
  import { iskToMillionInput, millionInputToIsk } from '@/utils/common'

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
    return roles.some((r) => ['super_admin', 'admin', 'senior_fc'].includes(r))
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

  function updateFittingSrpAmount(fit: InternalFitting, value: number | null | undefined) {
    fit.srp_amount = millionInputToIsk(value)
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

  // ─── 人物列表 ───
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
    selection: undefined as string | undefined
  })
  const userFittings = ref<Api.EveInfo.FittingResponse[]>([])

  function buildImportSelectionKey(characterID: number, fittingID: number) {
    return `${characterID}:${fittingID}`
  }

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
    if (!importForm.selection) {
      ElMessage.warning(t('fleetConfig.importSelectRequired'))
      return
    }
    const selected = userFittings.value.find(
      (f) => buildImportSelectionKey(f.character_id, f.fitting_id) === importForm.selection
    )
    if (!selected) {
      ElMessage.warning(t('fleetConfig.importSelectRequired'))
      return
    }
    importLoading.value = true
    try {
      const result = await importFittingFromUser({
        character_id: selected.character_id,
        fitting_id: selected.fitting_id
      })
      if (result) {
        formData.fittings.push({
          fitting_name: result.fitting_name ?? '',
          eft: result.eft ?? '',
          srp_amount: result.srp_amount ?? 0
        })
      }
      showImportDialog.value = false
      importForm.selection = undefined
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
  const itemSettingsActiveGroup = ref('')
  const replacementSearchId = ref<number | null>(null)
  const replacementEditorVisible = ref(false)
  const replacementEditorItemId = ref<number | null>(null)
  /** 本地缓存新增替代品的名称 type_id -> type_name */
  const replacementNameCache = ref<Record<number, string>>({})

  const itemSettingsItems = computed(() => itemSettingsData.value?.items ?? [])
  const replacementEditorItem = computed(
    () => itemSettingsItems.value.find((item) => item.id === replacementEditorItemId.value) ?? null
  )
  const replacementEditorEdit = computed(() => {
    if (!replacementEditorItemId.value) return null
    return itemEdits.value[replacementEditorItemId.value] ?? null
  })
  const replacementEditorTitle = computed(() =>
    t('fleetConfig.replacementEditorTitle', { name: replacementEditorItem.value?.type_name ?? '' })
  )

  async function openItemSettings(fittingId: number) {
    if (!props.editing) return
    itemSettingsFittingId.value = fittingId
    itemSettingsLoading.value = true
    itemSettingsVisible.value = true
    itemSettingsData.value = null
    itemEdits.value = {}
    replacementNameCache.value = {}
    closeReplacementEditor()
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
      itemSettingsActiveGroup.value = res?.items?.length ? res.items[0].flag_group || 'Other' : ''
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
      if (replacementEditorItemId.value === itemId) {
        closeReplacementEditor()
      }
    }
  }

  function importanceLabel(value: Api.FleetConfig.ItemSettingUpdate['importance']) {
    return t(`fleetConfig.importance.${value}`)
  }

  function penaltyLabel(value: Api.FleetConfig.ItemSettingUpdate['penalty']) {
    return t(`fleetConfig.penalty.${value}`)
  }

  function replacementPenaltyLabel(
    value: Api.FleetConfig.ItemSettingUpdate['replacement_penalty']
  ) {
    return t(`fleetConfig.replacementPenalty.${value}`)
  }

  function getReplacementLabels(itemId: number) {
    return (itemEdits.value[itemId]?.replacements ?? []).map((id) => {
      const name = replacementNameCache.value[id]
      return name ? `${name} #${id}` : `#${id}`
    })
  }

  function getReplacementSummary(itemId: number) {
    const labels = getReplacementLabels(itemId)
    if (!labels.length) return t('fleetConfig.noReplacements')
    const preview = labels.slice(0, 2).join(' / ')
    return labels.length > 2 ? `${preview} +${labels.length - 2}` : preview
  }

  function getReplacementTooltip(itemId: number) {
    const labels = getReplacementLabels(itemId)
    return labels.length ? labels.join(', ') : undefined
  }

  function openReplacementEditor(itemId: number) {
    replacementEditorItemId.value = itemId
    replacementSearchId.value = null
    replacementEditorVisible.value = true
  }

  function closeReplacementEditor() {
    replacementEditorVisible.value = false
    replacementEditorItemId.value = null
    replacementSearchId.value = null
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
    const items = itemSettingsItems.value
    const groups: Record<string, Api.FleetConfig.FittingItemDetail[]> = {}
    for (const item of items) {
      const g = item.flag_group || 'Other'
      if (!groups[g]) groups[g] = []
      groups[g].push(item)
    }
    return groups
  })

  const groupedItemEntries = computed(() =>
    Object.entries(groupedItems.value).map(([group, items]) => ({ group, items }))
  )

  function formatGroupTabLabel(group: string, count: number) {
    return `${t(`fleetConfig.slotGroup.${group}`, group)} (${count})`
  }

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

  watch(itemSettingsVisible, (visible) => {
    if (!visible) {
      closeReplacementEditor()
    }
  })

  watch(groupedItemEntries, (entries) => {
    if (!entries.length) {
      itemSettingsActiveGroup.value = ''
      return
    }

    if (!entries.some((entry) => entry.group === itemSettingsActiveGroup.value)) {
      itemSettingsActiveGroup.value = entries[0].group
    }
  })

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

  .item-setting-icon {
    width: 32px;
    height: 32px;
    border-radius: 4px;
  }

  .item-settings-dialog {
    min-height: 360px;
  }

  .item-settings-tabs :deep(.el-tabs__header) {
    margin-bottom: 10px;
  }

  .item-settings-tabs :deep(.el-tabs__content) {
    padding-top: 2px;
  }

  .item-settings-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-height: min(64vh, 640px);
    overflow: auto;
    padding-right: 4px;
  }

  .item-setting-row {
    display: grid;
    grid-template-columns: minmax(0, 240px) minmax(0, 1fr);
    gap: 12px;
    padding: 10px 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    background: var(--el-bg-color);
  }

  .item-setting-main,
  .replacement-editor-item {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .item-setting-body,
  .replacement-editor-item__body {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  .item-setting-name,
  .replacement-editor-item__name {
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .item-setting-meta,
  .replacement-editor-item__meta {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .item-setting-controls {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px 10px;
    align-items: start;
  }

  .item-setting-field {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  .item-setting-field--wide {
    grid-column: 1 / -1;
  }

  .item-setting-field__label {
    font-size: 11px;
    line-height: 1;
    color: var(--el-text-color-secondary);
  }

  .item-settings-inline-select {
    width: 100%;
  }

  .item-setting-replacements {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    min-height: 32px;
    padding: 0 8px;
    border: 1px dashed var(--el-border-color);
    border-radius: 8px;
    background: var(--el-fill-color-lighter);
  }

  .item-setting-replacements__summary {
    flex: 1;
    min-width: 0;
    font-size: 12px;
    color: var(--el-text-color-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .item-settings-empty {
    display: flex;
    align-items: center;
    min-height: 32px;
    padding: 0 8px;
    border: 1px dashed var(--el-border-color);
    border-radius: 8px;
    color: var(--el-text-color-placeholder);
    font-size: 12px;
  }

  .item-settings-empty-state {
    padding: 16px 4px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
  }

  .replacement-editor-item {
    margin-bottom: 14px;
    padding-bottom: 14px;
    border-bottom: 1px solid var(--el-border-color-extra-light);
  }

  .replacement-editor-tags {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    min-height: 32px;
  }

  .replacement-editor-search {
    width: 100%;
    margin-top: 12px;
  }

  .replacement-tag {
    margin: 0;
  }

  .million-isk-input {
    width: 100%;
    display: grid;
    grid-template-columns: minmax(0, 200px) auto;
    align-items: center;
    gap: 8px;
  }

  .million-isk-input__control {
    width: 100%;
  }

  .million-isk-input__suffix {
    font-size: 13px;
    color: var(--el-text-color-secondary);
    white-space: nowrap;
  }

  @media (max-width: 768px) {
    .item-setting-row {
      grid-template-columns: 1fr;
    }

    .item-setting-controls {
      grid-template-columns: 1fr;
    }
  }
</style>
