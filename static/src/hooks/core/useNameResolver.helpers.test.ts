import assert from 'node:assert/strict'
import test from 'node:test'
import {
  filterUnresolvedESIIDs,
  filterUnresolvedNameIDs,
  getResolvedName,
  mergeResolvedNames
} from './useNameResolver.helpers'

test('filterUnresolvedNameIDs deduplicates per namespace', () => {
  const result = filterUnresolvedNameIDs(
    {
      type: [1, 1, 2],
      solar_system: [1, 3]
    },
    {
      type: { 1: 'Rifter' },
      solar_system: {}
    }
  )

  assert.deepEqual(result, {
    type: [2],
    solar_system: [1, 3]
  })
})

test('filterUnresolvedESIIDs skips cached esi ids only', () => {
  const result = filterUnresolvedESIIDs([7, 7, 8, 0, -1], {
    esi: { 7: 'Pilot Seven' }
  })

  assert.deepEqual(result, [8])
})

test('mergeResolvedNames preserves namespace data and flat compatibility map', () => {
  const namesByNamespace: Record<string, Record<number, string>> = {}
  const flatMap: Record<number, string> = {}

  mergeResolvedNames(namesByNamespace, flatMap, {
    names: {
      type: { 1: 'Rifter' },
      solar_system: { 1: 'Jita' },
      esi: { 99: 'Capsuleer' }
    },
    flat: {
      1: 'Jita',
      99: 'Capsuleer'
    }
  })

  assert.equal(namesByNamespace.type[1], 'Rifter')
  assert.equal(namesByNamespace.solar_system[1], 'Jita')
  assert.equal(namesByNamespace.esi[99], 'Capsuleer')
  assert.equal(flatMap[1], 'Jita')
  assert.equal(flatMap[99], 'Capsuleer')
})

test('getResolvedName prefers namespace-specific result before flat fallback', () => {
  const namesByNamespace = {
    type: { 1: 'Rifter' },
    solar_system: { 1: 'Jita' }
  }
  const flatMap = { 1: 'Jita' }

  assert.equal(getResolvedName(1, '-', 'type', namesByNamespace, flatMap), 'Rifter')
  assert.equal(getResolvedName(1, '-', 'solar_system', namesByNamespace, flatMap), 'Jita')
  assert.equal(getResolvedName(2, 'fallback', 'type', namesByNamespace, flatMap), 'fallback')
})
