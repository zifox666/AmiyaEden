export interface UserManageDialogFormValue {
  nickname: string
  qq: string
  discordId: string
}

export type UserManageDialogValidationError =
  | 'nicknameRequired'
  | 'nicknameLength'
  | 'qqLength'
  | 'qqDigits'
  | 'discordLength'

const MAX_NICKNAME_LENGTH = 20
const MAX_CONTACT_LENGTH = 20

const getTextLength = (value: string) => Array.from(value.trim()).length

export function validateNicknameInput(value: string): UserManageDialogValidationError | null {
  const nickname = value.trim()
  if (!nickname) {
    return 'nicknameRequired'
  }
  if (getTextLength(nickname) > MAX_NICKNAME_LENGTH) {
    return 'nicknameLength'
  }
  return null
}

export function validateQQInput(value: string): UserManageDialogValidationError | null {
  const qq = value.trim()
  if (getTextLength(qq) > MAX_CONTACT_LENGTH) {
    return 'qqLength'
  }
  if (qq && !/^\d+$/.test(qq)) {
    return 'qqDigits'
  }
  return null
}

export function validateDiscordIdInput(value: string): UserManageDialogValidationError | null {
  const discordId = value.trim()
  if (getTextLength(discordId) > MAX_CONTACT_LENGTH) {
    return 'discordLength'
  }
  return null
}

export function buildUserManageUpdatePayload(
  form: UserManageDialogFormValue,
  canEditContacts: boolean
): { nickname: string; qq?: string; discord_id?: string } {
  const payload = {
    nickname: form.nickname.trim()
  }

  if (!canEditContacts) {
    return payload
  }

  return {
    ...payload,
    qq: form.qq.trim(),
    discord_id: form.discordId.trim()
  }
}
