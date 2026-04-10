import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { Sidebar } from './Sidebar'

function renderSidebar(path = '/dashboard') {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Sidebar />
    </MemoryRouter>
  )
}

describe('Sidebar', () => {
  it('renders the app name "Cereb-Ray"', () => {
    renderSidebar()
    expect(screen.getByText('Cereb-Ray')).toBeInTheDocument()
  })

  it('renders all navigation items', () => {
    renderSidebar()
    const labels = ['Dashboard', 'Inbox', 'Echoes', 'Codex', 'Index', 'Glossary', 'Settings']
    for (const label of labels) {
      expect(screen.getByText(label)).toBeInTheDocument()
    }
  })

  it('highlights the active navigation item based on current path', () => {
    renderSidebar('/codex')
    const codexLink = screen.getByText('Codex').closest('a')
    expect(codexLink).toHaveClass('bg-zinc-800', 'text-white')

    const dashboardLink = screen.getByText('Dashboard').closest('a')
    expect(dashboardLink).not.toHaveClass('bg-zinc-800')
    expect(dashboardLink).toHaveClass('text-zinc-400')
  })
})
