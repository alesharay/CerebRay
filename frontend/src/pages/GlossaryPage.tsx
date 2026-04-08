import { useEffect, useState } from 'react'
import {
  listGlossaryTerms,
  createGlossaryTerm,
  updateGlossaryTerm,
  deleteGlossaryTerm,
} from '../api/glossary'
import type { GlossaryTerm } from '../types'
import { BookA, Plus, Pencil, Trash2, Check } from 'lucide-react'

export function GlossaryPage() {
  const [terms, setTerms] = useState<GlossaryTerm[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editingId, setEditingId] = useState<number | null>(null)
  const [formTerm, setFormTerm] = useState('')
  const [formDef, setFormDef] = useState('')

  const load = () => {
    listGlossaryTerms()
      .then((t) => setTerms(t.sort((a, b) => a.term.localeCompare(b.term))))
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const resetForm = () => {
    setShowForm(false)
    setEditingId(null)
    setFormTerm('')
    setFormDef('')
  }

  const handleCreate = async () => {
    if (!formTerm.trim() || !formDef.trim()) return
    const term = await createGlossaryTerm(formTerm.trim(), formDef.trim()).catch(() => null)
    if (term) {
      setTerms((prev) => [...prev, term].sort((a, b) => a.term.localeCompare(b.term)))
      resetForm()
    }
  }

  const startEdit = (t: GlossaryTerm) => {
    setEditingId(t.id)
    setFormTerm(t.term)
    setFormDef(t.definition)
  }

  const handleUpdate = async () => {
    if (!editingId || !formTerm.trim() || !formDef.trim()) return
    const updated = await updateGlossaryTerm(editingId, formTerm.trim(), formDef.trim()).catch(() => null)
    if (updated) {
      setTerms((prev) =>
        prev.map((t) => (t.id === editingId ? updated : t)).sort((a, b) => a.term.localeCompare(b.term))
      )
      resetForm()
    }
  }

  const handleDelete = async (id: number) => {
    await deleteGlossaryTerm(id).catch(() => {})
    setTerms((prev) => prev.filter((t) => t.id !== id))
  }

  if (loading) return <div className="text-zinc-400">Loading...</div>

  // Group by first letter
  const grouped: Record<string, GlossaryTerm[]> = {}
  for (const t of terms) {
    const letter = t.term[0]?.toUpperCase() || '#'
    if (!grouped[letter]) grouped[letter] = []
    grouped[letter].push(t)
  }
  const letters = Object.keys(grouped).sort()

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <BookA className="h-6 w-6 text-sky-400" />
          <h1 className="text-2xl font-bold">Glossary</h1>
          <span className="text-sm text-zinc-500">{terms.length} terms</span>
        </div>
        <button
          onClick={() => { resetForm(); setShowForm(true) }}
          className="flex items-center gap-1.5 rounded-lg bg-zinc-800 px-3 py-2 text-sm font-medium transition-colors hover:bg-zinc-700"
        >
          <Plus className="h-4 w-4" />
          Add term
        </button>
      </div>

      {/* Create form */}
      {showForm && (
        <div className="rounded-lg border border-zinc-700 bg-zinc-900 p-4 space-y-3">
          <input
            type="text"
            value={formTerm}
            onChange={(e) => setFormTerm(e.target.value)}
            placeholder="Term"
            className="w-full rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 placeholder-zinc-500 focus:border-zinc-500 focus:outline-none"
            autoFocus
          />
          <textarea
            value={formDef}
            onChange={(e) => setFormDef(e.target.value)}
            placeholder="Definition"
            rows={3}
            className="w-full resize-y rounded border border-zinc-700 bg-zinc-950 px-3 py-2 text-sm text-zinc-100 placeholder-zinc-500 focus:border-zinc-500 focus:outline-none"
          />
          <div className="flex gap-2">
            <button
              onClick={handleCreate}
              className="flex items-center gap-1.5 rounded bg-white px-3 py-1.5 text-sm font-medium text-zinc-900 hover:bg-zinc-200"
            >
              <Check className="h-3.5 w-3.5" /> Save
            </button>
            <button onClick={resetForm} className="rounded px-3 py-1.5 text-sm text-zinc-400 hover:text-white">
              Cancel
            </button>
          </div>
        </div>
      )}

      {terms.length === 0 && !showForm ? (
        <p className="text-zinc-500">No glossary terms yet. Add your first term to start building your vocabulary.</p>
      ) : (
        <div className="space-y-6">
          {letters.map((letter) => (
            <div key={letter}>
              <h2 className="mb-2 text-sm font-bold text-zinc-500">{letter}</h2>
              <div className="space-y-2">
                {grouped[letter].map((t) => (
                  <div key={t.id} className="rounded-lg border border-zinc-800 bg-zinc-900 px-4 py-3">
                    {editingId === t.id ? (
                      <div className="space-y-2">
                        <input
                          type="text"
                          value={formTerm}
                          onChange={(e) => setFormTerm(e.target.value)}
                          className="w-full rounded border border-zinc-700 bg-zinc-950 px-2 py-1 text-sm text-zinc-100 focus:outline-none"
                        />
                        <textarea
                          value={formDef}
                          onChange={(e) => setFormDef(e.target.value)}
                          rows={2}
                          className="w-full resize-y rounded border border-zinc-700 bg-zinc-950 px-2 py-1 text-sm text-zinc-100 focus:outline-none"
                        />
                        <div className="flex gap-2">
                          <button onClick={handleUpdate} className="rounded bg-white px-2 py-1 text-xs font-medium text-zinc-900 hover:bg-zinc-200">
                            Save
                          </button>
                          <button onClick={resetForm} className="text-xs text-zinc-400 hover:text-white">Cancel</button>
                        </div>
                      </div>
                    ) : (
                      <div className="flex items-start justify-between gap-3">
                        <div>
                          <span className="font-medium text-zinc-100">{t.term}</span>
                          <p className="mt-0.5 text-sm text-zinc-400">{t.definition}</p>
                          {t.source_note_title && (
                            <span className="mt-1 inline-block text-xs text-zinc-600">
                              Source: {t.source_note_title}
                            </span>
                          )}
                        </div>
                        <div className="flex shrink-0 gap-1">
                          <button onClick={() => startEdit(t)} className="rounded p-1 text-zinc-600 hover:text-zinc-300">
                            <Pencil className="h-3.5 w-3.5" />
                          </button>
                          <button onClick={() => handleDelete(t.id)} className="rounded p-1 text-zinc-600 hover:text-red-400">
                            <Trash2 className="h-3.5 w-3.5" />
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
