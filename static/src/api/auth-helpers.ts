export function isUserProfileComplete(
  user?: Partial<Pick<Api.Auth.UserInfo, 'nickname' | 'qq' | 'discordId' | 'profileComplete'>>
): boolean {
  if (!user) return false
  if (typeof user.profileComplete === 'boolean') {
    return user.profileComplete
  }

  const nickname = user.nickname?.trim() ?? ''
  const qq = user.qq?.trim() ?? ''
  const discordId = user.discordId?.trim() ?? ''
  return nickname.length > 0 && (qq.length > 0 || discordId.length > 0)
}

export function hasInvalidCharacterToken(
  characters?: Array<Pick<Api.Auth.EveCharacter, 'token_invalid'>>
): boolean {
  return characters?.some((character) => character.token_invalid) ?? false
}
