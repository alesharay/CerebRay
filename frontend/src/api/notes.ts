import { api } from './client'
import type { Note } from '../types'

export interface ListNotesParams {
  status?: string
  type?: string
  q?: string
  limit?: number
  offset?: number
}

export function listNotes(params?: ListNotesParams): Promise<Note[]> {
  const search = new URLSearchParams()
  if (params?.status) search.set('status', params.status)
  if (params?.type) search.set('type', params.type)
  if (params?.q) search.set('q', params.q)
  if (params?.limit) search.set('limit', String(params.limit))
  if (params?.offset) search.set('offset', String(params.offset))
  const qs = search.toString()
  return api.get<Note[]>(`/notes${qs ? '?' + qs : ''}`)
}

export function getNote(id: number): Promise<Note> {
  return api.get<Note>(`/notes/${id}`)
}

export function createNote(note: Partial<Note>): Promise<Note> {
  return api.post<Note>('/notes', note)
}

export function updateNote(id: number, note: Partial<Note>): Promise<Note> {
  return api.put<Note>(`/notes/${id}`, note)
}

export function deleteNote(id: number): Promise<void> {
  return api.delete<void>(`/notes/${id}`)
}

export function promoteNote(id: number): Promise<Note> {
  return api.post<Note>(`/notes/${id}/promote`)
}

export function sleepNote(id: number): Promise<Note> {
  return api.post<Note>(`/notes/${id}/sleep`)
}

export function archiveNote(id: number): Promise<Note> {
  return api.post<Note>(`/notes/${id}/archive`)
}

export function searchNotes(q: string, limit?: number): Promise<Note[]> {
  return listNotes({ q, limit })
}
