import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { DashboardPage } from './DashboardPage'
import type { AnalyticsData } from '../types'

vi.mock('../api/dashboard', () => ({
  getAnalytics: vi.fn(),
}))

import { getAnalytics } from '../api/dashboard'

const mockedGetAnalytics = vi.mocked(getAnalytics)

function makeAnalytics(overrides: Partial<AnalyticsData> = {}): AnalyticsData {
  return {
    inbox: [
      { id: 1, title: 'Inbox Note', note_type: 'concept', age_seconds: 3600, readiness_score: 0.75, created_at: new Date().toISOString() },
    ],
    lifecycle: [
      { action: 'promoted', count: 5, avg_dwell_seconds: 7200 },
      { action: 'slept', count: 3, avg_dwell_seconds: 3600 },
      { action: 'archived', count: 2, avg_dwell_seconds: 1800 },
    ],
    lifecycle_trend: [],
    strength: {
      overall: 72,
      connection_density: 1.5,
      orphan_count: 2,
      active_notes: 10,
      total_connections: 15,
      glossary_count: 5,
      glossary_coverage: 0.5,
      sleeping_backlog: 3,
      sleeping_avg_age_seconds: 86400,
      type_distribution: [{ type: 'concept', count: 7 }, { type: 'theory', count: 3 }],
    },
    conversion: {
      total_conversations: 20,
      conversations_with_notes: 12,
      rate: 0.6,
    },
    ai_usage: {
      input_tokens: 5000,
      output_tokens: 3000,
      requests: 15,
    },
    stale_notes: [],
    ...overrides,
  }
}

function renderPage() {
  return render(
    <MemoryRouter>
      <DashboardPage />
    </MemoryRouter>
  )
}

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading state initially', () => {
    mockedGetAnalytics.mockReturnValue(new Promise(() => {}))
    renderPage()
    expect(screen.getByText('Loading...')).toBeInTheDocument()
  })

  it('renders dashboard sections after analytics load', async () => {
    mockedGetAnalytics.mockResolvedValueOnce(makeAnalytics())
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
      expect(screen.getByText('Your knowledge at a glance')).toBeInTheDocument()
    })
  })

  it('shows inbox items', async () => {
    mockedGetAnalytics.mockResolvedValueOnce(makeAnalytics())
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Inbox Note')).toBeInTheDocument()
      expect(screen.getByText('75%')).toBeInTheDocument()
    })
  })

  it('shows lifecycle section with activity data', async () => {
    mockedGetAnalytics.mockResolvedValueOnce(makeAnalytics())
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('Promoted')).toBeInTheDocument()
      expect(screen.getByText('Slept')).toBeInTheDocument()
      expect(screen.getByText('Archived')).toBeInTheDocument()
      expect(screen.getByText(/Triage Activity/)).toBeInTheDocument()
    })
  })

  it('shows strength score ring', async () => {
    mockedGetAnalytics.mockResolvedValueOnce(makeAnalytics({ strength: { ...makeAnalytics().strength, overall: 85 } }))
    renderPage()

    await waitFor(() => {
      expect(screen.getByText('85')).toBeInTheDocument()
      expect(screen.getByText('/ 100')).toBeInTheDocument()
    })
  })

  it('shows error message on API failure', async () => {
    mockedGetAnalytics.mockRejectedValueOnce(new Error('Network error'))
    renderPage()

    await waitFor(() => {
      expect(screen.getByText(/Failed to load dashboard/)).toBeInTheDocument()
      expect(screen.getByText(/Network error/)).toBeInTheDocument()
    })
  })
})
