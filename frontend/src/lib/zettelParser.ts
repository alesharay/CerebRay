import type { ZettelSuggestion, ConnectionSuggestion } from '../types'

// Strip suggestion blocks from text for clean chat display
export function stripSuggestionBlocks(text: string): string {
  return text
    .replace(/---ZETTEL_SUGGESTION---[\s\S]*?---END_ZETTEL---/g, '')
    .replace(/---CONNECTION_SUGGESTION---[\s\S]*?---END_CONNECTION---/g, '')
    .trim()
}

export function parseZettelSuggestions(text: string): ZettelSuggestion[] {
  const suggestions: ZettelSuggestion[] = []
  const regex = /---ZETTEL_SUGGESTION---([\s\S]*?)---END_ZETTEL---/g
  let match

  while ((match = regex.exec(text)) !== null) {
    const fields = parseBlock(match[1])

    if (fields.title) {
      suggestions.push({
        title: fields.title || '',
        type: (fields.type as ZettelSuggestion['type']) || 'concept',
        summary: fields.summary || '',
        laymans_terms: fields.laymans_terms || '',
        analogy: fields.analogy || '',
        core_idea: fields.core_idea || '',
        body: fields.body || '',
        components: fields.components || '',
        why_it_matters: fields.why_it_matters || '',
        examples: fields.examples || '',
        templates: fields.templates || '',
        tags: fields.tags ? fields.tags.split(',').map(t => t.trim()) : [],
      })
    }
  }

  return suggestions
}

export function parseConnectionSuggestions(text: string): ConnectionSuggestion[] {
  const suggestions: ConnectionSuggestion[] = []
  const regex = /---CONNECTION_SUGGESTION---([\s\S]*?)---END_CONNECTION---/g
  let match

  while ((match = regex.exec(text)) !== null) {
    const fields = parseBlock(match[1])

    if (fields.source_title && fields.target_title) {
      suggestions.push({
        source_title: fields.source_title,
        target_title: fields.target_title,
        label: fields.label || '',
        reason: fields.reason || '',
      })
    }
  }

  return suggestions
}

const knownFields = new Set([
  'title', 'type', 'summary', 'laymans_terms', 'analogy', 'core_idea',
  'body', 'components', 'why_it_matters', 'examples', 'templates', 'tags',
  'source_title', 'target_title', 'label', 'reason',
])

// Clean up field values: remove leading pipe chars, normalize indentation
function cleanFieldValue(value: string): string {
  let cleaned = value
  // Remove leading pipe character (YAML block indicator the AI sometimes adds)
  if (cleaned.startsWith('|')) {
    cleaned = cleaned.slice(1)
  }
  // Find the common leading whitespace across non-empty lines and strip it
  const lines = cleaned.split('\n')
  const nonEmptyLines = lines.filter(l => l.trim().length > 0)
  if (nonEmptyLines.length > 0) {
    const minIndent = Math.min(
      ...nonEmptyLines.map(l => l.match(/^(\s*)/)?.[1].length ?? 0)
    )
    if (minIndent > 0) {
      cleaned = lines.map(l => l.slice(minIndent)).join('\n')
    }
  }
  return cleaned.trim()
}

function parseBlock(block: string): Record<string, string> {
  const fields: Record<string, string> = {}
  let currentKey = ''
  let currentValue = ''

  for (const line of block.split('\n')) {
    const colonIdx = line.indexOf(':')
    if (colonIdx > 0) {
      const potentialKey = line.slice(0, colonIdx).trim()
      if (knownFields.has(potentialKey)) {
        // Save previous field
        if (currentKey) {
          fields[currentKey] = cleanFieldValue(currentValue)
        }
        currentKey = potentialKey
        currentValue = line.slice(colonIdx + 1).trim()
        continue
      }
    }
    // Continuation of current multi-line field
    if (currentKey) {
      currentValue += '\n' + line
    }
  }

  // Save the last field
  if (currentKey) {
    fields[currentKey] = cleanFieldValue(currentValue)
  }

  return fields
}
