package web

import (
	"embed"
	"io/fs"
)

//go:embed all:build
var buildFiles embed.FS

var BuildFS fs.FS

func init() {
	var err error
	BuildFS, err = fs.Sub(buildFiles, "build")
	if err != nil {
		panic("web: failed to get build sub-filesystem: " + err.Error())
	}
}

