import { ElMessage } from 'element-plus'

type Translator = (key: string, params?: Record<string, unknown>) => string

function formatCharacter(name?: string, id?: number) {
  const cleanName = name?.trim()
  if (cleanName && id) return `${cleanName} (${id})`
  if (cleanName) return cleanName
  if (id) return String(id)
  return ''
}

export function formatMailAttemptSuccess(
  result: Api.Common.MailActionResult,
  t: Translator
): string {
  const parts: string[] = []
  const sender = formatCharacter(result.mail_sender_character_name, result.mail_sender_character_id)
  const recipient = formatCharacter(
    result.mail_recipient_character_name,
    result.mail_recipient_character_id
  )

  if (sender) parts.push(`${t('mailAttempt.sender')}: ${sender}`)
  if (recipient) parts.push(`${t('mailAttempt.recipient')}: ${recipient}`)
  if (result.mail_id) parts.push(`${t('mailAttempt.mailId')}: ${result.mail_id}`)

  if (!parts.length) return ''
  return `${t('mailAttempt.sent')} ${parts.join(', ')}`
}

export function formatMailAttemptWarning(
  result: Api.Common.MailActionResult,
  t: Translator
): string {
  const error = result.mail_error?.trim()
  if (!error) return ''

  const parts = [error]
  const sender = formatCharacter(result.mail_sender_character_name, result.mail_sender_character_id)
  const recipient = formatCharacter(
    result.mail_recipient_character_name,
    result.mail_recipient_character_id
  )

  if (sender) parts.push(`${t('mailAttempt.sender')}: ${sender}`)
  if (recipient) parts.push(`${t('mailAttempt.recipient')}: ${recipient}`)
  if (result.mail_id) parts.push(`${t('mailAttempt.mailId')}: ${result.mail_id}`)

  return `${t('mailAttempt.failed')} ${parts.join(', ')}`
}

export function showMailAttemptMessage(result: Api.Common.MailActionResult, t: Translator): void {
  const mailWarning = formatMailAttemptWarning(result, t)
  if (mailWarning) {
    ElMessage.warning(mailWarning)
    return
  }

  const mailSuccess = formatMailAttemptSuccess(result, t)
  if (mailSuccess) {
    ElMessage.info(mailSuccess)
  }
}
