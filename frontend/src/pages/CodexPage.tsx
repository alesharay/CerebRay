import { useEffect, useRef, useState } from 'react'
import { listNotes, searchNotes, archiveNote } from '../api/notes'
import { listTags } from '../api/tags'
import type { Note, NoteType, Tag, SearchResult } from '../types'
import { NoteCard, Archive } from '../components/notes/NoteCard'
import { BookOpen, Search, X } from 'lucide-react'
import { cn } from '../lib/utils'

const noteTypes: { value: NoteType | ''; label: string }[] = [
  { value: '', label: 'All types' },
  { value: 'concept', label: 'Concept' },
  { value: 'theory', label: 'Theory' },
  { value: 'insight', label: 'Insight' },
  { value: 'quote', label: 'Quote' },
  { value: 'reference', label: 'Reference' },
  { value: 'question', label: 'Question' },
  { value: 'structure', label: 'Structure' },
  { value: 'guide', label: 'Guide' },
]

export function CodexPage() {
  const [notes, setNotes] = useState<Note[]>([])
  const [searchResults, setSearchResults] = useState<SearchResult[] | null>(null)
  const [tags, setTags] = useState<Tag[]>([])
  const [loading, setLoading] = useState(true)
  const [query, setQuery] = useState('')
  const [typeFilter, setTypeFilter] = useState<string>('')
  const [tagFilter, setTagFilter] = useState<string>('')
  const fetchIdRef = useRef(0)

  // Load tags once
  useEffect(() => {
    listTags().then(setTags).catch(() => {})
  }, [])

  // Load notes or search results
  useEffect(() => {
    const id = ++fetchIdRef.current

    if (query) {
      searchNotes(query).then((data) => {
        if (fetchIdRef.current === id) {
          setSearchResults(data)
          setNotes([])
          setLoading(false)
        }
      }).catch(() => {
        if (fetchIdRef.current === id) {
          setSearchResults([])
          setNotes([])
          setLoading(false)
        }
      })
    } else {
      listNotes({
        status: 'active',
        type: typeFilter || undefined,
        tag: tagFilter || undefined,
      }).then((data) => {
        if (fetchIdRef.current === id) {
          setSearchResults(null)
          setNotes(data)
          setLoading(false)
        }
      }).catch(() => {
        if (fetchIdRef.current === id) {
          setSearchResults(null)
          setNotes([])
          setLoading(false)
        }
      })
    }
  }, [query, typeFilter, tagFilter])

  const handleArchive = async (id: number) => {
    await archiveNote(id).catch(() => {})
    setNotes((prev) => prev.filter((n) => n.id !== id))
    setSearchResults((prev) => prev?.filter((n) => n.id !== id) ?? null)
  }

  const displayNotes = searchResults ?? notes
  const isSearching = searchResults !== null

  return (
    <div className="mx-auto max-w-3xl space-y-4">
      <div className="flex items-center gap-3">
        <BookOpen className="h-6 w-6 text-emerald-400" />
        <h1 className="text-2xl font-bold">Codex</h1>
        <span className="text-sm text-zinc-500">{displayNotes.length} notes</span>
      </div>

      {/* Search and filters */}
      <div className="flex gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-zinc-500" />
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search notes..."
            className="w-full rounded-lg border border-zinc-700 bg-zinc-900 py-2 pl-10 pr-4 text-sm text-zinc-100 placeholder-zinc-500 focus:border-zinc-500 focus:outline-none"
          />
        </div>
        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-300 focus:border-zinc-500 focus:outline-none"
        >
          {noteTypes.map((t) => (
            <option key={t.value} value={t.value}>
              {t.label}
            </option>
          ))}
        </select>
      </div>

      {/* Tag filter chips */}
      {tags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {tags.map((tag) => (
            <button
              key={tag.id}
              onClick={() => setTagFilter(tagFilter === tag.name ? '' : tag.name)}
              className={cn(
                'rounded-full px-3 py-1 text-xs transition-colors',
                tagFilter === tag.name
                  ? 'bg-emerald-800 text-emerald-200'
                  : 'bg-zinc-800 text-zinc-400 hover:bg-zinc-700'
              )}
            >
              {tag.name}
              {tagFilter === tag.name && <X className="ml-1 inline h-3 w-3" />}
            </button>
          ))}
        </div>
      )}

      {loading ? (
        <div className="text-zinc-400">Loading...</div>
      ) : displayNotes.length === 0 ? (
        <p className="text-zinc-500">
          {query ? 'No notes match your search.' : 'No active notes in the Codex yet.'}
        </p>
      ) : (
        <div className="space-y-2">
          {displayNotes.map((note) => (
            <NoteCard
              key={note.id}
              note={note}
              snippet={isSearching ? (note as SearchResult).snippet : undefined}
              actions={[
                {
                  label: 'Archive',
                  icon: Archive,
                  onClick: () => handleArchive(note.id),
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
