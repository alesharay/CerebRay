import { Link } from 'react-router-dom'
import type { Note } from '../../types'
import { cn } from '../../lib/utils'
import { Clock, ArrowUpCircle, Moon, Archive } from 'lucide-react'

const statusStyles: Record<string, string> = {
  fleeting: 'bg-amber-900/50 text-amber-300',
  sleeping: 'bg-purple-900/50 text-purple-300',
  active: 'bg-emerald-900/50 text-emerald-300',
  linked: 'bg-blue-900/50 text-blue-300',
  archived: 'bg-zinc-800 text-zinc-400',
}

const typeStyles: Record<string, string> = {
  concept: 'text-sky-400',
  theory: 'text-violet-400',
  insight: 'text-amber-400',
  quote: 'text-pink-400',
  reference: 'text-zinc-400',
  question: 'text-orange-400',
  structure: 'text-teal-400',
  guide: 'text-emerald-400',
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

interface ActionButton {
  label: string
  icon: React.ElementType
  onClick: () => void
  className: string
}

interface NoteCardProps {
  note: Note
  snippet?: string
  actions?: ActionButton[]
}

export function NoteCard({ note, snippet, actions }: NoteCardProps) {
  return (
    <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-4 transition-colors hover:border-zinc-700">
      <div className="flex items-start justify-between gap-3">
        <Link to={`/codex/${note.id}`} className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span className="truncate font-medium text-zinc-100">{note.title}</span>
            <span className={cn('text-xs', typeStyles[note.note_type] || 'text-zinc-500')}>
              {note.note_type}
            </span>
          </div>
          {snippet ? (
            <p
              className="mt-1 line-clamp-2 text-sm text-zinc-400 [&>mark]:bg-amber-700/40 [&>mark]:text-amber-200 [&>mark]:rounded [&>mark]:px-0.5"
              dangerouslySetInnerHTML={{ __html: snippet }}
            />
          ) : note.summary ? (
            <p className="mt-1 line-clamp-2 text-sm text-zinc-400">{note.summary}</p>
          ) : null}
          <div className="mt-2 flex items-center gap-3">
            <span className={cn('rounded px-2 py-0.5 text-xs font-medium', statusStyles[note.status])}>
              {note.status}
            </span>
            {note.tags && note.tags.length > 0 && (
              <div className="flex gap-1">
                {note.tags.slice(0, 3).map((tag) => (
                  <span key={tag} className="rounded bg-zinc-800 px-1.5 py-0.5 text-xs text-zinc-500">
                    {tag}
                  </span>
                ))}
                {note.tags.length > 3 && (
                  <span className="text-xs text-zinc-600">+{note.tags.length - 3}</span>
                )}
              </div>
            )}
            <span className="flex items-center gap-1 text-xs text-zinc-600">
              <Clock className="h-3 w-3" />
              {timeAgo(note.updated_at)}
            </span>
          </div>
        </Link>

        {actions && actions.length > 0 && (
          <div className="flex shrink-0 gap-1">
            {actions.map((action) => (
              <button
                key={action.label}
                onClick={action.onClick}
                title={action.label}
                className={cn('rounded p-1.5 transition-colors', action.className)}
              >
                <action.icon className="h-4 w-4" />
              </button>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

// Re-export action icons for convenience
export { ArrowUpCircle, Moon, Archive }
