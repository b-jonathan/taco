package stacks

import "context"

type Stack interface {
	Type() string
	Name() string
	Init(ctx context.Context, opts Options) error
	Generate(ctx context.Context, opts Options) error
	Post(ctx context.Context, opts Options) error
}

type Options struct {
	ProjectRoot    string
	AppName        string
	FrontendOrigin string
	BackendURL     string
	Port           int
}
