package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dockporter/internal/dockerops"
	"dockporter/internal/types" // Alias this if you have naming conflicts
)

// ShipBatch moves many containers with a concurrency limit and strict context control.
// FIX: Now accepts remoteToken to satisfy security requirements of the destination.
func ShipBatch(
	ctx context.Context,
	dm *dockerops.DockerManager,
	ids []string,
	remoteAddr string,
	remoteToken string,
	concurrency int,
	eventLog chan<- types.MigrationEvent,
) {
	var wg sync.WaitGroup

	// Ensure concurrency is at least 1
	if concurrency <= 0 {
		concurrency = 1
	}
	sem := make(chan struct{}, concurrency)

	// Helper to safely emit events and automatically attach timestamps
	emit := func(id string, status types.MigrationStatus, progress int, errStr string) {
		// Prevent panic if channel is closed elsewhere
		defer func() { recover() }()

		eventLog <- types.MigrationEvent{
			ContainerID: id,
			Status:      status,
			Progress:    progress,
			Error:       errStr,
			Timestamp:   time.Now(),
		}
	}

	for _, id := range ids {
		wg.Add(1)

		go func(cid string) {
			defer wg.Done()

			// 1. Mark as Pending immediately
			emit(cid, types.StatusPending, 0, "")

			// 2. Acquire Semaphore (Context-Aware)
			select {
			case sem <- struct{}{}:
				// Lock acquired
				defer func() { <-sem }()
			case <-ctx.Done():
				emit(cid, types.StatusFailed, 0, "Batch cancelled: waiting for slot")
				return
			}

			// 3. Final Pre-flight Check
			if err := ctx.Err(); err != nil {
				emit(cid, types.StatusFailed, 0, "Batch cancelled")
				return
			}

			// 4. Update Status to Sending (matches frontend state)
			emit(cid, types.StatusSending, 10, "")

			// 5. Execute the migration
			// FIX: Passing the remoteToken down to ShipContainer
			err := ShipContainer(ctx, dm, cid, remoteAddr, remoteToken)

			// 6. Handle Results
			if err != nil {
				emit(cid, types.StatusFailed, 10, err.Error())
			} else {
				emit(cid, types.StatusSuccess, 100, "")
			}
		}(id)
	}

	// 7. Final Sync
	// We wait for all goroutines to finish.
	// Note: The caller (the HTTP handler) is responsible for closing the eventLog channel.
	wg.Wait()
	fmt.Printf("✅ Batch process for %d containers finished.\n", len(ids))
}
