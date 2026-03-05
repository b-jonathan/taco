package templates

import "embed"

//go:embed express/* firebase/* mongodb/* nextjs/* postgresql/*
var FS embed.FS
