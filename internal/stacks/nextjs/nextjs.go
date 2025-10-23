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

func (nextjs) Name() string { return "express" }

func (nextjs) Init(ctx context.Context, opts *Options) error {
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
	eslintPath := filepath.Join(frontendDir, "eslint.config.mjs")
	eslintContent, err := fsutil.RenderTemplate("nextjs/eslint.config.mjs.tmpl")
	if err != nil {
		return err
	}
	eslint := fsutil.FileInfo{
		Path:    eslintPath,
		Content: eslintContent,
	}
	prettierPath := filepath.Join(frontendDir, ".prettierrc.json")
	prettierContent, err := fsutil.RenderTemplate("nextjs/.prettierrc.json.tmpl")
	if err != nil {
		return err
	}

	prettier := fsutil.FileInfo{
		Path:    prettierPath,
		Content: prettierContent,
	}

	prettierIgnorePath := filepath.Join(frontendDir, ".prettierignore")
	prettierIgnoreContent, err := fsutil.RenderTemplate("nextjs/.prettierignore.tmpl")
	if err != nil {
		return err
	}

	prettierIgnore := fsutil.FileInfo{
		Path:    prettierIgnorePath,
		Content: prettierIgnoreContent,
	}
	files := []fsutil.FileInfo{eslint, prettier, prettierIgnore}

	if err := fsutil.WriteMultipleFiles(files); err != nil {
		return fmt.Errorf("write files: %w", err)
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

func (nextjs) Post(ctx context.Context, opts *Options) error {
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
	pageContent, err := fsutil.RenderTemplate("nextjs/page.tsx.tmpl")
	if err != nil {
		return err
	}
	page := fsutil.FileInfo{
		Path:    pagePath,
		Content: pageContent,
	}

	if err := fsutil.WriteFile(page); err != nil {
		return err
	}
	return nil
}
