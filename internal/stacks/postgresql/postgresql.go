package postgresql

import (
	"context"
	"fmt"

	"github.com/b-jonathan/taco/internal/stacks"
)

type Stack = stacks.Stack
type Options = stacks.Options

type postgresql struct{}

func New() Stack { return &postgresql{} }

func (postgresql) Type() string { return "database" }
func (postgresql) Name() string { return "postgres" }

func (postgresql) Init(ctx context.Context, opts *Options) error {
	// TODO: Implement connection setup prompt (local vs custom connection string)
	fmt.Println("PostgreSQL Init - not yet implemented")
	return nil
}

func (postgresql) Generate(ctx context.Context, opts *Options) error {
	// TODO: Implement Prisma init, schema generation, and client generation
	fmt.Println("PostgreSQL Generate - not yet implemented")
	return nil
}

func (postgresql) Post(ctx context.Context, opts *Options) error {
	// TODO: Append DATABASE_URL to .env, update .gitignore
	fmt.Println("PostgreSQL Post - not yet implemented")
	return nil
}

// Seed implements the Seeder interface
func (postgresql) Seed(ctx context.Context, opts *Options) error {
	// TODO: Implement seed route similar to MongoDB's
	fmt.Println("PostgreSQL Seed - not yet implemented")
	return nil
}
