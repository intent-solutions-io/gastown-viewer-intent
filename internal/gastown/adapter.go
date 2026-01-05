package gastown

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Adapter provides access to Gas Town data.
type Adapter interface {
	// Status returns the overall town health status.
	Status(ctx context.Context) (*TownStatus, error)

	// Town returns the full town structure.
	Town(ctx context.Context) (*Town, error)

	// Rigs returns all rigs in the town.
	Rigs(ctx context.Context) ([]Rig, error)

	// Rig returns a specific rig by name.
	Rig(ctx context.Context, name string) (*Rig, error)

	// Agents returns all agents across all rigs.
	Agents(ctx context.Context) ([]Agent, error)

	// Convoys returns active convoys.
	Convoys(ctx context.Context) ([]Convoy, error)

	// Mail returns messages for an agent address.
	Mail(ctx context.Context, address string) ([]Message, error)
}

// FSAdapter reads Gas Town state from the filesystem and gt CLI.
type FSAdapter struct {
	townRoot string
}

// NewFSAdapter creates a new filesystem-based adapter.
func NewFSAdapter(townRoot string) *FSAdapter {
	if townRoot == "" {
		townRoot = filepath.Join(os.Getenv("HOME"), "gt")
	}
	return &FSAdapter{townRoot: townRoot}
}

// Status returns the overall town health status.
func (a *FSAdapter) Status(ctx context.Context) (*TownStatus, error) {
	status := &TownStatus{
		TownRoot: a.townRoot,
	}

	// Check if town exists
	if !a.townExists() {
		status.Healthy = false
		status.Error = fmt.Sprintf("Town not found at %s", a.townRoot)
		return status, nil
	}

	// Get town data
	town, err := a.Town(ctx)
	if err != nil {
		status.Healthy = false
		status.Error = err.Error()
		return status, nil
	}

	// Count agents
	status.ActiveRigs = len(town.Rigs)
	for _, rig := range town.Rigs {
		status.TotalAgents += len(rig.Polecats) + len(rig.Crew)
		if rig.Witness != nil {
			status.TotalAgents++
			if rig.Witness.Status == StatusActive {
				status.ActiveAgents++
			}
		}
		if rig.Refinery != nil {
			status.TotalAgents++
			if rig.Refinery.Status == StatusActive {
				status.ActiveAgents++
			}
		}
		for _, p := range rig.Polecats {
			if p.Status == StatusActive {
				status.ActiveAgents++
			}
		}
		for _, c := range rig.Crew {
			if c.Status == StatusActive {
				status.ActiveAgents++
			}
		}
	}

	if town.Mayor != nil {
		status.TotalAgents++
		if town.Mayor.Status == StatusActive {
			status.ActiveAgents++
		}
	}
	if town.Deacon != nil {
		status.TotalAgents++
		if town.Deacon.Status == StatusActive {
			status.ActiveAgents++
		}
	}

	status.OpenConvoys = len(town.Convoys)
	status.Healthy = true

	return status, nil
}

// Town returns the full town structure.
func (a *FSAdapter) Town(ctx context.Context) (*Town, error) {
	if !a.townExists() {
		return nil, fmt.Errorf("town not found at %s", a.townRoot)
	}

	town := &Town{
		Root: a.townRoot,
	}

	// Read town config
	config, err := a.readTownConfig()
	if err == nil {
		town.Name = config.Name
	}

	// Get tmux sessions to determine agent status
	sessions := a.getTmuxSessions()

	// Check mayor
	if a.dirExists(filepath.Join(a.townRoot, "mayor")) {
		mayor := &Agent{
			Role: RoleMayor,
			Name: "mayor",
		}
		a.enrichAgent(mayor, sessions)
		town.Mayor = mayor
	}

	// Check deacon (via daemon)
	if a.daemonRunning() {
		deacon := &Agent{
			Role:   RoleDeacon,
			Name:   "deacon",
			Status: StatusActive,
		}
		a.enrichAgent(deacon, sessions)
		town.Deacon = deacon
	}

	// Find rigs
	rigs, err := a.Rigs(ctx)
	if err == nil {
		town.Rigs = rigs
	}

	// Get convoys
	convoys, err := a.Convoys(ctx)
	if err == nil {
		town.Convoys = convoys
	}

	return town, nil
}

// Rigs returns all rigs in the town.
func (a *FSAdapter) Rigs(ctx context.Context) ([]Rig, error) {
	var rigs []Rig

	// Look for directories that have rig markers
	entries, err := os.ReadDir(a.townRoot)
	if err != nil {
		return nil, err
	}

	sessions := a.getTmuxSessions()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Skip non-rig directories
		if name == "mayor" || name == ".beads" || name == ".git" || strings.HasPrefix(name, ".") {
			continue
		}

		rigPath := filepath.Join(a.townRoot, name)

		// Check if it looks like a rig (has polecats/, witness/, or .beads/)
		if !a.dirExists(filepath.Join(rigPath, "polecats")) &&
			!a.dirExists(filepath.Join(rigPath, "witness")) &&
			!a.dirExists(filepath.Join(rigPath, ".beads")) {
			continue
		}

		rig := Rig{
			Name: name,
			Path: rigPath,
		}

		// Check witness
		if a.dirExists(filepath.Join(rigPath, "witness")) {
			witness := &Agent{
				Role: RoleWitness,
				Name: "witness",
				Rig:  name,
			}
			a.enrichAgent(witness, sessions)
			rig.Witness = witness
		}

		// Check refinery
		if a.dirExists(filepath.Join(rigPath, "refinery")) {
			refinery := &Agent{
				Role: RoleRefinery,
				Name: "refinery",
				Rig:  name,
			}
			a.enrichAgent(refinery, sessions)
			rig.Refinery = refinery
		}

		// Find polecats
		polecatsDir := filepath.Join(rigPath, "polecats")
		if a.dirExists(polecatsDir) {
			pEntries, err := os.ReadDir(polecatsDir)
			if err == nil {
				for _, pe := range pEntries {
					if pe.IsDir() && !strings.HasPrefix(pe.Name(), ".") {
						polecat := Agent{
							Role: RolePolecat,
							Name: pe.Name(),
							Rig:  name,
						}
						a.enrichAgent(&polecat, sessions)
						rig.Polecats = append(rig.Polecats, polecat)
					}
				}
			}
		}

		// Find crew
		crewDir := filepath.Join(rigPath, "crew")
		if a.dirExists(crewDir) {
			cEntries, err := os.ReadDir(crewDir)
			if err == nil {
				for _, ce := range cEntries {
					if ce.IsDir() && !strings.HasPrefix(ce.Name(), ".") {
						crew := Agent{
							Role: RoleCrew,
							Name: ce.Name(),
							Rig:  name,
						}
						a.enrichAgent(&crew, sessions)
						rig.Crew = append(rig.Crew, crew)
					}
				}
			}
		}

		rigs = append(rigs, rig)
	}

	return rigs, nil
}

// Rig returns a specific rig by name.
func (a *FSAdapter) Rig(ctx context.Context, name string) (*Rig, error) {
	rigs, err := a.Rigs(ctx)
	if err != nil {
		return nil, err
	}

	for _, rig := range rigs {
		if rig.Name == name {
			return &rig, nil
		}
	}

	return nil, fmt.Errorf("rig not found: %s", name)
}

// Agents returns all agents across all rigs.
func (a *FSAdapter) Agents(ctx context.Context) ([]Agent, error) {
	town, err := a.Town(ctx)
	if err != nil {
		return nil, err
	}

	var agents []Agent

	if town.Mayor != nil {
		agents = append(agents, *town.Mayor)
	}
	if town.Deacon != nil {
		agents = append(agents, *town.Deacon)
	}

	for _, rig := range town.Rigs {
		if rig.Witness != nil {
			agents = append(agents, *rig.Witness)
		}
		if rig.Refinery != nil {
			agents = append(agents, *rig.Refinery)
		}
		agents = append(agents, rig.Polecats...)
		agents = append(agents, rig.Crew...)
	}

	return agents, nil
}

// Convoys returns active convoys by running gt convoy list.
func (a *FSAdapter) Convoys(ctx context.Context) ([]Convoy, error) {
	// Try to run gt convoy list --json
	cmd := exec.CommandContext(ctx, "gt", "convoy", "list", "--json")
	cmd.Dir = a.townRoot
	output, err := cmd.Output()
	if err != nil {
		// gt might not be installed or convoy command might fail
		return nil, nil
	}

	var convoys []Convoy
	if err := json.Unmarshal(output, &convoys); err != nil {
		// Try parsing as single convoy
		var convoy Convoy
		if err := json.Unmarshal(output, &convoy); err != nil {
			return nil, nil
		}
		convoys = append(convoys, convoy)
	}

	return convoys, nil
}

// Mail returns messages for an agent address.
func (a *FSAdapter) Mail(ctx context.Context, address string) ([]Message, error) {
	// Run gt mail inbox for the address
	cmd := exec.CommandContext(ctx, "gt", "mail", "inbox", "--json")
	cmd.Dir = a.townRoot
	cmd.Env = append(os.Environ(), fmt.Sprintf("GT_ROLE=%s", address))
	output, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	var messages []Message
	if err := json.Unmarshal(output, &messages); err != nil {
		return nil, nil
	}

	return messages, nil
}

// Helper methods

func (a *FSAdapter) townExists() bool {
	return a.dirExists(filepath.Join(a.townRoot, "mayor"))
}

func (a *FSAdapter) dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (a *FSAdapter) readTownConfig() (*TownConfig, error) {
	configPath := filepath.Join(a.townRoot, "mayor", "town.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config TownConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (a *FSAdapter) getTmuxSessions() map[string]bool {
	sessions := make(map[string]bool)

	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		return sessions
	}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			sessions[line] = true
		}
	}

	return sessions
}

func (a *FSAdapter) daemonRunning() bool {
	// Check if gt daemon is running by looking for pid file or process
	pidFile := filepath.Join(a.townRoot, "mayor", "daemon.pid")
	if _, err := os.Stat(pidFile); err == nil {
		return true
	}

	// Also check via gt daemon status
	cmd := exec.Command("gt", "daemon", "status")
	cmd.Dir = a.townRoot
	err := cmd.Run()
	return err == nil
}

// LastActivity returns the last modification time of agent's workspace.
func (a *FSAdapter) LastActivity(rigName, agentName string) time.Time {
	var checkPath string
	if agentName == "witness" {
		checkPath = filepath.Join(a.townRoot, rigName, "witness")
	} else if agentName == "refinery" {
		checkPath = filepath.Join(a.townRoot, rigName, "refinery")
	} else {
		checkPath = filepath.Join(a.townRoot, rigName, "polecats", agentName)
		if !a.dirExists(checkPath) {
			checkPath = filepath.Join(a.townRoot, rigName, "crew", agentName)
		}
	}

	info, err := os.Stat(checkPath)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// getAgentWorkDir returns the working directory for an agent.
func (a *FSAdapter) getAgentWorkDir(rigName string, role Role, name string) string {
	switch role {
	case RoleMayor:
		return filepath.Join(a.townRoot, "mayor")
	case RoleDeacon:
		return filepath.Join(a.townRoot, "mayor") // Deacon runs from mayor dir
	case RoleWitness:
		return filepath.Join(a.townRoot, rigName, "witness")
	case RoleRefinery:
		return filepath.Join(a.townRoot, rigName, "refinery")
	case RolePolecat:
		return filepath.Join(a.townRoot, rigName, "polecats", name)
	case RoleCrew:
		return filepath.Join(a.townRoot, rigName, "crew", name)
	default:
		return ""
	}
}

// enrichAgent adds session, molecule, and hook info to an agent.
func (a *FSAdapter) enrichAgent(agent *Agent, sessions map[string]bool) {
	workDir := a.getAgentWorkDir(agent.Rig, agent.Role, agent.Name)
	if workDir == "" {
		return
	}
	agent.WorkDir = workDir

	// Get session name
	sessionName := a.getSessionName(agent)
	agent.Session = sessionName

	// Check if session is active
	if sessions[sessionName] {
		agent.Status = StatusActive
	} else {
		agent.Status = StatusOffline
	}

	// Read seance file for compaction level
	seancePath := filepath.Join(workDir, ".claude", "seance.json")
	if data, err := os.ReadFile(seancePath); err == nil {
		var seance struct {
			Compaction int    `json:"compaction"`
			Molecule   string `json:"molecule,omitempty"`
		}
		if json.Unmarshal(data, &seance) == nil {
			agent.Compaction = seance.Compaction
			if seance.Molecule != "" {
				agent.Molecule = seance.Molecule
			}
		}
	}

	// Check hook for attached molecule
	hookPath := filepath.Join(workDir, ".claude", "hook.json")
	if data, err := os.ReadFile(hookPath); err == nil {
		var hook struct {
			Molecule string `json:"molecule,omitempty"`
			Attached bool   `json:"attached,omitempty"`
		}
		if json.Unmarshal(data, &hook) == nil {
			agent.HookAttached = hook.Attached || hook.Molecule != ""
			if hook.Molecule != "" {
				agent.Molecule = hook.Molecule
			}
		}
	}

	// Also check molecule.json directly
	molPath := filepath.Join(workDir, ".beads", "molecule.json")
	if data, err := os.ReadFile(molPath); err == nil {
		var mol struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}
		if json.Unmarshal(data, &mol) == nil && mol.ID != "" {
			agent.Molecule = mol.ID
			agent.HookAttached = true
		}
	}

	// Get last activity time
	if info, err := os.Stat(workDir); err == nil {
		agent.LastActive = info.ModTime()
	}

	// Detect stuck status: active session but no activity for 10+ minutes
	if agent.Status == StatusActive && !agent.LastActive.IsZero() {
		if time.Since(agent.LastActive) > 10*time.Minute {
			agent.Status = StatusStuck
		} else if time.Since(agent.LastActive) > 2*time.Minute && !agent.HookAttached {
			agent.Status = StatusIdle
		}
	}
}

// getSessionName returns the expected tmux session name for an agent.
func (a *FSAdapter) getSessionName(agent *Agent) string {
	switch agent.Role {
	case RoleMayor:
		return "gt-mayor"
	case RoleDeacon:
		return "gt-deacon"
	case RoleWitness:
		return fmt.Sprintf("gt-%s-witness", agent.Rig)
	case RoleRefinery:
		return fmt.Sprintf("gt-%s-refinery", agent.Rig)
	case RolePolecat:
		return fmt.Sprintf("gt-%s-%s", agent.Rig, agent.Name)
	case RoleCrew:
		return fmt.Sprintf("gt-%s-crew-%s", agent.Rig, agent.Name)
	default:
		return ""
	}
}
