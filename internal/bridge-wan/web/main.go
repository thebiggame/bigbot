package web

import "embed"

type IndexPageData struct {
	Version string
}

//go:embed *
var Templates embed.FS
