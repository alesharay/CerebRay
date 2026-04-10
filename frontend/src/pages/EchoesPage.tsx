import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listNotes, promoteNote, archiveNote } from '../api/notes'
import type { Note } from '../types'
import { CloudFog, ArrowUpCircle, Archive } from 'lucide-react'

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h`
  const days = Math.floor(hours / 24)
  return `${days}d`
}

export function EchoesPage() {
  const [notes, setNotes] = useState<Note[]>([])
  const [loading, setLoading] = useState(true)

  const load = () => {
    listNotes({ status: 'sleeping' })
      .then(setNotes)
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleAction = async (id: number, action: (id: number) => Promise<Note>) => {
    await action(id).catch(() => {})
    setNotes((prev) => prev.filter((n) => n.id !== id))
  }

  if (loading) return <div className="text-zinc-400">Loading...</div>

  return (
    <div className="mx-auto max-w-3xl space-y-4">
      <div className="flex items-center gap-3">
        <CloudFog className="h-6 w-6 text-purple-400" />
        <h1 className="text-2xl font-bold">Echoes</h1>
        <span className="text-sm text-zinc-500">{notes.length} sleeping notes</span>
      </div>

      <p className="text-sm text-zinc-500">
        Notes resting here. When one resonates, promote it to the Codex. If it no longer serves you, archive it.
      </p>

      {notes.length === 0 ? (
        <p className="text-zinc-500">No sleeping notes.</p>
      ) : (
        <div className="space-y-1">
          {notes.map((note) => (
            <div
              key={note.id}
              className="flex items-center gap-3 rounded-lg border border-zinc-800 bg-zinc-900 px-4 py-3 transition-colors hover:border-zinc-700"
            >
              <Link
                to={`/codex/${note.id}`}
                className="min-w-0 flex-1 truncate font-medium text-zinc-100 hover:underline"
              >
                {note.title || 'Untitled'}
              </Link>
              <span className="shrink-0 rounded bg-zinc-800 px-2 py-0.5 text-xs text-zinc-500">
                {timeAgo(note.created_at)} old
              </span>
              <div className="flex shrink-0 gap-1">
                <button
                  onClick={() =>
                    handleAction(note.id, (id) => promoteNote(id).then((r) => r.note))
                  }
                  title="Promote to Codex"
                  className="rounded p-1.5 text-emerald-500 transition-colors hover:bg-emerald-900/30"
                >
                  <ArrowUpCircle className="h-4 w-4" />
                </button>
                <button
                  onClick={() => handleAction(note.id, archiveNote)}
                  title="Archive"
                  className="rounded p-1.5 text-zinc-500 transition-colors hover:bg-zinc-800"
                >
                  <Archive className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
