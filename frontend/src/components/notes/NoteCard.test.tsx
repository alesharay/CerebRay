import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { NoteCard, ArrowUpCircle } from './NoteCard'
import type { Note } from '../../types'

function makeNote(overrides: Partial<Note> = {}): Note {
  return {
    id: 1,
    user_id: 1,
    title: 'Test Note',
    summary: 'A brief summary',
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
    status: 'active',
    tlp: 'clear',
    connection_count: 0,
    tags: [],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    ...overrides,
  }
}

describe('NoteCard', () => {
  it('renders note title as a link to /codex/{id}', () => {
    const note = makeNote({ id: 42, title: 'My Zettel' })
    render(
      <MemoryRouter>
        <NoteCard note={note} />
      </MemoryRouter>
    )
    const link = screen.getByText('My Zettel').closest('a')
    expect(link).toHaveAttribute('href', '/codex/42')
  })

  it('displays note type', () => {
    const note = makeNote({ note_type: 'theory' })
    render(
      <MemoryRouter>
        <NoteCard note={note} />
      </MemoryRouter>
    )
    expect(screen.getByText('theory')).toBeInTheDocument()
  })

  it('shows summary text when present', () => {
    const note = makeNote({ summary: 'This is the summary' })
    render(
      <MemoryRouter>
        <NoteCard note={note} />
      </MemoryRouter>
    )
    expect(screen.getByText('This is the summary')).toBeInTheDocument()
  })

  it('shows status badge', () => {
    const note = makeNote({ status: 'fleeting' })
    render(
      <MemoryRouter>
        <NoteCard note={note} />
      </MemoryRouter>
    )
    expect(screen.getByText('fleeting')).toBeInTheDocument()
  })

  it('shows tags (first 3 + overflow count)', () => {
    const note = makeNote({ tags: ['alpha', 'beta', 'gamma', 'delta', 'epsilon'] })
    render(
      <MemoryRouter>
        <NoteCard note={note} />
      </MemoryRouter>
    )
    expect(screen.getByText('alpha')).toBeInTheDocument()
    expect(screen.getByText('beta')).toBeInTheDocument()
    expect(screen.getByText('gamma')).toBeInTheDocument()
    expect(screen.queryByText('delta')).not.toBeInTheDocument()
    expect(screen.getByText('+2')).toBeInTheDocument()
  })

  it('renders action buttons when provided', () => {
    const note = makeNote()
    const actions = [
      { label: 'Promote', icon: ArrowUpCircle, onClick: vi.fn(), className: 'text-green-500' },
    ]
    render(
      <MemoryRouter>
        <NoteCard note={note} actions={actions} />
      </MemoryRouter>
    )
    expect(screen.getByTitle('Promote')).toBeInTheDocument()
  })

  it('action button click calls onClick handler', () => {
    const note = makeNote()
    const handler = vi.fn()
    const actions = [
      { label: 'Promote', icon: ArrowUpCircle, onClick: handler, className: 'text-green-500' },
    ]
    render(
      <MemoryRouter>
        <NoteCard note={note} actions={actions} />
      </MemoryRouter>
    )
    fireEvent.click(screen.getByTitle('Promote'))
    expect(handler).toHaveBeenCalledTimes(1)
  })
})
