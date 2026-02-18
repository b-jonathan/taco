package gh

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/logx"
	"github.com/b-jonathan/taco/internal/prompt"
	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

// TODO: This whole file is p vibe coded, i think it works p well tho
type ctxKey struct{}

var ghClientKey = ctxKey{}

func NewClient(ctx context.Context, token string) *github.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return github.NewClient(oauth2.NewClient(ctx, tokenSource))
}

// WithContext stores the client in a derived context.
func WithContext(ctx context.Context, c *github.Client) context.Context {
	return context.WithValue(ctx, ghClientKey, c)
}

// FromContext returns the client or an error if missing.
func FromContext(ctx context.Context) (*github.Client, error) {
	v := ctx.Value(ghClientKey)
	if c, ok := v.(*github.Client); ok && c != nil {
		return c, nil
	}
	return nil, errors.New("github client missing in context")
}

// HasClient tells whether a client is present.
func HasClient(ctx context.Context) bool { return ctx.Value(ghClientKey) != nil }

func EnsureClient(ctx context.Context) (*github.Client, error) {
	if client, err := FromContext(ctx); err == nil {
		logx.Infof("Using existing GitHub client from context")
		return client, nil
	}

	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Println("GitHub CLI not found on your system.")

		// Detect OS
		osName := runtime.GOOS

		var installCmd string
		var installDescription string

		switch osName {
		case "windows":
			installCmd = "winget install GitHub.cli"
			installDescription = "winget install GitHub.cli"
		case "darwin":
			installCmd = "brew install gh"
			installDescription = "brew install gh"
		case "linux":
			installCmd = "sudo apt install gh"
			installDescription = "sudo apt install gh"
		default:
			return nil, fmt.Errorf("unsupported OS. Please install GitHub CLI manually")
		}

		// Ask if user wants auto installation
		shouldInstall, err := prompt.CreateSurveyConfirm(
			fmt.Sprintf("GitHub CLI is required. Would you like Taco to install it using %s?", installDescription),
			prompt.AskOpts{Default: true},
		)

		if err != nil {
			return nil, fmt.Errorf("failed to confirm installation: %w", err)
		}

		if shouldInstall {
			fmt.Println("Installing GitHub CLI...")

			if err := execx.RunCmdLive(ctx, "", installCmd); err != nil {
				return nil, fmt.Errorf("installation failed. Please install GitHub CLI manually using %s", installDescription)
			}

			fmt.Println("GitHub CLI installed successfully")
		} else {
			return nil, fmt.Errorf("GitHub CLI is required. Install using %s", installDescription)
		}
	}
	if err := execx.RunCmd(ctx, "", "gh auth status"); err != nil {
		fmt.Println("You are not authenticated with GitHub CLI.")
		shouldLogin, _ := prompt.CreateSurveyConfirm(
			"Would you like to authenticate now?",
			prompt.AskOpts{Default: true},
		)
		if shouldLogin {
			fmt.Println("Starting GitHub authentication...")

			if err := execx.RunCmdLive(ctx, "", "gh auth login"); err != nil {
				return nil, fmt.Errorf("failed to authenticate with GitHub CLI: %w", err)
			}
		} else {
			return nil, fmt.Errorf("GitHub authentication is required to proceed")
		}
	}

	tokenBytes, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve GitHub token: %w", err)
	}
	token := strings.TrimSpace(string(tokenBytes))

	client := NewClient(ctx, token)
	client.UserAgent = "taco-cli"
	return client, nil
}
