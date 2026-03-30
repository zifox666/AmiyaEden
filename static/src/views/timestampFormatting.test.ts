import assert from 'node:assert/strict'
import test from 'node:test'
import { readdirSync, readFileSync, statSync } from 'node:fs'
import { join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'

const TIMESTAMP_FIELDS = [
  'last_run',
  'next_run',
  'created_at',
  'updated_at',
  'last_login_at',
  'expires_at',
  'start_at',
  'end_at',
  'issued_at',
  'reviewed_at',
  'processed_at',
  'journal_at',
  'date_expired',
  'date_bid',
  'joined_at',
  'calculated_at',
  'started_at',
  'ended_at',
  'last_processed_at',
  'last_online_at',
  'killmail_time'
].join('|')

const UI_ROOT = fileURLToPath(new URL('..', import.meta.url))
const VUE_DIRECTORIES = [join(UI_ROOT, 'views'), join(UI_ROOT, 'components')]

const rawTimestampHFormatterPattern = new RegExp(
  String.raw`formatter:\s*\([^)]*\)\s*=>\s*h\('span',\s*\{\},\s*\w+\.(?:${TIMESTAMP_FIELDS})(?:\s*(?:\?\?|\|\|)\s*'-')?\s*\)`,
  'g'
)

const rawTimestampDirectFormatterPattern = new RegExp(
  String.raw`formatter:\s*\([^)]*\)\s*=>\s*\w+\.(?:${TIMESTAMP_FIELDS})(?:\s*(?:\?\?|\|\|)\s*'-')?(?!\s*\?)`,
  'g'
)

const rawTimestampInterpolationPattern = new RegExp(
  String.raw`\{\{\s*\w+\.(?:${TIMESTAMP_FIELDS})(?:\s*(?:\?\?|\|\|)\s*'-')?\s*\}\}`,
  'g'
)

function collectVueFiles(dir: string): string[] {
  return readdirSync(dir).flatMap((entry) => {
    const fullPath = join(dir, entry)

    if (statSync(fullPath).isDirectory()) {
      return collectVueFiles(fullPath)
    }

    return fullPath.endsWith('.vue') ? [fullPath] : []
  })
}

function compactSnippet(snippet: string): string {
  return snippet.replace(/\s+/g, ' ').trim()
}

function findViolations(filePath: string): string[] {
  const source = readFileSync(filePath, 'utf8')
  const violations: string[] = []

  for (const match of source.matchAll(rawTimestampHFormatterPattern)) {
    const snippet = match[0]

    violations.push(
      `${relative(UI_ROOT, filePath)} uses a raw timestamp formatter: ${compactSnippet(snippet)}`
    )
  }

  for (const match of source.matchAll(rawTimestampDirectFormatterPattern)) {
    const snippet = match[0]

    violations.push(
      `${relative(UI_ROOT, filePath)} uses a raw timestamp formatter: ${compactSnippet(snippet)}`
    )
  }

  for (const match of source.matchAll(rawTimestampInterpolationPattern)) {
    const snippet = match[0]

    violations.push(
      `${relative(UI_ROOT, filePath)} interpolates a raw timestamp: ${compactSnippet(snippet)}`
    )
  }

  return violations
}

test('timestamp-like UI fields use shared time formatters', () => {
  const violations = VUE_DIRECTORIES.flatMap((dir) => collectVueFiles(dir).flatMap(findViolations))

  assert.deepEqual(violations, [])
})
