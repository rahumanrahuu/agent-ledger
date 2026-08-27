import { useState, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import { getMemories } from './api/client';
import { useApi } from './hooks/useApi';
import type { Memory, MemoryType, ImportanceLevel } from './api/types';
import {
  Badge, memoryVariant, importanceVariant, SearchInput, FilterPill,
  InspectorPanel, SectionLabel, TagList, MetaRow,
  SkeletonRow, ErrorState, EmptyState, PageHeader,
  formatRelative, formatDateTime,
} from './components/ui';

const allTypes: MemoryType[] = ['fact', 'rule', 'pattern', 'entity', 'preference', 'insight'];
const allImportances: ImportanceLevel[] = ['critical', 'high', 'medium', 'low'];

const importanceOrder: Record<ImportanceLevel, number> = { critical: 0, high: 1, medium: 2, low: 3 };

function importanceDot(imp: ImportanceLevel) {
  switch (imp) {
    case 'critical': return 'bg-destructive';
    case 'high': return 'bg-warning';
    case 'medium': return 'bg-info';
    default: return 'bg-muted-foreground';
  }
}

export default function Memories() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [localSearch, setLocalSearch] = useState('');
  const [selectedTypes, setSelectedTypes] = useState<MemoryType[]>([]);
  const [selectedImportances, setSelectedImportances] = useState<ImportanceLevel[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(searchParams.get('id'));

  const { state, reload } = useApi(getMemories, []);

  const memories = state.status === 'ok' ? state.data : [];

  const filtered = useMemo(() => {
    let result = [...memories].sort(
      (a, b) => importanceOrder[a.importance] - importanceOrder[b.importance]
    );
    if (localSearch) {
      const q = localSearch.toLowerCase();
      result = result.filter(
        (m) => m.content.toLowerCase().includes(q) || m.tags.some((t) => t.toLowerCase().includes(q))
      );
    }
    if (selectedTypes.length) result = result.filter((m) => selectedTypes.includes(m.type as MemoryType));
    if (selectedImportances.length) result = result.filter((m) => selectedImportances.includes(m.importance));
    return result;
  }, [memories, localSearch, selectedTypes, selectedImportances]);

  const selected = memories.find((m) => m.id === selectedId) ?? null;

  function toggleType(t: MemoryType) {
    setSelectedTypes((prev) => prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]);
  }
  function toggleImportance(i: ImportanceLevel) {
    setSelectedImportances((prev) => prev.includes(i) ? prev.filter((x) => x !== i) : [...prev, i]);
  }
  function selectMemory(m: Memory) {
    setSelectedId(m.id);
    setSearchParams({ id: m.id });
  }
  function closeInspector() {
    setSelectedId(null);
    setSearchParams({});
  }

  return (
    <div className="max-w-6xl mx-auto flex flex-col h-full">
      <PageHeader
        title="Memories"
        description={state.status === 'ok' ? `${memories.length} memory entries` : 'Agent knowledge store'}
      />

      <div className="flex gap-4 flex-1 min-h-0" style={{ height: 'calc(100vh - 12rem)' }}>
        {/* Left panel */}
        <div className="w-96 shrink-0 flex flex-col bg-card border border-border rounded-xl overflow-hidden">
          {/* Search + filters */}
          <div className="p-3 border-b border-border space-y-2">
            <SearchInput
              value={localSearch}
              onChange={setLocalSearch}
              placeholder="Search memories…"
            />
            <div className="flex flex-wrap gap-1">
              {allTypes.map((t) => (
                <FilterPill
                  key={t} label={t} active={selectedTypes.includes(t)}
                  variant={memoryVariant[t]} onClick={() => toggleType(t)}
                />
              ))}
            </div>
            <div className="flex flex-wrap gap-1">
              {allImportances.map((imp) => (
                <FilterPill
                  key={imp} label={imp} active={selectedImportances.includes(imp)}
                  variant={importanceVariant[imp]} onClick={() => toggleImportance(imp)}
                />
              ))}
            </div>
          </div>

          {/* List */}
          <div className="flex-1 overflow-y-auto divide-y divide-border">
            {state.status === 'loading' && Array.from({ length: 6 }).map((_, i) => <SkeletonRow key={i} />)}
            {state.status === 'error' && (
              <ErrorState message={state.error} onRetry={reload} />
            )}
            {state.status === 'ok' && filtered.length === 0 && (
              <EmptyState
                title={memories.length === 0 ? 'No memories yet' : 'No matches'}
                description={
                  memories.length === 0
                    ? 'Agent decisions, discoveries, and other durable knowledge will appear here as they are recorded.'
                    : 'Try adjusting your search or filters.'
                }
              />
            )}
            {state.status === 'ok' && filtered.map((memory) => (
              <button
                key={memory.id}
                onClick={() => selectMemory(memory)}
                className={`w-full text-left px-4 py-3.5 hover:bg-muted/40 transition-colors ${
                  selected?.id === memory.id ? 'bg-primary-light/40 border-l-2 border-primary' : ''
                }`}
              >
                <div className="flex items-center gap-2 mb-1">
                  <div className={`w-1.5 h-1.5 rounded-full shrink-0 ${importanceDot(memory.importance)}`} />
                  <Badge variant={memoryVariant[memory.type as MemoryType]}>{memory.type}</Badge>
                  <span className="text-xs text-muted-foreground ml-auto shrink-0">{formatRelative(memory.timestamp)}</span>
                </div>
                <p className="text-sm text-foreground leading-snug line-clamp-2">{memory.content}</p>
                <div className="flex flex-wrap gap-1 mt-1.5">
                  {memory.tags.slice(0, 3).map((tag) => (
                    <span key={tag} className="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded font-mono">{tag}</span>
                  ))}
                  {memory.tags.length > 3 && (
                    <span className="text-xs text-muted-foreground">+{memory.tags.length - 3}</span>
                  )}
                </div>
              </button>
            ))}
          </div>
        </div>

        {/* Inspector */}
        {selected ? (
          <InspectorPanel
            title={selected.type.charAt(0).toUpperCase() + selected.type.slice(1)}
            subtitle={`Recorded ${formatDateTime(selected.timestamp)}`}
            badge={<Badge variant={importanceVariant[selected.importance]}>{selected.importance} importance</Badge>}
            onClose={closeInspector}
            width="flex-1"
          >
            <div>
              <SectionLabel>Content</SectionLabel>
              <p className="text-sm text-foreground leading-relaxed bg-muted/40 rounded-lg p-4 border border-border">
                {selected.content}
              </p>
            </div>

            {selected.source && (
              <div>
                <SectionLabel>Source</SectionLabel>
                <p className="text-xs text-muted-foreground italic">{selected.source}</p>
              </div>
            )}

            <div>
              <SectionLabel>Details</SectionLabel>
              <div className="rounded-lg border border-border overflow-hidden">
                <MetaRow label="Type" value={<Badge variant={memoryVariant[selected.type as MemoryType]}>{selected.type}</Badge>} />
                <MetaRow label="Importance" value={<Badge variant={importanceVariant[selected.importance]}>{selected.importance}</Badge>} />
                <MetaRow label="Session" value={selected.sessionName ?? selected.sessionId} />
                <MetaRow label="Recorded" value={formatDateTime(selected.timestamp)} />
                {selected.relevance !== undefined && (
                  <MetaRow label="Relevance" value={`${(selected.relevance * 100).toFixed(0)}%`} />
                )}
              </div>
            </div>

            {selected.tags.length > 0 && (
              <div>
                <SectionLabel>Tags</SectionLabel>
                <TagList tags={selected.tags} />
              </div>
            )}
          </InspectorPanel>
        ) : (
          <div className="flex-1 bg-card border border-border rounded-xl flex flex-col items-center justify-center text-muted-foreground">
            <svg width="28" height="28" viewBox="0 0 28 28" fill="none" className="mb-2 opacity-25">
              <path d="M14 2C9.6 2 6 5.6 6 10c0 2.4 1.1 4.5 2.8 5.9L8 22h12l-.8-6.1A7.97 7.97 0 0 0 22 10c0-4.4-3.6-8-8-8Z" stroke="currentColor" strokeWidth="1.5"/>
            </svg>
            <p className="text-sm">Select a memory to inspect</p>
          </div>
        )}
      </div>
    </div>
  );
}
