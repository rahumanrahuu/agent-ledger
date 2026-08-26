package ui

import "embed"

// Dist contains the embedded frontend assets (only populated in production builds)
// To embed the frontend:
// 1. cd ui/frontend && npm install
// 2. npm run build
// 3. The output will be in ../dist/
// 4. go build ./cmd/ledger will create the binary with embedded assets

//go:embed dist/*
var Dist embed.FS
