package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"
)

// TODO: There is absolutely no reason for init and push to be in one function, gonna have to refactor this for sure

func Init(ctx context.Context, projectRoot string) error {

	// If already a repo, skip init
	if _, err := fsutil.Fs.Stat(filepath.Join(projectRoot, ".git")); os.IsNotExist(err) {
		if err := execx.RunCmd(ctx, projectRoot, "git init"); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
	}

	// Set default branch to main
	if err := execx.RunCmd(ctx, projectRoot, "git checkout -B main"); err != nil {
		return fmt.Errorf("git checkout -B main: %w", err)
	}

	return nil
}

func Commit(ctx context.Context, projectRoot, commitMsg string) error {
	// Stage and commit
	if err := execx.RunCmd(ctx, projectRoot, "git add ."); err != nil {
		return fmt.Errorf("git add .: %w", err)
	}

	if err := execx.RunCmd(ctx, projectRoot, "git commit -m "+commitMsg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil

}

func Push(ctx context.Context, projectRoot, remoteURL, branch string) error {

	// Configure remote. If it already exists, update it.
	_ = execx.RunCmd(ctx, projectRoot, "git remote remove origin")
	if err := execx.RunCmd(ctx, projectRoot, "git remote add origin "+remoteURL); err != nil {
		return fmt.Errorf("git remote add: %w", err)
	}

	// Push upstream
	if err := execx.RunCmd(ctx, projectRoot, "git push -u origin main"); err != nil {
		return fmt.Errorf("git push: %w", err)
	}

	return nil
}

// Calls all 3 of the helper functions. If we don't want this then we will need to
// change the function call in cli/root.go
func InitAndPush(ctx context.Context, projectRoot, remoteURL, commitMsg string) error {
	if err := Init(ctx, projectRoot); err != nil {
		return err
	}

	if err := Commit(ctx, projectRoot, commitMsg); err != nil {
		return err
	}

	if err := Push(ctx, projectRoot, remoteURL, "main"); err != nil {
		return err
	}

	return nil
}
