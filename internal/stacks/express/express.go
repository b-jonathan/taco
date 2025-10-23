package express

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

type express struct{}

func New() Stack { return &express{} }

func (express) Type() string { return "backend" }
func (express) Name() string { return "express" }

func (express) Init(ctx context.Context, opts *Options) error {
	backendDir := filepath.Join(opts.ProjectRoot, "backend")
	srcDir := filepath.Join(backendDir, "src")

	if err := os.MkdirAll(srcDir, 0o755); err != nil {
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
	}
	//TODO: Prob can Refactor this somewhere, like keeping track of depencies to be installed, not urgent tho
	if err := execx.RunCmd(ctx, backendDir, "npm install -D "+strings.Join(devDependencies, " ")); err != nil {
		return fmt.Errorf("npm install dev deps: %w", err)
	}

	return nil
}

func (express) Generate(ctx context.Context, opts *Options) error {
	backendDir := filepath.Join(opts.ProjectRoot, "backend")
	files := []fsutil.FileInfo{}
	tsconfigPath := filepath.Join(backendDir, "tsconfig.json")
	tsconfigContent, err := fsutil.RenderTemplate("express/tsconfig.json.tmpl")
	if err != nil {
		return err
	}
	tsconfig := fsutil.FileInfo{
		Path:    tsconfigPath,
		Content: tsconfigContent,
	}
	indexPath := filepath.Join(backendDir, "src", "index.ts")
	indexContent, err := fsutil.RenderTemplate("express/src/index.ts.tmpl")
	if err != nil {
		return err
	}

	index := fsutil.FileInfo{
		Path:    indexPath,
		Content: indexContent,
	}

	eslintPath := filepath.Join(backendDir, "eslint.config.mjs")
	eslintContent, err := fsutil.RenderTemplate("express/eslint.config.mjs.tmpl")

	if err != nil {
		return err
	}
	eslint := fsutil.FileInfo{
		Path:    eslintPath,
		Content: eslintContent,
	}

	prettierPath := filepath.Join(backendDir, ".prettierrc.json")
	prettierContent, err := fsutil.RenderTemplate("express/.prettierrc.json.tmpl")
	if err != nil {
		return err
	}
	prettier := fsutil.FileInfo{
		Path:    prettierPath,
		Content: prettierContent,
	}

	prettierIgnorePath := filepath.Join(backendDir, ".prettierignore")
	prettierIgnoreContent, err := fsutil.RenderTemplate("express/.prettierignore.tmpl")
	if err != nil {
		return err
	}

	prettierIgnore := fsutil.FileInfo{
		Path:    prettierIgnorePath,
		Content: prettierIgnoreContent,
	}
	files = append(files, tsconfig, index, eslint, prettier, prettierIgnore)

	if err := fsutil.WriteMultipleFiles(files); err != nil {
		return fmt.Errorf("write files: %w", err)
	}

	packageParams := nodepkg.InitPackageParams{
		Name: "express",
		Main: "dist/index.js",
		Scripts: map[string]string{
			"dev":        "tsx watch src/index.ts",
			"build":      "tsc -p tsconfig.json",
			"start":      "node dist/index.js",
			"lint-check": "eslint . && prettier --check .",
			"lint-fix":   "eslint . --fix && prettier --write .",
		}}

	if err := nodepkg.InitPackage(backendDir, packageParams); err != nil {
		return fmt.Errorf("write src/index.ts: %w", err)
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
