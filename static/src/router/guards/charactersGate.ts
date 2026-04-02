import { hasInvalidCharacterToken, isUserProfileComplete } from '@/api/auth-helpers'

export const PROFILE_SETUP_PATH = '/dashboard/characters'

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
    | 'enforceCharacterESIRestriction'
  >
>

let inFlightCharactersGateRefresh: Promise<Api.Auth.UserInfo> | null = null

export function shouldRedirectToCharactersPage(
  context: CharactersGateContext,
  userInfo?: CharactersGateUserInfo
): boolean {
  if (!context.isLogin || context.path === PROFILE_SETUP_PATH) {
    return false
  }

  return (
    !isUserProfileComplete(userInfo) ||
    (userInfo?.enforceCharacterESIRestriction !== false &&
      hasInvalidCharacterToken(userInfo?.characters))
  )
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
