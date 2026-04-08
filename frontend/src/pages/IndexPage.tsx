import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { getGraphData } from '../api/connections'
import type { GraphData } from '../types'
import { Network } from 'lucide-react'

export function IndexPage() {
  const [data, setData] = useState<GraphData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getGraphData()
      .then(setData)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="text-zinc-400">Loading...</div>

  const nodes = data?.nodes || []
  const edges = data?.edges || []

  // Build adjacency for display
  const adjacency: Record<number, { id: number; title: string; label: string }[]> = {}
  for (const edge of edges) {
    if (!adjacency[edge.source]) adjacency[edge.source] = []
    if (!adjacency[edge.target]) adjacency[edge.target] = []
    const sourceNode = nodes.find((n) => n.id === edge.source)
    const targetNode = nodes.find((n) => n.id === edge.target)
    if (targetNode) adjacency[edge.source].push({ id: targetNode.id, title: targetNode.title, label: edge.label })
    if (sourceNode) adjacency[edge.target].push({ id: sourceNode.id, title: sourceNode.title, label: edge.label })
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div className="flex items-center gap-3">
        <Network className="h-6 w-6 text-teal-400" />
        <h1 className="text-2xl font-bold">Index</h1>
        <span className="text-sm text-zinc-500">{nodes.length} notes, {edges.length} connections</span>
      </div>

      {nodes.length === 0 ? (
        <p className="text-zinc-500">No notes with connections yet. Link notes together from the note detail page to see them here.</p>
      ) : (
        <div className="space-y-3">
          {nodes.map((node) => {
            const linked = adjacency[node.id] || []
            return (
              <div key={node.id} className="rounded-lg border border-zinc-800 bg-zinc-900 p-4">
                <Link to={`/codex/${node.id}`} className="font-medium text-zinc-100 hover:underline">
                  {node.title}
                </Link>
                <span className="ml-2 text-xs text-zinc-500">{node.type}</span>
                {linked.length > 0 && (
                  <div className="mt-2 flex flex-wrap gap-2">
                    {linked.map((l, i) => (
                      <Link
                        key={`${l.id}-${i}`}
                        to={`/codex/${l.id}`}
                        className="flex items-center gap-1 rounded bg-zinc-800 px-2 py-1 text-xs text-zinc-400 transition-colors hover:text-zinc-200"
                      >
                        {l.title}
                        {l.label && <span className="text-zinc-600">({l.label})</span>}
                      </Link>
                    ))}
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
