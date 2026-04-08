import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { getNote, updateNote, deleteNote, promoteNote, sleepNote, archiveNote } from '../api/notes'
import { listConnectionsForNote } from '../api/connections'
import { addTagToNote } from '../api/tags'
import type { Note, Connection, NoteType, NoteTLP } from '../types'
import { cn } from '../lib/utils'
import {
  ArrowLeft, Save, Trash2, ArrowUpCircle, Moon, Archive,
  Link2, Plus, Loader2,
} from 'lucide-react'

const noteTypes: NoteType[] = ['concept', 'theory', 'insight', 'quote', 'reference', 'question', 'structure', 'guide']
const tlpOptions: NoteTLP[] = ['clear', 'green', 'amber', 'red']

const sections: { key: keyof Note; label: string; multiline?: boolean }[] = [
  { key: 'summary', label: 'Summary', multiline: true },
  { key: 'core_idea', label: 'Core idea', multiline: true },
  { key: 'laymans_terms', label: "Layman's terms", multiline: true },
  { key: 'analogy', label: 'Analogy', multiline: true },
  { key: 'body', label: 'Body', multiline: true },
  { key: 'components', label: 'Components', multiline: true },
  { key: 'why_it_matters', label: 'Why it matters', multiline: true },
  { key: 'examples', label: 'Examples', multiline: true },
  { key: 'templates', label: 'Templates', multiline: true },
  { key: 'additional', label: 'Additional', multiline: true },
]

export function NoteDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [note, setNote] = useState<Note | null>(null)
  const [connections, setConnections] = useState<Connection[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [draft, setDraft] = useState<Partial<Note>>({})
  const [newTag, setNewTag] = useState('')

  useEffect(() => {
    if (!id) return
    const noteId = Number(id)
    Promise.all([
      getNote(noteId),
      listConnectionsForNote(noteId).catch(() => []),
    ]).then(([n, c]) => {
      setNote(n)
      setDraft(n)
      setConnections(c)
    }).catch(() => {
      navigate('/codex')
    }).finally(() => setLoading(false))
  }, [id, navigate])

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
    // Reload note to get updated tags
    const updated = await getNote(note.id).catch(() => null)
    if (updated) {
      setNote(updated)
      setDraft(updated)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center gap-2 text-zinc-400">
        <Loader2 className="h-4 w-4 animate-spin" /> Loading...
      </div>
    )
  }

  if (!note) return null

  return (
    <div className="mx-auto max-w-3xl space-y-6">
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
          note.status === 'fleeting' && 'bg-amber-900/50 text-amber-300',
          note.status === 'sleeping' && 'bg-purple-900/50 text-purple-300',
          note.status === 'active' && 'bg-emerald-900/50 text-emerald-300',
          note.status === 'linked' && 'bg-blue-900/50 text-blue-300',
          note.status === 'archived' && 'bg-zinc-800 text-zinc-400',
        )}>
          {note.status}
        </span>

        <div className="ml-auto flex gap-1">
          {note.status === 'fleeting' && (
            <>
              <button onClick={() => handleStatusAction(promoteNote)} className="rounded p-1.5 text-emerald-500 hover:bg-emerald-900/30" title="Promote">
                <ArrowUpCircle className="h-4 w-4" />
              </button>
              <button onClick={() => handleStatusAction(sleepNote)} className="rounded p-1.5 text-purple-400 hover:bg-purple-900/30" title="Sleep">
                <Moon className="h-4 w-4" />
              </button>
            </>
          )}
          {note.status === 'sleeping' && (
            <button onClick={() => handleStatusAction(promoteNote)} className="rounded p-1.5 text-emerald-500 hover:bg-emerald-900/30" title="Promote">
              <ArrowUpCircle className="h-4 w-4" />
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
        {sections.map(({ key, label }) => (
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
      {connections.length > 0 && (
        <div>
          <h3 className="mb-2 flex items-center gap-2 text-sm font-medium text-zinc-400">
            <Link2 className="h-4 w-4" />
            Connections ({connections.length})
          </h3>
          <div className="space-y-1">
            {connections.map((conn) => (
              <Link
                key={conn.id}
                to={`/codex/${conn.connected_id}`}
                className="flex items-center gap-2 rounded border border-zinc-800 bg-zinc-900 px-3 py-2 text-sm transition-colors hover:border-zinc-700"
              >
                <span className="text-zinc-100">{conn.connected_title}</span>
                {conn.label && (
                  <span className="text-xs text-zinc-500">{conn.label}</span>
                )}
              </Link>
            ))}
          </div>
        </div>
      )}

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
    </div>
  )
}
