import { useEffect, useState } from 'react'
import { listNotes, promoteNote, sleepNote, archiveNote } from '../api/notes'
import type { Note } from '../types'
import { NoteCard, ArrowUpCircle, Moon, Archive } from '../components/notes/NoteCard'
import { Inbox } from 'lucide-react'
import { Link } from 'react-router-dom'

export function InboxPage() {
  const [notes, setNotes] = useState<Note[]>([])
  const [loading, setLoading] = useState(true)

  const load = () => {
    listNotes({ status: 'fleeting' })
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
        <Inbox className="h-6 w-6 text-amber-400" />
        <h1 className="text-2xl font-bold">Inbox</h1>
        <span className="text-sm text-zinc-500">{notes.length} fleeting notes</span>
      </div>

      {notes.length === 0 ? (
        <p className="text-zinc-500">
          Inbox is empty. <Link to="/chat" className="text-white underline">Chat with AI</Link> to capture new ideas.
        </p>
      ) : (
        <div className="space-y-2">
          {notes.map((note) => (
            <NoteCard
              key={note.id}
              note={note}
              actions={[
                {
                  label: 'Promote to Codex',
                  icon: ArrowUpCircle,
                  onClick: () => handleAction(note.id, promoteNote),
                  className: 'text-emerald-500 hover:bg-emerald-900/30',
                },
                {
                  label: 'Sleep',
                  icon: Moon,
                  onClick: () => handleAction(note.id, sleepNote),
                  className: 'text-purple-400 hover:bg-purple-900/30',
                },
                {
                  label: 'Archive',
                  icon: Archive,
                  onClick: () => handleAction(note.id, archiveNote),
                  className: 'text-zinc-500 hover:bg-zinc-800',
                },
              ]}
            />
          ))}
        </div>
      )}
    </div>
  )
}
