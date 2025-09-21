package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/b-jonathan/taco/internal/gh"
	"github.com/google/go-github/v55/github"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type InitPackageParams struct {
	Name    string
	Main    string
	Scripts map[string]string // merged into existing
}

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
		if !isTTY() {
			return params, fmt.Errorf("name required in non-interactive mode")
		}
		name, err := createSurveyInput("Repository Name:", AskOpts{Help: "lowercase letters, numbers, dash, and underscore only", Validator: survey.Required})
		if err != nil {
			return params, err
		}
		params.Name = name
	}

	if f := cmd.Flags().Lookup("private"); f != nil && f.Changed {
		b, _ := strconv.ParseBool(f.Value.String())
		params.Private = b
	} else {
		b, err := createSurveyConfirm("Make repository private?", AskOpts{
			Default: false,
		})
		if err != nil && isTTY() {
			return params, err
		}
		if err == nil {
			params.Private = b
		}
	}

	if v, _ := cmd.Flags().GetString("remote"); v != "" {
		params.Remote = v
	} else {
		if isTTY() {
			r, err := createSurveySelect("Remote URL type", []string{"ssh", "https"}, AskOpts{
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
		if isTTY() {
			desc, err := createSurveyInput("Repository description", AskOpts{
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

			params, err := gatherInitParams(cmd, args)
			if err != nil {
				return err
			}

			stack, _ := createSurveySelect("Choose a Stack:\n", []string{"express", "TODO"}, AskOpts{})

			projectRoot := filepath.Join("..", params.Name)
			if stack == "express" {
				runInitExpress(cmd, projectRoot)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Name=%s Private=%t Remote=%s Desc = %q\n", params.Name, params.Private, params.Remote, params.Description)

			log.Println("Starting gh command")
			gh := gh.MustFromContext(cmd.Context())
			log.Println("GitHub client initialized")
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			newRepo := &github.Repository{
				Name:        github.String(params.Name),
				Private:     github.Bool(params.Private),
				Description: github.String(params.Description),
			}

			repo, _, err := gh.Repositories.Create(ctx, "", newRepo)
			if err != nil {
				return fmt.Errorf("create repo: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Created:", repo.GetHTMLURL())
			remoteURL := repo.GetSSHURL()
			if params.Remote == "https" {
				remoteURL = repo.GetCloneURL()
			}
			if err := gitInitAndPush(ctx, projectRoot, remoteURL, "chore: initial commit"); err != nil {
				return err
			}
			// Continue with local scaffold, git init, push, etc., using p.Remote to choose SSH or HTTPS
			return nil
		},
	}
	// Flags that feed into gatherInitParams
	cmd.Flags().Bool("private", false, "Make the repository private")
	cmd.Flags().String("remote", "ssh", "Remote URL type ssh or https")
	cmd.Flags().String("description", "", "Repository description")
	return cmd
}

func runInitExpress(cmd *cobra.Command, projectRoot string) error {
	ctx := cmd.Context()
	// <projectRoot>/backend/src
	backendDir := filepath.Join(projectRoot, "backend")
	srcDir := filepath.Join(backendDir, "src")

	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := runCmd(ctx, backendDir, "npm", "init", "-y"); err != nil {
		return fmt.Errorf("npm init: %w", err)
	}

	if err := runCmd(ctx, backendDir, "npm", "install", "express", "cors", "dotenv"); err != nil {
		return fmt.Errorf("npm install express: %w", err)
	}

	if err := runCmd(ctx, backendDir, "npm", "install", "-D",
		"typescript", "tsx", "@types/node", "@types/express", "@types/cors"); err != nil {
		return fmt.Errorf("npm install dev deps: %w", err)
	}

	tsconfigPath := filepath.Join(backendDir, "tsconfig.json")
	if err := ensureFile(tsconfigPath); err != nil {
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
	if err := ensureFile(indexPath); err != nil {
		return fmt.Errorf("ensure index file: %w", err)
	}
	// src/index.ts
	index := `import express from "express";

		const app = express();
		const PORT = process.env.PORT || 3000;

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
	if err := ensureFile(gitignorePath); err != nil {
		return fmt.Errorf("ensure gitignore file: %w", err)
	}

	_ = appendUniqueLines(gitignorePath,
		[]string{"backend/node_modules/", "backend/dist/", "backend/.env"})

	packageParams := InitPackageParams{
		Name: "express",
		Main: "src/index.ts",
		Scripts: map[string]string{
			"dev":   "tsx watch src/index.ts",
			"build": "tsc -p tsconfig.json",
			"start": "node dist/index.js",
		}}
	initPackage(backendDir, packageParams)

	return nil
}

func runCmd(ctx context.Context, dir, name string, args ...string) error {
	c := exec.CommandContext(ctx, name, args...)
	c.Dir = dir
	var out, errb bytes.Buffer
	c.Stdout, c.Stderr = &out, &errb
	if err := c.Run(); err != nil {
		return fmt.Errorf("%s %v failed: %v\nstdout:\n%s\nstderr:\n%s",
			name, args, err, out.String(), errb.String())
	}
	return nil
}

func gitInitAndPush(ctx context.Context, projectRoot, remoteURL, commitMsg string) error {
	// If already a repo, skip init
	if _, err := os.Stat(filepath.Join(projectRoot, ".git")); os.IsNotExist(err) {
		if err := runCmd(ctx, projectRoot, "git", "init"); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
	}

	// Set default branch to main
	if err := runCmd(ctx, projectRoot, "git", "checkout", "-B", "main"); err != nil {
		return fmt.Errorf("git checkout -B main: %w", err)
	}

	// Stage and commit
	if err := runCmd(ctx, projectRoot, "git", "add", "."); err != nil {
		return fmt.Errorf("git add .: %w", err)
	}
	if err := runCmd(ctx, projectRoot, "git", "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	// Configure remote. If it already exists, update it.
	_ = runCmd(ctx, projectRoot, "git", "remote", "remove", "origin")
	if err := runCmd(ctx, projectRoot, "git", "remote", "add", "origin", remoteURL); err != nil {
		return fmt.Errorf("git remote add: %w", err)
	}

	// Push upstream
	if err := runCmd(ctx, projectRoot, "git", "push", "-u", "origin", "main"); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}

/* Helpers */
func isTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func ensureFile(path string) error {
	// Create parent directories if needed.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	// Create the file if missing. O_EXCL prevents clobbering if a race happens.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		// If it already exists, that’s fine.
		if os.IsExist(err) {
			return nil
		}
		return err
	}
	return f.Close()
}

func initPackage(dir string, params InitPackageParams) error {
	path := filepath.Join(dir, "package.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var pkg map[string]any
	if err := json.Unmarshal(b, &pkg); err != nil {
		return err
	}
	scripts, _ := pkg["scripts"].(map[string]any)
	if scripts == nil {
		scripts = map[string]any{}
	}
	for k, v := range params.Scripts {
		scripts[k] = v
	}
	pkg["scripts"] = scripts

	pkg["name"] = params.Name
	pkg["main"] = params.Main

	out, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

func appendUniqueLines(path string, lines []string) error {
	buf, _ := os.ReadFile(path)
	for _, line := range lines {
		if !bytes.Contains(buf, []byte(line+"\n")) && !bytes.Equal(bytes.TrimSpace(buf), []byte(line)) {
			if len(buf) > 0 && buf[len(buf)-1] != '\n' {
				buf = append(buf, '\n')
			}
			buf = append(buf, []byte(line+"\n")...)
		}
	}
	return os.WriteFile(path, buf, 0o644)
}
