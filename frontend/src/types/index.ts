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
  direction: 'outgoing' | 'incoming'
  created_at: string
}

export interface SearchResult extends Note {
  snippet: string
}

export interface ConnectionSuggestion {
  source_title: string
  target_title: string
  label: string
  reason: string
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

export interface InboxItem {
  id: number
  title: string
  note_type: NoteType
  age_seconds: number
  readiness_score: number
  created_at: string
}

export interface LifecycleEntry {
  action: string
  count: number
  avg_dwell_seconds: number
}

export interface TypeCount {
  type: string
  count: number
}

export interface StrengthScore {
  overall: number
  connection_density: number
  orphan_count: number
  active_notes: number
  total_connections: number
  glossary_count: number
  glossary_coverage: number
  sleeping_backlog: number
  sleeping_avg_age_seconds: number
  type_distribution: TypeCount[]
}

export interface ConversionRate {
  total_conversations: number
  conversations_with_notes: number
  rate: number
}

export interface StaleNote {
  id: number
  title: string
  note_type: NoteType
  updated_at: string
  stale_days: number
}

export interface AnalyticsAIUsage {
  input_tokens: number
  output_tokens: number
  requests: number
}

export interface LifecycleTrendPoint {
  week: string
  action: string
  count: number
}

export interface AnalyticsData {
  inbox: InboxItem[]
  lifecycle: LifecycleEntry[]
  lifecycle_trend: LifecycleTrendPoint[]
  strength: StrengthScore
  conversion: ConversionRate
  ai_usage: AnalyticsAIUsage
  stale_notes: StaleNote[]
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
