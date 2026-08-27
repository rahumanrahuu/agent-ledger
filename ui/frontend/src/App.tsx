import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useApi } from './hooks/useApi';
import { useWebSocket } from './hooks/useWebSocket';
import { getOverview } from './api/client';
import Sidebar from './Sidebar';
import TopBar from './TopBar';
import Overview from './Overview';
import Memories from './Memories';
import Timeline from './Timeline';
import Sessions, { SessionDetail } from './Sessions';
import KnowledgeGraph from './KnowledgeGraph';

function AppShell() {
  const wsStatus = useWebSocket();
  const { state } = useApi(getOverview, []);

  const project = state.status === 'ok' ? state.data.project : undefined;

  return (
    <div className="flex h-full overflow-hidden bg-background">
      <Sidebar projectName={project?.name} version={project?.version} />
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        <TopBar
          projectName={project?.name}
          branch={project?.branch}
          commit={project?.commit}
          lastActivity={project?.lastActivity}
          wsStatus={wsStatus}
        />
        <main className="flex-1 overflow-auto p-5 lg:p-6">
          <Routes>
            <Route path="/" element={<Navigate to="/overview" replace />} />
            <Route path="/overview" element={<Overview />} />
            <Route path="/memories" element={<Memories />} />
            <Route path="/timeline" element={<Timeline />} />
            <Route path="/sessions" element={<Sessions />} />
            <Route path="/sessions/:id" element={<SessionDetail />} />
            <Route path="/knowledge-graph" element={<KnowledgeGraph />} />
            <Route path="*" element={<Navigate to="/overview" replace />} />
          </Routes>
        </main>
      </div>
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
