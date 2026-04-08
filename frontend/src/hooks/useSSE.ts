import { useCallback, useRef, useState } from 'react'
import { sendMessage, type StreamCallbacks } from '../api/chat'
import type { ZettelSuggestion } from '../types'

// Parse Zettel suggestions from AI response text
function parseZettelSuggestions(text: string): ZettelSuggestion[] {
  const suggestions: ZettelSuggestion[] = []
  const regex = /---ZETTEL_SUGGESTION---([\s\S]*?)---END_ZETTEL---/g
  let match

  while ((match = regex.exec(text)) !== null) {
    const block = match[1]
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

export function useChat() {
  const [streaming, setStreaming] = useState(false)
  const [streamedText, setStreamedText] = useState('')
  const [suggestions, setSuggestions] = useState<ZettelSuggestion[]>([])
  const [error, setError] = useState<string | null>(null)
  const cancelRef = useRef<(() => void) | null>(null)

  const send = useCallback((conversationId: number, content: string) => {
    setStreaming(true)
    setStreamedText('')
    setSuggestions([])
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

  return { streaming, streamedText, suggestions, error, send, cancel }
}
