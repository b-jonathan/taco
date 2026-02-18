package express

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"
	"github.com/b-jonathan/taco/internal/nodepkg"
	"github.com/spf13/afero"

	"github.com/b-jonathan/taco/internal/stacks"
)

type Stack = stacks.Stack
type Options = stacks.Options

type express struct{}

func New() Stack { return &express{} }

func (express) Type() string { return "backend" }
func (express) Name() string { return "express" }

func (express) Init(ctx context.Context, opts *Options) error {
	backendDir := filepath.Join(opts.ProjectRoot, "backend")
	srcDir := filepath.Join(backendDir, "src")

	if err := fsutil.Fs.MkdirAll(srcDir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := execx.RunCmd(ctx, backendDir, "npm init -y"); err != nil {
		return fmt.Errorf("npm init: %w", err)
	}
	dependencies := []string{
		"express",
		"cors",
		"dotenv",
	}
	if err := execx.RunCmd(ctx, backendDir, "npm install "+strings.Join(dependencies, " ")); err != nil {
		return fmt.Errorf("npm install express: %w", err)
	}
	devDependencies := []string{
		"typescript",
		"ts-node",
		"@types/node",
		"@types/express",
		"@types/cors",
		"eslint",
		"@eslint/js",
		"globals",
		"typescript-eslint",
		"eslint-plugin-n",
		"eslint-config-prettier",
		"prettier",
		"tsx",
	}
	//TODO: Prob can Refactor this somewhere, like keeping track of depencies to be installed, not urgent tho
	if err := execx.RunCmd(ctx, backendDir, "npm install -D "+strings.Join(devDependencies, " ")); err != nil {
		return fmt.Errorf("npm install dev deps: %w", err)
	}

	return nil
}

func (express) Generate(ctx context.Context, opts *Options) error {
	templateDir := "express"
	outputDir := filepath.Join(opts.ProjectRoot, "backend")

	if err := fsutil.GenerateFromTemplateDir(templateDir, outputDir); err != nil {
		return err
	}

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
	if err := fsutil.Fs.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	content := `PORT=4000
FRONTEND_ORIGIN=http://localhost:3000`
	if err := afero.WriteFile(fsutil.Fs, path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	params := nodepkg.InitPackageParams{
		Name: "backend",
		Main: "src/index.ts",
		Scripts: map[string]string{
			"build":      "tsc -p tsconfig.json",
			"dev":        "tsx watch src/index.ts",
			"lint-check": "eslint . && prettier --check .",
			"lint-fix":   "eslint . --fix && prettier --write .",
			"start":      "node dist/index.js",
			"test":       "echo \"Error: no test specified\" && exit 1",
		},
	}
	backendDir := filepath.Join(opts.ProjectRoot, "backend")
	if err := nodepkg.InitPackage(backendDir, params); err != nil {
		return fmt.Errorf("init express package.json: %w", err)
	}

	return nil
}

func (express) Rollback(ctx context.Context, opts *Options) error {
	backendDir := filepath.Join(opts.ProjectRoot, "backend")

	if err := fsutil.RemoveDir(backendDir); err != nil {
		return fmt.Errorf("remove backend dir: %w", err)
	}

	return nil
}
