import { useNavigate, useParams } from 'react-router-dom';
import { getSessions, getSession, getEvents } from './api/client';
import { useApi } from './hooks/useApi';
import type { Session } from './api/types';
import {
  Badge, eventVariant, sessionStatusVariant,
  SkeletonCard, SkeletonRow, ErrorState, EmptyState, PageHeader,
  Card, InspectorPanel, SectionLabel, MetaRow,
  formatRelative, formatDateTime,
} from './components/ui';

function modelBadge(model: string) {
  if (model.includes('claude')) return 'bg-primary-light text-primary';
  if (model.includes('gpt')) return 'bg-success-light text-success';
  if (model.includes('gemini')) return 'bg-info-light text-info';
  return 'bg-muted text-muted-foreground';
}

const statusDot: Record<string, string> = {
  running: 'bg-success', completed: 'bg-info', failed: 'bg-destructive', paused: 'bg-warning',
};

export function SessionDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { state: sessionState, reload: reloadSession } = useApi(() => getSession(id!), [id]);
  const { state: eventsState } = useApi(() => getEvents(), []);

  const session = sessionState.status === 'ok' ? sessionState.data : null;
  const allEvents = eventsState.status === 'ok' ? eventsState.data : [];
  const sessionEvents = allEvents
    .filter((e) => e.sessionId === id)
    .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

  if (sessionState.status === 'loading') {
    return (
      <div className="max-w-4xl mx-auto space-y-4">
        <div className="h-5 w-32 animate-pulse bg-muted rounded-md mb-4" />
        <div className="bg-card border border-border rounded-xl p-5 space-y-3">
          <div className="h-6 w-64 animate-pulse bg-muted rounded-md" />
          <div className="h-4 w-full animate-pulse bg-muted rounded-md" />
        </div>
      </div>
    );
  }
  if (sessionState.status === 'error') return <ErrorState message={sessionState.error} onRetry={reloadSession} />;

  return (
    <div className="max-w-4xl mx-auto">
      <PageHeader
        title={session?.name ?? 'Session'}
        breadcrumb={[
          { label: 'Sessions', onClick: () => navigate('/sessions') },
          { label: session?.name ?? id ?? '' },
        ]}
      />

      {session && (
        <>
          <Card className="p-5 mb-4">
            <div className="flex flex-col sm:flex-row sm:items-start gap-4">
              <div className="flex-1">
                <div className="flex flex-wrap items-center gap-2 mb-2">
                  <Badge variant={sessionStatusVariant[session.status]} dot pulse={session.status === 'running'}>
                    {session.status.charAt(0).toUpperCase() + session.status.slice(1)}
                  </Badge>
                  <span className={`text-xs px-2 py-0.5 rounded-md font-mono font-medium ${modelBadge(session.model)}`}>
                    {session.model}
                  </span>
                  <span className="text-xs px-2 py-0.5 rounded-md font-mono bg-muted text-muted-foreground">{session.branch}</span>
                  <span className="text-xs text-muted-foreground">{session.agent}</span>
                </div>
                <p className="text-sm text-muted-foreground leading-relaxed">{session.summary}</p>
              </div>
              <div className="shrink-0 space-y-1.5 text-right">
                <p className="text-xs text-muted-foreground">Started {formatDateTime(session.startedAt)}</p>
                {session.endedAt && <p className="text-xs text-muted-foreground">Ended {formatDateTime(session.endedAt)}</p>}
                {session.duration && <p className="text-xs font-medium text-foreground">{session.duration}</p>}
              </div>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-5 gap-3 mt-4 pt-4 border-t border-border">
              {[
                { label: 'Events', value: session.eventCount, color: 'text-foreground' },
                { label: 'Decisions', value: session.decisions, color: 'text-info' },
                { label: 'Discoveries', value: session.discoveries, color: 'text-success' },
                { label: 'Failures', value: session.failures, color: 'text-destructive' },
                { label: 'Constraints', value: session.constraints ?? 0, color: 'text-warning' },
              ].map((m) => (
                <div key={m.label} className="bg-muted/50 rounded-lg p-3">
                  <div className={`text-xl font-bold ${m.color}`}>{m.value}</div>
                  <div className="text-xs text-muted-foreground mt-0.5">{m.label}</div>
                </div>
              ))}
            </div>
          </Card>

          {/* Events */}
          <Card>
            <div className="px-5 py-4 border-b border-border">
              <h2 className="font-semibold text-sm text-foreground">Session Events</h2>
              <p className="text-xs text-muted-foreground mt-0.5">
                {eventsState.status === 'loading' ? 'Loading…' : `${sessionEvents.length} events recorded`}
              </p>
            </div>
            {eventsState.status === 'loading' && Array.from({ length: 4 }).map((_, i) => <SkeletonRow key={i} />)}
            {sessionEvents.length === 0 && eventsState.status === 'ok' && (
              <EmptyState title="No events for this session" />
            )}
            <div className="divide-y divide-border">
              {sessionEvents.map((event) => (
                <div key={event.id} className="px-5 py-3.5 hover:bg-muted/30 transition-colors">
                  <div className="flex items-start gap-3">
                    <div className={`w-2 h-2 rounded-full mt-1.5 shrink-0 ${
                      event.type === 'decision' ? 'bg-info' : event.type === 'discovery' ? 'bg-success' :
                      event.type === 'failure' ? 'bg-destructive' : event.type === 'constraint' ? 'bg-warning' : 'bg-primary'
                    }`} />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-0.5">
                        <Badge variant={eventVariant[event.type]}>{event.type}</Badge>
                        <span className="text-xs text-muted-foreground">{formatDateTime(event.timestamp)}</span>
                      </div>
                      <p className="text-sm font-medium text-foreground">{event.title}</p>
                      <p className="text-xs text-muted-foreground mt-0.5 leading-relaxed">{event.description}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </>
      )}
    </div>
  );
}

export default function Sessions() {
  const navigate = useNavigate();
  const { state, reload } = useApi(getSessions, []);

  const sessions = state.status === 'ok' ? state.data : [];

  const statusCounts = sessions.reduce<Record<string, number>>((acc, s) => {
    acc[s.status] = (acc[s.status] ?? 0) + 1;
    return acc;
  }, {});

  return (
    <div className="max-w-4xl mx-auto">
      <PageHeader
        title="Sessions"
        description={
          state.status === 'ok'
            ? `${statusCounts['running'] ?? 0} active · ${statusCounts['completed'] ?? 0} completed · ${statusCounts['failed'] ?? 0} failed`
            : 'Agent work sessions'
        }
      />

      {state.status === 'loading' && (
        <div className="space-y-3">
          {Array.from({ length: 4 }).map((_, i) => <SkeletonCard key={i} className="h-28" />)}
        </div>
      )}
      {state.status === 'error' && <ErrorState message={state.error} onRetry={reload} />}
      {state.status === 'ok' && sessions.length === 0 && (
        <EmptyState
          title="No sessions yet"
          description="Agent sessions will appear here once started. Run `agent-ledger start-session` to begin."
        />
      )}
      {state.status === 'ok' && (
        <div className="space-y-3">
          {sessions.map((session: Session) => (
            <button
              key={session.id}
              onClick={() => navigate(`/sessions/${session.id}`)}
              className="w-full text-left bg-card border border-border rounded-xl p-5 hover:border-foreground/20 hover:shadow-sm transition-all"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex-1 min-w-0">
                  <div className="flex flex-wrap items-center gap-2 mb-1.5">
                    <h3 className="text-sm font-semibold text-foreground">{session.name}</h3>
                    <Badge variant={sessionStatusVariant[session.status]} dot pulse={session.status === 'running'}>
                      {session.status.charAt(0).toUpperCase() + session.status.slice(1)}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground leading-snug line-clamp-1 mb-2.5">{session.summary}</p>
                  <div className="flex flex-wrap items-center gap-2">
                    <span className={`text-xs px-2 py-0.5 rounded-md font-mono font-medium ${modelBadge(session.model)}`}>
                      {session.model}
                    </span>
                    <span className="text-xs px-2 py-0.5 rounded-md font-mono bg-muted text-muted-foreground">{session.branch}</span>
                    <span className="text-xs text-muted-foreground">{session.agent}</span>
                  </div>
                </div>

                <div className="flex flex-col items-end gap-2 shrink-0">
                  <p className="text-xs text-muted-foreground">{formatRelative(session.startedAt)}</p>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <span>{session.eventCount} events</span>
                    {session.duration && <span>· {session.duration}</span>}
                  </div>
                  {/* Event type bar */}
                  <div className="flex gap-0.5 items-center mt-0.5">
                    {(['decision', 'discovery', 'failure', 'constraint', 'checkpoint'] as const).map((type) => {
                      const count = type === 'decision' ? session.decisions :
                        type === 'discovery' ? session.discoveries : type === 'failure' ? session.failures :
                        type === 'constraint' ? (session.constraints ?? 0) : (session.checkpoints ?? 0);
                      if (!count) return null;
                      return (
                        <div
                          key={type}
                          title={`${count} ${type}(s)`}
                          className={`h-1.5 rounded-full ${statusDot[type] || 'bg-primary'}`}
                          style={{ width: Math.max(count * 3, 6) }}
                        />
                      );
                    })}
                  </div>
                </div>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
