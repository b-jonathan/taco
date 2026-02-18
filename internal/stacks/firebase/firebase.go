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

type firebase struct{}

func New() Stack { return &firebase{} }

func (firebase) Type() string { return "auth" }
func (firebase) Name() string { return "firebase" }

func (firebase) Init(ctx context.Context, opts *Options) error {

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
	if err := execx.RunCmdLive(ctx, "", fmt.Sprintf("firebase projects:create %s --display-name %s", projectID, projectID)); err != nil {
		return fmt.Errorf("failed to create firebase project: %w", err)
	}

	appName := fmt.Sprintf("%s-web", opts.AppName)
	fmt.Printf("Creating Firebase Web App '%s' under project '%s'...\n", appName, projectID)

	if err := execx.RunCmdLive(ctx, "", fmt.Sprintf("firebase apps:create web %s --project %s", appName, projectID)); err != nil {
		return fmt.Errorf("failed to create firebase web app: %w", err)
	}

	// --- New step: Prompt to enable authentication providers ---
	fmt.Println("\nFirebase Authentication Setup:")
	fmt.Println("Recommended providers: Email/Password and Google Sign-In.")

	shouldOpen, err := prompt.CreateSurveyConfirm(
		"Would you like to open Firebase Console to enable these providers now?",
		prompt.AskOpts{Default: true},
	)
	if err != nil {
		return fmt.Errorf("failed to confirm provider setup: %w", err)
	}

	url := fmt.Sprintf("https://console.firebase.google.com/u/0/project/%s/authentication/providers", projectID)
	if shouldOpen {
		fmt.Println("Opening Firebase Authentication Providers page...")
		if err := execx.OpenBrowser(url); err != nil {
			fmt.Println("Could not open browser automatically. Please visit:")
			fmt.Println(url)
		}
	} else {
		fmt.Println("You can enable them later at:")
		fmt.Println(url)
	}

	fmt.Println("\nOnce you've enabled the Email/Password and Google providers in the Firebase Console, continue below.")

	done, err := prompt.CreateSurveyConfirm(
		"Have you finished enabling the recommended providers?",
		prompt.AskOpts{Default: false},
	)
	if err != nil {
		return fmt.Errorf("failed to confirm provider completion: %w", err)
	}

	if !done {
		return fmt.Errorf("firebase provider setup not completed. Please enable the providers and rerun the command")
	}
	return nil
}

func (firebase) Generate(ctx context.Context, opts *Options) error {
	if !fsutil.ValidateDependency("firebase", opts.Frontend) {
		return fmt.Errorf("firebase cannot be used with frontend '%s'", opts.Frontend)
	}

	frontendDir := filepath.Join(opts.ProjectRoot, "frontend")

	if err := execx.RunCmd(ctx, frontendDir, "npm install firebase"); err != nil {
		return fmt.Errorf("npm install firebase: %w", err)
	}

	templateDir := "firebase/nextjs"
	outputDir := filepath.Join(frontendDir)

	if err := fsutil.GenerateFromTemplateDir(templateDir, outputDir); err != nil {
		return fmt.Errorf("generate firebase nextjs templates: %w", err)
	}

	fmt.Println("Firebase Next.js frontend files successfully generated under frontend/src/")
	return nil
}

func (firebase) Post(ctx context.Context, opts *Options) error {
	// Target the .gitignore inside the frontend directory
	gitignorePath := filepath.Join(opts.ProjectRoot, "frontend", ".gitignore")
	if err := fsutil.EnsureFile(gitignorePath); err != nil {
		return fmt.Errorf("ensure frontend gitignore file: %w", err)
	}

	// Append only Firebase-specific ignores
	_ = fsutil.AppendUniqueLines(gitignorePath, []string{
		"# firebase",
		".firebase/",
		".firebasehosting.*",
		"firebase-debug.log",
		"firestore-debug.log",
		"ui-debug.log",
	})

	// Generate and append Firebase credentials to .env.local
	if err := createCredentials(ctx, opts.ProjectRoot, opts.AppName); err != nil {
		return fmt.Errorf("create credentials: %w", err)
	}

	fmt.Println("Firebase post-generation complete. Added Firebase ignores and credentials.")
	return nil
}

func (express) Rollback(ctx context.Context, opts *Options) error {
	return nil
}
