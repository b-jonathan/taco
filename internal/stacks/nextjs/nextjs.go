package nextjs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"
	"github.com/b-jonathan/taco/internal/nodepkg"
	"github.com/b-jonathan/taco/internal/stacks"
)

type Stack = stacks.Stack
type Options = stacks.Options

type nextjs struct{}

func New() Stack { return &nextjs{} }

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
	frontendDir := filepath.Join(opts.ProjectRoot, "frontend")
	if err := execx.RunCmd(ctx, frontendDir, "npm", "install", "-D",
		"eslint",
		"@eslint/js",
		"globals",
		"typescript",
		"typescript-eslint",
		"@next/eslint-plugin-next",
		"eslint-plugin-react-hooks",
		"eslint-config-prettier",
		"prettier",
		"prettier-plugin-tailwindcss"); err != nil {
		return fmt.Errorf("npm install dev deps: %w", err)
	}
	return nil
}

func (nextjs) Generate(ctx context.Context, opts Options) error {
	frontendDir := filepath.Join(opts.ProjectRoot, "frontend")
	eslintPath := filepath.Join(frontendDir, "eslint.config.mjs")
	eslint := `
	// eslint.config.mjs
	/* eslint-disable */
	import js from '@eslint/js';
	import globals from 'globals';
	import ts from 'typescript-eslint';
	import next from '@next/eslint-plugin-next';
	import reactHooks from 'eslint-plugin-react-hooks';

	export default [
	{ ignores: ['node_modules/**','**/.next/**','**/.turbo/**','**/dist/**','**/build/**','**/coverage/**','**/.vercel/**','**/.cache/**'] },
	js.configs.recommended,
	...ts.configs.recommendedTypeChecked,
	next.configs.recommended,
	{
		files: ['src/**/*.{ts,tsx,js,jsx}'],
		languageOptions: {
		globals: { ...globals.browser, ...globals.node },
		parserOptions: { projectService: true, tsconfigRootDir: import.meta.dirname }
		},
		plugins: { 'react-hooks': reactHooks },
		rules: {
		'react-hooks/rules-of-hooks': 'error',
		'react-hooks/exhaustive-deps': 'warn',
		}
	}
	];
	`
	if err := fsutil.WriteFile(eslintPath, []byte(eslint)); err != nil {
		return err
	}

	prettierPath := filepath.Join(frontendDir, ".prettierrc.json")
	prettier := `
	{
	"tabWidth": 2,
	"semi": true,
	"singleQuote": false,
	"trailingComma": "all",
	"plugins": ["prettier-plugin-tailwindcss"]
	}
	`
	if err := fsutil.WriteFile(prettierPath, []byte(prettier)); err != nil {
		return err
	}

	prettierIgnorePath := filepath.Join(frontendDir, ".prettierignore")
	prettierIgnore := `
	# Do not run Prettier on these paths. Customize as needed.
	.next/
	build/
	dist/
	out/
	public/


	# testing
	/coverage

	# misc
	.DS_Store
	.env.local
	.env.development.local
	.env.test.local
	.env.production.local

	npm-debug.log*
	yarn-debug.log*
	yarn-error.log*

	# logs
	/logs

	`

	if err := fsutil.WriteFile(prettierIgnorePath, []byte(prettierIgnore)); err != nil {
		return err
	}

	packageParams := nodepkg.InitPackageParams{
		Name: "express",
		Main: "dist/index.js",
		Scripts: map[string]string{
			"lint-check": "next lint && prettier --check .",
			"lint-fix":   "(next lint --fix || true) && prettier --write .",
		}}

	if err := nodepkg.InitPackage(frontendDir, packageParams); err != nil {
		return fmt.Errorf("write src/index.ts: %w", err)
	}

	return nil
}

func (nextjs) Post(ctx context.Context, opts Options) error {
	gitignorePath := filepath.Join(opts.ProjectRoot, ".gitignore")
	if err := fsutil.WithFileLock(gitignorePath, func() error {
		if err := fsutil.EnsureFile(gitignorePath); err != nil {
			return err
		}
		_ = fsutil.AppendUniqueLines(gitignorePath, []string{"backend/node_modules/", "backend/dist/", "backend/.env*"})
		return nil
	}); err != nil {
		return fmt.Errorf("file Lock: %w", err)
	}
	// Create an env placeholder

	frontendDir := filepath.Join(opts.ProjectRoot, "frontend")
	envPath := filepath.Join(frontendDir, ".env.local")
	if err := fsutil.EnsureFile(envPath); err != nil {
		return fmt.Errorf("ensure .env.local: %w", err)
	}

	dir := filepath.Dir(envPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	content := `NEXT_PUBLIC_BACKEND_URL=http://localhost:4000	
		`
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", envPath, err)
	}
	pagePath := filepath.Join(frontendDir, "src", "app", "page.tsx")
	page := `
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
	if err := fsutil.WriteFile(pagePath, []byte(page)); err != nil {
		return err
	}
	return nil
}
