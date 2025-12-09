package frontend

import "embed"

// DistFS contains the embedded frontend assets.
//
//go:embed dist/*
var DistFS embed.FS

// GetDistFS returns the embedded filesystem for the dist directory.
func GetDistFS() embed.FS {
	return DistFS
}
