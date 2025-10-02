package cli

import (
	"fmt"

	"github.com/b-jonathan/taco/internal/stacks"
	"github.com/b-jonathan/taco/internal/stacks/express"
	"github.com/b-jonathan/taco/internal/stacks/mongodb"
	"github.com/b-jonathan/taco/internal/stacks/nextjs"
)

type Stack = stacks.Stack

var Registry = map[string]Stack{
	"express": express.New(),
	"nextjs":  nextjs.New(),
	"mongodb": mongodb.New(),
}

func GetFactory(key string) (stacks.Stack, error) {
	if key == "" {
		return nil, nil
	}
	f, ok := Registry[key]
	if !ok {
		return nil, fmt.Errorf("unknown stack %q. available: %v", key, registryNames())
	}
	return f, nil
}

func registryNames() []string {
	names := make([]string, 0, len(Registry))
	for k := range Registry {
		names = append(names, k)
	}
	return names
}
