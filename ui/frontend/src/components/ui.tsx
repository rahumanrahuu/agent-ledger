/* Shared design-system primitives used across every page */
import type { ReactNode } from 'react';
import type { EventType, ImportanceLevel, MemoryType } from '../api/types';

// ─── Badge ──────────────────────────────────────────────────────────────────

type BadgeVariant = 'default' | 'success' | 'warning' | 'destructive' | 'info' | 'primary' | 'muted';

const badgeClasses: Record<BadgeVariant, string> = {
  default: 'bg-muted text-muted-foreground',
  muted: 'bg-muted text-muted-foreground',
  success: 'bg-success-light text-success',
  warning: 'bg-warning-light text-warning',
  destructive: 'bg-destructive-light text-destructive',
  info: 'bg-info-light text-info',
  primary: 'bg-primary-light text-primary',
};

interface BadgeProps {
  variant?: BadgeVariant;
  children: ReactNode;
  className?: string;
  dot?: boolean;
  pulse?: boolean;
}

export function Badge({ variant = 'default', children, className = '', dot, pulse }: BadgeProps) {
  return (
    <span className={`inline-flex items-center gap-1.5 text-xs px-2 py-0.5 rounded-full font-medium ${badgeClasses[variant]} ${className}`}>
      {dot && (
        <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${dotColorForVariant(variant)} ${pulse ? 'animate-pulse' : ''}`} />
      )}
      {children}
    </span>
  );
}

function dotColorForVariant(v: BadgeVariant) {
  switch (v) {
    case 'success': return 'bg-success';
    case 'warning': return 'bg-warning';
    case 'destructive': return 'bg-destructive';
    case 'info': return 'bg-info';
    case 'primary': return 'bg-primary';
    default: return 'bg-muted-foreground';
  }
}

// ─── Event type helpers ──────────────────────────────────────────────────────

export const eventVariant: Record<EventType, BadgeVariant> = {
  decision: 'info',
  discovery: 'success',
  failure: 'destructive',
  constraint: 'warning',
  checkpoint: 'primary',
};

export const eventDotColor: Record<EventType, string> = {
  decision: 'bg-info',
  discovery: 'bg-success',
  failure: 'bg-destructive',
  constraint: 'bg-warning',
  checkpoint: 'bg-primary',
};

export const memoryVariant: Record<MemoryType, BadgeVariant> = {
  fact: 'info',
  rule: 'destructive',
  pattern: 'primary',
  entity: 'info',
  preference: 'warning',
  insight: 'success',
  decision: 'info',
  discovery: 'success',
  failure: 'destructive',
  constraint: 'warning',
};

export const importanceVariant: Record<ImportanceLevel, BadgeVariant> = {
  critical: 'destructive',
  high: 'warning',
  medium: 'info',
  low: 'muted',
};

// ─── Session status ──────────────────────────────────────────────────────────

export const sessionStatusVariant: Record<string, BadgeVariant> = {
  running: 'success',
  completed: 'info',
  failed: 'destructive',
  paused: 'warning',
};

// ─── Skeleton ────────────────────────────────────────────────────────────────

export function Skeleton({ className = '' }: { className?: string }) {
  return <div className={`animate-pulse rounded-md bg-muted ${className}`} />;
}

export function SkeletonCard({ className = '' }: { className?: string }) {
  return (
    <div className={`bg-card border border-border rounded-xl p-4 space-y-3 ${className}`}>
      <Skeleton className="h-3 w-24" />
      <Skeleton className="h-7 w-16" />
      <Skeleton className="h-3 w-32" />
    </div>
  );
}

export function SkeletonRow() {
  return (
    <div className="flex gap-3 px-4 py-3">
      <Skeleton className="h-4 w-16 shrink-0" />
      <div className="flex-1 space-y-2">
        <Skeleton className="h-4 w-3/4" />
        <Skeleton className="h-3 w-1/2" />
      </div>
    </div>
  );
}

// ─── Empty state ─────────────────────────────────────────────────────────────

interface EmptyStateProps {
  title: string;
  description?: string;
  icon?: ReactNode;
  action?: ReactNode;
}

export function EmptyState({ title, description, icon, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-6 text-center">
      {icon && <div className="mb-3 text-muted-foreground opacity-30">{icon}</div>}
      <p className="text-sm font-medium text-foreground mb-1">{title}</p>
      {description && <p className="text-xs text-muted-foreground max-w-xs leading-relaxed">{description}</p>}
      {action && <div className="mt-4">{action}</div>}
    </div>
  );
}

// ─── Error state ─────────────────────────────────────────────────────────────

interface ErrorStateProps {
  message: string;
  onRetry?: () => void;
}

export function ErrorState({ message, onRetry }: ErrorStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 px-6 text-center">
      <div className="w-10 h-10 rounded-full bg-destructive-light flex items-center justify-center mb-3">
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
          <path d="M8 3v5M8 10.5v1" stroke="#DC2626" strokeWidth="1.6" strokeLinecap="round"/>
          <circle cx="8" cy="8" r="6.5" stroke="#DC2626" strokeWidth="1.4"/>
        </svg>
      </div>
      <p className="text-sm font-medium text-foreground mb-1">Something went wrong</p>
      <p className="text-xs text-muted-foreground mb-4 font-mono max-w-sm">{message}</p>
      {onRetry && (
        <button
          onClick={onRetry}
          className="text-xs font-medium px-3 py-1.5 rounded-lg bg-primary text-white hover:bg-primary-hover transition-colors"
        >
          Try again
        </button>
      )}
    </div>
  );
}

// ─── Page header ─────────────────────────────────────────────────────────────

interface PageHeaderProps {
  title: string;
  description?: string;
  actions?: ReactNode;
  breadcrumb?: { label: string; onClick?: () => void }[];
}

export function PageHeader({ title, description, actions, breadcrumb }: PageHeaderProps) {
  return (
    <div className="mb-5">
      {breadcrumb && breadcrumb.length > 0 && (
        <div className="flex items-center gap-1.5 mb-2">
          {breadcrumb.map((crumb, i) => (
            <span key={i} className="flex items-center gap-1.5">
              {i > 0 && <span className="text-border">/</span>}
              {crumb.onClick ? (
                <button
                  onClick={crumb.onClick}
                  className="text-xs text-muted-foreground hover:text-foreground transition-colors"
                >
                  {crumb.label}
                </button>
              ) : (
                <span className="text-xs text-muted-foreground">{crumb.label}</span>
              )}
            </span>
          ))}
        </div>
      )}
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-xl font-bold text-foreground tracking-tight">{title}</h1>
          {description && <p className="text-sm text-muted-foreground mt-0.5">{description}</p>}
        </div>
        {actions && <div className="flex items-center gap-2 shrink-0">{actions}</div>}
      </div>
    </div>
  );
}

// ─── Card ────────────────────────────────────────────────────────────────────

interface CardProps {
  children: ReactNode;
  className?: string;
  onClick?: () => void;
  hoverable?: boolean;
}

export function Card({ children, className = '', onClick, hoverable }: CardProps) {
  const base = 'bg-card border border-border rounded-xl';
  const interactive = hoverable || onClick ? 'cursor-pointer hover:border-foreground/20 hover:shadow-sm transition-all' : '';
  return onClick ? (
    <button onClick={onClick} className={`w-full text-left ${base} ${interactive} ${className}`}>
      {children}
    </button>
  ) : (
    <div className={`${base} ${interactive} ${className}`}>
      {children}
    </div>
  );
}

// ─── Metadata row ────────────────────────────────────────────────────────────

export function MetaRow({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="flex items-baseline gap-3 py-1.5 border-b border-border last:border-0">
      <span className="text-xs text-muted-foreground w-28 shrink-0">{label}</span>
      <span className="text-xs text-foreground font-medium flex-1">{value}</span>
    </div>
  );
}

// ─── Inspector panel ─────────────────────────────────────────────────────────

interface InspectorPanelProps {
  title: string;
  subtitle?: string;
  badge?: ReactNode;
  children: ReactNode;
  onClose: () => void;
  width?: string;
}

export function InspectorPanel({ title, subtitle, badge, children, onClose, width = 'w-80' }: InspectorPanelProps) {
  return (
    <div className={`${width} shrink-0 bg-card border border-border rounded-xl flex flex-col overflow-hidden h-full`}>
      <div className="px-4 py-3.5 border-b border-border shrink-0">
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            {badge && <div className="mb-1.5">{badge}</div>}
            <p className="text-sm font-semibold text-foreground leading-snug">{title}</p>
            {subtitle && <p className="text-xs text-muted-foreground mt-0.5">{subtitle}</p>}
          </div>
          <button
            onClick={onClose}
            className="text-muted-foreground hover:text-foreground transition-colors mt-0.5 shrink-0 p-0.5 rounded"
          >
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <path d="M2 2l10 10M12 2L2 12" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
            </svg>
          </button>
        </div>
      </div>
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {children}
      </div>
    </div>
  );
}

// ─── Section label ───────────────────────────────────────────────────────────

export function SectionLabel({ children }: { children: ReactNode }) {
  return (
    <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">{children}</p>
  );
}

// ─── Tag list ────────────────────────────────────────────────────────────────

export function TagList({ tags }: { tags: string[] }) {
  return (
    <div className="flex flex-wrap gap-1.5">
      {tags.map((tag) => (
        <span key={tag} className="text-xs text-foreground bg-muted border border-border px-2 py-0.5 rounded-md font-mono">
          {tag}
        </span>
      ))}
    </div>
  );
}

// ─── Search input ────────────────────────────────────────────────────────────

interface SearchInputProps {
  value: string;
  onChange: (v: string) => void;
  placeholder?: string;
  className?: string;
}

export function SearchInput({ value, onChange, placeholder = 'Search…', className = '' }: SearchInputProps) {
  return (
    <div className={`flex items-center gap-2 bg-muted rounded-lg px-3 py-2 ${className}`}>
      <svg width="13" height="13" viewBox="0 0 14 14" fill="none" className="text-muted-foreground shrink-0">
        <circle cx="6" cy="6" r="4.5" stroke="currentColor" strokeWidth="1.4"/>
        <path d="m9.5 9.5 2.5 2.5" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round"/>
      </svg>
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="flex-1 bg-transparent text-xs text-foreground placeholder:text-muted-foreground outline-none"
      />
      {value && (
        <button onClick={() => onChange('')} className="text-muted-foreground hover:text-foreground">
          <svg width="10" height="10" viewBox="0 0 10 10" fill="none">
            <path d="M1 1l8 8M9 1L1 9" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round"/>
          </svg>
        </button>
      )}
    </div>
  );
}

// ─── Filter pills ────────────────────────────────────────────────────────────

interface FilterPillProps {
  label: string;
  count?: number;
  active: boolean;
  variant?: BadgeVariant;
  onClick: () => void;
}

export function FilterPill({ label, count, active, variant = 'default', onClick }: FilterPillProps) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-full border font-medium transition-colors ${
        active
          ? `${badgeClasses[variant]} border-current`
          : 'bg-card border-border text-muted-foreground hover:border-foreground/20'
      }`}
    >
      {label}
      {count !== undefined && (
        <span className={`px-1.5 py-0.5 rounded-full text-xs font-normal ${active ? 'bg-current/10' : 'bg-muted'}`}>
          {count}
        </span>
      )}
    </button>
  );
}

// ─── Relative time ───────────────────────────────────────────────────────────

export function formatRelative(iso: string) {
  if (!iso) return 'recently';
  const d = new Date(iso);
  if (isNaN(d.getTime())) return 'recently';
  const now = new Date();
  const diffMs = Math.max(0, now.getTime() - d.getTime());
  const diffMin = Math.floor(diffMs / 60000);
  const diffH = Math.floor(diffMs / 3600000);
  const diffD = Math.floor(diffMs / 86400000);
  if (diffMin < 1) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffH < 24) return `${diffH}h ago`;
  if (diffD < 7) return `${diffD}d ago`;
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

export function formatDateTime(iso: string) {
  return new Date(iso).toLocaleString('en-US', {
    month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', hour12: true,
  });
}

export function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}
