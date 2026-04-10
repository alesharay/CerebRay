import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { getAnalytics } from '../api/dashboard'
import type { AnalyticsData, InboxItem, LifecycleEntry, LifecycleTrendPoint, StaleNote } from '../types'
import {
  Inbox, Brain, MessageSquare,
  Zap, Clock, TrendingUp, AlertCircle, Sparkles,
} from 'lucide-react'
import { cn } from '../lib/utils'

function formatAge(seconds: number): string {
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`
  const days = Math.floor(seconds / 86400)
  return days === 1 ? '1 day' : `${days} days`
}

function formatDwell(seconds: number): string {
  if (seconds < 3600) return `${Math.floor(seconds / 60)} min`
  if (seconds < 86400) return `${Math.round(seconds / 3600)} hr`
  return `${Math.round(seconds / 86400)} days`
}

function ReadinessBar({ score }: { score: number }) {
  const pct = Math.round(score * 100)
  const color = pct >= 70 ? 'bg-emerald-400' : pct >= 40 ? 'bg-amber-400' : 'bg-rose-400'
  return (
    <div className="flex items-center gap-2">
      <div className="h-1.5 w-16 rounded-full bg-zinc-800">
        <div className={cn('h-full rounded-full transition-all', color)} style={{ width: `${pct}%` }} />
      </div>
      <span className="text-xs text-zinc-500">{pct}%</span>
    </div>
  )
}

function StrengthRing({ score }: { score: number }) {
  const clamped = Math.min(100, Math.max(0, Math.round(score)))
  const radius = 54
  const circumference = 2 * Math.PI * radius
  const offset = circumference - (clamped / 100) * circumference
  const color = clamped >= 70 ? '#34d399' : clamped >= 40 ? '#fbbf24' : '#fb7185'

  return (
    <div className="relative inline-flex items-center justify-center">
      <svg width="132" height="132" viewBox="0 0 132 132">
        <circle cx="66" cy="66" r={radius} fill="none" stroke="#27272a" strokeWidth="10" />
        <circle
          cx="66" cy="66" r={radius} fill="none"
          stroke={color} strokeWidth="10" strokeLinecap="round"
          strokeDasharray={circumference} strokeDashoffset={offset}
          transform="rotate(-90 66 66)"
          className="transition-all duration-700"
        />
      </svg>
      <div className="absolute flex flex-col items-center">
        <span className="text-3xl font-bold" style={{ color }}>{clamped}</span>
        <span className="text-xs text-zinc-500">/ 100</span>
      </div>
    </div>
  )
}

function SectionCard({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={cn('rounded-xl border border-zinc-800/60 bg-zinc-900/50 p-5', className)}>
      {children}
    </div>
  )
}

function SectionTitle({ icon: Icon, children }: { icon: React.ComponentType<{ className?: string }>; children: React.ReactNode }) {
  return (
    <h2 className="mb-4 flex items-center gap-2 text-sm font-semibold uppercase tracking-wide text-zinc-400">
      <Icon className="h-4 w-4" />
      {children}
    </h2>
  )
}

function InboxSection({ items }: { items: InboxItem[] }) {
  if (items.length === 0) {
    return (
      <SectionCard>
        <SectionTitle icon={Inbox}>Inbox</SectionTitle>
        <p className="text-sm text-zinc-500">
          Nothing in the inbox. <Link to="/chat" className="text-amber-400 hover:underline">Start a conversation</Link> to capture new ideas.
        </p>
      </SectionCard>
    )
  }

  return (
    <SectionCard>
      <SectionTitle icon={Inbox}>Inbox ({items.length})</SectionTitle>
      <div className="space-y-3">
        {items.map((item) => (
          <Link
            key={item.id}
            to={`/codex/${item.id}`}
            className="flex items-center justify-between rounded-lg px-3 py-2.5 transition-colors hover:bg-zinc-800/50"
          >
            <div className="min-w-0 flex-1">
              <div className="truncate text-sm font-medium text-zinc-200">{item.title}</div>
              <div className="mt-0.5 flex items-center gap-3 text-xs text-zinc-500">
                <span className="rounded bg-zinc-800 px-1.5 py-0.5">{item.note_type}</span>
                <span className="flex items-center gap-1">
                  <Clock className="h-3 w-3" />
                  {formatAge(item.age_seconds)}
                </span>
              </div>
            </div>
            <ReadinessBar score={item.readiness_score} />
          </Link>
        ))}
      </div>
    </SectionCard>
  )
}

function TrendSparkline({ points, color }: { points: number[]; color: string }) {
  if (points.length < 2) return null
  const max = Math.max(...points, 1)
  const w = 64
  const h = 20
  const stepX = w / (points.length - 1)
  const path = points
    .map((v, i) => {
      const x = i * stepX
      const y = h - (v / max) * (h - 2) - 1
      return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')

  return (
    <svg width={w} height={h} className="inline-block align-middle">
      <path d={path} fill="none" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  )
}

function buildSparklineData(trend: LifecycleTrendPoint[], action: string): number[] {
  // Get all unique weeks across all actions for consistent x-axis
  const weeks = [...new Set(trend.map((t) => t.week))].sort()
  const countByWeek: Record<string, number> = {}
  for (const t of trend) {
    if (t.action === action) countByWeek[t.week] = t.count
  }
  return weeks.map((w) => countByWeek[w] || 0)
}

function LifecycleSection({ entries, trend }: { entries: LifecycleEntry[]; trend: LifecycleTrendPoint[] }) {
  const total = entries.reduce((sum, e) => sum + e.count, 0)
  if (total === 0) {
    return (
      <SectionCard>
        <SectionTitle icon={TrendingUp}>Triage Activity</SectionTitle>
        <p className="text-sm text-zinc-500">No triage activity yet. Promote, sleep, or archive notes from the inbox to see patterns here.</p>
      </SectionCard>
    )
  }

  const actionConfig: Record<string, { label: string; color: string; bg: string; hex: string }> = {
    promoted: { label: 'Promoted', color: 'text-emerald-400', bg: 'bg-emerald-400', hex: '#34d399' },
    slept: { label: 'Slept', color: 'text-purple-400', bg: 'bg-purple-400', hex: '#a78bfa' },
    archived: { label: 'Archived', color: 'text-zinc-400', bg: 'bg-zinc-500', hex: '#71717a' },
  }

  return (
    <SectionCard>
      <SectionTitle icon={TrendingUp}>Triage Activity (90 days)</SectionTitle>
      <div className="mb-3 flex h-3 overflow-hidden rounded-full bg-zinc-800">
        {entries.map((e) => {
          const config = actionConfig[e.action] || actionConfig.archived
          const pct = (e.count / total) * 100
          return (
            <div
              key={e.action}
              className={cn('h-full transition-all', config.bg)}
              style={{ width: `${pct}%` }}
              title={`${config.label}: ${e.count}`}
            />
          )
        })}
      </div>
      <div className="grid grid-cols-3 gap-3">
        {entries.map((e) => {
          const config = actionConfig[e.action] || actionConfig.archived
          const sparkData = buildSparklineData(trend, e.action)
          return (
            <div key={e.action} className="text-center">
              <div className={cn('text-xl font-bold', config.color)}>{e.count}</div>
              <div className="text-xs text-zinc-500">{config.label}</div>
              <div className="mt-0.5 text-xs text-zinc-600">avg {formatDwell(e.avg_dwell_seconds)}</div>
              {sparkData.length >= 2 && (
                <div className="mt-1.5 flex justify-center">
                  <TrendSparkline points={sparkData} color={config.hex} />
                </div>
              )}
            </div>
          )
        })}
      </div>
    </SectionCard>
  )
}

function StrengthSection({ data }: { data: AnalyticsData['strength'] }) {
  return (
    <SectionCard>
      <SectionTitle icon={Brain}>Zettelkasten Strength</SectionTitle>
      <div className="flex items-start gap-6">
        <StrengthRing score={data.overall} />
        <div className="flex-1 space-y-3 pt-1">
          <Metric label="Active notes" value={data.active_notes} />
          <Metric label="Connections" value={data.total_connections} detail={
            data.active_notes > 0 ? `${(data.connection_density).toFixed(1)} per note` : undefined
          } />
          <Metric label="Orphans" value={data.orphan_count} warn={data.orphan_count > 0} />
          <Metric label="Glossary terms" value={data.glossary_count} detail={
            data.active_notes > 0 ? `${Math.round(data.glossary_coverage * 100)}% coverage` : undefined
          } />
          {data.sleeping_backlog > 0 && (
            <Metric label="Sleeping backlog" value={data.sleeping_backlog} detail={
              `avg ${formatAge(data.sleeping_avg_age_seconds)} old`
            } />
          )}
        </div>
      </div>
      {data.type_distribution.length > 0 && (
        <div className="mt-4 flex flex-wrap gap-2">
          {data.type_distribution.map((t) => (
            <span key={t.type} className="rounded-full bg-zinc-800 px-2.5 py-1 text-xs text-zinc-300">
              {t.type} <span className="text-zinc-500">{t.count}</span>
            </span>
          ))}
        </div>
      )}
    </SectionCard>
  )
}

function Metric({ label, value, detail, warn }: { label: string; value: number; detail?: string; warn?: boolean }) {
  return (
    <div className="flex items-baseline justify-between text-sm">
      <span className="text-zinc-400">{label}</span>
      <span className="flex items-center gap-1.5">
        {warn && <AlertCircle className="h-3 w-3 text-amber-400" />}
        <span className={cn('font-semibold', warn ? 'text-amber-400' : 'text-zinc-200')}>{value}</span>
        {detail && <span className="text-xs text-zinc-600">{detail}</span>}
      </span>
    </div>
  )
}

function ConversionWidget({ data }: { data: AnalyticsData['conversion'] }) {
  if (data.total_conversations === 0) return null
  const pct = Math.round(data.rate * 100)
  return (
    <SectionCard className="flex items-center gap-4">
      <MessageSquare className="h-8 w-8 text-blue-400" />
      <div className="flex-1">
        <div className="text-sm text-zinc-400">Conversation yield</div>
        <div className="text-lg font-bold text-zinc-200">{pct}%</div>
      </div>
      <div className="text-right text-xs text-zinc-500">
        {data.conversations_with_notes} of {data.total_conversations} chats produced notes
      </div>
    </SectionCard>
  )
}

function AIBudgetWidget({ data }: { data: AnalyticsData['ai_usage'] }) {
  const totalTokens = data.input_tokens + data.output_tokens
  if (totalTokens === 0 && data.requests === 0) return null
  return (
    <SectionCard className="flex items-center gap-4">
      <Sparkles className="h-8 w-8 text-violet-400" />
      <div className="flex-1">
        <div className="text-sm text-zinc-400">AI usage this month</div>
        <div className="text-lg font-bold text-zinc-200">{totalTokens.toLocaleString()} tokens</div>
      </div>
      <div className="text-right text-xs text-zinc-500">
        {data.requests} requests
      </div>
    </SectionCard>
  )
}

function StaleSection({ notes }: { notes: StaleNote[] }) {
  if (notes.length === 0) return null
  return (
    <SectionCard>
      <SectionTitle icon={Zap}>Needs Attention</SectionTitle>
      <p className="mb-3 text-xs text-zinc-500">Notes that haven't been updated in a while. Connect them to something or archive them.</p>
      <div className="space-y-2">
        {notes.map((n) => (
          <Link
            key={n.id}
            to={`/codex/${n.id}`}
            className="flex items-center justify-between rounded-lg px-3 py-2 text-sm transition-colors hover:bg-zinc-800/50"
          >
            <span className="truncate text-zinc-300">{n.title}</span>
            <span className="ml-3 whitespace-nowrap text-xs text-zinc-600">{n.stale_days}d stale</span>
          </Link>
        ))}
      </div>
    </SectionCard>
  )
}

export function DashboardPage() {
  const [data, setData] = useState<AnalyticsData | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    getAnalytics()
      .then(setData)
      .catch((err) => setError(err.message))
  }, [])

  if (error) {
    return <div className="text-rose-400">Failed to load dashboard: {error}</div>
  }

  if (!data) {
    return <div className="text-zinc-500">Loading...</div>
  }

  return (
    <div className="mx-auto max-w-4xl space-y-6 pb-12">
      <div>
        <h1 className="text-2xl font-bold text-zinc-100">Dashboard</h1>
        <p className="mt-1 text-sm text-zinc-500">Your knowledge at a glance</p>
      </div>

      <InboxSection items={data.inbox} />

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <LifecycleSection entries={data.lifecycle} trend={data.lifecycle_trend || []} />
        <StrengthSection data={data.strength} />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <ConversionWidget data={data.conversion} />
        <AIBudgetWidget data={data.ai_usage} />
      </div>

      <StaleSection notes={data.stale_notes} />
    </div>
  )
}
