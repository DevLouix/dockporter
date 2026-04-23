package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"dockporter/internal/api"
	"dockporter/internal/config"
	"dockporter/internal/dockerops"
	"dockporter/internal/handlers"
	"dockporter/internal/orchestrator"
	"dockporter/internal/types"
	"dockporter/ui"
)

func main() {
	// 1. Define Flags
	mode := flag.String("mode", "server", "Mode: 'server' or 'ship'")
	containerIDs := flag.String("id", "", "Comma-separated IDs of containers to ship")
	remoteAddr := flag.String("to", "", "Remote agent address (e.g., 1.2.3.4:8080)")
	remoteToken := flag.String("token", "", "Auth token for the remote agent")
	port := flag.String("port", "8080", "Local port for server mode")
	showKey := flag.Bool("show-key", false, "Display this agent's Auth Token")
	flag.Parse()

	// 2. Load/Create Security Config
	cfg, err := config.GetOrCreateConfig(*port)
	if err != nil {
		log.Fatalf("❌ Configuration error: %v", err)
	}

	// Helper: Display key and exit
	if *showKey {
		fmt.Printf("\n🔑 YOUR AGENT AUTH TOKEN: %s\n", cfg.AuthToken)
		fmt.Println("Use this token when shipping containers TO this machine.")
		return
	}

	// 3. Initialize Docker
	ctx := context.Background()
	dm, err := dockerops.NewDockerManager(ctx)
	if err != nil {
		log.Fatalf("❌ Docker connection failed: %v", err)
	}
	defer dm.Close()

	// 4. Branch Logic (The CLI Switch)
	switch *mode {

	case "server":
		// Setup WebSocket Hub for real-time UI updates
		hub := api.NewHub()
		go hub.Run(ctx)

		mux := http.NewServeMux()

		// 2. Helper to apply AuthMiddleware ONLY to API routes
		protect := func(h http.HandlerFunc) http.Handler {
			return api.CorsMiddleware(api.AuthMiddleware(cfg.AuthToken, h))
		}

		// Wire handlers with dependencies
		mux.Handle("/ws", api.CorsMiddleware(http.HandlerFunc(hub.ServeWs)))

		mux.Handle("/receive", protect(handlers.HandleReceive(dm)))
		mux.Handle("/migrate", protect(handlers.HandleMigrate(dm, hub)))
		mux.Handle("/migrate-batch", protect(handlers.HandleBatchMigrate(dm, hub)))

		// Docker Explorer / Control Panel Endpoints
		mux.Handle("/containers", protect(handlers.HandleListContainers(dm)))
		mux.Handle("/containers/action", protect(handlers.HandleContainerActions(dm)))
		mux.Handle("/containers/rename", protect(handlers.HandleRename(dm)))

		// React UI Route (UNPROTECTED so the browser can load the page)
		// The file server acts as a "catch-all" for / and any static assets
		mux.Handle("/", http.FileServer(ui.GetFileSystem()))

		fmt.Printf("\n-----------------------------------------------------------\n")
		fmt.Printf("🐳 DockPorter by DevLouix...\n")
		fmt.Printf("🚀 Server Started Successfully!\n")
		fmt.Printf("🌍 Local Control Panel: http://localhost:%s?token=%s\n", *port, cfg.AuthToken)
		fmt.Printf("-----------------------------------------------------------\n\n")
		log.Fatal(http.ListenAndServe(":"+*port, mux))

	case "ship":
		// CLI MODE: Validation
		if *containerIDs == "" || *remoteAddr == "" || *remoteToken == "" {
			log.Fatal("❌ CLI 'ship' mode requires: -id, -to, AND -token")
		}

		ids := strings.Split(*containerIDs, ",")

		if len(ids) == 1 {
			// SINGLE SHIPMENT
			fmt.Printf("🚚 Shipping container %s to %s...\n", ids[0], *remoteAddr)
			err := orchestrator.ShipContainer(ctx, dm, ids[0], *remoteAddr, *remoteToken)
			if err != nil {
				log.Fatalf("❌ Shipment failed: %v", err)
			}
		} else {
			// BATCH SHIPMENT (CLI version)
			fmt.Printf("📦 Shipping batch of %d containers to %s...\n", len(ids), *remoteAddr)

			// We create a local channel to print progress to the terminal
			eventChan := make(chan types.MigrationEvent, 10)

			go func() {
				for ev := range eventChan {
					if ev.Error != "" {
						fmt.Printf("  - [%s] ❌ Failed: %v\n", ev.ContainerID, ev.Error)
					} else {
						fmt.Printf("  - [%s] %s (%d%%)\n", ev.ContainerID, ev.Status, ev.Progress)
					}
				}
			}()

			orchestrator.ShipBatch(ctx, dm, ids, *remoteAddr, *remoteToken, 3, eventChan)
			close(eventChan)
		}
		fmt.Println("✅ All CLI operations completed.")

	default:
		log.Fatalf("Unknown mode: %s. Use 'server' or 'ship'.", *mode)
	}
}
