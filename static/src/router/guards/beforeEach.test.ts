import assert from 'node:assert/strict'
import test, { afterEach } from 'node:test'
import { readFileSync } from 'node:fs'
import * as charactersGate from './charactersGate'

const {
  applyCharactersGateTransition,
  getCharactersGateReasons,
  refreshCharactersGateState,
  resetCharactersGateStateRefreshForTest,
  resolveCharactersGateTransition,
  shouldRedirectToCharactersPage
} = charactersGate

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

test('invalid primary character still redirects when enforcement is disabled', () => {
  assert.equal(
    shouldRedirectToCharactersPage(
      { isLogin: true, path: '/system/user' },
      {
        enforceCharacterESIRestriction: false,
        profileComplete: true,
        primaryCharacterId: 9001,
        characters: [
          { character_id: 9001, token_invalid: true } as Api.Auth.EveCharacter,
          { character_id: 9002, token_invalid: false } as Api.Auth.EveCharacter
        ]
      }
    ),
    true
  )
})

test('characters gate reports every active lock reason in priority order', () => {
  assert.equal(typeof getCharactersGateReasons, 'function')
  assert.deepEqual(
    getCharactersGateReasons(
      { isLogin: true, path: '/system/user' },
      {
        profileComplete: false,
        enforceCharacterESIRestriction: true,
        primaryCharacterId: 9001,
        characters: [
          { character_id: 9001, token_invalid: true } as Api.Auth.EveCharacter,
          { character_id: 9002, token_invalid: true } as Api.Auth.EveCharacter
        ]
      }
    ),
    ['profile_incomplete', 'primary_character_token_invalid', 'character_token_invalid']
  )
})

test('primary-character lock reason is not duplicated as all-character enforcement', () => {
  assert.equal(typeof getCharactersGateReasons, 'function')
  assert.deepEqual(
    getCharactersGateReasons(
      { isLogin: true, path: '/system/user' },
      {
        profileComplete: true,
        enforceCharacterESIRestriction: true,
        primaryCharacterId: 9001,
        characters: [{ character_id: 9001, token_invalid: true } as Api.Auth.EveCharacter]
      }
    ),
    ['primary_character_token_invalid']
  )
})

test('characters gate transition warns when locked navigation lands on characters page', () => {
  const transition = resolveCharactersGateTransition(
    { isLogin: true, path: '/system/user' },
    {
      profileComplete: false,
      enforceCharacterESIRestriction: true,
      primaryCharacterId: 9001,
      characters: [{ character_id: 9001, token_invalid: false } as Api.Auth.EveCharacter]
    }
  )

  let redirectCalls = 0
  const warningCalls: string[][] = []

  assert.equal(
    applyCharactersGateTransition(
      transition,
      () => {
        redirectCalls += 1
      },
      (reasons) => {
        warningCalls.push([...reasons])
      }
    ),
    true
  )
  assert.equal(redirectCalls, 1)
  assert.deepEqual(warningCalls, [['profile_incomplete']])
})

test('characters gate transition suppresses warning when permission fallback avoids characters page', () => {
  const transition = resolveCharactersGateTransition(
    { isLogin: true, path: '/system/user' },
    {
      profileComplete: true,
      enforceCharacterESIRestriction: true,
      primaryCharacterId: 9001,
      characters: [{ character_id: 9001, token_invalid: true } as Api.Auth.EveCharacter]
    },
    '/dashboard'
  )

  let redirectCalls = 0
  let warningCalls = 0

  assert.equal(
    applyCharactersGateTransition(
      transition,
      () => {
        redirectCalls += 1
      },
      () => {
        warningCalls += 1
      }
    ),
    true
  )
  assert.equal(redirectCalls, 1)
  assert.equal(warningCalls, 0)
})

test('characters gate transition does nothing when the user is not locked out', () => {
  let redirectCalls = 0
  let warningCalls = 0

  assert.equal(
    applyCharactersGateTransition(
      resolveCharactersGateTransition(
        { isLogin: true, path: '/system/user' },
        {
          profileComplete: true,
          enforceCharacterESIRestriction: true,
          primaryCharacterId: 9001,
          characters: [{ character_id: 9001, token_invalid: false } as Api.Auth.EveCharacter]
        }
      ),
      () => {
        redirectCalls += 1
      },
      () => {
        warningCalls += 1
      }
    ),
    false
  )
  assert.equal(redirectCalls, 0)
  assert.equal(warningCalls, 0)
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
