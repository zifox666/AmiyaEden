import assert from 'node:assert/strict'
import test from 'node:test'

async function loadOrderIskModule() {
  return await import(new URL('./order-isk.ts', import.meta.url).href)
}

test('resolveOrderIskTotal converts ISK orders from total fuxi coin into raw isk', async () => {
  const orderIskModule = await loadOrderIskModule()

  assert.ok(orderIskModule?.resolveOrderIskTotal, 'expected resolveOrderIskTotal export')
  assert.equal(
    orderIskModule.resolveOrderIskTotal({
      product_name: 'Large ISK Bundle',
      total_price: 1_200
    }),
    1_200_000_000
  )
})

test('resolveOrderIskTotal leaves non-ISK orders empty', async () => {
  const orderIskModule = await loadOrderIskModule()

  assert.ok(orderIskModule?.resolveOrderIskTotal, 'expected resolveOrderIskTotal export')
  assert.equal(
    orderIskModule.resolveOrderIskTotal({
      product_name: 'Faction Cruiser Hull',
      total_price: 1_200
    }),
    null
  )
})

test('resolveOrderIskTotal does not match product names containing isk as substring', async () => {
  const orderIskModule = await loadOrderIskModule()

  assert.ok(orderIskModule?.resolveOrderIskTotal, 'expected resolveOrderIskTotal export')
  assert.equal(
    orderIskModule.resolveOrderIskTotal({
      product_name: 'Risk Analysis Pack',
      total_price: 500
    }),
    null
  )
})
