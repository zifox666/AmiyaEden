const EVE_IMAGE_BASE_URL = 'https://images.evetech.net'

export function buildEveCharacterPortraitUrl(characterId: number, size = 128) {
  return characterId > 0
    ? `${EVE_IMAGE_BASE_URL}/characters/${characterId}/portrait?size=${size}`
    : ''
}
