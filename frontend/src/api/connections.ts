import { api } from './client'
import type { Connection, GraphData } from '../types'

export function listConnectionsForNote(noteId: number): Promise<Connection[]> {
  return api.get<Connection[]>(`/notes/${noteId}/connections`)
}

export function createConnection(sourceId: number, targetId: number, label: string): Promise<Connection> {
  return api.post<Connection>('/connections', { source_id: sourceId, target_id: targetId, label })
}

export function deleteConnection(id: number): Promise<void> {
  return api.delete<void>(`/connections/${id}`)
}

export function getGraphData(): Promise<GraphData> {
  return api.get<GraphData>('/index')
}
