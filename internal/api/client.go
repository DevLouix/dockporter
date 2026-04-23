package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"dockporter/internal/types"
)

func SendMigrationStream(ctx context.Context, remoteAddr string, token string, manifest types.ContainerManifest, stream io.Reader) error {
	// Support both http and https (defaults to http for now)
	url := fmt.Sprintf("http://%s/receive", remoteAddr)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, stream)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 1. AUTHENTICATION: Set the security token
	req.Header.Set("X-Auth-Token", token)

	// 2. METADATA: Encode the manifest
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to encode manifest: %w", err)
	}
	req.Header.Set("X-Container-Manifest", base64.StdEncoding.EncodeToString(manifestBytes))
	req.Header.Set("Content-Type", "application/x-tar-gz")

	// 3. HARDENED CLIENT
	client := &http.Client{
		Transport: &http.Transport{
			ResponseHeaderTimeout: 60 * time.Second, // Wait up to a minute for the receiver to accept
			DisableKeepAlives:     false,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network error during migration: %w", err)
	}
	defer resp.Body.Close()

	// 4. ERROR HANDLING
	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("remote rejected migration (Status %d): %s", resp.StatusCode, string(errBody))
	}

	return nil
}
