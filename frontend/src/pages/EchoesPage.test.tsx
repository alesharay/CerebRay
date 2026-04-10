import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { EchoesPage } from './EchoesPage'
import type { Note } from '../types'

vi.mock('../api/notes', () => ({
  listNotes: vi.fn(),
  promoteNote: vi.fn(),
  archiveNote: vi.fn(),
}))

import { listNotes, promoteNote, archiveNote } from '../api/notes'

const mockedListNotes = vi.mocked(listNotes)
const mockedPromoteNote = vi.mocked(promoteNote)
const mockedArchiveNote = vi.mocked(archiveNote)

function makeNote(overrides: Partial<Note> = {}): Note {
  return {
    id: 1,
    user_id: 1,
    title: 'Sleeping Note',
    summary: '',
    laymans_terms: '',
    analogy: '',
    core_idea: '',
    body: '',
    components: '',
    why_it_matters: '',
    examples: '',
    templates: '',
    additional: '',
    note_type: 'concept',
    status: 'sleeping',
    tlp: 'clear',
    connection_count: 0,
    tags: [],
    created_at: new Date(Date.now() - 3 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
    ...overrides,
  }
}

function renderPage() {
  return render(
    <MemoryRouter>
      <EchoesPage />
    </MemoryRouter>
  )
}

describe('EchoesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading state initially', () => {
    mockedListNotes.mockReturnValue(new Promise(() => {})) // never resolves
    renderPage()
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('renders sleeping notes after load', async () => {
    const notes = [
      makeNote({ id: 1, title: 'First Echo' }),
      makeNote({ id: 2, title: 'Second Echo' }),
    ]
    mockedListNotes.mockResolvedValueOnce(notes)
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('First Echo')).toBeInTheDocument()
      expect(screen.getByText('Second Echo')).toBeInTheDocument()
    })
  })

  it('shows "No sleeping notes" when empty', async () => {
    mockedListNotes.mockResolvedValueOnce([])
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('No sleeping notes.')).toBeInTheDocument()
    })
  })

  it('displays note title and age', async () => {
    const notes = [makeNote({ id: 1, title: 'Old Note' })]
    mockedListNotes.mockResolvedValueOnce(notes)
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Old Note')).toBeInTheDocument()
      expect(screen.getByText(/old$/)).toBeInTheDocument()
    })
  })

  it('promote button removes the note from the list', async () => {
    const notes = [makeNote({ id: 10, title: 'Promotable' })]
    mockedListNotes.mockResolvedValueOnce(notes)
    mockedPromoteNote.mockResolvedValueOnce({ note: { ...notes[0], status: 'active' }, conversation_id: 1 })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Promotable')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByTitle('Promote to Codex'))

    await waitFor(() => {
      expect(screen.queryByText('Promotable')).not.toBeInTheDocument()
    })
  })

  it('archive button removes the note from the list', async () => {
    const notes = [makeNote({ id: 20, title: 'Archivable' })]
    mockedListNotes.mockResolvedValueOnce(notes)
    mockedArchiveNote.mockResolvedValueOnce({ ...notes[0], status: 'archived' })
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Archivable')).toBeInTheDocument()
    })

    fireEvent.click(screen.getByTitle('Archive'))

    await waitFor(() => {
      expect(screen.queryByText('Archivable')).not.toBeInTheDocument()
    })
  })
})
