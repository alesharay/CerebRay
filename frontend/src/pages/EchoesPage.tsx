import { useEffect, useState } from 'react'
import { listNotes, promoteNote, archiveNote } from '../api/notes'
import type { Note } from '../types'
import { NoteCard, ArrowUpCircle, Archive } from '../components/notes/NoteCard'
import { CloudFog } from 'lucide-react'

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
