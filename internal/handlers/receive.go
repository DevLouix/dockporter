package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"dockporter/internal/dockerops"
	"dockporter/internal/types" // Import the types package containing the Manifest
)

// Response struct to send structured data back to the client
type MigrationResponse struct {
	Message       string `json:"message"`
	ContainerName string `json:"container_name"`
	Status        string `json:"status"`
}

func HandleReceive(dm *dockerops.DockerManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Method Validation
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 2. Extract and Decode the Manifest
		manifestB64 := r.Header.Get("X-Container-Manifest")
		if manifestB64 == "" {
			http.Error(w, "Missing X-Container-Manifest header", http.StatusBadRequest)
			return
		}

		// Decode Base64 string into raw bytes
		manifestBytes, err := base64.StdEncoding.DecodeString(manifestB64)
		if err != nil {
			log.Printf("❌ Invalid Base64 in manifest: %v", err)
			http.Error(w, "Invalid Base64 manifest", http.StatusBadRequest)
			return
		}

		// Unmarshal JSON bytes into our strong struct
		var manifest types.ContainerManifest
		if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
			log.Printf("❌ Invalid JSON in manifest: %v", err)
			http.Error(w, "Invalid JSON manifest payload", http.StatusBadRequest)
			return
		}

		// Basic validation of the manifest
		if manifest.Name == "" || manifest.Image == "" {
			http.Error(w, "Manifest missing crucial Name or Image fields", http.StatusBadRequest)
			return
		}

		// 3. Security Fix: Prevent Path Traversal by sanitizing the parsed name
		manifest.Name = filepath.Base(filepath.Clean(manifest.Name))

		ctx := r.Context()

		// 4. Name Conflict Resolution
		if _, err := dm.InspectContainer(ctx, manifest.Name); err == nil {
			newName := fmt.Sprintf("%s-migrated-%s", manifest.Name, time.Now().Format("150405"))
			log.Printf("⚠️ Name conflict for '%s'. Renaming to: '%s'\n", manifest.Name, newName)

			// Update the manifest so the rest of the logic uses the collision-free name
			manifest.Name = newName
		}

		log.Printf("📥 Incoming migration for: %s (Image: %s, Ports: %d)\n", manifest.Name, manifest.Image, len(manifest.Ports))

		// 5. Setup Paths
		restorePath := filepath.Join(os.TempDir(), fmt.Sprintf("dockporter-restored-%s", manifest.Name))

		// Cleanup: Ensure temporary extracted data is deleted when function exits
		defer func() {
			if err := os.RemoveAll(restorePath); err != nil {
				log.Printf("⚠️ Failed to clean up temp dir %s: %v", restorePath, err)
			}
		}()

		// 6. Streaming Import (Direct to disk, zero intermediate files)
		log.Println("📥 Streaming volume data directly to disk...")
		if err := dm.ImportVolumeFromStream(r.Body, restorePath); err != nil {
			log.Printf("❌ Streaming import failed: %v\n", err)
			http.Error(w, "Streaming import failed", http.StatusInternalServerError)
			return
		}

		// 7. Recreate Container using the FULL Manifest
		log.Printf("🏗️ Recreating container '%s' with preserved configuration...\n", manifest.Name)

		// Note: You will need to update your DockerManager to accept the `manifest` struct
		// instead of just strings, so it can actually apply the Ports, Env vars, and Labels!
		resp, err := dm.CreateMigratedContainer(ctx, manifest, restorePath)
		if err != nil {
			log.Printf("❌ Container creation failed: %v\n", err)
			http.Error(w, "Container creation failed", http.StatusInternalServerError)
			return
		}

		// 8. Start Container
		if err := dm.StartContainer(ctx, resp.ID); err != nil {
			log.Printf("❌ Container start failed: %v\n", err)
			http.Error(w, "Container start failed", http.StatusInternalServerError)
			return
		}

		log.Printf("✅ Container %s is now RUNNING!\n", manifest.Name)

		// 9. Structured JSON Response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(MigrationResponse{
			Message:       "Migration successful",
			ContainerName: manifest.Name,
			Status:        "RUNNING",
		})
	}
}
