import type { ZettelSuggestion, ConnectionSuggestion } from '../types'

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

function parseBlock(block: string): Record<string, string> {
  const fields: Record<string, string> = {}
  for (const line of block.split('\n')) {
    const colonIdx = line.indexOf(':')
    if (colonIdx === -1) continue
    const key = line.slice(0, colonIdx).trim()
    const value = line.slice(colonIdx + 1).trim()
    if (key && value) {
      fields[key] = value
    }
  }
  return fields
}
