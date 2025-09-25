package stacks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"
)

type nextjs struct{}

func NextJS() Stack { return &nextjs{} }

func (nextjs) Type() string { return "frontend" }

func (nextjs) Name() string { return "express" }

func (nextjs) Init(ctx context.Context, opts Options) error {
	if err := os.MkdirAll(opts.ProjectRoot, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	// 1) Scaffold Next.js in TS, without ESLint, noninteractive
	// Requires an execx.Npx() helper on Windows; use "npx" if you don't have one yet.
	nextFlags := []string{
		"--yes",
		"create-next-app@latest",
		"frontend",
		"--ts",
		"--no-eslint",
		"--app",
		"--tailwind",
		"--src-dir",
		"--import-alias", "@/*",
		"--use-npm",
		"--disable-git",
		"--turbopack",
	}

	if err := execx.RunCmd(ctx, opts.ProjectRoot, "npx", nextFlags...); err != nil {
		return fmt.Errorf("create-next-app: %w", err)
	}
	return nil
}

func (nextjs) Generate(ctx context.Context, opts Options) error {
	frontendDir := filepath.Join(opts.ProjectRoot, "frontend")
	// Repo root gitignore entries for the frontend
	gitignorePath := filepath.Join(opts.ProjectRoot, ".gitignore")
	if err := fsutil.EnsureFile(gitignorePath); err != nil {
		return fmt.Errorf("ensure .gitignore: %w", err)
	}
	if err := fsutil.AppendUniqueLines(gitignorePath, []string{
		"frontend/node_modules/",
		"frontend/.next/",
		"frontend/.env.local",
	}); err != nil {
		return err
	}

	// Create an env placeholder
	if err := fsutil.EnsureFile(filepath.Join(frontendDir, ".env.local")); err != nil {
		return fmt.Errorf("ensure .env.local: %w", err)
	}
	return nil
}

func (nextjs) Post(ctx context.Context, opts Options) error {
	// Create an env placeholder
	frontendDir := filepath.Join(opts.ProjectRoot, "frontend")
	if err := fsutil.EnsureFile(filepath.Join(frontendDir, ".env.local")); err != nil {
		return fmt.Errorf("ensure .env.local: %w", err)
	}

	pagePath := filepath.Join(frontendDir, "src", "app", "page.tsx")
	pageContent := `
		"use client";
		import { useEffect, useState } from "react";
		export default function Home() {
		const [message, setMessage] = useState<string>("loading...");
		useEffect(() => {
			fetch("http://localhost:4000/")
			.then((res) => res.text())
			.then(setMessage)
			.catch((err) => setMessage("error: " + err.message));
		}, []);
		return <div>{message}</div>;
		}
		`
	if err := os.WriteFile(pagePath, []byte(pageContent), 0o644); err != nil {
		return fmt.Errorf("write page.tsx: %w", err)
	}
	return nil
}
