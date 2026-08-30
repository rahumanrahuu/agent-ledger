import { useState, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useApi } from './hooks/useApi';
import { useWebSocket } from './hooks/useWebSocket';
import { useTheme } from './hooks/useTheme';
import { getOverview } from './api/client';
import Sidebar from './Sidebar';
import TopBar from './TopBar';
import CommandPalette from './components/CommandPalette';
import Overview from './Overview';
import ProjectContext from './Context';
import Memories from './Memories';
import Timeline from './Timeline';
import Sessions, { SessionDetail } from './Sessions';
import KnowledgeGraph from './KnowledgeGraph';

function AppShell() {
  const wsStatus = useWebSocket();
  const { theme, setTheme } = useTheme();
  const { state } = useApi(getOverview, []);
  const [isCmdPaletteOpen, setIsCmdPaletteOpen] = useState(false);
  const [liveToast, setLiveToast] = useState<string | null>(null);
  const [isCollapsed, setIsCollapsed] = useState(() => {
    return localStorage.getItem('al_sidebar_collapsed') === 'true';
  });

  const project = state.status === 'ok' ? state.data.project : undefined;

  const toggleSidebar = () => {
    setIsCollapsed((prev) => {
      const next = !prev;
      localStorage.setItem('al_sidebar_collapsed', String(next));
      return next;
    });
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
        e.preventDefault();
        setIsCmdPaletteOpen((prev) => !prev);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  useEffect(() => {
    const handleLiveMessage = () => {
      setLiveToast('Live update: Development context updated');
      const timer = setTimeout(() => setLiveToast(null), 3000);
      return () => clearTimeout(timer);
    };

    window.addEventListener('al-websocket-message', handleLiveMessage);
    return () => window.removeEventListener('al-websocket-message', handleLiveMessage);
  }, []);

  return (
    <div className="flex h-full overflow-hidden bg-background text-foreground transition-colors duration-150 relative">
      <Sidebar
        projectName={project?.name}
        version={project?.version}
        isCollapsed={isCollapsed}
        onToggleCollapse={toggleSidebar}
        theme={theme}
        setTheme={setTheme}
      />
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <TopBar
          projectName={project?.name}
          branch={project?.branch}
          commit={project?.commit}
          lastActivity={project?.lastActivity}
          wsStatus={wsStatus}
          onOpenCommandPalette={() => setIsCmdPaletteOpen(true)}
        />
        <main className="flex-1 overflow-auto p-5 lg:p-6">
          <Routes>
            <Route path="/" element={<Navigate to="/overview" replace />} />
            <Route path="/overview" element={<Overview />} />
            <Route path="/context" element={<ProjectContext />} />
            <Route path="/memories" element={<Memories />} />
            <Route path="/timeline" element={<Timeline />} />
            <Route path="/sessions" element={<Sessions />} />
            <Route path="/sessions/:id" element={<SessionDetail />} />
            <Route path="/knowledge-graph" element={<KnowledgeGraph />} />
            <Route path="*" element={<Navigate to="/overview" replace />} />
          </Routes>
        </main>
      </div>

      <CommandPalette
        isOpen={isCmdPaletteOpen}
        onClose={() => setIsCmdPaletteOpen(false)}
        theme={theme}
        setTheme={setTheme}
      />

      {/* Live Activity Toast Banner */}
      {liveToast && (
        <div className="fixed bottom-5 right-5 z-40 bg-card border border-primary/40 text-foreground px-4 py-2.5 rounded-xl shadow-xl flex items-center gap-2.5 animate-in fade-in slide-in-from-bottom-3 duration-200">
          <span className="w-2 h-2 rounded-full bg-success animate-pulse shrink-0" />
          <span className="text-xs font-medium">{liveToast}</span>
        </div>
      )}
    </div>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AppShell />
    </BrowserRouter>
  );
}
