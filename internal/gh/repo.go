package gh

import (
	"context"
	"fmt"
	"strings"
	"time"

	github "github.com/google/go-github/v55/github"
)

type CreateRepoOptions struct {
	Name        string
	Private     bool
	Description string
	Timeout     time.Duration
}

// Create Repo
func CreateRepo(ctx context.Context, opts CreateRepoOptions) (*github.Repository, error) {
	client, err := EnsureClient(ctx)
	if err != nil {
		return nil, err
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = 100 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	newRepo := &github.Repository{
		Name:        github.String(opts.Name),
		Private:     github.Bool(opts.Private),
		Description: github.String(opts.Description),
	}

	repo, _, err := client.Repositories.Create(ctx, "", newRepo)
	if err != nil {
		return nil, fmt.Errorf("create repo: %w", err)
	}

	return repo, nil
}

// For cleanup
func DeleteRepo(ctx context.Context, repo *github.Repository) error {
	client, err := EnsureClient(ctx)
	if err != nil {
		return err
	}

	if repo == nil {
		return nil
	}

	owner := ""
	if repo.GetOwner() != nil {
		owner = repo.GetOwner().GetLogin()
	}

	if owner == "" {
		parts := strings.Split(repo.GetFullName(), "/")
		if len(parts) == 2 {
			owner = parts[0]
		}
	}

	if owner == "" {
		return fmt.Errorf("cannot determine repo owner")
	}

	_, err = client.Repositories.Delete(ctx, owner, repo.GetName())
	if err != nil {
		return fmt.Errorf("delete repo: %w", err)
	}

	return nil
}
