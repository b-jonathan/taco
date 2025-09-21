package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/b-jonathan/taco/internal/execx"
)

func InitAndPush(ctx context.Context, projectRoot, remoteURL, commitMsg string) error {
	// If already a repo, skip init
	if _, err := os.Stat(filepath.Join(projectRoot, ".git")); os.IsNotExist(err) {
		if err := execx.RunCmd(ctx, projectRoot, "git", "init"); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
	}

	// Set default branch to main
	if err := execx.RunCmd(ctx, projectRoot, "git", "checkout", "-B", "main"); err != nil {
		return fmt.Errorf("git checkout -B main: %w", err)
	}

	// Stage and commit
	if err := execx.RunCmd(ctx, projectRoot, "git", "add", "."); err != nil {
		return fmt.Errorf("git add .: %w", err)
	}
	if err := execx.RunCmd(ctx, projectRoot, "git", "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	// Configure remote. If it already exists, update it.
	_ = execx.RunCmd(ctx, projectRoot, "git", "remote", "remove", "origin")
	if err := execx.RunCmd(ctx, projectRoot, "git", "remote", "add", "origin", remoteURL); err != nil {
		return fmt.Errorf("git remote add: %w", err)
	}

	// Push upstream
	if err := execx.RunCmd(ctx, projectRoot, "git", "push", "-u", "origin", "main"); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}
