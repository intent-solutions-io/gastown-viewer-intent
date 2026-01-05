// Package gastown provides integration with Gas Town multi-agent orchestrator.
package gastown

import "time"

// Role represents the type of Gas Town agent.
type Role string

const (
	RoleMayor    Role = "mayor"
	RoleDeacon   Role = "deacon"
	RoleWitness  Role = "witness"
	RoleRefinery Role = "refinery"
	RoleCrew     Role = "crew"
	RolePolecat  Role = "polecat"
)

// AgentStatus represents the current status of an agent.
type AgentStatus string

const (
	StatusActive   AgentStatus = "active"
	StatusIdle     AgentStatus = "idle"
	StatusStuck    AgentStatus = "stuck"
	StatusOffline  AgentStatus = "offline"
	StatusUnknown  AgentStatus = "unknown"
)

// Town represents the Gas Town workspace.
type Town struct {
	Root    string    `json:"root"`
	Name    string    `json:"name,omitempty"`
	Rigs    []Rig     `json:"rigs"`
	Mayor   *Agent    `json:"mayor,omitempty"`
	Deacon  *Agent    `json:"deacon,omitempty"`
	Convoys []Convoy  `json:"convoys,omitempty"`
}

// TownConfig is the configuration from mayor/town.json.
type TownConfig struct {
	Name    string   `json:"name,omitempty"`
	Rigs    []string `json:"rigs,omitempty"`
	Version string   `json:"version,omitempty"`
}

// Rig represents a project container with agents.
type Rig struct {
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	Remote    string   `json:"remote,omitempty"`
	Witness   *Agent   `json:"witness,omitempty"`
	Refinery  *Agent   `json:"refinery,omitempty"`
	Polecats  []Agent  `json:"polecats,omitempty"`
	Crew      []Agent  `json:"crew,omitempty"`
}

// Agent represents a Gas Town agent (polecat, witness, etc.).
type Agent struct {
	Role         Role        `json:"role"`
	Name         string      `json:"name"`
	Rig          string      `json:"rig,omitempty"`
	Status       AgentStatus `json:"status"`
	Session      string      `json:"session,omitempty"`
	Molecule     string      `json:"molecule,omitempty"`
	HookAttached bool        `json:"hook_attached,omitempty"`
	LastActive   time.Time   `json:"last_active,omitempty"`
	Compaction   int         `json:"compaction,omitempty"`
	WorkDir      string      `json:"work_dir,omitempty"`
}

// Address returns the mail-style address for this agent.
func (a *Agent) Address() string {
	switch a.Role {
	case RoleMayor:
		return "mayor/"
	case RoleDeacon:
		return "deacon/"
	case RoleWitness:
		return a.Rig + "/witness"
	case RoleRefinery:
		return a.Rig + "/refinery"
	case RoleCrew:
		return a.Rig + "/" + a.Name
	case RolePolecat:
		return a.Rig + "/" + a.Name
	default:
		return ""
	}
}

// Convoy represents a batch of tracked work.
type Convoy struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	Issues      []string  `json:"issues"`
	Progress    int       `json:"progress"`
	Total       int       `json:"total"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	Subscribers []string  `json:"subscribers,omitempty"`
}

// Message represents a mail message between agents.
type Message struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Timestamp time.Time `json:"timestamp"`
	Read      bool      `json:"read"`
	Priority  string    `json:"priority"`
	Type      string    `json:"type"`
}

// Molecule represents a workflow instance.
type Molecule struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Status   string         `json:"status"`
	Steps    []MoleculeStep `json:"steps"`
	Progress int            `json:"progress"`
	Total    int            `json:"total"`
}

// MoleculeStep represents a step in a molecule workflow.
type MoleculeStep struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Needs       []string `json:"needs,omitempty"`
}

// TownStatus provides a summary of town health.
type TownStatus struct {
	Healthy      bool   `json:"healthy"`
	TownRoot     string `json:"town_root"`
	ActiveAgents int    `json:"active_agents"`
	TotalAgents  int    `json:"total_agents"`
	ActiveRigs   int    `json:"active_rigs"`
	OpenConvoys  int    `json:"open_convoys"`
	Error        string `json:"error,omitempty"`
}
