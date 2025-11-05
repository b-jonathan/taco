package firebase

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"
	"github.com/b-jonathan/taco/internal/prompt"
	"github.com/b-jonathan/taco/internal/stacks"
)

type Stack = stacks.Stack
type Options = stacks.Options

type express struct{}

func New() Stack { return &express{} }

func (express) Type() string { return "auth" }
func (express) Name() string { return "firebase" }

func (express) Init(ctx context.Context, opts *Options) error {

	if _, err := exec.LookPath("firebase"); err != nil {
		fmt.Println("Firebase CLI not found.")

		// Ask before installing
		shouldInstall, perr := prompt.CreateSurveyConfirm(
			"Firebase CLI is required. Would you like me to install it now using npm -g firebase-tools?",
			prompt.AskOpts{Default: false},
		)
		if perr != nil {
			return fmt.Errorf("failed to get user confirmation: %w", perr)
		}

		if !shouldInstall {
			return fmt.Errorf("firebase-tools required but not installed; please install manually via `npm install -g firebase-tools`")
		}

		fmt.Println("Installing Firebase CLI globally...")
		if err := execx.RunCmdLive(ctx, "", "npm install -g firebase-tools"); err != nil {
			return fmt.Errorf("failed to install firebase-tools: %w", err)
		}
	}

	token := os.Getenv("FIREBASE_TOKEN")
	if token != "" {
		fmt.Println("Detected FIREBASE_TOKEN — skipping interactive login.")
		// verify authentication
		if err := execx.RunCmd(ctx, "", "firebase projects:list --non-interactive"); err != nil {
			return fmt.Errorf("token invalid or expired, please refresh via `firebase login:ci`: %w", err)
		}
	} else {
		// If no token, prompt the user for interactive login
		loggedIn := execx.RunCmd(ctx, "", "firebase projects:list --non-interactive") == nil
		if !loggedIn {
			shouldLogin, err := prompt.CreateSurveyConfirm(
				"No active Firebase session found. Would you like to log in now?",
				prompt.AskOpts{Default: true},
			)
			if err != nil {
				return fmt.Errorf("failed to confirm Firebase login: %w", err)
			}

			if !shouldLogin {
				return fmt.Errorf("firebase login required. You can also use non-interactive mode by setting FIREBASE_TOKEN")
			}

			fmt.Println("Opening Firebase login in browser...")
			if err := execx.RunCmdLive(ctx, "", "firebase login"); err != nil {
				return fmt.Errorf("firebase login failed: %w", err)
			}
		}
	}
	projectID := fmt.Sprintf("%s-taco", opts.AppName)
	fmt.Printf("Creating new Firebase project '%s'...\n", projectID)
	if err := execx.RunCmdLive(ctx, "", fmt.Sprintf("firebase projects:create %s", projectID)); err != nil {
		return fmt.Errorf("failed to create firebase project: %w", err)
	}

	appName := fmt.Sprintf("%s-web", opts.AppName)
	fmt.Printf("Creating Firebase Web App '%s' under project '%s'...\n", appName, projectID)

	if err := execx.RunCmdLive(ctx, "", fmt.Sprintf("firebase apps:create web %s --project %s", appName, projectID)); err != nil {
		return fmt.Errorf("failed to create firebase web app: %w", err)
	}
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
