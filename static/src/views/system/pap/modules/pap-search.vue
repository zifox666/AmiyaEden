<template>
  <ArtSearchBar
    ref="searchBarRef"
    v-model="formData"
    :items="formItems"
    @reset="handleReset"
    @search="handleSearch"
  >
    <template #fetchBtn>
      <ElButton type="primary" :loading="fetching" :icon="Download" @click="handleFetch">
        {{ t('alliancePap.fetchLatest') }}
      </ElButton>
      <ElButton type="primary" :loading="fetching" :icon="Download" @click="openImportSEATDialog">
        {{ t('alliancePap.importBtnSEAT') }}
      </ElButton>
      <ArtExcelImport @import-success="handleImportXLS">
        {{ t('alliancePap.importBtnXLS') }}
      </ArtExcelImport>
    </template>
  </ArtSearchBar>

  <!-- 从 SEAT 导入弹窗 -->
  <ElDialog
    v-model="dialogVisible"
    :title="$t('alliancePap.importBtnSEAT')"
    width="600px"
    destroy-on-close
  >
    <ElForm ref="formRef" :model="formDataSEAT" :rules="formRulesSEAT" label-width="150px">
      <ElFormItem
        :label="$t('alliancePap.importFormSEAT.fields.laravelSession')"
        prop="laravelSession"
      >
        <ElInput
          v-model="formDataSEAT.laravelSession"
          :placeholder="$t('alliancePap.importFormSEAT.fields.laravelSession')"
        />
      </ElFormItem>
      <ElFormItem :label="$t('alliancePap.importFormSEAT.fields.cfClearance')" prop="cfClearance">
        <ElInput
          v-model="formDataSEAT.cfClearance"
          :placeholder="$t('alliancePap.importFormSEAT.fields.cfClearance')"
        />
      </ElFormItem>
      <ElFormItem :label="$t('alliancePap.importFormSEAT.fields.UA')" prop="UA">
        <ElInput
          v-model="formDataSEAT.UA"
          :placeholder="$t('alliancePap.importFormSEAT.fields.UA')"
        />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <ElButton @click="dialogVisible = false">{{ $t('common.cancel') }}</ElButton>
      <ElButton type="primary" :loading="submitLoading" @click="handleImportSEAT">
        {{ $t('common.confirm') }}
      </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { Download } from '@element-plus/icons-vue'
  import { ElButton, type FormInstance, type FormRules } from 'element-plus'
  import { useI18n } from 'vue-i18n'
  import axios from 'axios'

  interface Props {
    modelValue: Record<string, any>
    fetching?: boolean
  }
  interface Emits {
    (e: 'update:modelValue', value: Record<string, any>): void
    (e: 'search', params: Record<string, any>): void
    (e: 'reset'): void
    (e: 'fetch'): void
  }

  const props = withDefaults(defineProps<Props>(), { fetching: false })
  const emit = defineEmits<Emits>()
  const { t } = useI18n()

  const searchBarRef = ref()
  const formData = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
  })

  const formItems = computed(() => [
    {
      label: t('alliancePap.selectMonth'),
      key: 'month',
      type: 'date',
      props: {
        type: 'month',
        format: 'YYYY-MM',
        valueFormat: 'YYYY-MM',
        clearable: false
      }
    },
    {
      key: 'fetchBtn',
      label: ''
    }
  ])

  function handleReset() {
    emit('reset')
  }

  async function handleSearch() {
    emit('search', formData.value)
  }

  function handleFetch() {
    emit('fetch')
  }

  function handleImportXLS(rows: Record<string, unknown>[]) {
    emit('import', rows)
  }

  // ─── 从 SEAT 导入 ───
  const dialogVisible = ref(false)
  const submitLoading = ref(false)
  const formRef = ref<FormInstance>()

  const formDataSEAT = reactive({
    laravelSession: '',
    cfClearance: '',
    UA: ''
  })

  const formRulesSEAT: FormRules = {
    laravelSession: [
      {
        required: true,
        message: t('alliancePap.importFormSEAT.fields.laravelSession'),
        trigger: 'blur'
      }
    ],
    cfClearance: [
      {
        required: false,
        message: t('alliancePap.importFormSEAT.fields.cfClearance'),
        trigger: 'blur'
      }
    ],
    UA: [{ required: false, message: t('alliancePap.importFormSEAT.fields.UA'), trigger: 'blur' }]
  }

  function resetFormDataSEAT() {
    formDataSEAT.laravelSession = ''
    formDataSEAT.cfClearance = ''
    formDataSEAT.UA = ''
  }

  function openImportSEATDialog() {
    resetFormDataSEAT()
    dialogVisible.value = true
  }

  const { VITE_API_URL } = import.meta.env

  async function handleImportSEAT() {
    if (!formRef.value) return
    await formRef.value.validate()

    let rows: Record<string, unknown>[] = []

    submitLoading.value = true
    try {
      /* 创建Axios实例 */
      const axiosInstance = axios.create({
        timeout: 10000,
        baseURL: VITE_API_URL,
        // SECURITY WARNING:
        // The X-Cookie header below intentionally contains real session tokens
        // (laravel_session and optionally cf_clearance) that are
        // proxied through our backend/Nginx to the SEAT server.
        //
        // These credentials MUST be treated as highly sensitive:
        //   - Backend/proxy access logs MUST NOT record the X-Cookie header.
        //   - Any request/response logging or tracing MUST redact or omit
        //     the values of this header.
        //   - Only HTTPS should be used for this request.
        //
        // Changing this behavior will break SEAT integration; if you need
        // to modify it, coordinate with security and infrastructure teams.
        headers: {
          Accept: 'application/json, text/javascript, */*; q=0.01',
          'X-Accept-Encoding': 'gzip, deflate, br, zstd',
          'Accept-Language': 'zh-CN,zh;q=0.9,en;q=0.8,en-GB;q=0.7,en-US;q=0.6',
          'Cache-Control': 'no-cache',
          'X-Cookie':
            `laravel_session=${formDataSEAT.laravelSession}` +
            (formDataSEAT.cfClearance === '' ? '' : `;cf_clearance=${formDataSEAT.cfClearance}`),
          Pragma: 'no-cache',
          Priority: 'u=1, i',
          'X-Sec-Ch-Ua': `"Not:A-Brand";v="99", "Microsoft Edge";v="145", "Chromium";v="145"`,
          'X-Sec-Ch-Ua-Mobile': '?0',
          'X-Sec-Ch-Ua-Platform': `"Windows"`,
          'X-Sec-Fetch-Dest': 'empty',
          'X-Sec-Fetch-Mode': 'cors',
          'X-Sec-Fetch-Site': 'same-origin',
          'X-User-Agent':
            formDataSEAT.UA === ''
              ? 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36'
              : formDataSEAT.UA,
          'X-Requested-With': 'XMLHttpRequest'
        }
      })

      const response = await axiosInstance.get('/seatproxy/tools/paptracking')

      if (response.status == 200 && response.data.data) {
        for (const item of response.data.data) {
          const temp: Record<string, unknown> = {
            主角色: item.character,
            '月 PAP': item.pap_count,
            数据时间: item.logoff_date
          }
          rows.push(temp)
        }
      } else {
        throw new Error(t('alliancePap.importFormSEAT.fetchPapError', { status: response.status }))
      }

      dialogVisible.value = false
      emit('import', rows)
    } catch (e: any) {
      ElMessage.error(e?.message ?? t('common.error'))
    } finally {
      submitLoading.value = false
    }
  }
</script>
