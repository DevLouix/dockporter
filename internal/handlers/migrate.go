package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"dockporter/internal/api"
	"dockporter/internal/dockerops"
	"dockporter/internal/orchestrator"
	"dockporter/internal/types"
)

type MigrateRequest struct {
	ContainerID string `json:"container_id"`
	RemoteAddr  string `json:"remote_addr"`
	RemoteToken string `json:"remote_token"` // Added this!
}

type BatchMigrateRequest struct {
	ContainerIDs []string `json:"container_ids"`
	RemoteAddr   string   `json:"remote_addr"`
	RemoteToken  string   `json:"remote_token"` // Added this!
	Concurrency  int      `json:"concurrency"`
}

// HandleMigrate triggers a single container migration
func HandleMigrate(dm *dockerops.DockerManager, hub *api.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req MigrateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// 2. Validate the new token field
		if req.ContainerID == "" || req.RemoteAddr == "" || req.RemoteToken == "" {
			http.Error(w, "Missing container_id, remote_addr, or remote_token", http.StatusBadRequest)
			return
		}

		jobID := uuid.New().String()

		go func() {
			log.Printf("🚀 [Job %s] Starting migration: %s -> %s", jobID, req.ContainerID, req.RemoteAddr)

			// Update UI: Pending
			hub.Publish(types.MigrationEvent{
				ContainerID: req.ContainerID,
				Status:      types.StatusPending,
				Timestamp:   time.Now(),
			})

			// 3. FIX THE ERROR: Pass req.RemoteToken as the 5th argument!
			err := orchestrator.ShipContainer(context.Background(), dm, req.ContainerID, req.RemoteAddr, req.RemoteToken)

			if err != nil {
				log.Printf("❌ [Job %s] Migration failed: %v", jobID, err)
				hub.Publish(types.MigrationEvent{
					ContainerID: req.ContainerID,
					Status:      types.StatusFailed,
					Error:       err.Error(),
					Timestamp:   time.Now(),
				})
			} else {
				log.Printf("✅ [Job %s] Migration successful", jobID)
				hub.Publish(types.MigrationEvent{
					ContainerID: req.ContainerID,
					Status:      types.StatusSuccess,
					Progress:    100,
					Timestamp:   time.Now(),
				})
			}
		}()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"job_id": jobID,
			"status": "initiated",
		})
	}
}

func HandleBatchMigrate(dm *dockerops.DockerManager, hub *api.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Enforce HTTP Method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 2. Strict JSON Decoding
		var req BatchMigrateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// 3. Input Validation (Now checks for RemoteToken!)
		if len(req.ContainerIDs) == 0 {
			http.Error(w, "No container IDs provided", http.StatusBadRequest)
			return
		}
		if req.RemoteAddr == "" || req.RemoteToken == "" {
			http.Error(w, "Remote address and token are required", http.StatusBadRequest)
			return
		}

		// 4. Safe Resource Clamping
		if req.Concurrency <= 0 {
			req.Concurrency = 3
		}
		if req.Concurrency > 10 {
			req.Concurrency = 10
		}

		// 5. Create a Trackable Job ID
		jobID := uuid.New().String()

		// 6. Setup Async Context & Channels
		jobCtx, cancel := context.WithCancel(context.Background())
		eventChan := make(chan types.MigrationEvent, 100)

		// ---------------------------------------------------------
		// 7. THE EVENT ROUTER (Console Logs + WebSocket Broadcast)
		// ---------------------------------------------------------
		go func() {
			for ev := range eventChan {
				statusLog := fmt.Sprintf("[Job: %s] 📦 %s -> %s (%d%%)",
					jobID, ev.ContainerID, ev.Status, ev.Progress)

				if ev.Error != "" {
					log.Printf("%s | ❌ Error: %s", statusLog, ev.Error)
				} else {
					log.Printf("%s", statusLog)
				}

				hub.Publish(ev)
			}
			log.Printf("📥 Event stream for Job %s closed.", jobID)
		}()

		// ---------------------------------------------------------
		// 8. THE EXECUTION ENGINE
		// ---------------------------------------------------------
		go func() {
			defer cancel()
			defer close(eventChan)

			log.Printf("🚀 Starting Batch Job %s (%d containers, concurrency: %d)",
				jobID, len(req.ContainerIDs), req.Concurrency)

			// FIX: Added req.RemoteToken as the 5th argument!
			orchestrator.ShipBatch(jobCtx, dm, req.ContainerIDs, req.RemoteAddr, req.RemoteToken, req.Concurrency, eventChan)

			log.Printf("🏁 Batch Job %s process complete.", jobID)

			hub.Publish(types.MigrationEvent{
				Status:    types.StatusSuccess,
				Timestamp: time.Now(),
				Error:     fmt.Sprintf("Job %s finished", jobID),
			})
		}()

		// 9. Respond to caller immediately
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"job_id": jobID,
			"status": "queued",
		})
	}
}
