package gh

import (
	"context"
	"errors"

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

// MustFromContext is optional if you like the panic behavior.
func MustFromContext(ctx context.Context) *github.Client {
	c, err := FromContext(ctx)
	if err != nil {
		panic(err)
	}
	return c
}
