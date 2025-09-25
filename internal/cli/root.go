package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"
	"github.com/b-jonathan/taco/internal/gh"
	"github.com/b-jonathan/taco/internal/nodepkg"
	"github.com/b-jonathan/taco/internal/prompt"
	"github.com/b-jonathan/taco/internal/stacks"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func Execute() error {
	_ = godotenv.Load()
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "taco",
		Short:         "Project Initializer",
		Long:          `taco is a CLI tool for initializing new projects that's language-agnostic.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if gh.HasClient(ctx) {
			return nil
		}
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return fmt.Errorf("set GITHUB_TOKEN")
		}
		client := gh.NewClient(ctx, token)
		client.UserAgent = "taco-cli"
		cmd.SetContext(gh.WithContext(ctx, client))
		return nil
	}
	cmd.AddCommand(initCmd())
	return cmd
}

func gatherInitParams(cmd *cobra.Command, args []string) (InitParams, error) {
	var params InitParams

	if len(args) > 0 && args[0] != "" {
		params.Name = args[0]
	} else {
		if !prompt.IsTTY() {
			return params, fmt.Errorf("name required in non-interactive mode")
		}
		name, err := prompt.CreateSurveyInput("Repository Name:", prompt.AskOpts{Help: "lowercase letters, numbers, dash, and underscore only", Validator: survey.Required})
		if err != nil {
			return params, err
		}
		params.Name = name
	}

	if f := cmd.Flags().Lookup("private"); f != nil && f.Changed {
		b, _ := strconv.ParseBool(f.Value.String())
		params.Private = b
	} else {
		b, err := prompt.CreateSurveyConfirm("Make repository private?", prompt.AskOpts{
			Default: false,
		})
		if err != nil && prompt.IsTTY() {
			return params, err
		}
		if err == nil {
			params.Private = b
		}
	}

	if v, _ := cmd.Flags().GetString("remote"); v != "" {
		params.Remote = v
	} else {
		if prompt.IsTTY() {
			r, err := prompt.CreateSurveySelect("Remote URL type", []string{"ssh", "https"}, prompt.AskOpts{
				Default:  "ssh",
				PageSize: 2,
			})
			if err != nil {
				return params, err
			}
			params.Remote = r
		}
	}

	if v, _ := cmd.Flags().GetString("description"); v != "" {
		params.Description = v
	} else {
		// optional field; allow empty in non-TTY
		if prompt.IsTTY() {
			desc, err := prompt.CreateSurveyInput("Repository description", prompt.AskOpts{
				Default: "",
				Help:    "you can leave this empty",
			})
			if err != nil {
				return params, err
			}
			params.Description = desc
		}
	}

	return params, nil
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Create repo and scaffold",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx := cmd.Context()
			stack := map[string]string{
				"frontend": "",
				"backend":  "",
			}
			params, err := gatherInitParams(cmd, args)
			if err != nil {
				return err
			}

			projectRoot := filepath.Join("..", params.Name)
			if err := os.MkdirAll(projectRoot, 0o755); err != nil {
				return fmt.Errorf("mkdir project root: %w", err)
			}

			stack["frontend"], _ = prompt.CreateSurveySelect("Choose a Frontend Stack:\n", []string{"NextJS", "None"}, prompt.AskOpts{})
			stack["frontend"] = strings.ToLower(stack["frontend"])
			stack["backend"], _ = prompt.CreateSurveySelect("Choose a Backend Stack:\n", []string{"Express", "None"}, prompt.AskOpts{})
			stack["backend"] = strings.ToLower(stack["backend"])

			frontend, err := GetFactory(stack["frontend"])
			if err != nil {
				return err
			}
			backend, err := GetFactory(stack["backend"])
			if err != nil {
				return err
			}

			opts := stacks.Options{
				ProjectRoot:    projectRoot,
				AppName:        params.Name,
				FrontendOrigin: "http://localhost:3000",
				BackendURL:     "http://localhost:4000",
				Port:           4000,
			}

			// Init phase
			if frontend != nil {
				if err := frontend.Init(ctx, opts); err != nil {
					return err
				}
			}
			if backend != nil {
				if err := backend.Init(ctx, opts); err != nil {
					return err
				}
			}

			// Generate phase
			if frontend != nil {
				if err := frontend.Generate(ctx, opts); err != nil {
					return err
				}
			}
			if backend != nil {
				if err := backend.Generate(ctx, opts); err != nil {
					return err
				}
			}

			// Post phase

			fmt.Fprintf(cmd.OutOrStdout(), "Name=%s Private=%t Remote=%s Desc = %q\n", params.Name, params.Private, params.Remote, params.Description)

			if frontend != nil && backend != nil {
				if err := frontend.Post(ctx, opts); err != nil {
					return err
				}
				if err := backend.Post(ctx, opts); err != nil {
					return err
				}
			}

			// log.Println("Starting gh command")
			// client := gh.MustFromContext(cmd.Context())
			// log.Println("GitHub client initialized")
			// ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			// defer cancel()

			// newRepo := &github.Repository{
			// 	Name:        github.String(params.Name),
			// 	Private:     github.Bool(params.Private),
			// 	Description: github.String(params.Description),
			// }

			// repo, _, err := client.Repositories.Create(ctx, "", newRepo)
			// if err != nil {
			// 	return fmt.Errorf("create repo: %w", err)
			// }

			// log.Println(cmd.OutOrStdout(), "Created:", repo.GetHTMLURL())
			// remoteURL := repo.GetSSHURL()
			// if params.Remote == "https" {
			// 	remoteURL = repo.GetCloneURL()
			// }
			// log.Println("Committing and Pushing to Github...")
			// if err := git.InitAndPush(ctx, projectRoot, remoteURL, "chore: initial commit"); err != nil {
			// 	_, err := client.Repositories.Delete(ctx, "", *newRepo.Name)
			// 	return err
			// }
			// log.Println("Pushed:", repo.GetHTMLURL())

			return nil
		},
	}
	// Flags that feed into gatherInitParams
	// cmd.Flags().Bool("private", false, "Make the repository private")
	// cmd.Flags().String("remote", "ssh", "Remote URL type ssh or https")
	// cmd.Flags().String("description", "", "Repository description")
	return cmd
}

func initEnvs(projectRoot string, stack map[string]string) error {
	// frontend
	if stack["frontend"] == "NextJS" {
		path := filepath.Join(projectRoot, "frontend", ".env.local")
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
		content := `NEXT_PUBLIC_BACKEND_URL=http://localhost:4000	
		`
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	// backend
	if stack["backend"] == "Express" {
		path := filepath.Join(projectRoot, "backend", ".env")
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
	}

	return nil
}

func runInitNextJS(cmd *cobra.Command, projectRoot string) error {
	ctx := cmd.Context()
	frontendDir := filepath.Join(projectRoot, "frontend")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
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

	if err := execx.RunCmd(ctx, projectRoot, "npx", nextFlags...); err != nil {
		return fmt.Errorf("create-next-app: %w", err)
	}

	// Repo root gitignore entries for the frontend
	gitignorePath := filepath.Join(projectRoot, ".gitignore")
	if err := fsutil.EnsureFile(gitignorePath); err != nil {
		return fmt.Errorf("ensure .gitignore: %w", err)
	}
	if err := fsutil.AppendUniqueLines(gitignorePath, []string{
		"frontend/node_modules/",
		"frontend/.next/",
		"frontend/.env.local",
	}); err != nil {
		return err
	}

	// Create an env placeholder
	if err := fsutil.EnsureFile(filepath.Join(frontendDir, ".env.local")); err != nil {
		return fmt.Errorf("ensure .env.local: %w", err)
	}

	pagePath := filepath.Join(frontendDir, "src", "app", "page.tsx")
	pageContent := `
		"use client";
		import { useEffect, useState } from "react";
		export default function Home() {
		const [message, setMessage] = useState<string>("loading...");
		useEffect(() => {
			fetch("http://localhost:4000/")
			.then((res) => res.text())
			.then(setMessage)
			.catch((err) => setMessage("error: " + err.message));
		}, []);
		return <div>{message}</div>;
		}
		`
	if err := os.WriteFile(pagePath, []byte(pageContent), 0o644); err != nil {
		return fmt.Errorf("write page.tsx: %w", err)
	}
	return nil
}

func runInitExpress(cmd *cobra.Command, projectRoot string) error {
	ctx := cmd.Context()
	// <projectRoot>/backend/src
	backendDir := filepath.Join(projectRoot, "backend")
	srcDir := filepath.Join(backendDir, "src")

	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

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
	gitignorePath := filepath.Join(projectRoot, ".gitignore")
	if err := fsutil.EnsureFile(gitignorePath); err != nil {
		return fmt.Errorf("ensure gitignore file: %w", err)
	}

	_ = fsutil.AppendUniqueLines(gitignorePath,
		[]string{"backend/node_modules/", "backend/dist/", "backend/.env*"})

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
