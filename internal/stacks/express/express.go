package express

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

type express struct{}

func New() Stack { return &express{} }

func (express) Type() string { return "backend" }
func (express) Name() string { return "express" }

func (express) Init(ctx context.Context, opts Options) error {
	backendDir := filepath.Join(opts.ProjectRoot, "backend")
	srcDir := filepath.Join(backendDir, "src")

	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return nil
}

func (express) Generate(ctx context.Context, opts Options) error {

	backendDir := filepath.Join(opts.ProjectRoot, "backend")
	if err := execx.RunCmd(ctx, backendDir, "npm", "init", "-y"); err != nil {
		return fmt.Errorf("npm init: %w", err)
	}

	if err := execx.RunCmd(ctx, backendDir, "npm", "install", "express", "cors", "dotenv"); err != nil {
		return fmt.Errorf("npm install express: %w", err)
	}

	if err := execx.RunCmd(ctx, backendDir, "npm", "install", "-D",
		"typescript", "tsx",
		"@types/node",
		"@types/express",
		"@types/cors",
		"eslint",
		"@eslint/js",
		"globals",
		"typescript",
		"typescript-eslint",
		"eslint-plugin-n",
		"eslint-config-prettier",
		"prettier"); err != nil {
		return fmt.Errorf("npm install dev deps: %w", err)
	}

	tsconfigPath := filepath.Join(backendDir, "tsconfig.json")
	tsconfig := `
		{
	"compilerOptions": {
		"target": "es2022",
		"module": "CommonJS",
		"strict": true,
		"esModuleInterop": true,
		"skipLibCheck": true,
		"forceConsistentCasingInFileNames": true,
		"outDir": "dist",
		"rootDir": "src",
		"noImplicitOverride": true,        
	},
	"include": ["src"],
	"exclude": ["node_modules", "dist"]
	}
	`
	if err := fsutil.WriteFile(tsconfigPath, []byte(tsconfig)); err != nil {
		return err
	}

	indexPath := filepath.Join(backendDir, "src", "index.ts")
	// src/index.ts
	index := `
		import "dotenv/config"; // auto-loads .env into process.env
		import express from "express"; 
		import cors from "cors"; // connects to frontend

		const app = express();
		const PORT = process.env.PORT || 3000;

		app.use(express.json());

		app.use(
		cors({
			origin: process.env.FRONTEND_ORIGIN,
		})
		);

		app.get("/", (_req, res) => {
		res.send("Hello, Express + TypeScript!");
		});

		app.listen(PORT, () => {
		console.log("Server listening on http://localhost:" + PORT);
		});
		`

	if err := fsutil.WriteFile(indexPath, []byte(index)); err != nil {
		return err
	}

	eslintPath := filepath.Join(backendDir, "eslint.config.mjs")
	eslint := `
	// eslint.config.mjs
	import js from '@eslint/js';
	import ts from 'typescript-eslint';
	import n from 'eslint-plugin-n';
	import globals from 'globals';
	import prettier from 'eslint-config-prettier';

	export default [
	{ ignores: ["**/node_modules/**","**/.next/**","**/.turbo/**","**/dist/**","**/build/**","**/coverage/**","**/.vercel/**","**/.cache/**"] },
	js.configs.recommended,
	...ts.configs.recommendedTypeChecked,
	n.configs['flat/recommended'],
	{
		files: ['src/**/*.{ts,tsx,js,cjs,mjs}'],
		languageOptions: {
		globals: { ...globals.node },
		parserOptions: {
			projectService: true,
			tsconfigRootDir: import.meta.dirname,
			ecmaVersion: 'latest',
			sourceType: 'module'
		}
		}
	},
	prettier
	];
	`
	if err := fsutil.WriteFile(eslintPath, []byte(eslint)); err != nil {
		return err
	}

	prettierPath := filepath.Join(backendDir, ".prettierrc.json")
	prettier := `
	{
	"tabWidth": 2,
	"semi": true,
	"singleQuote": false,
	"trailingComma": "all"
	}
	`
	if err := fsutil.WriteFile(prettierPath, []byte(prettier)); err != nil {
		return err
	}

	prettierIgnorePath := filepath.Join(backendDir, ".prettierignore")
	prettierIgnore := `
	# See https://help.github.com/articles/ignoring-files/ for more about ignoring files.

	# dependencies
	/node_modules
	/.pnp
	.pnp.js

	# testing
	/coverage

	# production
	/build

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

	/dist
	`

	if err := fsutil.WriteFile(prettierIgnorePath, []byte(prettierIgnore)); err != nil {
		return err
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

func (express) Post(ctx context.Context, opts Options) error {
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
