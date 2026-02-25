# Gastown Viewer Intent

> Mission Control dashboard for [Gastown](https://github.com/steveyegge/gastown) multi-agent workspaces.

[![Release](https://img.shields.io/github/v/release/intent-solutions-io/gastown-viewer-intent)](https://github.com/intent-solutions-io/gastown-viewer-intent/releases)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## What's New in v0.4.0

### Embedded Web UI
- **Single binary serves everything** — `gvid` now bundles the React dashboard via `go:embed`
- No more separate `npm install && npm run dev` — just run `gvid` and open `http://localhost:7070`
- `go install`, `brew install`, and direct downloads all include the full web UI
- Development workflow preserved: `make dev` still runs Vite hot reload with API proxy

### Previous Highlights (v0.3.0)

- **Convoy Dashboard**: Batch work progress with Done/Active/Blocked/Pending counts
- **Interactive Dependency Graph**: D3.js force-directed visualization of all 14 Beads edge types
- **Smart Agent Status**: Active/Idle/Stuck detection with tmux integration
- **Molecule Progress Tracker**: Workflow step-by-step completion tracking

---

## What It Does

**Gastown Viewer** provides real-time visibility into your Gas Town agent swarms:

- **Agent Dashboard**: See all agents (Mayor, Deacon, Witness, Refinery, Polecats, Crew) with live status
- **Dependency Graph**: Interactive visualization of issue relationships
- **Molecule Tracking**: Monitor workflow progress across agents
- **Rig Overview**: Monitor project rigs with agent health and activity
- **Convoy Tracking**: Track batch work progress across rigs
- **Beads Integration**: Kanban board view of issues managed by your agents
- **Web + TUI**: Browser dashboard or terminal interface

## Quickstart

### Install

**Homebrew (macOS/Linux)**
```bash
brew tap intent-solutions-io/tap
brew install gvid
```

**Direct Download**

Download binaries from [Releases](https://github.com/intent-solutions-io/gastown-viewer-intent/releases).

**From Source**
```bash
go install github.com/intent-solutions-io/gastown-viewer-intent/cmd/gvid@latest
```

### Prerequisites

- [Gastown](https://github.com/steveyegge/gastown) installed at `~/gt`
- [Beads](https://github.com/steveyegge/beads) (`bd` CLI in PATH)

For development:
- Go 1.22+
- Node.js 20+

### Run

```bash
# If installed via brew/binary:
gvid                          # Start daemon + web UI on :7070

# For development (hot reload):
make dev                      # Vite on :5173, API proxied to :7070
```

Open http://localhost:7070 (or http://localhost:5173 during development) and switch between tabs:
- **Board** - Kanban view of Beads issues
- **Graph** - Interactive dependency visualization
- **Gas Town** - Agent dashboard with molecules

### Verify

```bash
# Health check
curl http://localhost:7070/api/v1/health

# Gas Town status
curl http://localhost:7070/api/v1/town/status
# {"healthy":true,"active_agents":5,"total_agents":8,"active_rigs":2,"molecules":3}

# List agents with status
curl http://localhost:7070/api/v1/town/agents

# Get dependency graph as DOT
curl "http://localhost:7070/api/v1/graph?format=dot" | dot -Tsvg > deps.svg

# List active molecules
curl http://localhost:7070/api/v1/town/molecules

# List active convoys
curl http://localhost:7070/api/v1/town/convoys
```

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Gastown Viewer Intent                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────────┐      ┌──────────────┐      ┌──────────────┐  │
│   │   gvi-tui    │      │   Web UI     │      │  External    │  │
│   │  (Bubbletea) │      │ (React+Vite) │      │   Clients    │  │
│   └──────┬───────┘      └──────┬───────┘      └──────┬───────┘  │
│          │                     │                     │          │
│          └─────────────────────┼─────────────────────┘          │
│                                │                                 │
│                                ▼                                 │
│                    ┌───────────────────────┐                    │
│                    │       gvid Daemon     │                    │
│                    │     localhost:7070    │                    │
│                    └───────────┬───────────┘                    │
│                                │                                 │
│              ┌─────────────────┼─────────────────┐              │
│              ▼                                   ▼              │
│   ┌───────────────────────┐         ┌───────────────────────┐  │
│   │   Gastown Adapter     │         │    Beads Adapter      │  │
│   │   (reads ~/gt/)       │         │   (shells to `bd`)    │  │
│   └───────────┬───────────┘         └───────────┬───────────┘  │
│               │                                 │               │
│               ▼                                 ▼               │
│   ┌───────────────────────┐         ┌───────────────────────┐  │
│   │      Gas Town         │         │     .beads/ state     │  │
│   │  ~/gt (rigs, agents)  │         │   (issues, deps)      │  │
│   └───────────────────────┘         └───────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

## Gas Town Concepts

| Concept | Description |
|---------|-------------|
| **Town** | Workspace root (`~/gt`) containing all rigs and town-level agents |
| **Mayor** | Town coordinator - routes work across rigs |
| **Deacon** | Town patrol - monitors health and escalates issues |
| **Rig** | Project container with its own agent pool |
| **Witness** | Rig-level overseer - manages polecat lifecycle |
| **Refinery** | Merge queue processor for the rig |
| **Polecats** | Transient workers spawned for specific tasks |
| **Crew** | Persistent user-managed workers in a rig |
| **Convoy** | Batch work tracking across multiple rigs |
| **Molecule** | Workflow instance with steps, assigned to an agent |
| **Formula** | Template defining molecule structure and steps |

## API Endpoints

### Gas Town

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/town/status` | Town health, agent/rig counts |
| `GET /api/v1/town` | Full town structure |
| `GET /api/v1/town/rigs` | List all rigs |
| `GET /api/v1/town/rigs/:name` | Single rig details |
| `GET /api/v1/town/agents` | All agents with status |
| `GET /api/v1/town/convoys` | Active convoys |
| `GET /api/v1/town/convoys/:id` | Single convoy details |
| `GET /api/v1/town/molecules` | Active molecules across agents |
| `GET /api/v1/town/molecules/:id` | Single molecule details |
| `GET /api/v1/town/mail/:address` | Agent mail inbox |

### Beads (Issues)

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/health` | Health check |
| `GET /api/v1/board` | Kanban board view |
| `GET /api/v1/issues` | List issues |
| `GET /api/v1/issues/:id` | Issue details |
| `GET /api/v1/graph?format=json` | Dependency graph (JSON) |
| `GET /api/v1/graph?format=dot` | Dependency graph (Graphviz DOT) |
| `GET /api/v1/events` | SSE event stream |

## Configuration

```bash
# Custom Gas Town location
go run ./cmd/gvid --town /path/to/gt

# Custom port
go run ./cmd/gvid --port 8080

# All options
go run ./cmd/gvid --help
```

## Project Structure

```
gastown-viewer-intent/
├── cmd/
│   ├── gvid/              # Daemon
│   └── gvi-tui/           # TUI client
├── internal/
│   ├── api/               # HTTP handlers
│   ├── gastown/           # Gas Town adapter (reads ~/gt)
│   ├── beads/             # Beads adapter (bd CLI)
│   └── model/             # Domain types
├── web/                   # React + Vite frontend
└── Makefile
```

## License

MIT

## Related Projects

- [Gastown](https://github.com/steveyegge/gastown) - Multi-agent workspace orchestrator
- [Beads](https://github.com/steveyegge/beads) - Local-first issue tracking with dependencies
