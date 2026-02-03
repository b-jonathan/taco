package nextjs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func (nextjs) Name() string { return "nextjs" }

func (nextjs) Init(ctx context.Context, opts *Options) error {
	if err := os.MkdirAll(opts.ProjectRoot, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	// 1) Scaffold Next.js in TS, without ESLint, noninteractive
	// Requires an execx.Npx() helper on Windows; use "npx" if you don't have one yet.
	nextFlags := []string{
		"--yes",
		"create-next-app@16.0.0",
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
		"--no-react-compiler",
	}

	//TODO: This is a patch fix, prob need a helper in general to parse []string to string

	if err := execx.RunCmd(ctx, opts.ProjectRoot, "npx "+strings.Join(nextFlags, " ")); err != nil {
		return fmt.Errorf("create-next-app: %w", err)
	}
	frontendDir := filepath.Join(opts.ProjectRoot, "frontend")
	frontendDeps := []string{
		"eslint",
		"@eslint/js",
		"globals",
		"typescript",
		"typescript-eslint",
		"@next/eslint-plugin-next",
		"eslint-plugin-react-hooks",
		"eslint-config-prettier",
		"prettier",
		"prettier-plugin-tailwindcss",
	}
	if err := execx.RunCmd(ctx, frontendDir, "npm install -D "+strings.Join(frontendDeps, " ")); err != nil {
		return fmt.Errorf("npm install dev deps: %w", err)
	}
	return nil
}

func (nextjs) Generate(ctx context.Context, opts *Options) error {
	frontendDir := filepath.Join(opts.ProjectRoot, "frontend")

	templateDir := "nextjs"
	outputDir := filepath.Join(frontendDir)

	if err := fsutil.GenerateFromTemplateDir(templateDir, outputDir); err != nil {
		return fmt.Errorf("generate nextjs templates: %w", err)
	}

	packageParams := nodepkg.InitPackageParams{
		Name: "nextjs",
		Scripts: map[string]string{
			"lint-check": "next lint && prettier --check .",
			"lint-fix":   "(next lint --fix || true) && prettier --write .",
		}}

	if err := nodepkg.InitPackage(frontendDir, packageParams); err != nil {
		return fmt.Errorf("init nextjs package.json: %w", err)
	}

	return nil
}

func (nextjs) Post(ctx context.Context, opts *Options) error {
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
	content := `NEXT_PUBLIC_BACKEND_URL=http://localhost:4000`
	if err := os.WriteFile(envPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", envPath, err)
	}

	return nil
}
