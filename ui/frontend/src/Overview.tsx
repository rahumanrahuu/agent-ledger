import { useNavigate } from 'react-router-dom';
import { getOverview, getSessions } from './api/client';
import { useApi } from './hooks/useApi';
import {
  Badge, eventVariant, sessionStatusVariant, SkeletonCard, SkeletonRow,
  ErrorState, EmptyState, formatRelative, formatDateTime, Card, PageHeader,
} from './components/ui';

const nodeTypeColors: Record<string, string> = {
  session: '#EA580C', decision: '#2563EB', discovery: '#16A34A',
  failure: '#DC2626', constraint: '#D97706', checkpoint: '#7C3AED',
  file: '#64748B', function: '#7C3AED', module: '#2563EB', agent: '#EA580C',
  service: '#16A34A', concept: '#9333EA', entity: '#0891B2',
};

export default function Overview() {
  const { state, reload } = useApi(getOverview);
  const { state: sessionsState } = useApi(getSessions, []);
  const navigate = useNavigate();

  if (state.status === 'loading') {
    return (
      <div className="max-w-5xl mx-auto space-y-6">
        <div className="bg-card border border-border rounded-xl p-5 space-y-3">
          <div className="h-6 w-48 animate-pulse bg-muted rounded-md" />
          <div className="h-4 w-80 animate-pulse bg-muted rounded-md" />
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-3">
          {Array.from({ length: 7 }).map((_, i) => <SkeletonCard key={i} />)}
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
          <div className="lg:col-span-3 bg-card border border-border rounded-xl">
            {Array.from({ length: 5 }).map((_, i) => <SkeletonRow key={i} />)}
          </div>
          <div className="lg:col-span-2 h-48 bg-card border border-border rounded-xl animate-pulse" />
        </div>
      </div>
    );
  }

  if (state.status === 'error') {
    return <ErrorState message={state.error} onRetry={reload} />;
  }

  const { project, stats, recentEvents, activeSession } = state.data;

  const statCards = [
    { label: 'Sessions', value: stats.sessions, icon: '⬡', to: '/sessions', variant: 'primary' as const },
    { label: 'Decisions', value: stats.decisions, icon: '◆', to: '/timeline?type=decision', variant: 'info' as const },
    { label: 'Discoveries', value: stats.discoveries, icon: '◉', to: '/timeline?type=discovery', variant: 'success' as const },
    { label: 'Failures', value: stats.failures, icon: '✕', to: '/timeline?type=failure', variant: 'destructive' as const },
    { label: 'Constraints', value: stats.constraints, icon: '⊘', to: '/timeline?type=constraint', variant: 'warning' as const },
    { label: 'Checkpoints', value: stats.checkpoints, icon: '◈', to: '/timeline?type=checkpoint', variant: 'primary' as const },
    { label: 'Memories', value: stats.memories, icon: '◎', to: '/memories', variant: 'success' as const },
  ];

  return (
    <div className="max-w-5xl mx-auto space-y-5">
      {/* Project header */}
      <Card className="p-5">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <div className="flex items-center gap-2 mb-1">
              <h1 className="text-xl font-bold text-foreground tracking-tight">{project.name}</h1>
              {activeSession && (
                <Badge variant="success" dot pulse>Running</Badge>
              )}
            </div>
            {project.description && (
              <p className="text-sm text-muted-foreground">{project.description}</p>
            )}
            {project.path && (
              <p className="text-xs text-muted-foreground font-mono mt-0.5">{project.path}</p>
            )}
          </div>
          <div className="flex flex-wrap gap-2">
            {project.branch && (
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground font-mono bg-muted px-3 py-1.5 rounded-lg">
                <svg width="10" height="10" viewBox="0 0 10 12" fill="none">
                  <circle cx="3" cy="2" r="1.5" stroke="currentColor" strokeWidth="1.2"/>
                  <circle cx="3" cy="10" r="1.5" stroke="currentColor" strokeWidth="1.2"/>
                  <circle cx="8" cy="5" r="1.5" stroke="currentColor" strokeWidth="1.2"/>
                  <path d="M3 3.5V8.5M3 3.5C3 3.5 3 5 8 5" stroke="currentColor" strokeWidth="1.2" strokeLinecap="round"/>
                </svg>
                {project.branch}
              </div>
            )}
            {project.commit && (
              <span className="text-xs font-mono text-muted-foreground bg-muted px-3 py-1.5 rounded-lg" title={project.commit}>{project.commit.slice(0, 8)}</span>
            )}
            {project.lastActivity && (
              <span className="text-xs text-muted-foreground bg-muted px-3 py-1.5 rounded-lg">Updated {formatRelative(project.lastActivity)}</span>
            )}
          </div>
        </div>
      </Card>

      {/* Stat cards — clickable */}
      <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-3">
        {statCards.map((card) => (
          <button
            key={card.label}
            onClick={() => navigate(card.to)}
            className="bg-card border border-border rounded-xl p-4 text-left hover:border-foreground/20 hover:shadow-sm transition-all group"
          >
            <div className={`text-base mb-1 ${card.variant === 'info' ? 'text-info' : card.variant === 'success' ? 'text-success' : card.variant === 'destructive' ? 'text-destructive' : card.variant === 'warning' ? 'text-warning' : 'text-primary'}`}>
              {card.icon}
            </div>
            <div className="text-2xl font-bold text-foreground tabular-nums">{card.value}</div>
            <div className="text-xs text-muted-foreground mt-0.5 group-hover:text-foreground transition-colors">{card.label}</div>
          </button>
        ))}
      </div>

      {/* Lower grid */}
      <div className="grid grid-cols-1 lg:grid-cols-5 gap-4">
        {/* Recent activity */}
        <div className="lg:col-span-3 bg-card border border-border rounded-xl">
          <div className="px-5 py-4 border-b border-border flex items-center justify-between">
            <h2 className="font-semibold text-sm text-foreground">Recent Activity</h2>
            <button onClick={() => navigate('/timeline')} className="text-xs text-primary hover:underline">
              View all
            </button>
          </div>
          {recentEvents.length === 0 ? (
            <EmptyState
              title="No recent activity"
              description="Events will appear here as agent sessions record decisions, discoveries, and failures."
            />
          ) : (
            <div className="divide-y divide-border">
              {recentEvents.slice(0, 8).map((event) => (
                <button
                  key={event.id}
                  onClick={() => navigate(`/timeline?event=${event.id}`)}
                  className="w-full text-left px-5 py-3.5 hover:bg-muted/50 transition-colors"
                >
                  <div className="flex items-start gap-3">
                    <div className={`w-1.5 h-1.5 rounded-full mt-1.5 shrink-0 ${
                      event.type === 'decision' ? 'bg-info' :
                      event.type === 'discovery' ? 'bg-success' :
                      event.type === 'failure' ? 'bg-destructive' :
                      event.type === 'constraint' ? 'bg-warning' : 'bg-primary'
                    }`} />
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2 mb-0.5">
                        <Badge variant={eventVariant[event.type]}>{event.type}</Badge>
                        <span className="text-xs text-muted-foreground">{formatRelative(event.timestamp)}</span>
                      </div>
                      <p className="text-sm text-foreground font-medium leading-snug truncate">{event.title}</p>
                      <p className="text-xs text-muted-foreground leading-snug mt-0.5 truncate">{event.description}</p>
                    </div>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Right column */}
        <div className="lg:col-span-2 space-y-4">
          {/* Active session */}
          {activeSession && (
            <Card hoverable onClick={() => navigate(`/sessions/${activeSession.id}`)} className="p-4">
              <div className="flex items-center justify-between mb-3">
                <h2 className="font-semibold text-sm text-foreground">Active Session</h2>
                <Badge variant="success" dot pulse>Running</Badge>
              </div>
              <p className="text-sm font-medium text-foreground mb-1">{activeSession.name}</p>
              <p className="text-xs text-muted-foreground mb-3 line-clamp-2">{activeSession.summary}</p>
              <div className="grid grid-cols-3 gap-2 mb-3">
                {[
                  { label: 'decisions', value: activeSession.decisions },
                  { label: 'discoveries', value: activeSession.discoveries },
                  { label: 'failures', value: activeSession.failures },
                ].map((m) => (
                  <div key={m.label} className="bg-muted rounded-lg p-2 text-center">
                    <div className="text-base font-bold text-foreground">{m.value}</div>
                    <div className="text-xs text-muted-foreground">{m.label}</div>
                  </div>
                ))}
              </div>
              <div className="flex flex-wrap gap-1.5">
                <span className="text-xs font-mono text-muted-foreground bg-muted px-2 py-0.5 rounded">{activeSession.model}</span>
                <span className="text-xs font-mono text-muted-foreground bg-muted px-2 py-0.5 rounded">{activeSession.branch}</span>
              </div>
            </Card>
          )}

          {/* Sessions summary */}
          <Card className="p-4">
            <div className="flex items-center justify-between mb-3">
              <h2 className="font-semibold text-sm text-foreground">Sessions</h2>
              <button onClick={() => navigate('/sessions')} className="text-xs text-primary hover:underline">View all</button>
            </div>
            <div className="grid grid-cols-2 gap-2">
              {(['running', 'completed', 'failed', 'paused'] as const).map((status) => (
                <div key={status} className="bg-muted/50 rounded-lg p-2.5">
                  <Badge variant={sessionStatusVariant[status]} className="mb-1 capitalize">{status}</Badge>
                  <div className="text-xl font-bold text-foreground">
                    {sessionsState.status === 'ok'
                      ? sessionsState.data.filter((session) => session.status === status).length
                      : '—'}
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
}
