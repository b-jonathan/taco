package templates

import "embed"

//go:embed express/* firebase/* mongodb/* nextjs/* fastapi/*
var FS embed.FS
