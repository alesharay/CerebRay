import { useEffect, useRef, useState } from 'react'
import { useParams, useNavigate, useSearchParams, Link } from 'react-router-dom'
import { getNote, updateNote, deleteNote, promoteNote, sleepNote, archiveNote, searchNotes } from '../api/notes'
import { listConnectionsForNote, createConnection } from '../api/connections'
import { addTagToNote } from '../api/tags'
import { getConversation } from '../api/chat'
import { useChat } from '../hooks/useSSE'
import { parseZettelSuggestions, stripSuggestionBlocks } from '../lib/zettelParser'
import type { Note, Connection, NoteType, NoteTLP, Message } from '../types'
import { cn } from '../lib/utils'
import {
  ArrowLeft, Save, Trash2, Moon, Archive, CloudFog,
  Link2, Plus, Loader2, ArrowRight, ArrowDownLeft, Search,
  Send, Sparkles, StopCircle, MessageSquare,
} from 'lucide-react'

const noteTypes: NoteType[] = ['concept', 'theory', 'insight', 'quote', 'reference', 'question', 'structure', 'guide']
const tlpOptions: NoteTLP[] = ['clear', 'green', 'amber', 'red']

const contentSections: { key: keyof Note; label: string }[] = [
  { key: 'summary', label: 'Summary' },
  { key: 'core_idea', label: 'Core idea' },
  { key: 'laymans_terms', label: "Layman's terms" },
  { key: 'analogy', label: 'Analogy' },
  { key: 'body', label: 'Body' },
  { key: 'components', label: 'Components' },
  { key: 'why_it_matters', label: 'Why it matters' },
  { key: 'examples', label: 'Examples' },
  { key: 'templates', label: 'Templates' },
  { key: 'additional', label: 'Additional' },
]

export function NoteDetailPage() {
  const { id } = useParams<{ id: string }>()
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const promoted = searchParams.get('promoted') === 'true'

  const [note, setNote] = useState<Note | null>(null)
  const [connections, setConnections] = useState<Connection[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [draft, setDraft] = useState<Partial<Note>>({})
  const [newTag, setNewTag] = useState('')

  // Connection builder state
  const [linkQuery, setLinkQuery] = useState('')
  const [linkResults, setLinkResults] = useState<Note[]>([])
  const [linkLabel, setLinkLabel] = useState('')
  const [showLinkBuilder, setShowLinkBuilder] = useState(false)

  // Chat state
  const [chatMessages, setChatMessages] = useState<Message[]>([])
  const [chatLoadId, setChatLoadId] = useState<number | null>(null)
  const [chatInput, setChatInput] = useState('')
  const expandTriggered = useRef(false)
  const chatEndRef = useRef<HTMLDivElement>(null)
  const { streaming, streamedText, error: chatError, send, cancel } = useChat()

  // Load note, connections, and chat history together
  useEffect(() => {
    if (!id) return
    expandTriggered.current = false

    const noteId = Number(id)
    getNote(noteId).then(async (n) => {
      setNote(n)
      setDraft(n)
      const conns = await listConnectionsForNote(noteId).catch(() => [] as Connection[])
      setConnections(conns)

      // Load chat if note has a linked conversation
      if (n.source_chat_id) {
        const data = await getConversation(n.source_chat_id).catch(() => null)
        setChatMessages(data?.messages || [])
      } else {
        setChatMessages([])
      }
      setChatLoadId(noteId)
    }).catch(() => {
      navigate('/codex')
    }).finally(() => setLoading(false))
  }, [id, navigate])

  // Auto-expand on promote: trigger AI research after chat history loads
  const triggerExpand = (convoId: number, title: string, body: string, noteId: number) => {
    const thought = body ? `${title}\n\n${body}` : title
    const userMsg: Message = {
      id: Date.now() - 1,
      conversation_id: convoId,
      role: 'user',
      content: `Expand this thought into a full Zettel: ${thought}`,
      input_tokens: 0,
      output_tokens: 0,
      model: '',
      created_at: new Date().toISOString(),
    }
    setChatMessages(prev => [...prev, userMsg])

    send(convoId, `Expand this thought into a full Zettel: ${thought}`, (fullText) => {
      const suggestions = parseZettelSuggestions(fullText)
      console.log('[cerebray] parsed suggestions:', suggestions.length)
      if (suggestions.length > 0) {
        const s = suggestions[0]
        console.log('[cerebray] applying suggestion:', s.title, '| body length:', s.body.length)
        // Must include status/tlp/additional - backend does a full field replace
        getNote(noteId).then((current) => {
          const updates: Partial<Note> = {
            title: s.title || title,
            summary: s.summary,
            core_idea: s.core_idea,
            laymans_terms: s.laymans_terms,
            analogy: s.analogy,
            body: s.body,
            components: s.components,
            why_it_matters: s.why_it_matters,
            examples: s.examples,
            templates: s.templates,
            note_type: s.type as NoteType,
            status: current.status,
            tlp: current.tlp,
            additional: current.additional,
          }
          console.log('[cerebray] calling updateNote with status:', updates.status, 'tlp:', updates.tlp)
          return updateNote(noteId, updates)
        }).then((updated) => {
          console.log('[cerebray] note updated, title:', updated.title)
          setNote(updated)
          setDraft(updated)
        }).catch((err) => {
          console.error('[cerebray] auto-apply failed:', err)
        })
      }
      setChatMessages(prev => [...prev, {
        id: Date.now(),
        conversation_id: convoId,
        role: 'assistant' as const,
        content: fullText,
        input_tokens: 0,
        output_tokens: 0,
        model: '',
        created_at: new Date().toISOString(),
      }])
    })
  }

  useEffect(() => {
    if (!promoted || !note?.source_chat_id || expandTriggered.current) return
    if (chatLoadId !== note.id) return // wait for chat history to load for this note
    // Only skip if note already has structured content from a previous expansion
    if (note.summary) return

    expandTriggered.current = true
    queueMicrotask(() => triggerExpand(note.source_chat_id!, note.title, note.body || '', note.id))
  }, [promoted, chatLoadId, note]) // eslint-disable-line react-hooks/exhaustive-deps

  // Scroll chat to bottom when streaming
  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [streamedText, chatMessages])

  // Search for notes to link (debounced)
  useEffect(() => {
    const timer = setTimeout(() => {
      if (!linkQuery.trim()) {
        setLinkResults([])
        return
      }
      searchNotes(linkQuery, 5).then((results) => {
        setLinkResults(results.filter((r) => r.id !== note?.id))
      }).catch(() => {})
    }, 300)
    return () => clearTimeout(timer)
  }, [linkQuery, note?.id])

  const handleChange = (key: string, value: string) => {
    setDraft((prev) => ({ ...prev, [key]: value }))
  }

  const handleSave = async () => {
    if (!note) return
    setSaving(true)
    try {
      const updated = await updateNote(note.id, draft)
      setNote(updated)
      setDraft(updated)
    } catch {
      // ignore
    }
    setSaving(false)
  }

  const handleDelete = async () => {
    if (!note) return
    await deleteNote(note.id).catch(() => {})
    navigate('/codex')
  }

  const handleStatusAction = async (action: (id: number) => Promise<Note>) => {
    if (!note) return
    const updated = await action(note.id).catch(() => null)
    if (updated) {
      setNote(updated)
      setDraft(updated)
    }
  }

  const handleAddTag = async () => {
    if (!note || !newTag.trim()) return
    await addTagToNote(note.id, newTag.trim()).catch(() => {})
    setNewTag('')
    const updated = await getNote(note.id).catch(() => null)
    if (updated) {
      setNote(updated)
      setDraft(updated)
    }
  }

  const handleCreateConnection = async (targetId: number) => {
    if (!note) return
    await createConnection(note.id, targetId, linkLabel.trim()).catch(() => {})
    setLinkQuery('')
    setLinkLabel('')
    setLinkResults([])
    setShowLinkBuilder(false)
    const conns = await listConnectionsForNote(note.id).catch(() => [])
    setConnections(conns)
  }

  const handleSendChat = () => {
    if (!note?.source_chat_id || !chatInput.trim() || streaming) return
    const content = chatInput.trim()
    setChatInput('')

    // Add user message to display
    setChatMessages(prev => [...prev, {
      id: Date.now(),
      conversation_id: note.source_chat_id!,
      role: 'user' as const,
      content,
      input_tokens: 0,
      output_tokens: 0,
      model: '',
      created_at: new Date().toISOString(),
    }])

    send(note.source_chat_id, content, (fullText) => {
      setChatMessages(prev => [...prev, {
        id: Date.now(),
        conversation_id: note.source_chat_id!,
        role: 'assistant' as const,
        content: fullText,
        input_tokens: 0,
        output_tokens: 0,
        model: '',
        created_at: new Date().toISOString(),
      }])
    })
  }

  if (loading) {
    return (
      <div className="flex items-center gap-2 text-zinc-400">
        <Loader2 className="h-4 w-4 animate-spin" /> Loading...
      </div>
    )
  }

  if (!note) return null

  const expanding = promoted && streaming
  const isFleeting = note.status === 'fleeting'

  const outgoing = connections.filter((c) => c.direction === 'outgoing')
  const incoming = connections.filter((c) => c.direction === 'incoming')

  const handlePromote = async () => {
    const result = await promoteNote(note.id)
    navigate(`/codex/${result.note.id}?promoted=true`)
  }

  const handleDiscard = async () => {
    await sleepNote(note.id)
    navigate('/inbox')
  }

  // Fleeting note: simple scratchpad view
  if (isFleeting) {
    return (
      <div className="mx-auto max-w-2xl space-y-6 pb-12">
        <div className="flex items-center gap-3">
          <Link to="/inbox" className="rounded p-1 text-zinc-400 hover:text-white">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <span className="rounded bg-amber-900/50 px-2 py-0.5 text-xs font-medium text-amber-300">fleeting</span>
        </div>

        <input
          type="text"
          value={draft.title || ''}
          onChange={(e) => handleChange('title', e.target.value)}
          className="w-full bg-transparent text-2xl font-bold text-zinc-100 focus:outline-none"
          placeholder="What's the thought?"
          autoFocus
        />

        <textarea
          value={(draft.body as string) || ''}
          onChange={(e) => handleChange('body', e.target.value)}
          rows={8}
          className="w-full resize-y rounded-lg border border-zinc-800 bg-zinc-900/50 px-4 py-3 text-sm text-zinc-200 placeholder-zinc-600 focus:border-zinc-600 focus:outline-none"
          placeholder="Jot down your thoughts, context, questions... anything that helps capture what you're thinking about."
        />

        <div className="flex items-center gap-3">
          <button
            onClick={handleSave}
            disabled={saving}
            className="flex items-center gap-2 rounded-lg bg-zinc-800 px-4 py-2 text-sm font-medium text-zinc-200 transition-colors hover:bg-zinc-700 disabled:opacity-50"
          >
            {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
            Save
          </button>
          <button
            onClick={handlePromote}
            className="flex items-center gap-2 rounded-lg bg-emerald-900/30 px-4 py-2 text-sm font-medium text-emerald-400 transition-colors hover:bg-emerald-900/50"
          >
            <Sparkles className="h-4 w-4" />
            Promote
          </button>
          <button
            onClick={handleDiscard}
            className="flex items-center gap-2 rounded-lg border border-zinc-800 px-4 py-2 text-sm text-zinc-500 transition-colors hover:bg-zinc-800"
          >
            <CloudFog className="h-4 w-4" />
            Discard
          </button>
          <button
            onClick={handleDelete}
            className="ml-auto flex items-center gap-2 rounded-lg border border-red-800/50 px-4 py-2 text-sm text-red-400 transition-colors hover:bg-red-900/20"
          >
            <Trash2 className="h-4 w-4" />
          </button>
        </div>
      </div>
    )
  }

  // Active/linked/sleeping/archived: full Zettel view
  return (
    <div className="mx-auto max-w-3xl space-y-6 pb-12">
      {/* Header */}
      <div className="flex items-center gap-3">
        <Link to="/codex" className="rounded p-1 text-zinc-400 hover:text-white">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <input
          type="text"
          value={draft.title || ''}
          onChange={(e) => handleChange('title', e.target.value)}
          className="flex-1 bg-transparent text-2xl font-bold text-zinc-100 focus:outline-none"
          placeholder="Note title"
        />
      </div>

      {/* Expanding indicator */}
      {expanding && (
        <div className="flex items-center gap-2 rounded-lg border border-emerald-800/40 bg-emerald-900/20 px-4 py-3">
          <Sparkles className="h-4 w-4 animate-pulse text-emerald-400" />
          <span className="text-sm text-emerald-300">AI is expanding your thought...</span>
        </div>
      )}

      {/* Meta row */}
      <div className="flex flex-wrap items-center gap-3">
        <select
          value={draft.note_type || note.note_type}
          onChange={(e) => handleChange('note_type', e.target.value)}
          className="rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm text-zinc-300 focus:outline-none"
        >
          {noteTypes.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>

        <select
          value={draft.tlp || note.tlp}
          onChange={(e) => handleChange('tlp', e.target.value)}
          className="rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-sm text-zinc-300 focus:outline-none"
        >
          {tlpOptions.map((t) => (
            <option key={t} value={t}>TLP: {t}</option>
          ))}
        </select>

        <span className={cn(
          'rounded px-2 py-0.5 text-xs font-medium',
          note.status === 'sleeping' && 'bg-purple-900/50 text-purple-300',
          note.status === 'active' && 'bg-emerald-900/50 text-emerald-300',
          note.status === 'linked' && 'bg-blue-900/50 text-blue-300',
          note.status === 'archived' && 'bg-zinc-800 text-zinc-400',
        )}>
          {note.status}
        </span>

        <div className="ml-auto flex gap-1">
          {note.status === 'sleeping' && (
            <button onClick={() => handleStatusAction(sleepNote)} className="rounded p-1.5 text-purple-400 hover:bg-purple-900/30" title="Sleep">
              <Moon className="h-4 w-4" />
            </button>
          )}
          {(note.status === 'active' || note.status === 'linked') && (
            <button onClick={() => handleStatusAction(archiveNote)} className="rounded p-1.5 text-zinc-500 hover:bg-zinc-800" title="Archive">
              <Archive className="h-4 w-4" />
            </button>
          )}
        </div>
      </div>

      {/* Tags */}
      <div className="flex flex-wrap items-center gap-2">
        {note.tags?.map((tag) => (
          <span key={tag} className="rounded bg-zinc-800 px-2 py-0.5 text-xs text-zinc-400">
            {tag}
          </span>
        ))}
        <div className="flex items-center gap-1">
          <input
            type="text"
            value={newTag}
            onChange={(e) => setNewTag(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAddTag()}
            placeholder="Add tag..."
            className="w-24 rounded border border-zinc-800 bg-transparent px-2 py-0.5 text-xs text-zinc-400 placeholder-zinc-600 focus:border-zinc-600 focus:outline-none"
          />
          <button onClick={handleAddTag} className="text-zinc-600 hover:text-zinc-400">
            <Plus className="h-3 w-3" />
          </button>
        </div>
      </div>

      {/* Content sections */}
      <div className="space-y-4">
        {contentSections.map(({ key, label }) => (
          <div key={key}>
            <label className="mb-1 block text-xs font-medium uppercase tracking-wide text-zinc-500">
              {label}
            </label>
            <textarea
              value={(draft[key] as string) || ''}
              onChange={(e) => handleChange(key, e.target.value)}
              rows={3}
              className="w-full resize-y rounded-lg border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm text-zinc-200 placeholder-zinc-600 focus:border-zinc-600 focus:outline-none"
              placeholder={`${label}...`}
            />
          </div>
        ))}
      </div>

      {/* Connections */}
      <div>
        <div className="mb-3 flex items-center justify-between">
          <h3 className="flex items-center gap-2 text-sm font-medium text-zinc-400">
            <Link2 className="h-4 w-4" />
            Connections ({connections.length})
          </h3>
          <button
            onClick={() => setShowLinkBuilder(!showLinkBuilder)}
            className="flex items-center gap-1 rounded bg-zinc-800 px-2.5 py-1 text-xs text-zinc-300 transition-colors hover:bg-zinc-700"
          >
            <Plus className="h-3 w-3" />
            Link note
          </button>
        </div>

        {showLinkBuilder && (
          <div className="mb-3 rounded-lg border border-zinc-700 bg-zinc-900 p-3 space-y-2">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-zinc-500" />
              <input
                type="text"
                value={linkQuery}
                onChange={(e) => setLinkQuery(e.target.value)}
                placeholder="Search for a note to link..."
                className="w-full rounded border border-zinc-700 bg-zinc-950 py-1.5 pl-8 pr-3 text-sm text-zinc-100 placeholder-zinc-500 focus:border-zinc-500 focus:outline-none"
                autoFocus
              />
            </div>
            <input
              type="text"
              value={linkLabel}
              onChange={(e) => setLinkLabel(e.target.value)}
              placeholder="Relationship label (optional, e.g. 'builds on')"
              className="w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-1.5 text-sm text-zinc-100 placeholder-zinc-500 focus:border-zinc-500 focus:outline-none"
            />
            {linkResults.length > 0 && (
              <div className="space-y-1">
                {linkResults.map((r) => (
                  <button
                    key={r.id}
                    onClick={() => handleCreateConnection(r.id)}
                    className="flex w-full items-center gap-2 rounded bg-zinc-800 px-3 py-2 text-left text-sm transition-colors hover:bg-zinc-700"
                  >
                    <Link2 className="h-3.5 w-3.5 shrink-0 text-zinc-500" />
                    <span className="truncate text-zinc-200">{r.title}</span>
                    <span className="ml-auto shrink-0 text-xs text-zinc-500">{r.note_type}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
        )}

        {outgoing.length > 0 && (
          <div className="mb-3">
            <div className="mb-1 flex items-center gap-1.5 text-xs text-zinc-500">
              <ArrowRight className="h-3 w-3" />
              Links from this note
            </div>
            <div className="space-y-1">
              {outgoing.map((conn) => (
                <Link
                  key={conn.id}
                  to={`/codex/${conn.connected_id}`}
                  className="flex items-center gap-2 rounded border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm transition-colors hover:border-zinc-700"
                >
                  <span className="text-zinc-100">{conn.connected_title}</span>
                  {conn.label && <span className="text-xs text-zinc-500">{conn.label}</span>}
                </Link>
              ))}
            </div>
          </div>
        )}

        {incoming.length > 0 && (
          <div className="mb-3">
            <div className="mb-1 flex items-center gap-1.5 text-xs text-zinc-500">
              <ArrowDownLeft className="h-3 w-3" />
              Backlinks
            </div>
            <div className="space-y-1">
              {incoming.map((conn) => (
                <Link
                  key={conn.id}
                  to={`/codex/${conn.connected_id}`}
                  className="flex items-center gap-2 rounded border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm transition-colors hover:border-zinc-700"
                >
                  <span className="text-zinc-100">{conn.connected_title}</span>
                  {conn.label && <span className="text-xs text-zinc-500">{conn.label}</span>}
                </Link>
              ))}
            </div>
          </div>
        )}

        {connections.length === 0 && !showLinkBuilder && (
          <p className="text-sm text-zinc-600">No connections yet.</p>
        )}
      </div>

      {/* Action bar */}
      <div className="flex items-center gap-3 border-t border-zinc-800 pt-4">
        <button
          onClick={handleSave}
          disabled={saving}
          className="flex items-center gap-2 rounded-lg bg-white px-4 py-2 text-sm font-medium text-zinc-900 transition-colors hover:bg-zinc-200 disabled:opacity-50"
        >
          {saving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
          Save
        </button>
        <button
          onClick={handleDelete}
          className="flex items-center gap-2 rounded-lg border border-red-800/50 px-4 py-2 text-sm text-red-400 transition-colors hover:bg-red-900/20"
        >
          <Trash2 className="h-4 w-4" />
          Delete
        </button>
      </div>

      {/* Chat section - only for notes with a linked conversation */}
      {note.source_chat_id && (
        <div className="border-t border-zinc-800 pt-6">
          <h3 className="mb-4 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-zinc-400">
            <MessageSquare className="h-4 w-4" />
            Research Chat
          </h3>

          <div className="space-y-3">
            {chatMessages.map((msg) => (
              <div
                key={msg.id}
                className={cn(
                  'rounded-lg px-4 py-3 text-sm',
                  msg.role === 'user'
                    ? 'ml-8 bg-zinc-800 text-zinc-200'
                    : 'mr-8 border border-zinc-800/60 bg-zinc-900/50 text-zinc-300'
                )}
              >
                <div className="mb-1 text-xs font-medium text-zinc-500">
                  {msg.role === 'user' ? 'You' : 'AI'}
                </div>
                <div className="whitespace-pre-wrap">
                  {msg.role === 'assistant' ? stripSuggestionBlocks(msg.content) : msg.content}
                </div>
              </div>
            ))}

            {/* Streaming indicator */}
            {streaming && (
              <div className="mr-8 rounded-lg border border-zinc-800/60 bg-zinc-900/50 px-4 py-3 text-sm text-zinc-300">
                <div className="mb-1 text-xs font-medium text-zinc-500">AI</div>
                <div className="whitespace-pre-wrap">{stripSuggestionBlocks(streamedText) || 'Thinking...'}</div>
              </div>
            )}

            {chatError && (
              <div className="rounded-lg border border-red-800/40 bg-red-900/20 px-4 py-2 text-sm text-red-400">
                {chatError}
              </div>
            )}

            <div ref={chatEndRef} />
          </div>

          {/* Chat input */}
          <div className="mt-4 flex gap-2">
            <input
              type="text"
              value={chatInput}
              onChange={(e) => setChatInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && !e.shiftKey) {
                  e.preventDefault()
                  handleSendChat()
                }
              }}
              placeholder="Ask a follow-up question..."
              disabled={streaming}
              className="flex-1 rounded-lg border border-zinc-800 bg-zinc-900/50 px-4 py-2.5 text-sm text-zinc-100 placeholder-zinc-600 outline-none transition-colors focus:border-zinc-600 disabled:opacity-50"
            />
            {streaming ? (
              <button
                onClick={cancel}
                className="rounded-lg bg-red-900/30 px-3 py-2.5 text-red-400 transition-colors hover:bg-red-900/50"
              >
                <StopCircle className="h-4 w-4" />
              </button>
            ) : (
              <button
                onClick={handleSendChat}
                disabled={!chatInput.trim()}
                className="rounded-lg bg-zinc-800 px-3 py-2.5 text-zinc-300 transition-colors hover:bg-zinc-700 disabled:opacity-40"
              >
                <Send className="h-4 w-4" />
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
