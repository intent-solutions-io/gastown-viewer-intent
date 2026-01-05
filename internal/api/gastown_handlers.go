package api

import (
	"net/http"

	"github.com/intent-solutions-io/gastown-viewer-intent/internal/gastown"
)

// handleTownStatus handles GET /api/v1/town/status.
func (s *Server) handleTownStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status, err := s.gtAdapter.Status(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GASTOWN_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, status)
}

// handleTown handles GET /api/v1/town.
func (s *Server) handleTown(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	town, err := s.gtAdapter.Town(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GASTOWN_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, town)
}

// handleRigs handles GET /api/v1/town/rigs.
func (s *Server) handleRigs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rigs, err := s.gtAdapter.Rigs(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GASTOWN_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"rigs":  rigs,
		"total": len(rigs),
	})
}

// handleRig handles GET /api/v1/town/rigs/{name}.
func (s *Server) handleRig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")

	if name == "" {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "rig name required")
		return
	}

	rig, err := s.gtAdapter.Rig(ctx, name)
	if err != nil {
		writeError(w, http.StatusNotFound, "RIG_NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, rig)
}

// handleAgents handles GET /api/v1/town/agents.
func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	agents, err := s.gtAdapter.Agents(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GASTOWN_ERROR", err.Error())
		return
	}

	// Group by status
	var active, offline []gastown.Agent
	for _, a := range agents {
		if a.Status == gastown.StatusActive {
			active = append(active, a)
		} else {
			offline = append(offline, a)
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"agents":  agents,
		"total":   len(agents),
		"active":  len(active),
		"offline": len(offline),
	})
}

// handleConvoys handles GET /api/v1/town/convoys.
func (s *Server) handleConvoys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	convoys, err := s.gtAdapter.Convoys(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GASTOWN_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"convoys": convoys,
		"total":   len(convoys),
	})
}

// handleMail handles GET /api/v1/town/mail/{address}.
func (s *Server) handleMail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	address := r.PathValue("address")

	if address == "" {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "address required")
		return
	}

	messages, err := s.gtAdapter.Mail(ctx, address)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GASTOWN_ERROR", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"messages": messages,
		"total":    len(messages),
	})
}

// handleMolecules handles GET /api/v1/town/molecules.
func (s *Server) handleMolecules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	molecules, err := s.gtAdapter.Molecules(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "GASTOWN_ERROR", err.Error())
		return
	}

	// Group by status
	var inProgress, pending, complete, blocked int
	for _, m := range molecules {
		switch m.Status {
		case gastown.MolStatusInProgress:
			inProgress++
		case gastown.MolStatusPending:
			pending++
		case gastown.MolStatusComplete:
			complete++
		case gastown.MolStatusBlocked, gastown.MolStatusFailed:
			blocked++
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"molecules":   molecules,
		"total":       len(molecules),
		"in_progress": inProgress,
		"pending":     pending,
		"complete":    complete,
		"blocked":     blocked,
	})
}

// handleMolecule handles GET /api/v1/town/molecules/{id}.
func (s *Server) handleMolecule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	if id == "" {
		writeError(w, http.StatusBadRequest, "INVALID_PARAM", "molecule ID required")
		return
	}

	molecule, err := s.gtAdapter.Molecule(ctx, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "MOLECULE_NOT_FOUND", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, molecule)
}
