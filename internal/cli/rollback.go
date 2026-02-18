package cli

import (
	"context"
	"log"

	"github.com/b-jonathan/taco/internal/stacks"
)

func rollbackStacks(ctx context.Context, opts *stacks.Options, ss ...stacks.Stack) {
	for _, s := range ss {
		if s == nil {
			continue
		}
		log.Printf("Rolling back %s stack...", s.Name())

		if err := s.Rollback(ctx, opts); err != nil {
			log.Printf("rollback %s failed: %v", s.Name(), err)
		}
	}
}
