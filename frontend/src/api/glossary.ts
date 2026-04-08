import { api } from './client'
import type { GlossaryTerm } from '../types'

export function listGlossaryTerms(): Promise<GlossaryTerm[]> {
  return api.get<GlossaryTerm[]>('/glossary')
}

export function createGlossaryTerm(term: string, definition: string, sourceNoteId?: number): Promise<GlossaryTerm> {
  return api.post<GlossaryTerm>('/glossary', { term, definition, source_note_id: sourceNoteId ?? null })
}

export function updateGlossaryTerm(id: number, term: string, definition: string, sourceNoteId?: number): Promise<GlossaryTerm> {
  return api.put<GlossaryTerm>(`/glossary/${id}`, { term, definition, source_note_id: sourceNoteId ?? null })
}

export function deleteGlossaryTerm(id: number): Promise<void> {
  return api.delete<void>(`/glossary/${id}`)
}
