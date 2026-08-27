/* Types matching the Agent Ledger backend API contract */

export interface Overview {
  project: ProjectInfo;
  stats: Stats;
  recentEvents: Event[];
  activeSession: Session | null;
}

export interface ProjectInfo {
  name: string;
  path: string;
  branch: string;
  commit: string;
  lastActivity: string;
  version?: string;
  description?: string;
}

export interface Stats {
  sessions: number;
  decisions: number;
  discoveries: number;
  failures: number;
  constraints: number;
  checkpoints: number;
  memories: number;
}

export interface Session {
  id: string;
  name: string;
  agent: string;
  model: string;
  branch: string;
  status: 'running' | 'completed' | 'failed' | 'paused';
  eventCount: number;
  startedAt: string;
  endedAt?: string;
  duration?: string;
  summary: string;
  decisions: number;
  discoveries: number;
  failures: number;
  constraints?: number;
  checkpoints?: number;
}

export type EventType = 'decision' | 'discovery' | 'failure' | 'constraint' | 'checkpoint';

export interface Event {
  id: string;
  type: EventType;
  title: string;
  description: string;
  timestamp: string;
  sessionId: string;
  sessionName?: string;
  metadata?: Record<string, string>;
}

export type MemoryType = 'fact' | 'rule' | 'pattern' | 'entity' | 'preference' | 'insight' | 'decision' | 'discovery' | 'failure' | 'constraint';
export type ImportanceLevel = 'critical' | 'high' | 'medium' | 'low';

export interface Memory {
  id: string;
  type: MemoryType;
  content: string;
  importance: ImportanceLevel;
  timestamp: string;
  sessionId: string;
  sessionName?: string;
  tags: string[];
  source?: string;
  relevance?: number;
}

export interface GraphNode {
  id: string;
  label: string;
  type: string;
  description: string;
  connections: number;
  metadata?: Record<string, string>;
}

export interface GraphEdge {
  source: string;
  target: string;
  label?: string;
  type?: string;
}

export interface GraphData {
  nodes: GraphNode[];
  edges: GraphEdge[];
}

export interface SearchResult {
  id: string;
  type: string;
  title: string;
  snippet: string;
  timestamp: string;
  sessionId?: string;
  sessionName?: string;
}

export interface LiveEvent {
  type: 'event' | 'session' | 'memory' | 'graph' | 'ping';
  payload?: unknown;
}
