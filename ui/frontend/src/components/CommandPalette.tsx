import { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { search as apiSearch } from '../api/client';
import type { SearchResult } from '../api/types';
import type { Theme } from '../hooks/useTheme';

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  theme: Theme;
  setTheme: (t: Theme) => void;
}

interface CommandItem {
  id: string;
  category: 'Navigation' | 'Filter' | 'Action' | 'Search Result';
  title: string;
  subtitle?: string;
  icon?: string;
  action: () => void;
}

export default function CommandPalette({ isOpen, onClose, theme, setTheme }: CommandPaletteProps) {
  const [query, setQuery] = useState('');
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const navigate = useNavigate();
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isOpen) {
      setTimeout(() => inputRef.current?.focus(), 50);
      setQuery('');
      setSearchResults([]);
      setSelectedIndex(0);
    }
  }, [isOpen]);

  useEffect(() => {
    if (!query.trim()) {
      setSearchResults([]);
      return;
    }
    const timer = setTimeout(async () => {
      try {
        const res = await apiSearch(query);
        setSearchResults(res);
      } catch (err) {
        console.error('Search error:', err);
      }
    }, 150);
    return () => clearTimeout(timer);
  }, [query]);

  const defaultCommands: CommandItem[] = [
    {
      id: 'nav-overview',
      category: 'Navigation',
      title: 'Go to Overview',
      subtitle: 'Dashboard KPI stats & recent activity',
      icon: '⬡',
      action: () => { navigate('/overview'); onClose(); },
    },
    {
      id: 'nav-context',
      category: 'Navigation',
      title: 'Go to Project Context',
      subtitle: 'Compiled development context, architecture & facts',
      icon: '📋',
      action: () => { navigate('/context'); onClose(); },
    },
    {
      id: 'nav-memories',
      category: 'Navigation',
      title: 'Go to Memories',
      subtitle: 'Agent facts, rules & insights',
      icon: '◎',
      action: () => { navigate('/memories'); onClose(); },
    },
    {
      id: 'nav-timeline',
      category: 'Navigation',
      title: 'Go to Timeline',
      subtitle: 'Chronological development event log',
      icon: '◈',
      action: () => { navigate('/timeline'); onClose(); },
    },
    {
      id: 'nav-sessions',
      category: 'Navigation',
      title: 'Go to Sessions',
      subtitle: 'Agent sessions & status breakdown',
      icon: '⬡',
      action: () => { navigate('/sessions'); onClose(); },
    },
    {
      id: 'nav-graph',
      category: 'Navigation',
      title: 'Go to Knowledge Graph',
      subtitle: 'Interactive relationship node graph',
      icon: '◉',
      action: () => { navigate('/knowledge-graph'); onClose(); },
    },
    {
      id: 'filter-decisions',
      category: 'Filter',
      title: 'View Decisions',
      subtitle: 'Filter timeline by agent decisions',
      icon: '◆',
      action: () => { navigate('/timeline?type=decision'); onClose(); },
    },
    {
      id: 'filter-discoveries',
      category: 'Filter',
      title: 'View Discoveries',
      subtitle: 'Filter timeline by discoveries & findings',
      icon: '◉',
      action: () => { navigate('/timeline?type=discovery'); onClose(); },
    },
    {
      id: 'filter-failures',
      category: 'Filter',
      title: 'View Failures',
      subtitle: 'Filter timeline by errors & failures',
      icon: '✕',
      action: () => { navigate('/timeline?type=failure'); onClose(); },
    },
    {
      id: 'filter-constraints',
      category: 'Filter',
      title: 'View Constraints',
      subtitle: 'Filter timeline by rules & constraints',
      icon: '⊘',
      action: () => { navigate('/timeline?type=constraint'); onClose(); },
    },
    {
      id: 'theme-toggle',
      category: 'Action',
      title: `Switch Theme (Current: ${theme})`,
      subtitle: 'Cycle between Light, Dark, and System theme',
      icon: theme === 'dark' ? '🌙' : '☀️',
      action: () => {
        const nextTheme = theme === 'dark' ? 'light' : theme === 'light' ? 'system' : 'dark';
        setTheme(nextTheme);
        onClose();
      },
    },
  ];

  const searchCommands: CommandItem[] = searchResults.map((r) => ({
    id: `sr-${r.id}`,
    category: 'Search Result',
    title: r.title,
    subtitle: `${r.type.toUpperCase()} · ${r.snippet || ''}`,
    icon: r.type === 'session' ? '⬡' : r.type === 'decision' ? '◆' : r.type === 'failure' ? '✕' : '◎',
    action: () => {
      if (r.type === 'session') navigate(`/sessions/${r.id}`);
      else if (['decision', 'discovery', 'failure', 'constraint', 'checkpoint'].includes(r.type))
        navigate(`/timeline?event=${r.id}`);
      else navigate(`/memories?id=${r.id}`);
      onClose();
    },
  }));

  const allFilteredCommands = query.trim()
    ? [
        ...searchCommands,
        ...defaultCommands.filter(
          (c) =>
            c.title.toLowerCase().includes(query.toLowerCase()) ||
            c.subtitle?.toLowerCase().includes(query.toLowerCase())
        ),
      ]
    : defaultCommands;

  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((prev) => (prev + 1) % Math.max(1, allFilteredCommands.length));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((prev) => (prev - 1 + allFilteredCommands.length) % Math.max(1, allFilteredCommands.length));
    } else if (e.key === 'Enter') {
      e.preventDefault();
      if (allFilteredCommands[selectedIndex]) {
        allFilteredCommands[selectedIndex].action();
      }
    } else if (e.key === 'Escape') {
      e.preventDefault();
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs flex items-start justify-center pt-20 px-4 animate-in fade-in duration-150"
      onClick={onClose}
    >
      <div
        className="w-full max-w-xl bg-card border border-border rounded-xl shadow-2xl overflow-hidden flex flex-col"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        {/* Search header */}
        <div className="flex items-center gap-3 px-4 py-3 border-b border-border bg-muted/30">
          <svg width="16" height="16" viewBox="0 0 14 14" fill="none" className="text-muted-foreground shrink-0">
            <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.4" />
            <path d="m9.5 9.5 2.5 2.5" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" />
          </svg>
          <input
            ref={inputRef}
            type="text"
            placeholder="Type a command, page, or search query..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="flex-1 bg-transparent text-sm text-foreground placeholder:text-muted-foreground outline-none font-medium"
          />
          <kbd className="text-[10px] font-mono bg-muted text-muted-foreground px-1.5 py-0.5 rounded border border-border">
            ESC
          </kbd>
        </div>

        {/* List of items */}
        <div className="max-h-80 overflow-y-auto p-2 divide-y divide-border/40">
          {allFilteredCommands.length === 0 ? (
            <div className="px-4 py-8 text-center text-xs text-muted-foreground">
              No matching commands or search results found.
            </div>
          ) : (
            allFilteredCommands.map((cmd, idx) => {
              const isSelected = idx === selectedIndex;
              return (
                <button
                  key={cmd.id}
                  onClick={cmd.action}
                  onMouseEnter={() => setSelectedIndex(idx)}
                  className={`w-full text-left px-3 py-2.5 rounded-lg flex items-center gap-3 transition-colors ${
                    isSelected ? 'bg-primary-light text-primary' : 'hover:bg-muted/60 text-foreground'
                  }`}
                >
                  <span className={`text-base shrink-0 w-6 text-center ${isSelected ? 'text-primary' : 'text-muted-foreground'}`}>
                    {cmd.icon}
                  </span>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <p className={`text-xs font-semibold truncate ${isSelected ? 'text-primary' : 'text-foreground'}`}>
                        {cmd.title}
                      </p>
                      <span className="text-[10px] font-medium px-1.5 py-0.2 rounded bg-muted text-muted-foreground uppercase tracking-wider shrink-0">
                        {cmd.category}
                      </span>
                    </div>
                    {cmd.subtitle && (
                      <p className="text-[11px] text-muted-foreground truncate mt-0.5">{cmd.subtitle}</p>
                    )}
                  </div>
                  {isSelected && (
                    <span className="text-xs text-primary font-mono shrink-0">↵</span>
                  )}
                </button>
              );
            })
          )}
        </div>

        {/* Footer */}
        <div className="px-4 py-2 border-t border-border bg-muted/40 flex items-center justify-between text-[11px] text-muted-foreground">
          <div className="flex items-center gap-3">
            <span><kbd className="font-mono bg-muted px-1 rounded border border-border">↑</kbd> <kbd className="font-mono bg-muted px-1 rounded border border-border">↓</kbd> Navigate</span>
            <span><kbd className="font-mono bg-muted px-1 rounded border border-border">↵</kbd> Select</span>
          </div>
          <span>Agent Ledger Quick Actions</span>
        </div>
      </div>
    </div>
  );
}
