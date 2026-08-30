import type {
  Overview, Session, Event, Memory, GraphData, SearchResult, EventType, ImportanceLevel,
} from './types';

const API_BASE = '/api';

// Minimal fallback data for when API is unavailable
const fallbackOverview: Overview = {
  project: { name: 'Agent Ledger', path: '', branch: '', commit: '', lastActivity: '' },
  stats: { sessions: 0, decisions: 0, discoveries: 0, failures: 0, constraints: 0, checkpoints: 0, memories: 0 },
  recentEvents: [],
  activeSession: null,
};

const fallbackSession: Session = {
  id: '1',
  name: 'Sample Session',
  agent: 'agent',
  model: 'gpt-4',
  branch: 'main',
  status: 'completed',
  eventCount: 0,
  startedAt: new Date().toISOString(),
  summary: 'No session data available',
  decisions: 0,
  discoveries: 0,
  failures: 0,
};

async function request<T>(path: string, fallback: T): Promise<T> {
  try {
    const res = await fetch(`${API_BASE}${path}`);
    if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
    return res.json() as Promise<T>;
  } catch (error) {
    console.warn(`API request failed for ${path}, using fallback data:`, error);
    return fallback;
  }
}

export async function getOverview(): Promise<Overview> {
  const [data, events, sessions, memories] = await Promise.all([
    request<BackendOverview | null>('/overview', null),
    getEvents(),
    getSessions(),
    getMemories(),
  ]);
  if (!data) return fallbackOverview;

  // The Go API exposes overview fields as a flat, snake_case object. Keep the
  // React-facing model stable by adapting that wire format at the API boundary.
  return {
    project: {
      name: data.project_name && data.project_name !== 'Project'
        ? data.project_name
        : data.repository_root.split(/[\\/]/).filter(Boolean).pop() || fallbackOverview.project.name,
      path: data.repository_root || '',
      branch: data.current_branch || '',
      commit: data.current_commit || '',
      lastActivity: data.last_activity_time || '',
      version: data.version || '',
    },
    stats: {
      sessions: data.session_count ?? 0,
      decisions: data.decision_count ?? 0,
      discoveries: data.discovery_count ?? 0,
      failures: events.filter((event) => event.type === 'failure').length,
      constraints: events.filter((event) => event.type === 'constraint').length,
      checkpoints: data.checkpoint_count ?? 0,
      memories: memories.length,
    },
    recentEvents: events.slice(0, 8),
    activeSession: sessions.find((session) => session.status === 'running') || null,
  };
}

interface BackendOverview {
  project_name: string;
  repository_root: string;
  current_branch: string;
  current_commit: string;
  version: string;
  session_count: number;
  decision_count: number;
  discovery_count: number;
  checkpoint_count: number;
  last_activity_time?: string;
}

export async function getSessions(): Promise<Session[]> {
  const [data, events] = await Promise.all([
    request<BackendSessionList>('/sessions', { sessions: [] }),
    getEvents(),
  ]);
  return (data.sessions || []).map((session) => {
    const sessionEvents = events.filter((event) => event.sessionId === session.id);
    const count = (type: EventType) => sessionEvents.filter((event) => event.type === type).length;
    return normalizeSession(session, {
      decisions: count('decision'),
      discoveries: count('discovery'),
      failures: count('failure'),
      constraints: count('constraint'),
      checkpoints: count('checkpoint'),
    });
  });
}

export async function getSession(id: string): Promise<Session> {
  const data = await request<BackendSessionDetail | null>(`/sessions/${id}`, null);
  if (!data?.session) return fallbackSession;
  return normalizeSession(data.session, {
    decisions: data.decision_count,
    discoveries: data.discovery_count,
    failures: data.failure_count,
    constraints: data.constraint_count,
    checkpoints: data.checkpoint_count,
  });
}

export async function getEvents(type?: EventType): Promise<Event[]> {
  const params = type ? `?type=${type}` : '';
  const data = await request(`/events${params}`, { events: [] });
  return (data.events || [])
    .map(normalizeEvent)
    .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
}

export async function getMemories(query?: string, type?: string): Promise<Memory[]> {
  const params = new URLSearchParams();
  if (query) params.set('q', query);
  if (type) params.set('type', type);
  const path = `/memories${params.toString() ? '?' + params.toString() : ''}`;
  const data = await request<BackendMemory[]>(path, []);
  return (Array.isArray(data) ? data : []).map((memory) => ({
    id: memory.id,
    type: memory.type as Memory['type'],
    content: memory.content || memory.title,
    importance: normalizeImportance(memory.importance),
    timestamp: memory.created_at,
    sessionId: memory.session_id || '',
    tags: memory.keywords ? memory.keywords.split(',').map((tag) => tag.trim()).filter(Boolean) : [],
    source: memory.path || undefined,
  }));
}

export async function getGraph(): Promise<GraphData> {
  return request('/graph', { nodes: [], edges: [] });
}

export async function search(query: string): Promise<SearchResult[]> {
  if (!query.trim()) return [];
  const data = await request<{ results: BackendSearchResult[] }>(`/search?q=${encodeURIComponent(query)}`, { results: [] });
  return (data.results || []).map((result) => ({
    id: result.id,
    type: result.type,
    title: result.title,
    snippet: result.excerpt,
    timestamp: '',
  }));
}

interface BackendSession {
  id: string;
  agent?: string;
  model?: string;
  branch: string;
  head: string;
  start_time: string;
  end_time?: string;
  status: string;
}

interface BackendSessionList { sessions: BackendSession[] }
interface BackendSessionDetail {
  session: BackendSession;
  checkpoint_count: number;
  decision_count: number;
  discovery_count: number;
  failure_count: number;
  constraint_count: number;
}
interface SessionCounts {
  decisions?: number;
  discoveries?: number;
  failures?: number;
  constraints?: number;
  checkpoints?: number;
}

function normalizeSession(session: BackendSession, counts: SessionCounts = {}): Session {
  const status: Session['status'] = session.status === 'active' ? 'running' :
    session.status === 'ended' ? 'completed' :
    ['running', 'completed', 'failed', 'paused'].includes(session.status)
      ? session.status as Session['status'] : 'completed';
  const total = Object.values(counts).reduce((sum, count) => sum + (count || 0), 0);

  // Dynamic session naming based on agent & model parameters passed by MCP tools
  const agentName = session.agent && session.agent !== 'unknown' ? session.agent : 'AI Agent';
  const displayName = session.model ? `${agentName} (${session.model})` : agentName;

  return {
    id: session.id,
    name: displayName,
    agent: agentName,
    model: session.model || '',
    branch: session.branch || '',
    status,
    eventCount: total,
    startedAt: session.start_time,
    endedAt: session.end_time,
    summary: status === 'running' ? 'This agent session is currently active.' : 'Agent session completed.',
    decisions: counts.decisions || 0,
    discoveries: counts.discoveries || 0,
    failures: counts.failures || 0,
    constraints: counts.constraints || 0,
    checkpoints: counts.checkpoints || 0,
  };
}

interface BackendEvent { id: string; type: EventType; title: string; content: string; timestamp: string }
interface BackendMemory {
  id: string; type: string; title: string; content: string; keywords: string;
  created_at: string; importance: number; session_id?: string; path?: string;
}
interface BackendSearchResult { id: string; type: string; title: string; excerpt: string }

function normalizeImportance(value: number): ImportanceLevel {
  if (value >= 0.9) return 'critical';
  if (value >= 0.7) return 'high';
  if (value >= 0.4) return 'medium';
  return 'low';
}

function normalizeEvent(event: BackendEvent): Event {
  const createdAt = event.content?.match(/\*Created:\s*([^*\s]+)\*/i)?.[1];
  const meaningfulLine = event.content
    ?.split('\n')
    .map((line) => line.trim())
    .find((line) => /^\*\*(Decision|Finding|Constraint|Attempted|Rationale):\*\*/i.test(line));
  return {
    id: event.id,
    type: event.type,
    title: event.title.replace(/^[0-9a-f]{8}-/, '').replaceAll('-', ' '),
    description: meaningfulLine
      ? meaningfulLine.replace(/^\*\*[^:]+:\*\*\s*/, '')
      : event.content || '',
    timestamp: createdAt || event.timestamp,
    sessionId: event.content?.match(/(?:\*Session:\s*|session\s+)([0-9a-f-]{36})/i)?.[1] || '',
  };
}

export function createWebSocket(onEvent: (e: MessageEvent) => void, onStatusChange: (status: WsStatus) => void) {
  const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
  const url = `${proto}://${window.location.host}/api/live`;
  let ws: WebSocket | null = null;
  let retryTimer: ReturnType<typeof setTimeout> | null = null;
  let closed = false;

  function connect() {
    if (closed) return;
    onStatusChange('connecting');
    try {
      ws = new WebSocket(url);
      ws.addEventListener('open', () => onStatusChange('connected'));
      ws.addEventListener('message', onEvent);
      ws.addEventListener('close', () => {
        if (!closed) {
          onStatusChange('reconnecting');
          retryTimer = setTimeout(connect, 3000);
        }
      });
      ws.addEventListener('error', () => {
        onStatusChange('disconnected');
      });
    } catch {
      onStatusChange('disconnected');
    }
  }

  connect();

  return () => {
    closed = true;
    if (retryTimer) clearTimeout(retryTimer);
    ws?.close();
  };
}

export type WsStatus = 'connected' | 'connecting' | 'reconnecting' | 'disconnected';

export async function getContext(task?: string) {
  const params = task ? `?task=${encodeURIComponent(task)}` : '';
  return request(`/context${params}`, {} as import('./types').ProjectContextData);
}

