import assert from 'node:assert/strict'
import test, { afterEach } from 'node:test'
import { readFileSync } from 'node:fs'
import {
  refreshCharactersGateState,
  resetCharactersGateStateRefreshForTest,
  shouldRedirectToCharactersPage
} from './charactersGate'

const source = readFileSync(new URL('./beforeEach.ts', import.meta.url), 'utf8')

afterEach(() => {
  resetCharactersGateStateRefreshForTest()
})

test('dynamic route initialization fetches badge counts after storing the menu list', () => {
  assert.match(source, /menuStore\.setMenuList\(menuList\)[\s\S]*loadBadgeCounts\(badgeStore\)/)
})

test('invalid non-primary character still requires redirect to characters page', () => {
  assert.equal(
    shouldRedirectToCharactersPage(
      { isLogin: true, path: '/system/user' },
      {
        enforceCharacterESIRestriction: true,
        profileComplete: true,
        characters: [
          { token_invalid: false } as Api.Auth.EveCharacter,
          { token_invalid: true } as Api.Auth.EveCharacter
        ]
      }
    ),
    true
  )
})

test('characters page is allowed even when there are invalid tokens', () => {
  assert.equal(
    shouldRedirectToCharactersPage(
      { isLogin: true, path: '/dashboard/characters' },
      {
        enforceCharacterESIRestriction: true,
        profileComplete: true,
        characters: [{ token_invalid: true } as Api.Auth.EveCharacter]
      }
    ),
    false
  )
})

test('invalid non-primary character does not redirect when enforcement is disabled', () => {
  assert.equal(
    shouldRedirectToCharactersPage(
      { isLogin: true, path: '/system/user' },
      {
        enforceCharacterESIRestriction: false,
        profileComplete: true,
        characters: [{ token_invalid: true } as Api.Auth.EveCharacter]
      }
    ),
    false
  )
})

test('guard refresh stores authoritative user info before redirect evaluation', async () => {
  const refreshedUser = {
    userId: 7,
    roles: ['user'],
    userName: 'Amiya',
    avatar: '',
    nickname: 'Amiya',
    qq: '12345',
    discordId: '',
    profileComplete: true,
    enforceCharacterESIRestriction: true,
    characters: [{ token_invalid: true } as Api.Auth.EveCharacter]
  } as Api.Auth.UserInfo

  let storedUser: Api.Auth.UserInfo | undefined

  const result = await refreshCharactersGateState(
    async () => refreshedUser,
    (value) => {
      storedUser = value
    }
  )

  assert.equal(storedUser?.characters?.[0]?.token_invalid, true)
  assert.equal(result.characters?.[0]?.token_invalid, true)
})

test('concurrent gate refreshes share one current-user request', async () => {
  let calls = 0
  const refreshedUser = {
    userId: 7,
    roles: ['user'],
    userName: 'Amiya',
    avatar: '',
    nickname: 'Amiya',
    qq: '12345',
    discordId: '',
    profileComplete: true,
    enforceCharacterESIRestriction: true,
    characters: []
  } as Api.Auth.UserInfo

  const fetchUserInfo = async () => {
    calls += 1
    return refreshedUser
  }

  await Promise.all([
    refreshCharactersGateState(fetchUserInfo, () => {}),
    refreshCharactersGateState(fetchUserInfo, () => {})
  ])

  assert.equal(calls, 1)
})
