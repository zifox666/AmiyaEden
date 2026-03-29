const SYSTEM_DISPLAY_NAME = 'Fuxi Legion'
const SYSTEM_CORPORATION_ID = 98185110

export const SYSTEM_IDENTITY = Object.freeze({
  corporationId: SYSTEM_CORPORATION_ID,
  displayName: SYSTEM_DISPLAY_NAME,
  description: `${SYSTEM_DISPLAY_NAME} operations platform built with Vue 3, TypeScript, and Element Plus.`
})

export const SYSTEM_IDENTITY_I18N = Object.freeze({
  corporationId: SYSTEM_IDENTITY.corporationId,
  corporationName: SYSTEM_IDENTITY.displayName
})
