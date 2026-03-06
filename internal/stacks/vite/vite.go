package vite

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"
	"github.com/b-jonathan/taco/internal/nodepkg"
	"github.com/b-jonathan/taco/internal/stacks"
	"github.com/spf13/afero"
)

type Stack = stacks.Stack
type Options = stacks.Options

type vite struct{}

func New() Stack { return &vite{} }

func (vite) Type() string { return "frontend" }

func (vite) Name() string { return "vite" }

func (vite) Init(ctx context.Context, opts *Options) error {
	if err := fsutil.Fs.MkdirAll(opts.ProjectRoot, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// 1) Scaffold Vite with React + TypeScript template
	viteFlags := []string{
		"--yes",
		"create-vite@latest",
		"frontend",
		"--template", "react-ts",
	}

	if err := execx.RunCmd(ctx, opts.ProjectRoot, "npx "+strings.Join(viteFlags, " ")); err != nil {
		return fmt.Errorf("create-vite: %w", err)
	}

	frontendDir := filepath.Join(opts.ProjectRoot, "frontend")

	// 2) Install base dependencies
	if err := execx.RunCmd(ctx, frontendDir, "npm install"); err != nil {
		return fmt.Errorf("npm install: %w", err)
	}

	// 3) Install React Router
	if err := execx.RunCmd(ctx, frontendDir, "npm install react-router"); err != nil {
		return fmt.Errorf("npm install react-router: %w", err)
	}

	// 4) Install Tailwind CSS and its dependencies
	tailwindDeps := []string{
		"tailwindcss",
		"@tailwindcss/vite",
	}
	if err := execx.RunCmd(ctx, frontendDir, "npm install -D "+strings.Join(tailwindDeps, " ")); err != nil {
		return fmt.Errorf("npm install tailwind deps: %w", err)
	}

	// 5) Install ESLint + Prettier
	devDeps := []string{
		"eslint",
		"@eslint/js",
		"globals",
		"typescript",
		"typescript-eslint",
		"eslint-plugin-react-hooks",
		"eslint-plugin-react-refresh",
		"eslint-config-prettier",
		"prettier",
		"prettier-plugin-tailwindcss",
	}
	if err := execx.RunCmd(ctx, frontendDir, "npm install -D "+strings.Join(devDeps, " ")); err != nil {
		return fmt.Errorf("npm install dev deps: %w", err)
	}

	return nil
}

func (vite) Generate(ctx context.Context, opts *Options) error {
	frontendDir := filepath.Join(opts.ProjectRoot, "frontend")

	templateDir := "vite"
	outputDir := filepath.Join(frontendDir)

	if err := fsutil.GenerateFromTemplateDir(templateDir, outputDir); err != nil {
		return fmt.Errorf("generate vite templates: %w", err)
	}

	packageParams := nodepkg.InitPackageParams{
		Name: "frontend",
		Scripts: map[string]string{
			"lint-check": "eslint . && prettier --check .",
			"lint-fix":   "eslint . --fix && prettier --write .",
		},
	}

	if err := nodepkg.InitPackage(frontendDir, packageParams); err != nil {
		return fmt.Errorf("init vite package.json: %w", err)
	}

	return nil
}

func (vite) Post(ctx context.Context, opts *Options) error {
	frontendDir := filepath.Join(opts.ProjectRoot, "frontend")

	// Create .env file with VITE_BACKEND_URL
	envPath := filepath.Join(frontendDir, ".env")
	if err := fsutil.EnsureFile(envPath); err != nil {
		return fmt.Errorf("ensure .env: %w", err)
	}

	dir := filepath.Dir(envPath)
	if err := fsutil.Fs.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	content := `VITE_BACKEND_URL=http://localhost:4000`
	if err := afero.WriteFile(fsutil.Fs, envPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", envPath, err)
	}

	// Update root .gitignore
	gitignorePath := filepath.Join(opts.ProjectRoot, ".gitignore")
	if err := fsutil.EnsureFile(gitignorePath); err != nil {
		return fmt.Errorf("ensure gitignore file: %w", err)
	}

	_ = fsutil.AppendUniqueLines(gitignorePath,
		[]string{"frontend/node_modules/", "frontend/dist/", "frontend/.env*"})

	return nil
}
