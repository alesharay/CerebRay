import { useEffect, useState } from 'react'
import { useAuthStore } from '../store/authStore'
import { getUsage } from '../api/settings'
import type { AIUsage } from '../types'
import { Settings, User, Cpu, LogOut } from 'lucide-react'

const MONTHLY_BUDGET = 2_000_000

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(0)}K`
  return String(n)
}

export function SettingsPage() {
  const { user, logout } = useAuthStore()
  const [usage, setUsage] = useState<AIUsage | null>(null)

  useEffect(() => {
    getUsage().then(setUsage).catch(() => {})
  }, [])

  const totalTokens = usage ? usage.total_input_tokens + usage.total_output_tokens : 0
  const pct = Math.min(100, Math.round((totalTokens / MONTHLY_BUDGET) * 100))

  return (
    <div className="mx-auto max-w-2xl space-y-8">
      <div className="flex items-center gap-3">
        <Settings className="h-6 w-6 text-zinc-400" />
        <h1 className="text-2xl font-bold">Settings</h1>
      </div>

      {/* Profile */}
      <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-5">
        <h2 className="mb-4 flex items-center gap-2 font-semibold">
          <User className="h-4 w-4 text-zinc-400" />
          Profile
        </h2>
        {user ? (
          <div className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-zinc-400">Name</span>
              <span>{user.name}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-400">Email</span>
              <span>{user.email}</span>
            </div>
            <div className="flex justify-between">
              <span className="text-zinc-400">Member since</span>
              <span>{new Date(user.created_at).toLocaleDateString()}</span>
            </div>
          </div>
        ) : (
          <p className="text-sm text-zinc-500">Not signed in.</p>
        )}
      </div>

      {/* AI Usage */}
      <div className="rounded-lg border border-zinc-800 bg-zinc-900 p-5">
        <h2 className="mb-4 flex items-center gap-2 font-semibold">
          <Cpu className="h-4 w-4 text-zinc-400" />
          AI Usage (this month)
        </h2>
        {usage ? (
          <div className="space-y-4">
            <div>
              <div className="mb-1 flex justify-between text-sm">
                <span className="text-zinc-400">Tokens used</span>
                <span>{formatTokens(totalTokens)} / {formatTokens(MONTHLY_BUDGET)}</span>
              </div>
              <div className="h-2 rounded-full bg-zinc-800">
                <div
                  className="h-2 rounded-full bg-emerald-500 transition-all"
                  style={{ width: `${pct}%` }}
                />
              </div>
            </div>
            <div className="grid grid-cols-3 gap-4 text-center text-sm">
              <div>
                <div className="font-medium">{formatTokens(usage.total_input_tokens)}</div>
                <div className="text-zinc-500">Input</div>
              </div>
              <div>
                <div className="font-medium">{formatTokens(usage.total_output_tokens)}</div>
                <div className="text-zinc-500">Output</div>
              </div>
              <div>
                <div className="font-medium">{usage.total_requests}</div>
                <div className="text-zinc-500">Requests</div>
              </div>
            </div>
          </div>
        ) : (
          <p className="text-sm text-zinc-500">Loading usage data...</p>
        )}
      </div>

      {/* Logout */}
      <button
        onClick={logout}
        className="flex items-center gap-2 rounded-lg border border-zinc-800 px-4 py-2.5 text-sm text-zinc-400 transition-colors hover:border-red-800/50 hover:text-red-400"
      >
        <LogOut className="h-4 w-4" />
        Sign out
      </button>
    </div>
  )
}
