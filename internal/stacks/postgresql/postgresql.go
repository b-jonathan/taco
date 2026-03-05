package postgresql

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"
	"github.com/b-jonathan/taco/internal/prompt"
	"github.com/b-jonathan/taco/internal/stacks"
)

type Stack = stacks.Stack
type Options = stacks.Options

type postgresql struct{}

func New() Stack { return &postgresql{} }

func (postgresql) Type() string { return "database" }
func (postgresql) Name() string { return "postgres" }

// EnsurePostgresURI validates that the URI starts with postgresql:// or postgres://
func EnsurePostgresURI(uri string) error {
	if !strings.HasPrefix(uri, "postgresql://") && !strings.HasPrefix(uri, "postgres://") {
		return fmt.Errorf("invalid PostgreSQL URI: must start with postgresql:// or postgres://")
	}
	return nil
}

func (postgresql) Init(ctx context.Context, opts *Options) error {
	var postgresURI string

	// Step 1: Ask local vs custom
	var choice string
	if prompt.IsTTY() {
		c, _ := prompt.CreateSurveySelect(
			"How do you want to connect to PostgreSQL?",
			[]string{"Local (default localhost:5432)", "Custom (connection string)"},
			prompt.AskOpts{
				Default:  "Local (default localhost:5432)",
				PageSize: 2,
			},
		)
		choice = c
	}

	// Step 2: Set URI if local
	if strings.HasPrefix(choice, "Local") {
		postgresURI = fmt.Sprintf("postgresql://localhost:5432/%s", opts.AppName)
		fmt.Println("Using default local PostgreSQL URI:", postgresURI)
	} else {
		// Step 3: Interactive loop for custom URI with "undo" option
		for {
			if prompt.IsTTY() {
				uri, _ := prompt.CreateSurveyInput(
					"Enter your PostgreSQL connection URI (type 'undo' to go back):",
					prompt.AskOpts{
						Help:      "Example: postgresql://username:password@host:5432/database",
						Validator: survey.Required,
					},
				)
				postgresURI = strings.TrimSpace(uri)
			}

			// Allow undo: re-ask local vs custom
			if postgresURI == "undo" {
				c, _ := prompt.CreateSurveySelect(
					"How do you want to connect to PostgreSQL?",
					[]string{"Local (default localhost:5432)", "Custom (connection string)"},
					prompt.AskOpts{
						Default:  "Local (default localhost:5432)",
						PageSize: 2,
					},
				)
				choice = c
				if strings.HasPrefix(choice, "Local") {
					postgresURI = fmt.Sprintf("postgresql://localhost:5432/%s", opts.AppName)
					fmt.Println("Using default local PostgreSQL URI:", postgresURI)
					break
				}
				continue // go back to asking for URI
			}

			if err := EnsurePostgresURI(postgresURI); err != nil {
				fmt.Println("Invalid PostgreSQL URI:", err)
				continue
			}
			break
		}
	}

	opts.DatabaseURI = postgresURI
	fmt.Println("Final PostgreSQL URI set:", opts.DatabaseURI)

	return nil
}

func (postgresql) Generate(ctx context.Context, opts *Options) error {
	// Validate that the backend is compatible (Express only)
	if !fsutil.ValidateDependency("postgresql", opts.Backend) {
		return fmt.Errorf("postgresql can only be used with Express backend, got '%s'", opts.Backend)
	}

	backendDir := filepath.Join(opts.ProjectRoot, "backend")

	// Install Prisma dependencies
	if err := execx.RunCmd(ctx, backendDir, "npm install prisma @prisma/client"); err != nil {
		return fmt.Errorf("npm install prisma: %w", err)
	}

	// Initialize Prisma (creates prisma/schema.prisma)
	if err := execx.RunCmd(ctx, backendDir, "npx prisma init"); err != nil {
		return fmt.Errorf("prisma init: %w", err)
	}

	// Overwrite schema.prisma with our starter schema containing User model
	// Prisma 7 no longer supports url in datasource block
	schemaPath := filepath.Join(backendDir, "prisma", "schema.prisma")
	schemaContent := `// This is your Prisma schema file,
// learn more about it in the docs: https://pris.ly/d/prisma-schema

generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
}

model User {
  id        Int      @id @default(autoincrement())
  email     String   @unique
  name      String?
  createdAt DateTime @default(now())
}
`
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0o644); err != nil {
		return fmt.Errorf("write schema.prisma: %w", err)
	}

	// Create prisma.config.ts for Prisma 7 migrations
	configPath := filepath.Join(backendDir, "prisma.config.ts")
	configContent := `import path from "node:path";
import { defineConfig } from "prisma/config";

export default defineConfig({
  earlyAccess: true,
  schema: path.join(__dirname, "prisma/schema.prisma"),
  migrate: {
    adapter: async () => {
      const { PrismaPg } = await import("@prisma/adapter-pg");
      const { Pool } = await import("pg");
      const pool = new Pool({ connectionString: process.env.DATABASE_URL });
      return new PrismaPg(pool);
    },
  },
});
`
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		return fmt.Errorf("write prisma.config.ts: %w", err)
	}

	// Install pg adapter for Prisma 7
	if err := execx.RunCmd(ctx, backendDir, "npm install @prisma/adapter-pg pg"); err != nil {
		return fmt.Errorf("npm install pg adapter: %w", err)
	}
	if err := execx.RunCmd(ctx, backendDir, "npm install -D @types/pg"); err != nil {
		return fmt.Errorf("npm install @types/pg: %w", err)
	}

	// Generate Prisma client
	if err := execx.RunCmd(ctx, backendDir, "npx prisma generate"); err != nil {
		return fmt.Errorf("prisma generate: %w", err)
	}

	// Copy templates from postgresql/express/
	templateDir := "postgresql/express"
	outputDir := filepath.Join(backendDir, "src")

	if err := fsutil.GenerateFromTemplateDir(templateDir, outputDir); err != nil {
		return fmt.Errorf("generate postgresql templates: %w", err)
	}

	// Inject database import and seed route into index.ts
	indexPath := filepath.Join(backendDir, "src", "index.ts")
	indexBytes, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("read index.ts: %w", err)
	}

	src := string(indexBytes)

	// Inject Prisma client import
	if !strings.Contains(src, "prisma") {
		src = strings.Replace(src, "// [DATABASE IMPORT]", `
import { prisma } from "./db/client";`, 1)
	}

	// Inject seed route
	if !strings.Contains(src, "/seed") {
		route, err := fsutil.RenderTemplate("postgresql/express/seed.tmpl")
		if err != nil {
			return fmt.Errorf("render seed route template: %w", err)
		}
		src = strings.Replace(src, "// [DATABASE ROUTE]", string(route), 1)
	}

	updated := fsutil.FileInfo{
		Path:    indexPath,
		Content: []byte(src),
	}

	if err := fsutil.WriteFile(updated); err != nil {
		return err
	}

	// Push schema to database (requires database to be running)
	fmt.Println("Pushing Prisma schema to database...")
	if err := execx.RunCmd(ctx, backendDir, "npx prisma db push"); err != nil {
		fmt.Println("Warning: prisma db push failed. Make sure PostgreSQL is running and try manually: cd backend && npx prisma db push")
	}

	return nil
}

func (postgresql) Post(ctx context.Context, opts *Options) error {
	// Append DATABASE_URL to backend .env
	envPath := filepath.Join(opts.ProjectRoot, "backend", ".env")
	content := fmt.Sprintf("\nDATABASE_URL=%s", opts.DatabaseURI)
	if err := fsutil.AppendUniqueLines(envPath, []string{content}); err != nil {
		return fmt.Errorf("append DATABASE_URL to .env: %w", err)
	}

	// Add prisma/ and .env to .gitignore
	gitignorePath := filepath.Join(opts.ProjectRoot, ".gitignore")
	if err := fsutil.EnsureFile(gitignorePath); err != nil {
		return fmt.Errorf("ensure gitignore file: %w", err)
	}

	if err := fsutil.AppendUniqueLines(gitignorePath, []string{
		"prisma/.env",
		"backend/prisma/.env",
	}); err != nil {
		return fmt.Errorf("update .gitignore: %w", err)
	}

	return nil
}

// Seed implements the Seeder interface
func (postgresql) Seed(ctx context.Context, opts *Options) error {
	if opts.DatabaseURI == "" {
		return fmt.Errorf("DatabaseURI is empty — did Init() run?")
	}
	
	// For PostgreSQL with Prisma, seeding requires the schema to exist first.
	// The schema is created in Generate(), which runs after Seed() in the current flow.
	// So we just log that seeding will be available after setup completes.
	fmt.Println("PostgreSQL configured with URI:", opts.DatabaseURI)
	fmt.Println("After setup completes, run 'npx prisma db push' in the backend directory to sync your schema.")
	
	return nil
}
