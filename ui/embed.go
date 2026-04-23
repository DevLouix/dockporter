package ui

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed dist/*
var staticFiles embed.FS

// GetFileSystem strips the "dist" prefix so the web server serves the files correctly
func GetFileSystem() http.FileSystem {
	// Extract the "dist" folder from the embedded filesystem
	distFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		log.Fatalf("❌ Failed to load embedded UI. Did you run 'npm run build' inside the ui folder?: %v", err)
	}
	return http.FS(distFS)
}