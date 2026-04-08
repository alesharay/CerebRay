import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import { getGraphData } from '../api/connections'
import type { GraphData } from '../types'
import { Network, List, GitBranch } from 'lucide-react'
import { cn } from '../lib/utils'
import { forceSimulation, forceLink, forceManyBody, forceCenter, forceCollide, type SimulationNodeDatum, type SimulationLinkDatum } from 'd3-force'
import { select } from 'd3-selection'

interface GraphNode extends SimulationNodeDatum {
  id: number
  title: string
  type: string
}

interface GraphLink extends SimulationLinkDatum<GraphNode> {
  label: string
}

function ForceGraph({ data }: { data: GraphData }) {
  const svgRef = useRef<SVGSVGElement>(null)

  useEffect(() => {
    if (!svgRef.current || data.nodes.length === 0) return

    const width = svgRef.current.clientWidth
    const height = svgRef.current.clientHeight

    const nodes: GraphNode[] = data.nodes.map((n) => ({ ...n }))
    const links: GraphLink[] = data.edges.map((e) => ({
      source: e.source,
      target: e.target,
      label: e.label,
    }))

    const svg = select(svgRef.current)
    svg.selectAll('*').remove()

    const g = svg.append('g')

    // Zoom
    svg.call(
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (select(svgRef.current) as any).call.bind(
        svg,
        // Simple pan via drag
      )
    )

    const link = g
      .selectAll('line')
      .data(links)
      .join('line')
      .attr('stroke', '#3f3f46')
      .attr('stroke-width', 1.5)

    const node = g
      .selectAll('circle')
      .data(nodes)
      .join('circle')
      .attr('r', 8)
      .attr('fill', (d) => {
        const colors: Record<string, string> = {
          concept: '#38bdf8',
          theory: '#a78bfa',
          insight: '#fbbf24',
          quote: '#f472b6',
          reference: '#a1a1aa',
          question: '#fb923c',
          structure: '#2dd4bf',
          guide: '#34d399',
        }
        return colors[d.type] || '#71717a'
      })
      .attr('stroke', '#18181b')
      .attr('stroke-width', 2)
      .attr('cursor', 'pointer')

    const label = g
      .selectAll('text')
      .data(nodes)
      .join('text')
      .text((d) => d.title.length > 20 ? d.title.slice(0, 20) + '...' : d.title)
      .attr('font-size', 11)
      .attr('fill', '#a1a1aa')
      .attr('dx', 12)
      .attr('dy', 4)

    const simulation = forceSimulation(nodes)
      .force('link', forceLink<GraphNode, GraphLink>(links).id((d) => d.id).distance(100))
      .force('charge', forceManyBody().strength(-200))
      .force('center', forceCenter(width / 2, height / 2))
      .force('collide', forceCollide(30))

    simulation.on('tick', () => {
      link
        .attr('x1', (d) => (d.source as GraphNode).x || 0)
        .attr('y1', (d) => (d.source as GraphNode).y || 0)
        .attr('x2', (d) => (d.target as GraphNode).x || 0)
        .attr('y2', (d) => (d.target as GraphNode).y || 0)

      node.attr('cx', (d) => d.x || 0).attr('cy', (d) => d.y || 0)
      label.attr('x', (d) => d.x || 0).attr('y', (d) => d.y || 0)
    })

    // Click to navigate
    node.on('click', (_event, d) => {
      window.location.hash = ''
      window.location.pathname = `/codex/${d.id}`
    })

    // Tooltip on hover
    node.append('title').text((d) => `${d.title} (${d.type})`)

    return () => { simulation.stop() }
  }, [data])

  return (
    <svg
      ref={svgRef}
      className="h-[500px] w-full rounded-lg border border-zinc-800 bg-zinc-950"
    />
  )
}

export function IndexPage() {
  const [data, setData] = useState<GraphData | null>(null)
  const [loading, setLoading] = useState(true)
  const [view, setView] = useState<'graph' | 'list'>('graph')

  useEffect(() => {
    getGraphData()
      .then(setData)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="text-zinc-400">Loading...</div>

  const nodes = data?.nodes || []
  const edges = data?.edges || []

  // Build adjacency for list view
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
    <div className="mx-auto max-w-4xl space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Network className="h-6 w-6 text-teal-400" />
          <h1 className="text-2xl font-bold">Index</h1>
          <span className="text-sm text-zinc-500">{nodes.length} notes, {edges.length} connections</span>
        </div>
        <div className="flex gap-1 rounded-lg border border-zinc-800 p-0.5">
          <button
            onClick={() => setView('graph')}
            className={cn(
              'rounded-md px-3 py-1 text-xs transition-colors',
              view === 'graph' ? 'bg-zinc-800 text-white' : 'text-zinc-500 hover:text-zinc-300'
            )}
          >
            <GitBranch className="mr-1 inline h-3 w-3" />
            Graph
          </button>
          <button
            onClick={() => setView('list')}
            className={cn(
              'rounded-md px-3 py-1 text-xs transition-colors',
              view === 'list' ? 'bg-zinc-800 text-white' : 'text-zinc-500 hover:text-zinc-300'
            )}
          >
            <List className="mr-1 inline h-3 w-3" />
            List
          </button>
        </div>
      </div>

      {nodes.length === 0 ? (
        <p className="text-zinc-500">No notes with connections yet. Link notes together from the note detail page to see them here.</p>
      ) : view === 'graph' && data ? (
        <ForceGraph data={data} />
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
