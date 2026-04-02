import { ElMessage } from 'element-plus'
import { useClipboard } from '@vueuse/core'
import { useI18n } from 'vue-i18n'

type ClipboardCopyOptions = {
  successMessage?: string
  failureMessage?: string
}

type ClipboardCopyDeps = {
  writeText: (text: string) => Promise<void>
  successMessage: string
  failureMessage: string
  notifySuccess: (message: string) => void
  notifyFailure: (message: string) => void
}

export function createClipboardCopy(deps: ClipboardCopyDeps) {
  const copyText = async (text: string, options: ClipboardCopyOptions = {}) => {
    try {
      await deps.writeText(text)
      deps.notifySuccess(options.successMessage ?? deps.successMessage)
    } catch {
      deps.notifyFailure(options.failureMessage ?? deps.failureMessage)
    }
  }

  return { copyText }
}

export function useClipboardCopy() {
  const { copy } = useClipboard()
  const { t } = useI18n()

  return createClipboardCopy({
    writeText: (text) => copy(text),
    successMessage: t('common.copied'),
    failureMessage: t('common.copyFailed'),
    notifySuccess: (message) => ElMessage.success(message),
    notifyFailure: (message) => ElMessage.warning(message)
  })
}
