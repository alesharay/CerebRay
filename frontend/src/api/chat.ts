import { api } from './client'
import type { Conversation, Message } from '../types'

export function listConversations(): Promise<Conversation[]> {
  return api.get<Conversation[]>('/conversations')
}

export function createConversation(title: string, topic: string): Promise<Conversation> {
  return api.post<Conversation>('/conversations', { title, topic })
}

export function getConversation(id: number): Promise<{ conversation: Conversation; messages: Message[] }> {
  return api.get<{ conversation: Conversation; messages: Message[] }>(`/conversations/${id}`)
}

export function deleteConversation(id: number): Promise<void> {
  return api.delete<void>(`/conversations/${id}`)
}

export interface StreamCallbacks {
  onDelta: (text: string) => void
  onDone: () => void
  onError: (error: string) => void
}

export function sendMessage(conversationId: number, content: string, callbacks: StreamCallbacks): () => void {
  const controller = new AbortController()

  fetch(`/api/v1/conversations/${conversationId}/messages`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content }),
    signal: controller.signal,
  }).then(async (res) => {
    if (!res.ok) {
      callbacks.onError(`Request failed: ${res.status}`)
      return
    }

    const reader = res.body?.getReader()
    if (!reader) {
      callbacks.onError('No response body')
      return
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        if (!line.startsWith('data: ')) continue
        const data = line.slice(6)
        try {
          const parsed = JSON.parse(data)
          if (parsed.done) {
            callbacks.onDone()
          } else if (parsed.error) {
            callbacks.onError(parsed.error)
          } else if (parsed.delta) {
            callbacks.onDelta(parsed.delta)
          }
        } catch {
          // ignore parse errors
        }
      }
    }
  }).catch((err) => {
    if (err.name !== 'AbortError') {
      callbacks.onError(err.message)
    }
  })

  return () => controller.abort()
}
