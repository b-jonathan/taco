package stacks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"
	"github.com/b-jonathan/taco/internal/nodepkg"
)

type express struct{}

func Express() Stack { return &express{} }

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
		"typescript", "tsx", "@types/node", "@types/express", "@types/cors"); err != nil {
		return fmt.Errorf("npm install dev deps: %w", err)
	}

	tsconfigPath := filepath.Join(backendDir, "tsconfig.json")
	if err := fsutil.EnsureFile(tsconfigPath); err != nil {
		return fmt.Errorf("ensure tsconfig file: %w", err)
	}

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
	if err := os.WriteFile(tsconfigPath, []byte(tsconfig), 0o644); err != nil {
		return fmt.Errorf("write tsconfig.json: %w", err)
	}

	indexPath := filepath.Join(backendDir, "src", "index.ts")
	if err := fsutil.EnsureFile(indexPath); err != nil {
		return fmt.Errorf("ensure index file: %w", err)
	}
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

	if err := os.WriteFile(indexPath, []byte(index), 0o644); err != nil {
		return fmt.Errorf("write src/index.ts: %w", err)
	}

	packageParams := nodepkg.InitPackageParams{
		Name: "express",
		Main: "dist/index.js",
		Scripts: map[string]string{
			"dev":   "tsx watch src/index.ts",
			"build": "tsc -p tsconfig.json",
			"start": "node dist/index.js",
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
