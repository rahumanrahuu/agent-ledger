import { useState, useMemo } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { getEvents } from './api/client';
import { useApi } from './hooks/useApi';
import type { EventType } from './api/types';
import {
  Badge, eventVariant, FilterPill,
  InspectorPanel, SectionLabel, MetaRow,
  SkeletonRow, ErrorState, EmptyState, PageHeader,
  formatRelative, formatDateTime,
} from './components/ui';

const allTypes: EventType[] = ['decision', 'discovery', 'failure', 'constraint', 'checkpoint'];

const typeLabels: Record<EventType, string> = {
  decision: 'Decision', discovery: 'Discovery', failure: 'Failure',
  constraint: 'Constraint', checkpoint: 'Checkpoint',
};

const typeDot: Record<EventType, string> = {
  decision: 'bg-info', discovery: 'bg-success', failure: 'bg-destructive',
  constraint: 'bg-warning', checkpoint: 'bg-primary',
};

const typeLineColor: Record<EventType, string> = {
  decision: 'border-info/40', discovery: 'border-success/40',
  failure: 'border-destructive/40', constraint: 'border-warning/40', checkpoint: 'border-primary/40',
};

export default function Timeline() {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const typeParam = searchParams.get('type') as EventType | null;
  const eventParam = searchParams.get('event');

  const [activeTypes, setActiveTypes] = useState<EventType[]>(typeParam ? [typeParam] : []);
  const [copiedPrompt, setCopiedPrompt] = useState(false);
  const { state, reload } = useApi(getEvents, []);

  const events = state.status === 'ok' ? state.data : [];

  const filtered = useMemo(() => {
    const sorted = [...events].sort(
      (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
    );
    if (!activeTypes.length) return sorted;
    return sorted.filter((e) => activeTypes.includes(e.type));
  }, [events, activeTypes]);

  const selectedEvent = events.find((e) => e.id === eventParam) ?? null;

  function toggleType(t: EventType) {
    const next = activeTypes.includes(t) ? activeTypes.filter((x) => x !== t) : [...activeTypes, t];
    setActiveTypes(next);
    if (next.length === 1) setSearchParams({ type: next[0] });
    else setSearchParams({});
  }

  function selectEvent(id: string) {
    setSearchParams(activeTypes.length === 1 ? { type: activeTypes[0], event: id } : { event: id });
  }

  function closeInspector() {
    const p: Record<string, string> = {};
    if (activeTypes.length === 1) p.type = activeTypes[0];
    setSearchParams(p);
  }

  const typeCounts = useMemo(() => {
    const c: Record<string, number> = {};
    events.forEach((e) => { c[e.type] = (c[e.type] ?? 0) + 1; });
    return c;
  }, [events]);

  return (
    <div className="max-w-5xl mx-auto flex flex-col">
      <PageHeader
        title="Timeline"
        description={state.status === 'ok' ? `${events.length} events across all sessions` : 'Chronological event feed'}
      />

      {/* Filters */}
      <div className="flex items-center gap-2 mb-6 flex-wrap">
        <FilterPill label="All" count={events.length} active={activeTypes.length === 0} onClick={() => { setActiveTypes([]); setSearchParams({}); }} />
        {allTypes.map((type) => (
          <FilterPill
            key={type} label={typeLabels[type]}
            count={typeCounts[type] ?? 0}
            active={activeTypes.includes(type)}
            variant={eventVariant[type]}
            onClick={() => toggleType(type)}
          />
        ))}
        {activeTypes.length > 0 && (
          <button onClick={() => { setActiveTypes([]); setSearchParams({}); }} className="text-xs text-muted-foreground hover:text-foreground px-2 py-1">
            Clear
          </button>
        )}
      </div>

      <div className="flex gap-4 flex-1">
        {/* Feed */}
        <div className="flex-1 min-w-0">
          {state.status === 'loading' && (
            <div className="space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="flex gap-4">
                  <div className="w-3 h-3 rounded-full bg-muted mt-4 shrink-0 animate-pulse" />
                  <div className="flex-1 bg-card border border-border rounded-xl p-4 space-y-2">
                    <SkeletonRow />
                  </div>
                </div>
              ))}
            </div>
          )}
          {state.status === 'error' && <ErrorState message={state.error} onRetry={reload} />}
          {state.status === 'ok' && filtered.length === 0 && (
            <EmptyState title="No events" description="No events match the current filters." />
          )}
          {state.status === 'ok' && (
            <div className="relative">
              <div className="absolute left-3.5 top-0 bottom-0 w-px bg-border" />
              <div className="space-y-1">
                {filtered.map((event) => {
                  const isSelected = event.id === eventParam;
                  return (
                    <div key={event.id} className="relative pl-12">
                      <div className={`absolute left-2.5 top-4 w-2.5 h-2.5 rounded-full border-2 border-card ${typeDot[event.type]}`} />
                      <button
                        onClick={() => selectEvent(event.id)}
                        className={`w-full text-left bg-card border rounded-xl px-4 py-3.5 transition-all hover:border-foreground/20 ${
                          isSelected ? 'border-foreground/20 shadow-sm' : 'border-border'
                        }`}
                      >
                        <div className="flex flex-wrap items-center gap-2 mb-1.5">
                          <Badge variant={eventVariant[event.type]}>{typeLabels[event.type]}</Badge>
                          <span className="text-xs text-muted-foreground">{formatDateTime(event.timestamp)}</span>
                          {event.sessionName && (
                            <button
                              onClick={(ev) => { ev.stopPropagation(); navigate(`/sessions/${event.sessionId}`); }}
                              className="ml-auto text-xs text-muted-foreground hover:text-primary font-mono"
                            >
                              {event.sessionName}
                            </button>
                          )}
                        </div>
                        <p className="text-sm font-semibold text-foreground leading-snug">{event.title}</p>
                        {!isSelected && (
                          <p className="text-xs text-muted-foreground mt-0.5 line-clamp-1">{event.description}</p>
                        )}
                        {isSelected && (
                          <div className="mt-2.5 space-y-2.5">
                            <p className="text-sm text-muted-foreground leading-relaxed">{event.description}</p>
                            {event.metadata && Object.keys(event.metadata).length > 0 && (
                              <div className="bg-muted/50 rounded-lg p-3 border border-border">
                                <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">Metadata</p>
                                <div className="grid grid-cols-2 gap-1.5">
                                  {Object.entries(event.metadata).map(([k, v]) => (
                                    <div key={k} className="text-xs">
                                      <span className="text-muted-foreground font-mono">{k}: </span>
                                      <span className="text-foreground font-mono font-medium">{v}</span>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            )}
                          </div>
                        )}
                      </button>
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </div>

        {/* Inspector (only on selected event with full detail) */}
        {selectedEvent && (
          <InspectorPanel
            title={selectedEvent.title}
            badge={<Badge variant={eventVariant[selectedEvent.type]}>{typeLabels[selectedEvent.type]}</Badge>}
            subtitle={formatDateTime(selectedEvent.timestamp)}
            onClose={closeInspector}
          >
            <div>
              <div className="flex items-center justify-between mb-2">
                <SectionLabel>Description</SectionLabel>
                <button
                  onClick={() => {
                    const promptText = `[Agent Event - ${selectedEvent.type.toUpperCase()}] Title: ${selectedEvent.title}\nDescription: ${selectedEvent.description}\nTimestamp: ${selectedEvent.timestamp}`;
                    navigator.clipboard.writeText(promptText);
                    setCopiedPrompt(true);
                    setTimeout(() => setCopiedPrompt(false), 2000);
                  }}
                  className="text-xs flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-primary-light text-primary hover:bg-primary/20 transition-colors font-medium cursor-pointer"
                >
                  <svg width="12" height="12" viewBox="0 0 14 14" fill="none">
                    <rect x="4" y="4" width="8" height="8" rx="1.5" stroke="currentColor" strokeWidth="1.3"/>
                    <path d="M2.5 9.5H2a1 1 0 0 1-1-1v-6a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1v.5" stroke="currentColor" strokeWidth="1.3"/>
                  </svg>
                  {copiedPrompt ? '✓ Copied Context!' : 'Copy for AI Prompt'}
                </button>
              </div>
              <p className="text-sm text-foreground leading-relaxed bg-muted/40 rounded-lg p-3 border border-border">{selectedEvent.description}</p>
            </div>
            <div>
              <SectionLabel>Details</SectionLabel>
              <div className="rounded-lg border border-border overflow-hidden">
                <MetaRow label="Type" value={<Badge variant={eventVariant[selectedEvent.type]}>{typeLabels[selectedEvent.type]}</Badge>} />
                <MetaRow label="Session" value={
                  <button onClick={() => navigate(`/sessions/${selectedEvent.sessionId}`)} className="text-primary hover:underline">
                    {selectedEvent.sessionName ?? selectedEvent.sessionId}
                  </button>
                } />
                <MetaRow label="Timestamp" value={formatDateTime(selectedEvent.timestamp)} />
                <MetaRow label="ID" value={<span className="font-mono">{selectedEvent.id}</span>} />
              </div>
            </div>
            {selectedEvent.metadata && Object.keys(selectedEvent.metadata).length > 0 && (
              <div>
                <SectionLabel>Metadata</SectionLabel>
                <div className="rounded-lg border border-border overflow-hidden">
                  {Object.entries(selectedEvent.metadata).map(([k, v]) => (
                    <MetaRow key={k} label={k} value={<span className="font-mono">{v}</span>} />
                  ))}
                </div>
              </div>
            )}
          </InspectorPanel>
        )}
      </div>
    </div>
  );
}
