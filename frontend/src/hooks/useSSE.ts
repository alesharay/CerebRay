import { useCallback, useRef, useState } from 'react'
import { sendMessage, type StreamCallbacks } from '../api/chat'
import type { ZettelSuggestion, ConnectionSuggestion } from '../types'

function parseZettelSuggestions(text: string): ZettelSuggestion[] {
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

function parseConnectionSuggestions(text: string): ConnectionSuggestion[] {
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

export function useChat() {
  const [streaming, setStreaming] = useState(false)
  const [streamedText, setStreamedText] = useState('')
  const [suggestions, setSuggestions] = useState<ZettelSuggestion[]>([])
  const [connectionSuggestions, setConnectionSuggestions] = useState<ConnectionSuggestion[]>([])
  const [error, setError] = useState<string | null>(null)
  const cancelRef = useRef<(() => void) | null>(null)

  const send = useCallback((conversationId: number, content: string, onComplete?: (text: string) => void) => {
    setStreaming(true)
    setStreamedText('')
    setSuggestions([])
    setConnectionSuggestions([])
    setError(null)

    let accumulated = ''

    const callbacks: StreamCallbacks = {
      onDelta: (text) => {
        accumulated += text
        setStreamedText(accumulated)
      },
      onDone: () => {
        setStreaming(false)
        setSuggestions(parseZettelSuggestions(accumulated))
        setConnectionSuggestions(parseConnectionSuggestions(accumulated))
        onComplete?.(accumulated)
      },
      onError: (err) => {
        setStreaming(false)
        setError(err)
      },
    }

    cancelRef.current = sendMessage(conversationId, content, callbacks)
  }, [])

  const cancel = useCallback(() => {
    cancelRef.current?.()
    setStreaming(false)
  }, [])

  return { streaming, streamedText, suggestions, connectionSuggestions, error, send, cancel }
}
