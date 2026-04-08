export type NoteType = 'concept' | 'theory' | 'insight' | 'quote' | 'reference' | 'question' | 'structure' | 'guide'
export type NoteStatus = 'fleeting' | 'sleeping' | 'active' | 'linked' | 'archived'
export type NoteTLP = 'clear' | 'green' | 'amber' | 'red'

export interface User {
  id: number
  oidc_subject: string
  email: string
  name: string
  avatar_url: string
  created_at: string
  updated_at: string
}

export interface Note {
  id: number
  user_id: number
  title: string
  summary: string
  laymans_terms: string
  analogy: string
  core_idea: string
  body: string
  components: string
  why_it_matters: string
  examples: string
  templates: string
  additional: string
  note_type: NoteType
  status: NoteStatus
  tlp: NoteTLP
  source_chat_id?: number
  tags?: string[]
  connection_count: number
  created_at: string
  updated_at: string
}

export interface Connection {
  id: number
  source_id: number
  target_id: number
  label: string
  connected_id: number
  connected_title: string
  created_at: string
}

export interface Tag {
  id: number
  name: string
  note_count: number
}

export interface Conversation {
  id: number
  user_id: number
  title: string
  topic: string
  created_at: string
  updated_at: string
}

export interface Message {
  id: number
  conversation_id: number
  role: 'user' | 'assistant' | 'system'
  content: string
  input_tokens: number
  output_tokens: number
  model: string
  created_at: string
}

export interface GlossaryTerm {
  id: number
  user_id: number
  term: string
  definition: string
  source_note_id?: number
  source_note_title?: string
  created_at: string
  updated_at: string
}

export interface DashboardStats {
  stats: {
    inbox: number
    echoes: number
    codex: number
  }
  recent: Note[]
}

export interface GraphData {
  nodes: { id: number; title: string; type: NoteType }[]
  edges: { source: number; target: number; label: string }[]
}

export interface AIUsage {
  total_input_tokens: number
  total_output_tokens: number
  total_requests: number
}

// Zettel suggestion parsed from AI response
export interface ZettelSuggestion {
  title: string
  type: NoteType
  summary: string
  laymans_terms: string
  analogy: string
  core_idea: string
  body: string
  components: string
  why_it_matters: string
  examples: string
  templates: string
  tags: string[]
}
