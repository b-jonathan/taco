package firebase

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/b-jonathan/taco/internal/fsutil"
	"github.com/b-jonathan/taco/internal/stacks"
)

type Stack = stacks.Stack
type Options = stacks.Options

type express struct{}

func New() Stack { return &express{} }

func (express) Type() string { return "auth" }
func (express) Name() string { return "firebase" }

func (express) Init(ctx context.Context, opts *Options) error {
	return nil
}

func (express) Generate(ctx context.Context, opts *Options) error {
	return nil
}

func (express) Post(ctx context.Context, opts *Options) error {
	gitignorePath := filepath.Join(opts.ProjectRoot, ".gitignore")
	if err := fsutil.EnsureFile(gitignorePath); err != nil {
		return fmt.Errorf("ensure gitignore file: %w", err)
	}

	_ = fsutil.AppendUniqueLines(gitignorePath,
		[]string{"backend/node_modules/", "backend/dist/", "backend/.env*"})
	path := filepath.Join(opts.ProjectRoot, "backend", ".env")
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	content := `
		PORT=4000
		FRONTEND_ORIGIN=http://localhost:3000
		`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
