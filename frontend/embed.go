package frontend

import (
	"embed"
	"io/fs"
)

//go:embed dist/*
var dist embed.FS

func GetDistFS() (fs.FS, error) {
	return fs.Sub(dist, "dist")
}
