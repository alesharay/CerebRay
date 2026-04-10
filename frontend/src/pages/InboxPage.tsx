import { useEffect, useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { listNotes, createNote, promoteNote, sleepNote } from '../api/notes'
import type { Note } from '../types'
import { Plus, Sparkles, CloudFog, Clock } from 'lucide-react'
import { cn } from '../lib/utils'

function formatAge(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return days === 1 ? '1 day ago' : `${days} days ago`
}

export function InboxPage() {
  const navigate = useNavigate()
  const [notes, setNotes] = useState<Note[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [promoting, setPromoting] = useState<number | null>(null)

  useEffect(() => {
    listNotes({ status: 'fleeting' })
      .then(setNotes)
      .finally(() => setLoading(false))
  }, [])

  const handleCapture = async () => {
    const title = input.trim()
    if (!title || submitting) return

    setSubmitting(true)
    try {
      const note = await createNote({ title, status: 'fleeting' })
      setNotes(prev => [note, ...prev])
      setInput('')
    } finally {
      setSubmitting(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleCapture()
    }
  }

  const handlePromote = async (id: number) => {
    setPromoting(id)
    try {
      const result = await promoteNote(id)
      navigate(`/codex/${result.note.id}?promoted=true`)
    } catch {
      setPromoting(null)
    }
  }

  const handleDiscard = async (id: number) => {
    setNotes(prev => prev.filter(n => n.id !== id))
    await sleepNote(id)
  }

  if (loading) {
    return <div className="text-zinc-500">Loading...</div>
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6 pb-12">
      <div>
        <h1 className="text-2xl font-bold text-zinc-100">Inbox</h1>
        <p className="mt-1 text-sm text-zinc-500">Capture a thought, promote it later</p>
      </div>

      <div className="flex gap-3">
        <input
          type="text"
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="What's on your mind?"
          className="flex-1 rounded-lg border border-zinc-800 bg-zinc-900/50 px-4 py-3 text-sm text-zinc-100 placeholder-zinc-600 outline-none transition-colors focus:border-zinc-600"
          autoFocus
        />
        <button
          onClick={handleCapture}
          disabled={!input.trim() || submitting}
          className="flex items-center gap-2 rounded-lg bg-zinc-800 px-4 py-3 text-sm font-medium text-zinc-200 transition-colors hover:bg-zinc-700 disabled:opacity-40 disabled:hover:bg-zinc-800"
        >
          <Plus className="h-4 w-4" />
          Capture
        </button>
      </div>

      {notes.length === 0 ? (
        <div className="rounded-xl border border-zinc-800/60 bg-zinc-900/30 p-8 text-center">
          <p className="text-sm text-zinc-500">
            No thoughts captured yet. Type something above and hit Enter.
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          {notes.map(note => (
            <div
              key={note.id}
              className="flex items-center justify-between rounded-lg border border-zinc-800/60 bg-zinc-900/50 px-4 py-3 transition-colors hover:border-zinc-700/60"
            >
              <Link to={`/codex/${note.id}`} className="min-w-0 flex-1">
                <div className="truncate text-sm font-medium text-zinc-200 hover:text-white">{note.title}</div>
                <div className="mt-0.5 flex items-center gap-1 text-xs text-zinc-600">
                  <Clock className="h-3 w-3" />
                  {formatAge(note.created_at)}
                </div>
              </Link>

              <div className="ml-4 flex items-center gap-2">
                <button
                  onClick={() => handleDiscard(note.id)}
                  className="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs text-zinc-500 transition-colors hover:bg-zinc-800 hover:text-zinc-300"
                  title="Move to Echoes"
                >
                  <CloudFog className="h-3.5 w-3.5" />
                  Discard
                </button>
                <button
                  onClick={() => handlePromote(note.id)}
                  disabled={promoting === note.id}
                  className={cn(
                    'flex items-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-colors',
                    promoting === note.id
                      ? 'bg-emerald-900/30 text-emerald-400'
                      : 'bg-emerald-900/20 text-emerald-400 hover:bg-emerald-900/40'
                  )}
                  title="Promote with AI expansion"
                >
                  <Sparkles className="h-3.5 w-3.5" />
                  {promoting === note.id ? 'Promoting...' : 'Promote'}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
