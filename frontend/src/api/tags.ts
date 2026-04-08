import { api } from './client'
import type { Tag } from '../types'

export function listTags(): Promise<Tag[]> {
  return api.get<Tag[]>('/tags')
}

export function addTagToNote(noteId: number, name: string): Promise<void> {
  return api.post<void>(`/notes/${noteId}/tags`, { name })
}

export function removeTagFromNote(noteId: number, tagId: number): Promise<void> {
  return api.delete<void>(`/notes/${noteId}/tags/${tagId}`)
}
