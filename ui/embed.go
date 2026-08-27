package ui

import "embed"

// Dist contains the embedded frontend assets (only populated in production builds)
//
// IMPORTANT: The ui/dist directory must exist before building the Go binary.
// To embed the frontend:
// 1. cd ui/frontend && npm install
// 2. npm run build
// 3. The output will be in ../dist/
// 4. go build ./cmd/ledger will create the binary with embedded assets
//
// The GoReleaser configuration automatically builds the frontend before Go compilation.
// For local development builds, manually build the frontend first.
//
//go:embed all:dist/*
var Dist embed.FS
