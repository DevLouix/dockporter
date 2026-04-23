package handlers

import (
	"encoding/json"
	"net/http"
	"sync"

	"dockporter/internal/dockerops"
)

// HandleListContainers returns all containers for the UI table
func HandleListContainers(dm *dockerops.DockerManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Use GET", http.StatusMethodNotAllowed)
			return
		}

		containers, err := dm.ListContainers(r.Context())
		if err != nil {
			http.Error(w, "Failed to list containers", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(containers)
	}
}

// ActionRequest is used for Multi-Select Start/Stop/Delete
type ActionRequest struct {
	Action       string   `json:"action"` // "start", "stop", "delete"
	ContainerIDs []string `json:"container_ids"`
	Force        bool     `json:"force"` // Used for forced deletion
}

// HandleContainerActions processes multi-select actions concurrently
func HandleContainerActions(dm *dockerops.DockerManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Use POST", http.StatusMethodNotAllowed)
			return
		}

		var req ActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Process concurrently for fast multi-select handling
		var wg sync.WaitGroup
		results := make(map[string]string)
		var mu sync.Mutex

		for _, id := range req.ContainerIDs {
			wg.Add(1)
			go func(cID string) {
				defer wg.Done()
				var err error

				switch req.Action {
				case "start":
					err = dm.StartContainer(r.Context(), cID)
				case "stop":
					err = dm.StopContainer(r.Context(), cID)
				case "delete":
					err = dm.RemoveContainer(r.Context(), cID, req.Force)
				default:
					err = http.ErrNotSupported
				}

				mu.Lock()
				if err != nil {
					results[cID] = err.Error()
				} else {
					results[cID] = "success"
				}
				mu.Unlock()
			}(id)
		}

		wg.Wait()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"action":  req.Action,
			"results": results,
		})
	}
}

// HandleRename is a dedicated endpoint for renaming single containers
func HandleRename(dm *dockerops.DockerManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Use POST", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ContainerID string `json:"container_id"`
			NewName     string `json:"new_name"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		if err := dm.RenameContainer(r.Context(), req.ContainerID, req.NewName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}
