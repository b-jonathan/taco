package templates

import "embed"

//go:embed express/* firebase/* mongodb/* nextjs/*
var FS embed.FS
