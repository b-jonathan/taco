package templates

import "embed"

//go:embed express/* firebase/* mongodb/* nextjs/* vite/*
var FS embed.FS
