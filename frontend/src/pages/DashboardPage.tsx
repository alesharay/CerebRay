import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { getDashboardStats } from '../api/dashboard'
import type { DashboardStats, Note } from '../types'
import { Inbox, CloudFog, BookOpen, Clock, ArrowRight } from 'lucide-react'
import { cn } from '../lib/utils'

const statCards = [
  { key: 'inbox' as const, label: 'Inbox', icon: Inbox, color: 'text-amber-400', href: '/inbox' },
  { key: 'echoes' as const, label: 'Echoes', icon: CloudFog, color: 'text-purple-400', href: '/echoes' },
  { key: 'codex' as const, label: 'Codex', icon: BookOpen, color: 'text-emerald-400', href: '/codex' },
]

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

function statusBadge(status: string) {
  const styles: Record<string, string> = {
    fleeting: 'bg-amber-900/50 text-amber-300',
    sleeping: 'bg-purple-900/50 text-purple-300',
    active: 'bg-emerald-900/50 text-emerald-300',
    linked: 'bg-blue-900/50 text-blue-300',
    archived: 'bg-zinc-800 text-zinc-400',
  }
  return (
    <span className={cn('rounded px-2 py-0.5 text-xs font-medium', styles[status] || styles.archived)}>
      {status}
    </span>
  )
}

export function DashboardPage() {
  const [data, setData] = useState<DashboardStats | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    getDashboardStats()
      .then(setData)
      .catch((err) => setError(err.message))
  }, [])

  if (error) {
    return <div className="text-red-400">Failed to load dashboard: {error}</div>
  }

  if (!data) {
    return <div className="text-zinc-400">Loading...</div>
  }

  return (
    <div className="mx-auto max-w-4xl space-y-8">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      <div className="grid grid-cols-3 gap-4">
        {statCards.map((card) => (
          <Link
            key={card.key}
            to={card.href}
            className="group flex items-center gap-4 rounded-lg border border-zinc-800 bg-zinc-900 p-5 transition-colors hover:border-zinc-700"
          >
            <card.icon className={cn('h-8 w-8', card.color)} />
            <div>
              <div className="text-3xl font-bold">{data.stats[card.key]}</div>
              <div className="text-sm text-zinc-400">{card.label}</div>
            </div>
            <ArrowRight className="ml-auto h-4 w-4 text-zinc-600 transition-colors group-hover:text-zinc-400" />
          </Link>
        ))}
      </div>

      <div>
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-lg font-semibold">Recent Notes</h2>
          <Link to="/codex" className="text-sm text-zinc-400 hover:text-white">
            View all
          </Link>
        </div>

        {data.recent.length === 0 ? (
          <p className="text-zinc-500">
            No notes yet. <Link to="/chat" className="text-white underline">Start a conversation</Link> to create your first note.
          </p>
        ) : (
          <div className="space-y-2">
            {data.recent.map((note: Note) => (
              <Link
                key={note.id}
                to={`/codex/${note.id}`}
                className="flex items-center justify-between rounded-lg border border-zinc-800 bg-zinc-900 px-4 py-3 transition-colors hover:border-zinc-700"
              >
                <div className="min-w-0 flex-1">
                  <div className="truncate font-medium">{note.title}</div>
                  {note.summary && (
                    <div className="mt-0.5 truncate text-sm text-zinc-400">{note.summary}</div>
                  )}
                </div>
                <div className="ml-4 flex items-center gap-3">
                  {statusBadge(note.status)}
                  <span className="flex items-center gap-1 text-xs text-zinc-500">
                    <Clock className="h-3 w-3" />
                    {timeAgo(note.updated_at)}
                  </span>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
