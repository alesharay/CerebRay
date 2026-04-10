import { useCallback, useRef, useState } from 'react'
import { sendMessage, type StreamCallbacks } from '../api/chat'
import { parseZettelSuggestions, parseConnectionSuggestions } from '../lib/zettelParser'
import type { ZettelSuggestion, ConnectionSuggestion } from '../types'

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
