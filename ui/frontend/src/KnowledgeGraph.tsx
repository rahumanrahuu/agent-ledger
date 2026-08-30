import { useState, useRef, useEffect, useMemo, useCallback } from 'react';
import * as d3 from 'd3-force';
import { getGraph } from './api/client';
import { useApi } from './hooks/useApi';
import type { GraphNode } from './api/types';
import {
  SearchInput, ErrorState, EmptyState, PageHeader,
  SectionLabel, MetaRow, InspectorPanel,
} from './components/ui';

const nodeTypeColors: Record<string, string> = {
  session: '#EA580C',
  decision: '#3B82F6',
  discovery: '#22C55E',
  failure: '#EF4444',
  constraint: '#F59E0B',
  checkpoint: '#8B5CF6',
  memory: '#06B6D4',
  fact: '#06B6D4',
  rule: '#EF4444',
  pattern: '#8B5CF6',
  entity: '#3B82F6',
  preference: '#F59E0B',
  insight: '#22C55E',
};

function colorFor(type: string) {
  return nodeTypeColors[type.toLowerCase()] ?? '#A1A1AA';
}

interface SimNode extends d3.SimulationNodeDatum {
  id: string;
  label: string;
  type: string;
  description?: string;
  connections: number;
  degree: number;
  radius: number;
  data?: any;
  x?: number;
  y?: number;
  vx?: number;
  vy?: number;
  fx?: number | null;
  fy?: number | null;
}

interface SimLink extends d3.SimulationLinkDatum<SimNode> {
  source: string | SimNode;
  target: string | SimNode;
  label?: string;
  type?: string;
}

export default function KnowledgeGraph() {
  const { state, reload } = useApi(getGraph, []);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [zoom, setZoom] = useState(1);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [hoveredId, setHoveredId] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [activeTypes, setActiveTypes] = useState<string[]>([]);
  const [isInitialFitted, setIsInitialFitted] = useState(false);

  // Interaction tracking refs
  const isPanningRef = useRef(false);
  const isDraggingNodeRef = useRef<SimNode | null>(null);
  const dragStartRef = useRef({ x: 0, y: 0 });
  const panStartRef = useRef({ x: 0, y: 0 });
  const mouseMovedRef = useRef(false);

  const graphData = state.status === 'ok' ? state.data : null;

  // Build simulation nodes and links with degree calculation
  const { nodes, links, nodeMap, degreeMap } = useMemo(() => {
    if (!graphData || !graphData.nodes.length) {
      return { nodes: [], links: [], nodeMap: new Map(), degreeMap: new Map() };
    }

    const degMap = new Map<string, number>();
    graphData.edges.forEach((e) => {
      degMap.set(e.source, (degMap.get(e.source) || 0) + 1);
      degMap.set(e.target, (degMap.get(e.target) || 0) + 1);
    });

    const simNodes: SimNode[] = graphData.nodes.map((n) => {
      const degree = degMap.get(n.id) || 0;
      // Restrained radius scaling: 4px to 18px based on degree
      const radius = Math.min(18, Math.max(4.5, 4.5 + Math.sqrt(degree) * 2.2));
      return {
        ...n,
        degree,
        radius,
      };
    });

    const nMap = new Map<string, SimNode>();
    simNodes.forEach((n) => nMap.set(n.id, n));

    const simLinks: SimLink[] = graphData.edges
      .filter((e) => nMap.has(e.source) && nMap.has(e.target))
      .map((e) => ({
        source: e.source,
        target: e.target,
        label: e.label,
        type: e.type,
      }));

    return { nodes: simNodes, links: simLinks, nodeMap: nMap, degreeMap: degMap };
  }, [graphData]);

  // Unique node types
  const nodeTypes = useMemo(() => {
    const types = new Set(nodes.map((n) => n.type));
    return Array.from(types);
  }, [nodes]);

  // Neighbor lookup map for fast highlighting
  const neighborMap = useMemo(() => {
    const map = new Map<string, Set<string>>();
    links.forEach((l) => {
      const sId = typeof l.source === 'object' ? (l.source as SimNode).id : (l.source as string);
      const tId = typeof l.target === 'object' ? (l.target as SimNode).id : (l.target as string);
      if (!map.has(sId)) map.set(sId, new Set());
      if (!map.has(tId)) map.set(tId, new Set());
      map.get(sId)!.add(tId);
      map.get(tId)!.add(sId);
    });
    return map;
  }, [links]);

  // d3-force Simulation Reference
  const simulationRef = useRef<d3.Simulation<SimNode, SimLink> | null>(null);

  // Auto-fit graph bounding box into viewport
  const fitGraph = useCallback((simNodesToFit: SimNode[]) => {
    if (!simNodesToFit.length || !containerRef.current) return;
    const width = containerRef.current.clientWidth || 800;
    const height = containerRef.current.clientHeight || 600;

    let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity;
    simNodesToFit.forEach((n) => {
      const nx = n.x ?? 0;
      const ny = n.y ?? 0;
      if (nx < minX) minX = nx;
      if (nx > maxX) maxX = nx;
      if (ny < minY) minY = ny;
      if (ny > maxY) maxY = ny;
    });

    const graphW = maxX - minX || 100;
    const graphH = maxY - minY || 100;
    const centerX = (minX + maxX) / 2;
    const centerY = (minY + maxY) / 2;

    const scale = Math.min(
      2.2,
      Math.max(0.35, Math.min((width * 0.78) / graphW, (height * 0.78) / graphH))
    );

    setZoom(scale);
    setPan({
      x: width / 2 - centerX * scale,
      y: height / 2 - centerY * scale,
    });
  }, []);

  // Initialize and run d3 force simulation
  useEffect(() => {
    if (!nodes.length) return;

    // Reset initial fit flag
    setIsInitialFitted(false);

    const simulation = d3.forceSimulation<SimNode>(nodes)
      .force(
        'link',
        d3
          .forceLink<SimNode, SimLink>(links)
          .id((d) => d.id)
          .distance(58)
          .strength(0.55)
      )
      .force('charge', d3.forceManyBody().strength(-180).distanceMax(360))
      .force(
        'collision',
        d3.forceCollide<SimNode>().radius((d) => d.radius + 12).strength(1.0).iterations(3)
      )
      .force('center', d3.forceCenter(0, 0))
      .force('x', d3.forceX(0).strength(0.03))
      .force('y', d3.forceY(0).strength(0.03))
      .alphaDecay(0.025);

    simulationRef.current = simulation;

    // Warm up simulation for 120 ticks so graph loads reasonably stabilized
    for (let i = 0; i < 120; ++i) simulation.tick();

    // Perform initial auto-fit
    fitGraph(nodes);
    setIsInitialFitted(true);

    return () => {
      simulation.stop();
    };
  }, [nodes, links, fitGraph]);

  // Main Canvas Rendering Loop
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !containerRef.current) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animId: number;

    const render = () => {
      const width = containerRef.current?.clientWidth || 800;
      const height = containerRef.current?.clientHeight || 600;

      // Handle HiDPI displays
      const dpr = window.devicePixelRatio || 1;
      if (canvas.width !== width * dpr || canvas.height !== height * dpr) {
        canvas.width = width * dpr;
        canvas.height = height * dpr;
      }

      ctx.save();
      ctx.scale(dpr, dpr);
      ctx.clearRect(0, 0, width, height);

      // Apply Pan & Zoom transformations
      ctx.save();
      ctx.translate(pan.x, pan.y);
      ctx.scale(zoom, zoom);

      // Determine active search filter matches
      const searchLower = search.toLowerCase().trim();
      const hasSearch = searchLower.length > 0;
      const hasTypeFilter = activeTypes.length > 0;

      const activeFocusId = hoveredId || selectedId;
      const activeNeighbors = activeFocusId ? neighborMap.get(activeFocusId) || new Set() : null;

      // Helper to check node visibility & dimming
      const isNodeMatch = (n: SimNode) => {
        if (hasTypeFilter && !activeTypes.includes(n.type)) return false;
        if (hasSearch) {
          return (
            n.label.toLowerCase().includes(searchLower) ||
            (n.description && n.description.toLowerCase().includes(searchLower))
          );
        }
        return true;
      };

      // ─── 1. DRAW EDGES ──────────────────────────────────────────────────────────
      links.forEach((l) => {
        const sourceNode = l.source as SimNode;
        const targetNode = l.target as SimNode;
        if (!sourceNode || !targetNode || sourceNode.x === undefined || targetNode.x === undefined) return;

        const sId = sourceNode.id;
        const tId = targetNode.id;

        const isEdgeConnectedToActive =
          activeFocusId && (sId === activeFocusId || tId === activeFocusId);
        const isEdgeConnectedToNeighbor =
          activeFocusId &&
          (activeNeighbors?.has(sId) || sId === activeFocusId) &&
          (activeNeighbors?.has(tId) || tId === activeFocusId);

        let opacity = 0.25;
        let strokeColor = '#3F3F46'; // Neutral dark zinc
        let lineWidth = 0.9;

        if (activeFocusId) {
          if (isEdgeConnectedToActive) {
            opacity = 0.95;
            strokeColor = colorFor(sourceNode.type);
            lineWidth = 1.8;
          } else if (isEdgeConnectedToNeighbor) {
            opacity = 0.6;
            strokeColor = '#A1A1AA';
            lineWidth = 1.2;
          } else {
            opacity = 0.04; // Fade unrelated edges
          }
        } else if (hasSearch || hasTypeFilter) {
          const sMatch = isNodeMatch(sourceNode);
          const tMatch = isNodeMatch(targetNode);
          if (sMatch && tMatch) {
            opacity = 0.6;
            lineWidth = 1.2;
          } else {
            opacity = 0.05;
          }
        }

        ctx.beginPath();
        ctx.moveTo(sourceNode.x!, sourceNode.y!);
        ctx.lineTo(targetNode.x!, targetNode.y!);
        ctx.strokeStyle = strokeColor;
        ctx.globalAlpha = opacity;
        ctx.lineWidth = lineWidth;
        ctx.stroke();
      });

      ctx.globalAlpha = 1.0;

      // ─── 2. DRAW NODES ──────────────────────────────────────────────────────────
      nodes.forEach((n) => {
        if (n.x === undefined || n.y === undefined) return;

        const isSelected = n.id === selectedId;
        const isHovered = n.id === hoveredId;
        const isNeighbor = activeNeighbors ? activeNeighbors.has(n.id) : false;
        const matchesFilter = isNodeMatch(n);

        let opacity = 1.0;
        if (activeFocusId) {
          if (n.id === activeFocusId) opacity = 1.0;
          else if (isNeighbor) opacity = 0.9;
          else opacity = 0.15; // Dim unrelated nodes
        } else if (hasSearch || hasTypeFilter) {
          opacity = matchesFilter ? 1.0 : 0.12;
        }

        ctx.globalAlpha = opacity;
        const baseColor = colorFor(n.type);
        const radius = isHovered || isSelected ? n.radius + 2 : n.radius;

        // Outer glow/ring for selected or hovered node
        if (isSelected || isHovered) {
          ctx.beginPath();
          ctx.arc(n.x, n.y, radius + 4, 0, 2 * Math.PI);
          ctx.fillStyle = baseColor;
          ctx.globalAlpha = opacity * 0.25;
          ctx.fill();

          ctx.beginPath();
          ctx.arc(n.x, n.y, radius + 2, 0, 2 * Math.PI);
          ctx.strokeStyle = '#FFFFFF';
          ctx.lineWidth = 1.5;
          ctx.globalAlpha = opacity;
          ctx.stroke();
        }

        // Solid Node Circle
        ctx.beginPath();
        ctx.arc(n.x, n.y, radius, 0, 2 * Math.PI);
        ctx.fillStyle = baseColor;
        ctx.globalAlpha = opacity;
        ctx.fill();

        // Subtle node border
        ctx.strokeStyle = 'rgba(255, 255, 255, 0.3)';
        ctx.lineWidth = 0.8;
        ctx.stroke();
      });

      // ─── 3. DRAW LABELS WITH SEMANTIC ZOOM ─────────────────────────────────────
      nodes.forEach((n) => {
        if (n.x === undefined || n.y === undefined) return;

        const isSelected = n.id === selectedId;
        const isHovered = n.id === hoveredId;
        const isNeighbor = activeNeighbors ? activeNeighbors.has(n.id) : false;
        const matchesFilter = isNodeMatch(n);

        // Semantic Zoom label logic:
        // - Zoom < 0.75: Only major hubs (degree >= 4)
        // - 0.75 <= Zoom < 1.4: Hubs (degree >= 2) + selected/hovered/neighbors
        // - Zoom >= 1.4: All labels
        let showLabel = false;
        if (isHovered || isSelected) {
          showLabel = true;
        } else if (zoom < 0.75) {
          showLabel = n.degree >= 4;
        } else if (zoom < 1.4) {
          showLabel = n.degree >= 2 || isNeighbor;
        } else {
          showLabel = true;
        }

        if (!showLabel && !hasSearch) return;
        if (hasSearch && !matchesFilter) return;

        let opacity = 0.85;
        if (activeFocusId) {
          if (n.id === activeFocusId) opacity = 1.0;
          else if (isNeighbor) opacity = 0.85;
          else opacity = 0.15;
        }

        ctx.globalAlpha = opacity;
        ctx.font = `${isHovered || isSelected ? '600' : '400'} 10px Plus Jakarta Sans, system-ui, sans-serif`;
        ctx.fillStyle = isSelected || isHovered ? '#FFFFFF' : '#E4E4E7';

        // Human-readable label truncation
        let displayLabel = n.label;
        if (!isHovered && displayLabel.length > 22) {
          displayLabel = displayLabel.slice(0, 21) + '…';
        }

        const labelX = n.x + n.radius + 5;
        const labelY = n.y + 3;

        ctx.fillText(displayLabel, labelX, labelY);
      });

      ctx.restore();
      ctx.restore();

      animId = requestAnimationFrame(render);
    };

    render();

    return () => {
      cancelAnimationFrame(animId);
    };
  }, [nodes, links, pan, zoom, selectedId, hoveredId, search, activeTypes, neighborMap]);

  // Helper to convert Canvas Mouse Coordinates to World Coordinates
  const getWorldCoords = useCallback(
    (clientX: number, clientY: number) => {
      const canvas = canvasRef.current;
      if (!canvas) return { x: 0, y: 0 };
      const rect = canvas.getBoundingClientRect();
      const screenX = clientX - rect.left;
      const screenY = clientY - rect.top;
      return {
        x: (screenX - pan.x) / zoom,
        y: (screenY - pan.y) / zoom,
      };
    },
    [pan, zoom]
  );

  // Find node at world coordinate
  const getNodeAtCoords = useCallback(
    (worldX: number, worldY: number) => {
      for (let i = nodes.length - 1; i >= 0; i--) {
        const n = nodes[i];
        if (n.x === undefined || n.y === undefined) continue;
        const dx = worldX - n.x;
        const dy = worldY - n.y;
        const dist = Math.sqrt(dx * dx + dy * dy);
        if (dist <= n.radius + 4) {
          return n;
        }
      }
      return null;
    },
    [nodes]
  );

  // Mouse Handlers
  const handleMouseDown = (e: React.MouseEvent) => {
    mouseMovedRef.current = false;
    dragStartRef.current = { x: e.clientX, y: e.clientY };
    panStartRef.current = { ...pan };

    const world = getWorldCoords(e.clientX, e.clientY);
    const hitNode = getNodeAtCoords(world.x, world.y);

    if (hitNode) {
      isDraggingNodeRef.current = hitNode;
      hitNode.fx = hitNode.x;
      hitNode.fy = hitNode.y;
      if (simulationRef.current) {
        simulationRef.current.alphaTarget(0.15).restart();
      }
    } else {
      isPanningRef.current = true;
    }
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    const dx = e.clientX - dragStartRef.current.x;
    const dy = e.clientY - dragStartRef.current.y;
    if (Math.abs(dx) > 3 || Math.abs(dy) > 3) {
      mouseMovedRef.current = true;
    }

    if (isDraggingNodeRef.current) {
      const world = getWorldCoords(e.clientX, e.clientY);
      isDraggingNodeRef.current.fx = world.x;
      isDraggingNodeRef.current.fy = world.y;
      if (simulationRef.current) {
        simulationRef.current.alphaTarget(0.15).restart();
      }
    } else if (isPanningRef.current) {
      setPan({
        x: panStartRef.current.x + dx,
        y: panStartRef.current.y + dy,
      });
    } else {
      // Hover hit test
      const world = getWorldCoords(e.clientX, e.clientY);
      const hoverHit = getNodeAtCoords(world.x, world.y);
      setHoveredId(hoverHit ? hoverHit.id : null);
    }
  };

  const handleMouseUp = (e: React.MouseEvent) => {
    if (isDraggingNodeRef.current) {
      isDraggingNodeRef.current.fx = null;
      isDraggingNodeRef.current.fy = null;
      if (simulationRef.current) {
        simulationRef.current.alphaTarget(0);
      }
      if (!mouseMovedRef.current) {
        setSelectedId(isDraggingNodeRef.current.id);
      }
      isDraggingNodeRef.current = null;
    } else if (isPanningRef.current) {
      isPanningRef.current = false;
      if (!mouseMovedRef.current) {
        setSelectedId(null);
      }
    }
  };

  const handleWheel = (e: React.WheelEvent) => {
    e.preventDefault();
    const zoomFactor = e.deltaY < 0 ? 1.12 : 0.88;
    const newZoom = Math.min(4.5, Math.max(0.25, zoom * zoomFactor));

    const canvas = canvasRef.current;
    if (!canvas) return;
    const rect = canvas.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;

    // Zoom centered around cursor position
    setPan({
      x: mouseX - (mouseX - pan.x) * (newZoom / zoom),
      y: mouseY - (mouseY - pan.y) * (newZoom / zoom),
    });
    setZoom(newZoom);
  };

  const handleDoubleClick = (e: React.MouseEvent) => {
    const world = getWorldCoords(e.clientX, e.clientY);
    const hitNode = getNodeAtCoords(world.x, world.y);
    if (hitNode && containerRef.current && hitNode.x !== undefined && hitNode.y !== undefined) {
      const width = containerRef.current.clientWidth;
      const height = containerRef.current.clientHeight;
      const targetZoom = 1.8;
      setZoom(targetZoom);
      setPan({
        x: width / 2 - hitNode.x * targetZoom,
        y: height / 2 - hitNode.y * targetZoom,
      });
      setSelectedId(hitNode.id);
    }
  };

  // Toggle node type filter
  function toggleType(t: string) {
    setActiveTypes((prev) => (prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]));
  }

  // Selected node details for Inspector panel
  const selectedNode = nodes.find((n) => n.id === selectedId) ?? null;

  const connectedEdges = useMemo(() => {
    if (!selectedNode) return [];
    return links.filter((l) => {
      const sId = typeof l.source === 'object' ? (l.source as SimNode).id : (l.source as string);
      const tId = typeof l.target === 'object' ? (l.target as SimNode).id : (l.target as string);
      return sId === selectedNode.id || tId === selectedNode.id;
    });
  }, [selectedNode, links]);

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
      {/* Left Sidebar Controls */}
      <div className="w-56 shrink-0 flex flex-col gap-3">
        <div>
          <h1 className="text-xl font-bold text-foreground tracking-tight">Knowledge Graph</h1>
          <p className="text-xs text-muted-foreground mt-0.5 font-mono">
            {nodes.length} nodes · {links.length} relationships
          </p>
        </div>

        <SearchInput
          value={search}
          onChange={setSearch}
          placeholder="Search nodes..."
          className="bg-card border border-border"
        />

        {/* Node Type Filter */}
        <div className="bg-card border border-border rounded-xl p-3 flex-1 overflow-y-auto space-y-1">
          <p className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider mb-2">
            Node Types
          </p>
          {nodeTypes.map((type) => {
            const count = nodes.filter((n) => n.type === type).length;
            const active = activeTypes.includes(type);
            return (
              <button
                key={type}
                onClick={() => toggleType(type)}
                className={`w-full flex items-center gap-2 px-2.5 py-1.5 rounded-lg text-xs transition-colors cursor-pointer ${
                  active ? 'bg-primary-light text-primary font-medium' : 'hover:bg-muted/60 text-foreground'
                }`}
              >
                <div
                  className="w-2.5 h-2.5 rounded-full shrink-0 shadow-xs"
                  style={{ backgroundColor: colorFor(type) }}
                />
                <span className="capitalize truncate flex-1 text-left">{type}</span>
                <span className="text-[11px] font-mono text-muted-foreground">{count}</span>
              </button>
            );
          })}
          {activeTypes.length > 0 && (
            <button
              onClick={() => setActiveTypes([])}
              className="w-full mt-2 text-xs text-muted-foreground hover:text-foreground py-1 text-center font-medium cursor-pointer"
            >
              Clear filters
            </button>
          )}
        </div>

        {/* Viewport Zoom & Fit Controls */}
        <div className="bg-card border border-border rounded-xl p-2 flex items-center gap-1.5 shadow-xs">
          <button
            onClick={() => setZoom((z) => Math.max(0.25, z - 0.15))}
            className="w-7 h-7 flex items-center justify-center rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground text-sm font-bold transition-colors cursor-pointer"
          >
            −
          </button>
          <span className="flex-1 text-center text-xs text-muted-foreground font-mono font-medium">
            {Math.round(zoom * 100)}%
          </span>
          <button
            onClick={() => setZoom((z) => Math.min(4.5, z + 0.15))}
            className="w-7 h-7 flex items-center justify-center rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground text-sm font-bold transition-colors cursor-pointer"
          >
            +
          </button>
          <button
            onClick={() => fitGraph(nodes)}
            title="Auto-Fit Graph"
            className="w-7 h-7 flex items-center justify-center rounded-lg hover:bg-muted text-muted-foreground hover:text-foreground transition-colors cursor-pointer"
          >
            <svg width="13" height="13" viewBox="0 0 14 14" fill="none">
              <path d="M1.5 5.5V2.5A1 1 0 0 1 2.5 1.5H5.5M8.5 1.5H11.5A1 1 0 0 1 12.5 2.5V5.5M12.5 8.5V11.5A1 1 0 0 1 11.5 12.5H8.5M5.5 12.5H2.5A1 1 0 0 1 1.5 11.5V8.5" stroke="currentColor" strokeWidth="1.3" strokeLinecap="round"/>
            </svg>
          </button>
        </div>
      </div>

      {/* Main Graph Canvas Container */}
      <div ref={containerRef} className="flex-1 bg-[#09090B] border border-border rounded-xl overflow-hidden relative shadow-inner">
        {graphData && nodes.length === 0 && (
          <EmptyState
            title="Empty graph"
            description="No nodes have been recorded yet. The knowledge graph builds as agent sessions discover relationships."
          />
        )}
        <canvas
          ref={canvasRef}
          className={`w-full h-full block ${isPanningRef.current ? 'cursor-grabbing' : 'cursor-grab'}`}
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          onWheel={handleWheel}
          onDoubleClick={handleDoubleClick}
        />

        <div className="absolute bottom-3 left-3 text-[11px] text-muted-foreground bg-card/85 backdrop-blur-md px-3 py-1.5 rounded-lg border border-border pointer-events-none shadow-xs">
          Scroll to zoom · Drag canvas to pan · Drag node to reposition · Dbl-click to focus
        </div>
      </div>

      {/* Right Inspector Panel */}
      {selectedNode ? (
        <InspectorPanel
          title={selectedNode.label}
          badge={
            <span
              className="inline-flex items-center gap-1.5 text-xs px-2.5 py-0.5 rounded-full font-medium capitalize"
              style={{
                backgroundColor: colorFor(selectedNode.type) + '25',
                color: colorFor(selectedNode.type),
              }}
            >
              <span
                className="w-1.5 h-1.5 rounded-full"
                style={{ backgroundColor: colorFor(selectedNode.type) }}
              />
              {selectedNode.type}
            </span>
          }
          onClose={() => setSelectedId(null)}
        >
          {selectedNode.description && (
            <div>
              <SectionLabel>Description</SectionLabel>
              <p className="text-xs text-foreground leading-relaxed bg-muted/40 rounded-lg p-3 border border-border">
                {selectedNode.description}
              </p>
            </div>
          )}

          <div>
            <SectionLabel>Details</SectionLabel>
            <div className="rounded-lg border border-border overflow-hidden">
              <MetaRow
                label="Type"
                value={
                  <span className="capitalize font-medium" style={{ color: colorFor(selectedNode.type) }}>
                    {selectedNode.type}
                  </span>
                }
              />
              <MetaRow label="Degree / Connections" value={selectedNode.degree} />
              <MetaRow label="ID" value={<span className="font-mono text-[11px]">{selectedNode.id}</span>} />
            </div>
          </div>

          {connectedEdges.length > 0 && (
            <div>
              <SectionLabel>Relationships ({connectedEdges.length})</SectionLabel>
              <div className="space-y-1.5 max-h-56 overflow-y-auto">
                {connectedEdges.map((edge, i) => {
                  const sId = typeof edge.source === 'object' ? (edge.source as SimNode).id : (edge.source as string);
                  const isSource = sId === selectedNode.id;
                  const otherId = isSource
                    ? typeof edge.target === 'object' ? (edge.target as SimNode).id : (edge.target as string)
                    : sId;
                  const otherNode = nodeMap.get(otherId);
                  if (!otherNode) return null;
                  return (
                    <div
                      key={i}
                      className="flex items-center gap-2 text-xs p-2 rounded-lg bg-muted/40 hover:bg-muted/70 transition-colors border border-border/50"
                    >
                      <span className="text-muted-foreground font-mono text-[11px]">{isSource ? '→' : '←'}</span>
                      {edge.label && (
                        <span className="text-muted-foreground font-mono text-[10px] bg-muted px-1.5 py-0.5 rounded">
                          {edge.label}
                        </span>
                      )}
                      <button
                        onClick={() => setSelectedId(otherNode.id)}
                        className="text-primary hover:underline truncate font-medium text-left flex-1"
                      >
                        {otherNode.label}
                      </button>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </InspectorPanel>
      ) : (
        <div className="w-80 shrink-0 bg-card border border-border rounded-xl flex flex-col items-center justify-center text-muted-foreground">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" className="mb-2 opacity-25">
            <circle cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="1.5" />
            <circle cx="12" cy="12" r="3" fill="currentColor" />
          </svg>
          <p className="text-xs font-medium">Click any node to inspect details</p>
        </div>
      )}
    </div>
  );
}
