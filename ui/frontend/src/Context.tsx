import { useState } from 'react';
import { getContext } from './api/client';
import { useApi } from './hooks/useApi';
import {
  PageHeader, Card, SectionLabel, SkeletonRow, ErrorState, Badge,
} from './components/ui';

export default function ProjectContext() {
  const [taskQuery, setTaskQuery] = useState('');
  const [copiedPrompt, setCopiedPrompt] = useState(false);

  const { state, reload } = useApi(() => getContext(taskQuery), [taskQuery]);

  const ctxData = state.status === 'ok' ? state.data : null;

  const handleCopyMarkdown = () => {
    if (!ctxData) return;
    let markdown = `# PROJECT CONTEXT\n\n`;

    if (ctxData.Project) markdown += `## Project Facts\n\`\`\`\n${ctxData.Project}\n\`\`\`\n\n`;
    if (ctxData.CurrentState) markdown += `## Current Git State\n\`\`\`\n${ctxData.CurrentState}\n\`\`\`\n\n`;
    if (ctxData.Architecture) markdown += `## Architecture & Packages\n\`\`\`\n${ctxData.Architecture}\n\`\`\`\n\n`;

    if (ctxData.ImportantDecisions?.length) {
      markdown += `## Key Decisions\n`;
      ctxData.ImportantDecisions.forEach((d) => {
        markdown += `- **${d.Title}**: ${d.Decision} (Rationale: ${d.Rationale || 'N/A'})\n`;
      });
      markdown += `\n`;
    }

    if (ctxData.Discoveries?.length) {
      markdown += `## Discoveries\n`;
      ctxData.Discoveries.forEach((d) => {
        markdown += `- **${d.Title}**: ${d.Finding}\n`;
      });
      markdown += `\n`;
    }

    if (ctxData.Constraints?.length) {
      markdown += `## Constraints & Rules\n`;
      ctxData.Constraints.forEach((c) => {
        markdown += `- **${c.Title}**: ${c.Constraint}\n`;
      });
      markdown += `\n`;
    }

    if (ctxData.LatestHandoff) {
      markdown += `## Latest Handoff Notes\n${ctxData.LatestHandoff}\n\n`;
    }

    if (ctxData.Recommendations?.length) {
      markdown += `## Recommendations\n`;
      ctxData.Recommendations.forEach((r) => {
        markdown += `- ${r}\n`;
      });
      markdown += `\n`;
    }

    navigator.clipboard.writeText(markdown);
    setCopiedPrompt(true);
    setTimeout(() => setCopiedPrompt(false), 2000);
  };

  return (
    <div className="max-w-5xl mx-auto space-y-5 pb-10">
      <PageHeader
        title="Project Context"
        description="Compiled development context, architecture facts, and agent knowledge."
        actions={
          <button
            onClick={handleCopyMarkdown}
            disabled={state.status !== 'ok'}
            className="text-xs flex items-center gap-2 px-3.5 py-2 rounded-lg bg-primary text-white hover:bg-primary-hover transition-colors font-medium cursor-pointer shadow-xs disabled:opacity-50"
          >
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <rect x="4" y="4" width="8" height="8" rx="1.5" stroke="currentColor" strokeWidth="1.3"/>
              <path d="M2.5 9.5H2a1 1 0 0 1-1-1v-6a1 1 0 0 1 1-1h6a1 1 0 0 1 1 1v.5" stroke="currentColor" strokeWidth="1.3"/>
            </svg>
            {copiedPrompt ? '✓ Copied Markdown Context!' : 'Copy Context for AI Prompt'}
          </button>
        }
      />

      {/* Task Focus Input */}
      <Card className="p-4">
        <div className="flex flex-col sm:flex-row items-start sm:items-center gap-3">
          <div className="flex-1 min-w-0">
            <p className="text-xs font-semibold text-foreground">Task-Specific Context Filter</p>
            <p className="text-xs text-muted-foreground mt-0.5">
              Focus compiled context on a specific task or feature area
            </p>
          </div>
          <div className="w-full sm:w-80">
            <input
              type="text"
              value={taskQuery}
              onChange={(e) => setTaskQuery(e.target.value)}
              placeholder="e.g. refactor knowledge graph, fix auth..."
              className="w-full px-3 py-1.5 rounded-lg border border-border bg-muted text-xs text-foreground placeholder:text-muted-foreground outline-none focus:border-primary transition-colors"
            />
          </div>
        </div>
      </Card>

      {state.status === 'loading' && (
        <div className="space-y-4">
          <Card className="p-5 space-y-3">
            <SkeletonRow />
            <SkeletonRow />
          </Card>
        </div>
      )}

      {state.status === 'error' && <ErrorState message={state.error} onRetry={reload} />}

      {state.status === 'ok' && ctxData && (
        <div className="space-y-5">
          {/* Objective Facts & Git State Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
            {/* Git & Repo Facts */}
            {ctxData.Project && (
              <Card className="p-5">
                <div className="flex items-center justify-between mb-3">
                  <SectionLabel>Repository & Project Facts</SectionLabel>
                  <Badge variant="info">Objective Facts</Badge>
                </div>
                <pre className="text-xs text-foreground font-mono leading-relaxed bg-muted/50 p-3.5 rounded-lg border border-border overflow-x-auto whitespace-pre-wrap">
                  {ctxData.Project}
                </pre>
              </Card>
            )}

            {/* Current State */}
            {ctxData.CurrentState && (
              <Card className="p-5">
                <div className="flex items-center justify-between mb-3">
                  <SectionLabel>Current Git & File State</SectionLabel>
                  <Badge variant="warning">Working Tree</Badge>
                </div>
                <pre className="text-xs text-foreground font-mono leading-relaxed bg-muted/50 p-3.5 rounded-lg border border-border overflow-x-auto whitespace-pre-wrap">
                  {ctxData.CurrentState}
                </pre>
              </Card>
            )}
          </div>

          {/* Architecture & Packages */}
          {ctxData.Architecture && (
            <Card className="p-5">
              <div className="flex items-center justify-between mb-3">
                <SectionLabel>Architecture & Package Structure</SectionLabel>
                <Badge variant="primary">Codebase Map</Badge>
              </div>
              <pre className="text-xs text-foreground font-mono leading-relaxed bg-muted/50 p-4 rounded-lg border border-border overflow-x-auto whitespace-pre-wrap max-h-72">
                {ctxData.Architecture}
              </pre>
            </Card>
          )}

          {/* Recorded Knowledge Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Decisions */}
            {ctxData.ImportantDecisions && ctxData.ImportantDecisions.length > 0 && (
              <Card className="p-5 space-y-3">
                <SectionLabel>Key Decisions ({ctxData.ImportantDecisions.length})</SectionLabel>
                <div className="space-y-2.5">
                  {ctxData.ImportantDecisions.map((d, i) => (
                    <div key={i} className="p-3 bg-muted/40 rounded-lg border border-border space-y-1">
                      <p className="text-xs font-semibold text-foreground">{d.Title}</p>
                      <p className="text-xs text-muted-foreground">{d.Decision}</p>
                      {d.Rationale && (
                        <p className="text-[11px] text-muted-foreground font-mono">Rationale: {d.Rationale}</p>
                      )}
                    </div>
                  ))}
                </div>
              </Card>
            )}

            {/* Constraints */}
            {ctxData.Constraints && ctxData.Constraints.length > 0 && (
              <Card className="p-5 space-y-3">
                <SectionLabel>Active Constraints & Rules ({ctxData.Constraints.length})</SectionLabel>
                <div className="space-y-2.5">
                  {ctxData.Constraints.map((c, i) => (
                    <div key={i} className="p-3 bg-muted/40 rounded-lg border border-border space-y-1">
                      <p className="text-xs font-semibold text-destructive">{c.Title}</p>
                      <p className="text-xs text-muted-foreground">{c.Constraint}</p>
                    </div>
                  ))}
                </div>
              </Card>
            )}
          </div>

          {/* Latest Handoff */}
          {ctxData.LatestHandoff && (
            <Card className="p-5">
              <SectionLabel>Latest Agent Handoff</SectionLabel>
              <div className="mt-2 text-xs text-foreground font-mono leading-relaxed bg-muted/50 p-4 rounded-lg border border-border whitespace-pre-wrap">
                {ctxData.LatestHandoff}
              </div>
            </Card>
          )}

          {/* Recommendations */}
          {ctxData.Recommendations && ctxData.Recommendations.length > 0 && (
            <Card className="p-5">
              <SectionLabel>Agent Guidance & Recommendations</SectionLabel>
              <ul className="mt-2 space-y-1.5 list-disc list-inside text-xs text-foreground">
                {ctxData.Recommendations.map((r, i) => (
                  <li key={i} className="leading-relaxed">{r}</li>
                ))}
              </ul>
            </Card>
          )}
        </div>
      )}
    </div>
  );
}
