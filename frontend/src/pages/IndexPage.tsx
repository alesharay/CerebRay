import { useEffect, useRef, useState, useCallback } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { getGraphData } from '../api/connections'
import type { GraphData, NoteType } from '../types'
import { Network, List, GitBranch, Search } from 'lucide-react'
import { cn } from '../lib/utils'
import {
  forceSimulation, forceLink, forceManyBody, forceCenter, forceCollide,
  type SimulationNodeDatum, type SimulationLinkDatum,
} from 'd3-force'
import { select } from 'd3-selection'
import { zoom as d3Zoom, type ZoomBehavior } from 'd3-zoom'
import { drag as d3Drag } from 'd3-drag'

interface GraphNode extends SimulationNodeDatum {
  id: number
  title: string
  type: string
}

interface GraphLink extends SimulationLinkDatum<GraphNode> {
  label: string
}

const typeColors: Record<string, string> = {
  concept: '#38bdf8',
  theory: '#a78bfa',
  insight: '#fbbf24',
  quote: '#f472b6',
  reference: '#a1a1aa',
  question: '#fb923c',
  structure: '#2dd4bf',
  guide: '#34d399',
}

function ForceGraph({
  data,
  searchQuery,
  typeFilter,
}: {
  data: GraphData
  searchQuery: string
  typeFilter: NoteType | ''
}) {
  const svgRef = useRef<SVGSVGElement>(null)
  const wrapperRef = useRef<HTMLDivElement>(null)
  const navigate = useNavigate()
  const [tooltip, setTooltip] = useState<{
    x: number
    y: number
    title: string
    type: string
    connections: number
  } | null>(null)

  // Build connection count per node
  const connectionCount = useCallback(() => {
    const counts: Record<number, number> = {}
    for (const e of data.edges) {
      counts[e.source] = (counts[e.source] || 0) + 1
      counts[e.target] = (counts[e.target] || 0) + 1
    }
    return counts
  }, [data.edges])

  useEffect(() => {
    if (!svgRef.current || data.nodes.length === 0) return

    const width = svgRef.current.clientWidth
    const height = svgRef.current.clientHeight
    const counts = connectionCount()

    const nodes: GraphNode[] = data.nodes.map((n) => ({ ...n }))
    const links: GraphLink[] = data.edges.map((e) => ({
      source: e.source,
      target: e.target,
      label: e.label,
    }))

    // Build adjacency set for hover highlighting
    const adjacency = new Set<string>()
    for (const e of data.edges) {
      adjacency.add(`${e.source}-${e.target}`)
      adjacency.add(`${e.target}-${e.source}`)
    }

    const svg = select(svgRef.current)
    svg.selectAll('*').remove()

    const g = svg.append('g')

    // Links
    const link = g
      .selectAll<SVGLineElement, GraphLink>('line')
      .data(links)
      .join('line')
      .attr('stroke', '#3f3f46')
      .attr('stroke-width', 1.5)

    // Nodes
    const node = g
      .selectAll<SVGCircleElement, GraphNode>('circle')
      .data(nodes)
      .join('circle')
      .attr('r', 8)
      .attr('fill', (d) => typeColors[d.type] || '#71717a')
      .attr('stroke', '#18181b')
      .attr('stroke-width', 2)
      .attr('cursor', 'pointer')

    // Labels
    const label = g
      .selectAll<SVGTextElement, GraphNode>('text')
      .data(nodes)
      .join('text')
      .text((d) => (d.title.length > 20 ? d.title.slice(0, 20) + '...' : d.title))
      .attr('font-size', 11)
      .attr('fill', '#a1a1aa')
      .attr('dx', 12)
      .attr('dy', 4)
      .attr('pointer-events', 'none')

    // Simulation
    const simulation = forceSimulation(nodes)
      .force(
        'link',
        forceLink<GraphNode, GraphLink>(links)
          .id((d) => d.id)
          .distance(100),
      )
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

    // Zoom and pan
    const zoomBehavior: ZoomBehavior<SVGSVGElement, unknown> = d3Zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.3, 4])
      .on('zoom', (event) => {
        g.attr('transform', event.transform)
      })
    svg.call(zoomBehavior)

    // Drag behavior
    const dragBehavior = d3Drag<SVGCircleElement, GraphNode>()
      .on('start', (event, d) => {
        if (!event.active) simulation.alphaTarget(0.3).restart()
        d.fx = d.x
        d.fy = d.y
      })
      .on('drag', (event, d) => {
        d.fx = event.x
        d.fy = event.y
      })
      .on('end', (event, d) => {
        if (!event.active) simulation.alphaTarget(0)
        d.fx = null
        d.fy = null
      })
    node.call(dragBehavior)

    // Hover highlighting
    node.on('mouseover', (event, d) => {
      // Dim unconnected elements
      node.attr('opacity', (o) =>
        o.id === d.id || adjacency.has(`${d.id}-${o.id}`) ? 1 : 0.15,
      )
      link.attr('opacity', (l) => {
        const src = (l.source as GraphNode).id
        const tgt = (l.target as GraphNode).id
        return src === d.id || tgt === d.id ? 1 : 0.08
      })
      label.attr('opacity', (o) =>
        o.id === d.id || adjacency.has(`${d.id}-${o.id}`) ? 1 : 0.15,
      )

      // Show tooltip
      const svgRect = svgRef.current!.getBoundingClientRect()
      const wrapperRect = wrapperRef.current!.getBoundingClientRect()
      // Get the current transform from the zoom
      const transform = svg.node()!.__zoom || { x: 0, y: 0, k: 1 }
      const screenX = (d.x || 0) * transform.k + transform.x
      const screenY = (d.y || 0) * transform.k + transform.y
      setTooltip({
        x: svgRect.left - wrapperRect.left + screenX,
        y: svgRect.top - wrapperRect.top + screenY,
        title: d.title,
        type: d.type,
        connections: counts[d.id] || 0,
      })
    })

    node.on('mouseout', () => {
      node.attr('opacity', 1)
      link.attr('opacity', 1)
      label.attr('opacity', 1)
      setTooltip(null)
    })

    // Click to navigate (SPA)
    node.on('click', (_event, d) => {
      navigate(`/codex/${d.id}`)
    })

    return () => {
      simulation.stop()
    }
  }, [data, navigate, connectionCount])

  // Apply search/filter highlighting as a separate effect
  useEffect(() => {
    if (!svgRef.current) return
    const svg = select(svgRef.current)
    const q = searchQuery.toLowerCase()
    const hasFilter = typeFilter !== '' || q !== ''

    svg.selectAll<SVGCircleElement, GraphNode>('circle').attr('opacity', (d) => {
      if (!hasFilter) return 1
      const matchesSearch = !q || d.title.toLowerCase().includes(q)
      const matchesType = !typeFilter || d.type === typeFilter
      return matchesSearch && matchesType ? 1 : 0.12
    })

    svg.selectAll<SVGTextElement, GraphNode>('text').attr('opacity', (d) => {
      if (!hasFilter) return 1
      const matchesSearch = !q || d.title.toLowerCase().includes(q)
      const matchesType = !typeFilter || d.type === typeFilter
      return matchesSearch && matchesType ? 1 : 0.12
    })

    svg.selectAll<SVGLineElement, GraphLink>('line').attr('opacity', (l) => {
      if (!hasFilter) return 1
      const src = l.source as GraphNode
      const tgt = l.target as GraphNode
      const srcMatch =
        (!q || src.title.toLowerCase().includes(q)) &&
        (!typeFilter || src.type === typeFilter)
      const tgtMatch =
        (!q || tgt.title.toLowerCase().includes(q)) &&
        (!typeFilter || tgt.type === typeFilter)
      return srcMatch || tgtMatch ? 0.6 : 0.06
    })
  }, [searchQuery, typeFilter])

  return (
    <div ref={wrapperRef} className="relative">
      <svg
        ref={svgRef}
        className="h-[600px] w-full rounded-lg border border-zinc-800 bg-zinc-950"
      />
      {tooltip && (
        <div
          className="pointer-events-none absolute z-10 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 shadow-lg"
          style={{ left: tooltip.x + 16, top: tooltip.y - 12 }}
        >
          <div className="text-sm font-medium text-zinc-100">{tooltip.title}</div>
          <div className="mt-0.5 flex items-center gap-2 text-xs">
            <span style={{ color: typeColors[tooltip.type] || '#71717a' }}>
              {tooltip.type}
            </span>
            <span className="text-zinc-500">
              {tooltip.connections} connection{tooltip.connections !== 1 ? 's' : ''}
            </span>
          </div>
        </div>
      )}
    </div>
  )
}

function ColorLegend({ types }: { types: string[] }) {
  return (
    <div className="flex flex-wrap gap-3 px-1">
      {types.map((t) => (
        <div key={t} className="flex items-center gap-1.5 text-xs text-zinc-400">
          <span
            className="inline-block h-2.5 w-2.5 rounded-full"
            style={{ backgroundColor: typeColors[t] || '#71717a' }}
          />
          {t}
        </div>
      ))}
    </div>
  )
}

export function IndexPage() {
  const [data, setData] = useState<GraphData | null>(null)
  const [loading, setLoading] = useState(true)
  const [view, setView] = useState<'graph' | 'list'>('graph')
  const [searchQuery, setSearchQuery] = useState('')
  const [typeFilter, setTypeFilter] = useState<NoteType | ''>('')

  useEffect(() => {
    getGraphData()
      .then(setData)
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="text-zinc-400">Loading...</div>

  const nodes = data?.nodes || []
  const edges = data?.edges || []

  // Unique types present in graph data
  const presentTypes = [...new Set(nodes.map((n) => n.type))].sort()

  // Build adjacency for list view
  const adjacency: Record<number, { id: number; title: string; label: string }[]> = {}
  for (const edge of edges) {
    if (!adjacency[edge.source]) adjacency[edge.source] = []
    if (!adjacency[edge.target]) adjacency[edge.target] = []
    const sourceNode = nodes.find((n) => n.id === edge.source)
    const targetNode = nodes.find((n) => n.id === edge.target)
    if (targetNode)
      adjacency[edge.source].push({
        id: targetNode.id,
        title: targetNode.title,
        label: edge.label,
      })
    if (sourceNode)
      adjacency[edge.target].push({
        id: sourceNode.id,
        title: sourceNode.title,
        label: edge.label,
      })
  }

  return (
    <div className="mx-auto max-w-4xl space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Network className="h-6 w-6 text-teal-400" />
          <h1 className="text-2xl font-bold">Index</h1>
          <span className="text-sm text-zinc-500">
            {nodes.length} notes, {edges.length} connections
          </span>
        </div>
        <div className="flex gap-1 rounded-lg border border-zinc-800 p-0.5">
          <button
            onClick={() => setView('graph')}
            className={cn(
              'rounded-md px-3 py-1 text-xs transition-colors',
              view === 'graph'
                ? 'bg-zinc-800 text-white'
                : 'text-zinc-500 hover:text-zinc-300',
            )}
          >
            <GitBranch className="mr-1 inline h-3 w-3" />
            Graph
          </button>
          <button
            onClick={() => setView('list')}
            className={cn(
              'rounded-md px-3 py-1 text-xs transition-colors',
              view === 'list'
                ? 'bg-zinc-800 text-white'
                : 'text-zinc-500 hover:text-zinc-300',
            )}
          >
            <List className="mr-1 inline h-3 w-3" />
            List
          </button>
        </div>
      </div>

      {/* Search and filter bar (graph view only) */}
      {view === 'graph' && nodes.length > 0 && (
        <div className="flex items-center gap-3">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-zinc-500" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search nodes..."
              className="w-full rounded-lg border border-zinc-800 bg-zinc-900/50 py-2 pl-9 pr-3 text-sm text-zinc-100 placeholder-zinc-600 outline-none transition-colors focus:border-zinc-600"
            />
          </div>
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value as NoteType | '')}
            className="rounded-lg border border-zinc-800 bg-zinc-900/50 px-3 py-2 text-sm text-zinc-300 outline-none focus:border-zinc-600"
          >
            <option value="">All types</option>
            {presentTypes.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
        </div>
      )}

      {nodes.length === 0 ? (
        <p className="text-zinc-500">
          No notes with connections yet. Link notes together from the note detail
          page to see them here.
        </p>
      ) : view === 'graph' && data ? (
        <div className="space-y-3">
          <ForceGraph
            data={data}
            searchQuery={searchQuery}
            typeFilter={typeFilter}
          />
          <ColorLegend types={presentTypes} />
        </div>
      ) : (
        <div className="space-y-3">
          {nodes.map((node) => {
            const linked = adjacency[node.id] || []
            return (
              <div
                key={node.id}
                className="rounded-lg border border-zinc-800 bg-zinc-900 p-4"
              >
                <Link
                  to={`/codex/${node.id}`}
                  className="font-medium text-zinc-100 hover:underline"
                >
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
                        {l.label && (
                          <span className="text-zinc-600">({l.label})</span>
                        )}
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
