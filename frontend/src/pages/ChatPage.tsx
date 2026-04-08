import { useCallback, useEffect, useRef, useState } from 'react'
import { listConversations, createConversation, getConversation, deleteConversation } from '../api/chat'
import { createNote } from '../api/notes'
import { useChat } from '../hooks/useSSE'
import type { Conversation, Message, ZettelSuggestion } from '../types'
import { cn } from '../lib/utils'
import { Plus, Trash2, Send, Square, Sparkles, BookOpen, Link2, Loader2 } from 'lucide-react'

export function ChatPage() {
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [activeId, setActiveId] = useState<number | null>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [creating, setCreating] = useState(false)
  const [loadingConvo, setLoadingConvo] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const { streaming, streamedText, suggestions, connectionSuggestions, error, send, cancel } = useChat()

  // Load conversation list
  useEffect(() => {
    listConversations().then(setConversations).catch(() => {})
  }, [])

  // Auto-scroll to bottom on new messages or streaming
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, streamedText])

  const loadConversation = useCallback(async (id: number) => {
    setActiveId(id)
    setLoadingConvo(true)
    try {
      const data = await getConversation(id)
      setMessages(data.messages)
    } catch {
      setMessages([])
    }
    setLoadingConvo(false)
  }, [])

  const handleNewConversation = async () => {
    setCreating(true)
    try {
      const convo = await createConversation('New conversation', '')
      setConversations((prev) => [convo, ...prev])
      setActiveId(convo.id)
      setMessages([])
    } catch {
      // ignore
    }
    setCreating(false)
  }

  const handleDeleteConversation = async (id: number, e: React.MouseEvent) => {
    e.stopPropagation()
    await deleteConversation(id).catch(() => {})
    setConversations((prev) => prev.filter((c) => c.id !== id))
    if (activeId === id) {
      setActiveId(null)
      setMessages([])
    }
  }

  const handleSend = () => {
    if (!activeId || !input.trim() || streaming) return
    const content = input.trim()
    setInput('')

    // Add user message to display immediately
    const userMsg: Message = {
      id: Date.now(),
      conversation_id: activeId,
      role: 'user',
      content,
      input_tokens: 0,
      output_tokens: 0,
      model: '',
      created_at: new Date().toISOString(),
    }
    setMessages((prev) => [...prev, userMsg])

    send(activeId, content, (fullText) => {
      const assistantMsg: Message = {
        id: Date.now() + 1,
        conversation_id: activeId,
        role: 'assistant',
        content: fullText,
        input_tokens: 0,
        output_tokens: 0,
        model: '',
        created_at: new Date().toISOString(),
      }
      setMessages((prev) => [...prev, assistantMsg])
    })
  }

  const handleSaveNote = async (suggestion: ZettelSuggestion) => {
    try {
      await createNote({
        title: suggestion.title,
        note_type: suggestion.type,
        summary: suggestion.summary,
        laymans_terms: suggestion.laymans_terms,
        analogy: suggestion.analogy,
        core_idea: suggestion.core_idea,
        body: suggestion.body,
        components: suggestion.components,
        why_it_matters: suggestion.why_it_matters,
        examples: suggestion.examples,
        templates: suggestion.templates,
        source_chat_id: activeId || undefined,
      })
    } catch {
      // ignore
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex h-full gap-0 -m-6">
      {/* Conversation sidebar */}
      <div className="flex w-64 flex-col border-r border-zinc-800 bg-zinc-950">
        <div className="border-b border-zinc-800 p-3">
          <button
            onClick={handleNewConversation}
            disabled={creating}
            className="flex w-full items-center justify-center gap-2 rounded-md bg-zinc-800 px-3 py-2 text-sm font-medium transition-colors hover:bg-zinc-700 disabled:opacity-50"
          >
            <Plus className="h-4 w-4" />
            New chat
          </button>
        </div>
        <div className="flex-1 overflow-auto">
          {conversations.map((convo) => (
            <button
              key={convo.id}
              onClick={() => loadConversation(convo.id)}
              className={cn(
                'group flex w-full items-center justify-between px-3 py-2.5 text-left text-sm transition-colors',
                activeId === convo.id
                  ? 'bg-zinc-800 text-white'
                  : 'text-zinc-400 hover:bg-zinc-900 hover:text-white'
              )}
            >
              <span className="min-w-0 truncate">{convo.title || 'Untitled'}</span>
              <button
                onClick={(e) => handleDeleteConversation(convo.id, e)}
                className="ml-2 shrink-0 opacity-0 transition-opacity group-hover:opacity-100"
              >
                <Trash2 className="h-3.5 w-3.5 text-zinc-500 hover:text-red-400" />
              </button>
            </button>
          ))}
          {conversations.length === 0 && (
            <p className="p-4 text-center text-xs text-zinc-600">No conversations yet</p>
          )}
        </div>
      </div>

      {/* Main chat area */}
      <div className="flex flex-1 flex-col">
        {!activeId ? (
          <div className="flex flex-1 items-center justify-center text-zinc-500">
            Select a conversation or start a new one
          </div>
        ) : (
          <>
            {/* Messages */}
            <div className="flex-1 overflow-auto p-6 space-y-4">
              {loadingConvo && (
                <div className="flex justify-center py-8">
                  <Loader2 className="h-5 w-5 animate-spin text-zinc-500" />
                </div>
              )}

              {messages.map((msg) => (
                <div
                  key={msg.id}
                  className={cn(
                    'max-w-2xl rounded-lg px-4 py-3 text-sm',
                    msg.role === 'user'
                      ? 'ml-auto bg-zinc-800 text-zinc-100'
                      : 'bg-zinc-900 text-zinc-300'
                  )}
                >
                  <div className="whitespace-pre-wrap">{msg.content}</div>
                </div>
              ))}

              {/* Streaming response */}
              {streaming && streamedText && (
                <div className="max-w-2xl rounded-lg bg-zinc-900 px-4 py-3 text-sm text-zinc-300">
                  <div className="whitespace-pre-wrap">{streamedText}</div>
                  <Loader2 className="mt-2 h-3 w-3 animate-spin text-zinc-500" />
                </div>
              )}

              {/* Zettel suggestions */}
              {suggestions.length > 0 && (
                <div className="space-y-3">
                  <div className="flex items-center gap-2 text-xs font-medium text-zinc-400">
                    <Sparkles className="h-3.5 w-3.5" />
                    Suggested notes
                  </div>
                  {suggestions.map((s, i) => (
                    <div
                      key={i}
                      className="rounded-lg border border-zinc-700 bg-zinc-900 p-4"
                    >
                      <div className="mb-1 flex items-center justify-between">
                        <span className="font-medium">{s.title}</span>
                        <span className="rounded bg-zinc-800 px-2 py-0.5 text-xs text-zinc-400">
                          {s.type}
                        </span>
                      </div>
                      {s.summary && (
                        <p className="mb-3 text-sm text-zinc-400">{s.summary}</p>
                      )}
                      <button
                        onClick={() => handleSaveNote(s)}
                        className="flex items-center gap-1.5 rounded bg-emerald-800/50 px-3 py-1.5 text-xs font-medium text-emerald-300 transition-colors hover:bg-emerald-800"
                      >
                        <BookOpen className="h-3 w-3" />
                        Save to Inbox
                      </button>
                    </div>
                  ))}
                </div>
              )}

              {/* Connection suggestions */}
              {connectionSuggestions.length > 0 && (
                <div className="space-y-3">
                  <div className="flex items-center gap-2 text-xs font-medium text-zinc-400">
                    <Link2 className="h-3.5 w-3.5" />
                    Suggested connections
                  </div>
                  {connectionSuggestions.map((cs, i) => (
                    <div
                      key={i}
                      className="rounded-lg border border-zinc-700 bg-zinc-900 p-4"
                    >
                      <div className="mb-1 flex items-center gap-2 text-sm">
                        <span className="font-medium">{cs.source_title}</span>
                        <span className="rounded bg-zinc-800 px-1.5 py-0.5 text-xs text-zinc-500">{cs.label}</span>
                        <span className="font-medium">{cs.target_title}</span>
                      </div>
                      {cs.reason && (
                        <p className="text-sm text-zinc-400">{cs.reason}</p>
                      )}
                    </div>
                  ))}
                </div>
              )}

              {error && (
                <div className="rounded-lg bg-red-900/30 px-4 py-3 text-sm text-red-300">
                  {error}
                </div>
              )}

              <div ref={messagesEndRef} />
            </div>

            {/* Input */}
            <div className="border-t border-zinc-800 p-4">
              <div className="mx-auto flex max-w-2xl items-end gap-2">
                <textarea
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="Ask about a topic you're learning..."
                  rows={1}
                  className="flex-1 resize-none rounded-lg border border-zinc-700 bg-zinc-900 px-4 py-2.5 text-sm text-zinc-100 placeholder-zinc-500 focus:border-zinc-500 focus:outline-none"
                />
                {streaming ? (
                  <button
                    onClick={cancel}
                    className="rounded-lg bg-red-800/50 p-2.5 text-red-300 transition-colors hover:bg-red-800"
                  >
                    <Square className="h-4 w-4" />
                  </button>
                ) : (
                  <button
                    onClick={handleSend}
                    disabled={!input.trim()}
                    className="rounded-lg bg-white p-2.5 text-zinc-900 transition-colors hover:bg-zinc-200 disabled:opacity-30"
                  >
                    <Send className="h-4 w-4" />
                  </button>
                )}
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
