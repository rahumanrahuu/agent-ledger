import { useState, useRef, useMemo, useCallback } from 'react';
import { getGraph } from './api/client';
import { useApi } from './hooks/useApi';
import type { GraphNode } from './api/types';
import {
  Badge, SearchInput, ErrorState, EmptyState, PageHeader,
  SectionLabel, MetaRow, InspectorPanel,
} from './components/ui';

const nodeTypeColors: Record<string, string> = {
  session: '#EA580C', decision: '#2563EB', discovery: '#16A34A',
  failure: '#DC2626', constraint: '#D97706', checkpoint: '#7C3AED',
  file: '#64748B', function: '#7C3AED', module: '#2563EB',
  agent: '#EA580C', service: '#16A34A', concept: '#9333EA', entity: '#0891B2',
};

function colorFor(type: string) {
  return nodeTypeColors[type.toLowerCase()] ?? '#71717A';
}

interface SimNode extends GraphNode {
  x: number;
  y: number;
  vx: number;
  vy: number;
}

function runSimulation(nodes: GraphNode[], edges: { source: string; target: string }[]): SimNode[] {
  const result: SimNode[] = nodes.map((n, i) => {
    const angle = (i / nodes.length) * 2 * Math.PI;
    const r = nodes.length < 8 ? 120 : 200;
    return { ...n, x: Math.cos(angle) * r, y: Math.sin(angle) * r, vx: 0, vy: 0 };
  });

  for (let iter = 0; iter < 300; iter++) {
    const alpha = 1 - iter / 300;
    for (let i = 0; i < result.length; i++) {
      for (let j = i + 1; j < result.length; j++) {
        const a = result[i], b = result[j];
        const dx = b.x - a.x || 0.001, dy = b.y - a.y || 0.001;
        const dist2 = dx * dx + dy * dy, dist = Math.sqrt(dist2) || 0.1;
        const force = (3500 / dist2) * alpha;
        const fx = (dx / dist) * force, fy = (dy / dist) * force;
        a.vx -= fx; a.vy -= fy; b.vx += fx; b.vy += fy;
      }
    }
    for (const edge of edges) {
      const a = result.find((n) => n.id === edge.source);
      const b = result.find((n) => n.id === edge.target);
      if (!a || !b) continue;
      const dx = b.x - a.x, dy = b.y - a.y;
      const dist = Math.sqrt(dx * dx + dy * dy) || 0.1;
      const force = (dist - 150) * 0.05;
      const fx = (dx / dist) * force, fy = (dy / dist) * force;
      a.vx += fx; a.vy += fy; b.vx -= fx; b.vy -= fy;
    }
    for (const node of result) {
      node.x += node.vx * alpha; node.y += node.vy * alpha;
      node.vx *= 0.65; node.vy *= 0.65;
    }
  }
  return result;
}

export default function KnowledgeGraph() {
  const { state, reload } = useApi(getGraph, []);
  const svgRef = useRef<SVGSVGElement>(null);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [panStart, setPanStart] = useState({ x: 0, y: 0 });
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [activeTypes, setActiveTypes] = useState<string[]>([]);

  const graphData = state.status === 'ok' ? state.data : null;

  const simNodes = useMemo(() => {
    if (!graphData) return [];
    return runSimulation(graphData.nodes, graphData.edges);
  }, [graphData]);

  const nodeTypes = useMemo(() => {
    const types = new Set(simNodes.map((n) => n.type));
    return Array.from(types);
  }, [simNodes]);

  const visibleNodes = useMemo(() => {
    let result = simNodes;
    if (activeTypes.length) result = result.filter((n) => activeTypes.includes(n.type));
    if (search) {
      const q = search.toLowerCase();
      result = result.filter((n) => n.label.toLowerCase().includes(q) || n.description.toLowerCase().includes(q));
    }
    return result;
  }, [simNodes, activeTypes, search]);

  const visibleIds = useMemo(() => new Set(visibleNodes.map((n) => n.id)), [visibleNodes]);
  const visibleEdges = useMemo(
    () => (graphData?.edges ?? []).filter((e) => visibleIds.has(e.source) && visibleIds.has(e.target)),
    [graphData, visibleIds]
  );

  const selectedNode = visibleNodes.find((n) => n.id === selectedId) ?? null;
  const connectedEdges = selectedNode
    ? visibleEdges.filter((e) => e.source === selectedNode.id || e.target === selectedNode.id)
    : [];
  const connectedIds = new Set(connectedEdges.flatMap((e) => [e.source, e.target]));

  const handleWheel = useCallback((e: React.WheelEvent) => {
    e.preventDefault();
    setZoom((z) => Math.min(4, Math.max(0.2, z * (e.deltaY > 0 ? 0.9 : 1.1))));
  }, []);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if ((e.target as Element).closest('circle.node')) return;
    setIsDragging(true);
    setDragStart({ x: e.clientX, y: e.clientY });
    setPanStart({ x: pan.x, y: pan.y });
  }, [pan]);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!isDragging) return;
    setPan({ x: panStart.x + e.clientX - dragStart.x, y: panStart.y + e.clientY - dragStart.y });
  }, [isDragging, dragStart, panStart]);

  const handleMouseUp = useCallback(() => setIsDragging(false), []);

  function toggleType(t: string) {
    setActiveTypes((prev) => prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]);
  }

  const svgCx = (svgRef.current?.clientWidth ?? 700) / 2;
  const svgCy = (svgRef.current?.clientHeight ?? 500) / 2;

  if (state.status === 'loading') {
    return (
      <div className="max-w-full flex flex-col gap-4 h-[calc(100vh-7.5rem)]">
        <PageHeader title="Knowledge Graph" />
        <div className="flex-1 bg-card border border-border rounded-xl animate-pulse" />
      </div>
    );
  }

  if (state.status === 'error') return <ErrorState message={state.error} onRetry={reload} />;

  return (
    <div className="max-w-full flex gap-4" style={{ height: 'calc(100vh - 7.5rem)' }}>
      {/* Left controls */}
      <div className="w-52 shrink-0 flex flex-col gap-3">
        <div>
          <h1 className="text-xl font-bold text-foreground tracking-tight">Knowledge Graph</h1>
          <p className="text-xs text-muted-foreground mt-0.5">
            {visibleNodes.length} nodes · {visibleEdges.length} edges
          </p>
        </div>

        <SearchInput value={search} onChange={setSearch} placeholder="Search nodes…" className="bg-card border border-border" />

        {/* Node type filter */}
        <div className="bg-card border border-border rounded-lg p-3 flex-1 overflow-y-auto">
          <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">Node Types</p>
          <div className="space-y-1">
            {nodeTypes.map((type) => {
              const count = simNodes.filter((n) => n.type === type).length;
              const active = activeTypes.includes(type);
              return (
                <button
                  key={type}
                  onClick={() => toggleType(type)}
                  className={`w-full flex items-center gap-2 px-2 py-1.5 rounded-md text-xs transition-colors ${active ? 'bg-muted' : 'hover:bg-muted/50'}`}
                >
                  <div className="w-2.5 h-2.5 rounded-full shrink-0" style={{ backgroundColor: colorFor(type), opacity: active ? 1 : 0.5 }} />
                  <span className={active ? 'text-foreground font-medium' : 'text-muted-foreground capitalize'}>{type}</span>
                  <span className="ml-auto text-muted-foreground text-xs">{count}</span>
                </button>
              );
            })}
          </div>
          {activeTypes.length > 0 && (
            <button onClick={() => setActiveTypes([])} className="w-full mt-2 text-xs text-muted-foreground hover:text-foreground py-1 text-center">
              Clear filters
            </button>
          )}
        </div>

        {/* Zoom controls */}
        <div className="bg-card border border-border rounded-lg p-2.5 flex items-center gap-2">
          <button onClick={() => setZoom((z) => Math.max(0.2, z - 0.15))} className="w-7 h-7 flex items-center justify-center rounded hover:bg-muted text-muted-foreground hover:text-foreground text-base font-medium transition-colors">−</button>
          <span className="flex-1 text-center text-xs text-muted-foreground font-mono">{Math.round(zoom * 100)}%</span>
          <button onClick={() => setZoom((z) => Math.min(4, z + 0.15))} className="w-7 h-7 flex items-center justify-center rounded hover:bg-muted text-muted-foreground hover:text-foreground text-base font-medium transition-colors">+</button>
          <button
            onClick={() => { setZoom(1); setPan({ x: 0, y: 0 }); }}
            title="Reset view"
            className="w-7 h-7 flex items-center justify-center rounded hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
          >
            <svg width="11" height="11" viewBox="0 0 12 12" fill="none">
              <path d="M10 6A4 4 0 1 1 6 2" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round"/>
              <path d="M6 1l1.5 1.5L6 4" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round"/>
            </svg>
          </button>
        </div>
      </div>

      {/* Canvas */}
      <div className="flex-1 bg-card border border-border rounded-xl overflow-hidden relative">
        {graphData && graphData.nodes.length === 0 && (
          <EmptyState title="Empty graph" description="No nodes have been recorded yet. The knowledge graph builds as agent sessions discover relationships." />
        )}
        <svg
          ref={svgRef}
          className={`w-full h-full ${isDragging ? 'cursor-grabbing' : 'cursor-grab'}`}
          onWheel={handleWheel}
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
        >
          <defs>
            <marker id="kg-arrow" markerWidth="6" markerHeight="6" refX="5" refY="3" orient="auto">
              <path d="M0,0 L6,3 L0,6 Z" fill="#D4D4D8"/>
            </marker>
            <marker id="kg-arrow-active" markerWidth="6" markerHeight="6" refX="5" refY="3" orient="auto">
              <path d="M0,0 L6,3 L0,6 Z" fill="#7C3AED"/>
            </marker>
          </defs>
          <g transform={`translate(${pan.x + svgCx}, ${pan.y + svgCy}) scale(${zoom})`}>
            {/* Edges */}
            {visibleEdges.map((edge, i) => {
              const src = visibleNodes.find((n) => n.id === edge.source);
              const tgt = visibleNodes.find((n) => n.id === edge.target);
              if (!src || !tgt) return null;
              const isActive = selectedNode && connectedEdges.includes(edge);
              const dx = tgt.x - src.x, dy = tgt.y - src.y;
              const dist = Math.sqrt(dx * dx + dy * dy) || 1;
              const ex = tgt.x - (dx / dist) * 14, ey = tgt.y - (dy / dist) * 14;
              return (
                <g key={i} opacity={selectedNode && !isActive ? 0.15 : 1}>
                  <line
                    x1={src.x} y1={src.y} x2={ex} y2={ey}
                    stroke={isActive ? '#7C3AED' : '#E4E4E7'}
                    strokeWidth={isActive ? 1.5 : 1}
                    markerEnd={`url(#kg-arrow${isActive ? '-active' : ''})`}
                  />
                  {isActive && edge.label && (
                    <text x={(src.x + tgt.x) / 2} y={(src.y + tgt.y) / 2 - 6}
                      textAnchor="middle" fontSize="8" fill="#7C3AED" fontFamily="var(--font-sans)" opacity="0.9">
                      {edge.label}
                    </text>
                  )}
                </g>
              );
            })}

            {/* Nodes */}
            {visibleNodes.map((node) => {
              const isSelected = node.id === selectedId;
              const isConnected = connectedIds.has(node.id);
              const dimmed = selectedNode && !isSelected && !isConnected;
              const color = colorFor(node.type);
              return (
                <g
                  key={node.id}
                  transform={`translate(${node.x}, ${node.y})`}
                  onClick={() => setSelectedId(isSelected ? null : node.id)}
                  className="cursor-pointer"
                  opacity={dimmed ? 0.18 : 1}
                >
                  {isSelected && <circle r="20" fill={color} opacity="0.12" />}
                  <circle
                    r="12" className="node"
                    fill={isSelected ? color : 'white'}
                    stroke={color} strokeWidth={isSelected ? 0 : 2}
                  />
                  <text textAnchor="middle" y="26" fontSize="9"
                    fontFamily="var(--font-sans)"
                    fontWeight={isSelected ? '600' : '400'}
                    fill={isSelected ? color : '#71717A'}>
                    {node.label.length > 16 ? node.label.slice(0, 15) + '…' : node.label}
                  </text>
                  {/* Type badge */}
                  <circle cx="8" cy="-8" r="5" fill={color} />
                  <text x="8" y="-5.5" textAnchor="middle" fontSize="5.5" fill="white" fontFamily="var(--font-sans)" fontWeight="600">
                    {node.type[0].toUpperCase()}
                  </text>
                </g>
              );
            })}
          </g>
        </svg>

        <div className="absolute bottom-3 left-3 text-xs text-muted-foreground bg-card/90 backdrop-blur-sm px-2.5 py-1.5 rounded-lg border border-border pointer-events-none">
          Scroll to zoom · Drag to pan · Click to inspect
        </div>
      </div>

      {/* Inspector */}
      {selectedNode ? (
        <InspectorPanel
          title={selectedNode.label}
          badge={
            <span className="inline-flex items-center gap-1.5 text-xs px-2 py-0.5 rounded-full font-medium capitalize"
              style={{ backgroundColor: colorFor(selectedNode.type) + '20', color: colorFor(selectedNode.type) }}>
              <span className="w-1.5 h-1.5 rounded-full" style={{ backgroundColor: colorFor(selectedNode.type) }} />
              {selectedNode.type}
            </span>
          }
          onClose={() => setSelectedId(null)}
        >
          <div>
            <SectionLabel>Description</SectionLabel>
            <p className="text-xs text-foreground leading-relaxed">{selectedNode.description}</p>
          </div>

          <div>
            <SectionLabel>Details</SectionLabel>
            <div className="rounded-lg border border-border overflow-hidden">
              <MetaRow label="Type" value={
                <span className="capitalize font-medium" style={{ color: colorFor(selectedNode.type) }}>{selectedNode.type}</span>
              } />
              <MetaRow label="ID" value={<span className="font-mono text-xs">{selectedNode.id}</span>} />
              <MetaRow label="Connections" value={selectedNode.connections} />
            </div>
          </div>

          {connectedEdges.length > 0 && (
            <div>
              <SectionLabel>Relations ({connectedEdges.length})</SectionLabel>
              <div className="space-y-1.5">
                {connectedEdges.map((edge, i) => {
                  const isSource = edge.source === selectedNode.id;
                  const otherId = isSource ? edge.target : edge.source;
                  const other = visibleNodes.find((n) => n.id === otherId);
                  if (!other) return null;
                  return (
                    <div key={i} className="flex items-center gap-1.5 text-xs p-1.5 rounded bg-muted/50">
                      <span className="text-muted-foreground">{isSource ? '→' : '←'}</span>
                      {edge.label && <span className="text-muted-foreground font-mono text-xs">{edge.label}</span>}
                      <button
                        onClick={() => setSelectedId(other.id)}
                        className="text-primary hover:underline truncate"
                      >
                        {other.label}
                      </button>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {selectedNode.metadata && Object.keys(selectedNode.metadata).length > 0 && (
            <div>
              <SectionLabel>Metadata</SectionLabel>
              <div className="rounded-lg border border-border overflow-hidden">
                {Object.entries(selectedNode.metadata).map(([k, v]) => (
                  <MetaRow key={k} label={k} value={<span className="font-mono">{v}</span>} />
                ))}
              </div>
            </div>
          )}
        </InspectorPanel>
      ) : (
        <div className="w-80 shrink-0 bg-card border border-border rounded-xl flex flex-col items-center justify-center text-muted-foreground">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" className="mb-2 opacity-25">
            <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="1.5"/>
            <circle cx="12" cy="12" r="3" fill="currentColor"/>
          </svg>
          <p className="text-xs">Click a node to inspect</p>
        </div>
      )}
    </div>
  );
}
