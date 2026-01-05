import { useEffect, useState } from 'react';
import type { BoardResponse, Issue, Column, IssueSummary, Town, TownStatus, Agent, Rig } from './api';
import { fetchBoard, fetchIssue, fetchTown, fetchTownStatus } from './api';
import DependencyGraph from './components/DependencyGraph';
import './App.css';

type ViewMode = 'beads' | 'graph' | 'gastown';

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    pending: '#6b7280',
    in_progress: '#f59e0b',
    done: '#10b981',
    blocked: '#ef4444',
    active: '#10b981',
    offline: '#6b7280',
    idle: '#f59e0b',
    stuck: '#ef4444',
  };
  return (
    <span
      className="status-badge"
      style={{ backgroundColor: colors[status] || '#6b7280' }}
    >
      {status.replace('_', ' ')}
    </span>
  );
}

function IssueCard({
  issue,
  onClick,
}: {
  issue: IssueSummary;
  onClick: () => void;
}) {
  return (
    <div className="issue-card" onClick={onClick}>
      <div className="issue-title">{issue.title}</div>
      <div className="issue-meta">
        <StatusBadge status={issue.status} />
        <span className="issue-priority">{issue.priority}</span>
      </div>
    </div>
  );
}

function BoardColumn({
  column,
  onIssueClick,
}: {
  column: Column;
  onIssueClick: (id: string) => void;
}) {
  return (
    <div className="board-column">
      <div className="column-header">
        <span className="column-title">{column.label}</span>
        <span className="column-count">{column.count}</span>
      </div>
      <div className="column-issues">
        {column.issues.map((issue) => (
          <IssueCard
            key={issue.id}
            issue={issue}
            onClick={() => onIssueClick(issue.id)}
          />
        ))}
      </div>
    </div>
  );
}

function IssueDetail({
  issue,
  onClose,
}: {
  issue: Issue;
  onClose: () => void;
}) {
  return (
    <div className="issue-detail-overlay" onClick={onClose}>
      <div className="issue-detail" onClick={(e) => e.stopPropagation()}>
        <button className="close-btn" onClick={onClose}>
          &times;
        </button>
        <h2>{issue.title}</h2>
        <div className="issue-detail-meta">
          <StatusBadge status={issue.status} />
          <span className="issue-priority">[{issue.priority}]</span>
          <span className="issue-id">{issue.id}</span>
        </div>

        {issue.description && (
          <div className="issue-section">
            <h3>Description</h3>
            <p className="issue-description">{issue.description}</p>
          </div>
        )}

        {issue.done_when && issue.done_when.length > 0 && (
          <div className="issue-section">
            <h3>Done When</h3>
            <ul>
              {issue.done_when.map((item, i) => (
                <li key={i}>{item}</li>
              ))}
            </ul>
          </div>
        )}

        {issue.blocks && issue.blocks.length > 0 && (
          <div className="issue-section">
            <h3>Blocks</h3>
            <ul>
              {issue.blocks.map((dep) => (
                <li key={dep.id}>
                  {dep.title} <span className="dep-id">({dep.id})</span>
                </li>
              ))}
            </ul>
          </div>
        )}

        {issue.blocked_by && issue.blocked_by.length > 0 && (
          <div className="issue-section">
            <h3>Blocked By</h3>
            <ul>
              {issue.blocked_by.map((dep) => (
                <li key={dep.id}>
                  {dep.title} <span className="dep-id">({dep.id})</span>
                </li>
              ))}
            </ul>
          </div>
        )}

        {issue.children && issue.children.length > 0 && (
          <div className="issue-section">
            <h3>Children</h3>
            <ul>
              {issue.children.map((child) => (
                <li key={child.id}>
                  {child.title} <span className="dep-id">({child.id})</span>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </div>
  );
}

// Gas Town Components

function AgentCard({ agent }: { agent: Agent }) {
  const roleIcons: Record<string, string> = {
    mayor: '👑',
    deacon: '⚙️',
    witness: '👁️',
    refinery: '🏭',
    polecat: '🦨',
    crew: '👷',
  };

  return (
    <div className="agent-card">
      <div className="agent-icon">{roleIcons[agent.role] || '🤖'}</div>
      <div className="agent-info">
        <div className="agent-name">{agent.name}</div>
        <div className="agent-meta">
          <StatusBadge status={agent.status} />
          <span className="agent-role">{agent.role}</span>
          {agent.rig && <span className="agent-rig">{agent.rig}</span>}
        </div>
      </div>
    </div>
  );
}

function RigCard({ rig }: { rig: Rig }) {
  const agentCount = (rig.polecats?.length || 0) + (rig.crew?.length || 0) +
    (rig.witness ? 1 : 0) + (rig.refinery ? 1 : 0);
  const activeCount = [
    ...(rig.polecats || []),
    ...(rig.crew || []),
    rig.witness,
    rig.refinery
  ].filter(a => a && a.status === 'active').length;

  return (
    <div className="rig-card">
      <div className="rig-header">
        <span className="rig-name">{rig.name}</span>
        <span className="rig-stats">{activeCount}/{agentCount} active</span>
      </div>
      <div className="rig-agents">
        {rig.witness && <AgentCard agent={rig.witness} />}
        {rig.refinery && <AgentCard agent={rig.refinery} />}
        {rig.polecats?.map((p, i) => <AgentCard key={`p-${i}`} agent={p} />)}
        {rig.crew?.map((c, i) => <AgentCard key={`c-${i}`} agent={c} />)}
      </div>
    </div>
  );
}

function TownView({ town, status }: { town: Town | null; status: TownStatus | null }) {
  if (!town) {
    return (
      <div className="town-empty">
        <h2>Gas Town Not Found</h2>
        <p>No Gas Town workspace found at {status?.town_root || '~/gt'}</p>
        <p>Run <code>gt install ~/gt</code> to create one.</p>
      </div>
    );
  }

  return (
    <div className="town-view">
      {/* Town Status Bar */}
      <div className="town-status-bar">
        <div className="status-item">
          <span className="status-label">Status</span>
          <StatusBadge status={status?.healthy ? 'active' : 'offline'} />
        </div>
        <div className="status-item">
          <span className="status-label">Agents</span>
          <span className="status-value">{status?.active_agents || 0}/{status?.total_agents || 0}</span>
        </div>
        <div className="status-item">
          <span className="status-label">Rigs</span>
          <span className="status-value">{status?.active_rigs || 0}</span>
        </div>
        <div className="status-item">
          <span className="status-label">Convoys</span>
          <span className="status-value">{status?.open_convoys || 0}</span>
        </div>
      </div>

      {/* Town-level agents */}
      <div className="town-agents">
        <h3>Town Agents</h3>
        <div className="agents-grid">
          {town.mayor && <AgentCard agent={town.mayor} />}
          {town.deacon && <AgentCard agent={town.deacon} />}
        </div>
      </div>

      {/* Rigs */}
      <div className="town-rigs">
        <h3>Rigs ({town.rigs?.length || 0})</h3>
        {town.rigs?.length === 0 ? (
          <p className="empty-message">No rigs configured. Run <code>gt rig add &lt;name&gt;</code></p>
        ) : (
          <div className="rigs-grid">
            {town.rigs?.map((rig) => <RigCard key={rig.name} rig={rig} />)}
          </div>
        )}
      </div>

      {/* Convoys */}
      {town.convoys && town.convoys.length > 0 && (
        <div className="town-convoys">
          <h3>Active Convoys</h3>
          <div className="convoys-list">
            {town.convoys.map((convoy) => (
              <div key={convoy.id} className="convoy-card">
                <div className="convoy-header">
                  <span className="convoy-title">{convoy.title}</span>
                  <span className="convoy-progress">{convoy.progress}/{convoy.total}</span>
                </div>
                <div className="convoy-id">{convoy.id}</div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function App() {
  const [viewMode, setViewMode] = useState<ViewMode>('beads');
  const [board, setBoard] = useState<BoardResponse | null>(null);
  const [selectedIssue, setSelectedIssue] = useState<Issue | null>(null);
  const [town, setTown] = useState<Town | null>(null);
  const [townStatus, setTownStatus] = useState<TownStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 5000);
    return () => clearInterval(interval);
  }, []);

  async function loadData() {
    try {
      const [boardData, townData, statusData] = await Promise.all([
        fetchBoard().catch(() => null),
        fetchTown().catch(() => null),
        fetchTownStatus().catch(() => null),
      ]);
      if (boardData) setBoard(boardData);
      setTown(townData);
      setTownStatus(statusData);
      setError(null);
    } catch {
      setError('Failed to connect to daemon. Is gvid running on localhost:7070?');
    } finally {
      setLoading(false);
    }
  }

  async function handleIssueClick(id: string) {
    try {
      const issue = await fetchIssue(id);
      setSelectedIssue(issue);
    } catch (e) {
      console.error('Failed to fetch issue:', e);
    }
  }

  if (loading) {
    return (
      <div className="app">
        <div className="loading">Loading...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="app">
        <div className="error">
          <h2>Connection Error</h2>
          <p>{error}</p>
          <button onClick={loadData}>Retry</button>
        </div>
      </div>
    );
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>Gastown Viewer Intent</h1>
        <div className="view-tabs">
          <button
            className={`tab ${viewMode === 'beads' ? 'active' : ''}`}
            onClick={() => setViewMode('beads')}
          >
            Board ({board?.total || 0})
          </button>
          <button
            className={`tab ${viewMode === 'graph' ? 'active' : ''}`}
            onClick={() => setViewMode('graph')}
          >
            Graph
          </button>
          <button
            className={`tab ${viewMode === 'gastown' ? 'active' : ''}`}
            onClick={() => setViewMode('gastown')}
          >
            Gas Town {townStatus?.healthy ? '●' : '○'}
          </button>
        </div>
      </header>

      {viewMode === 'beads' && (
        <div className="board">
          {board?.columns.map((column) => (
            <BoardColumn
              key={column.status}
              column={column}
              onIssueClick={handleIssueClick}
            />
          ))}
        </div>
      )}

      {viewMode === 'graph' && (
        <DependencyGraph
          onNodeClick={handleIssueClick}
          width={window.innerWidth - 32}
          height={window.innerHeight - 200}
        />
      )}

      {viewMode === 'gastown' && (
        <TownView town={town} status={townStatus} />
      )}

      {selectedIssue && (
        <IssueDetail
          issue={selectedIssue}
          onClose={() => setSelectedIssue(null)}
        />
      )}
    </div>
  );
}

export default App;
