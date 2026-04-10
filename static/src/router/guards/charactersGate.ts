import {
  hasInvalidCharacterToken,
  hasInvalidPrimaryCharacterToken,
  isUserProfileComplete
} from '@/api/auth-helpers'

export const PROFILE_SETUP_PATH = '/dashboard/characters'

export type CharactersGateReason =
  | 'profile_incomplete'
  | 'primary_character_token_invalid'
  | 'character_token_invalid'

export type CharactersGateTransition = {
  reasons: CharactersGateReason[]
  redirectPath: string | null
  shouldWarn: boolean
}

type CharactersGateContext = {
  isLogin: boolean
  path: string
}

type CharactersGateUserInfo = Partial<
  Pick<
    Api.Auth.UserInfo,
    | 'profileComplete'
    | 'nickname'
    | 'qq'
    | 'discordId'
    | 'characters'
    | 'primaryCharacterId'
    | 'enforceCharacterESIRestriction'
  >
>

let inFlightCharactersGateRefresh: Promise<Api.Auth.UserInfo> | null = null

function hasInvalidSecondaryCharacterToken(
  primaryCharacterId?: number,
  characters?: Array<Pick<Api.Auth.EveCharacter, 'character_id' | 'token_invalid'>>
): boolean {
  if (!primaryCharacterId) {
    return hasInvalidCharacterToken(characters)
  }

  return (
    characters?.some(
      (character) =>
        character.character_id !== primaryCharacterId && character.token_invalid === true
    ) ?? false
  )
}

export function getCharactersGateReasons(
  context: CharactersGateContext,
  userInfo?: CharactersGateUserInfo
): CharactersGateReason[] {
  if (!context.isLogin || context.path === PROFILE_SETUP_PATH) {
    return []
  }

  const reasons: CharactersGateReason[] = []

  if (!isUserProfileComplete(userInfo)) {
    reasons.push('profile_incomplete')
  }

  if (hasInvalidPrimaryCharacterToken(userInfo?.primaryCharacterId, userInfo?.characters)) {
    reasons.push('primary_character_token_invalid')
  }

  if (
    userInfo?.enforceCharacterESIRestriction !== false &&
    hasInvalidSecondaryCharacterToken(userInfo?.primaryCharacterId, userInfo?.characters)
  ) {
    reasons.push('character_token_invalid')
  }

  return reasons
}

export function resolveCharactersGateTransition(
  context: CharactersGateContext,
  userInfo?: CharactersGateUserInfo,
  destinationPath: string = PROFILE_SETUP_PATH
): CharactersGateTransition {
  const reasons = getCharactersGateReasons(context, userInfo)

  return {
    reasons,
    redirectPath: reasons.length > 0 ? PROFILE_SETUP_PATH : null,
    shouldWarn: reasons.length > 0 && destinationPath === PROFILE_SETUP_PATH
  }
}

export function applyCharactersGateTransition(
  transition: CharactersGateTransition,
  redirect: () => void,
  notify: (reasons: CharactersGateReason[]) => void
): boolean {
  if (!transition.redirectPath) {
    return false
  }

  redirect()

  if (transition.shouldWarn) {
    notify(transition.reasons)
  }

  return true
}

export function shouldRedirectToCharactersPage(
  context: CharactersGateContext,
  userInfo?: CharactersGateUserInfo
): boolean {
  return resolveCharactersGateTransition(context, userInfo).redirectPath !== null
}

export async function refreshCharactersGateState(
  fetchUserInfo: () => Promise<Api.Auth.UserInfo>,
  storeUserInfo: (userInfo: Api.Auth.UserInfo) => void
): Promise<Api.Auth.UserInfo> {
  if (!inFlightCharactersGateRefresh) {
    inFlightCharactersGateRefresh = fetchUserInfo().finally(() => {
      inFlightCharactersGateRefresh = null
    })
  }

  const userInfo = await inFlightCharactersGateRefresh
  storeUserInfo(userInfo)
  return userInfo
}

export function resetCharactersGateStateRefreshForTest(): void {
  inFlightCharactersGateRefresh = null
}
