package cli

import (
	"context"
	"fmt"

	"github.com/b-jonathan/taco/internal/stacks"
)

func rollbackStacks(ctx context.Context, opts *stacks.Options, ss ...stacks.Stack) {
	for _, s := range ss {
		if s == nil {
			continue
		}
		fmt.Printf("Rolling back %s stack...\n", s.Name())

		if err := s.Rollback(ctx, opts); err != nil {
			fmt.Printf("rollback %s failed: %v\n", s.Name(), err)
		}
	}
}
