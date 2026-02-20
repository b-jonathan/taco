package fastapi

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"

	"github.com/b-jonathan/taco/internal/stacks"
)

type Stack = stacks.Stack
type Options = stacks.Options

type fastapi struct{}

func New() Stack { return &fastapi{} }

func (fastapi) Type() string { return "backend" }
func (fastapi) Name() string { return "fastapi" }

func (fastapi) Init(ctx context.Context, opts *Options) error {
	backendDir := filepath.Join(opts.ProjectRoot, "backend")

	if err := os.MkdirAll(backendDir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	// initialize python environment
	if err := execx.RunCmd(ctx, backendDir, "python3 -m venv venv"); err != nil {
		return fmt.Errorf("create venv: %w", err)
	}

	return nil
}

func (fastapi) Generate(ctx context.Context, opts *Options) error {
	templateDir := "fastapi"
	outputDir := filepath.Join(opts.ProjectRoot, "backend")

	// generate files from template
	if err := fsutil.GenerateFromTemplateDir(templateDir, outputDir); err != nil {
		return err
	}

	return nil
}

func (fastapi) Post(ctx context.Context, opts *Options) error {
	// install dependencies
	backendDir := filepath.Join(opts.ProjectRoot, "backend")
	if err := execx.RunCmd(ctx, backendDir, "venv/bin/python -m pip install -r requirements.txt"); err != nil {
		return fmt.Errorf("install dependencies: %w", err)
	}

	gitignorePath := filepath.Join(opts.ProjectRoot, ".gitignore")
	if err := fsutil.EnsureFile(gitignorePath); err != nil {
		return fmt.Errorf("ensure gitignore file %w", err)
	}

	_ = fsutil.AppendUniqueLines(gitignorePath,
		[]string{"backend/__pycache__/", "backend/venv/", "backend/.env*"})

	path := filepath.Join(opts.ProjectRoot, "backend", ".env")
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	content := `PORT=4000 
FRONTEND_ORIGIN=http://localhost:3000`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return nil
}
