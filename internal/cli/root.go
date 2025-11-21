package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/b-jonathan/taco/internal/gh"
	"github.com/b-jonathan/taco/internal/git"
	"github.com/b-jonathan/taco/internal/prompt"
	"github.com/b-jonathan/taco/internal/stacks"
	github "github.com/google/go-github/v55/github"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func Execute() error {
	_ = godotenv.Load()
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "taco",
		Short:         "Project Initializer",
		Long:          `taco is a CLI tool for initializing new projects that's language-agnostic.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		if gh.HasClient(ctx) {
			return nil
		}
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return fmt.Errorf("set GITHUB_TOKEN")
		}
		client := gh.NewClient(ctx, token)
		client.UserAgent = "taco-cli"
		cmd.SetContext(gh.WithContext(ctx, client))
		return nil
	}
	cmd.AddCommand(initCmd())
	return cmd
}

func gatherInitParams(cmd *cobra.Command, args []string) (InitParams, error) {
	var params InitParams

	if len(args) > 0 && args[0] != "" {
		params.Name = args[0]
	} else {
		if !prompt.IsTTY() {
			return params, fmt.Errorf("name required in non-interactive mode")
		}
		name, err := prompt.CreateSurveyInput("Repository Name:", prompt.AskOpts{Help: "lowercase letters, numbers, dash, and underscore only", Validator: survey.Required})
		if err != nil {
			return params, err
		}
		params.Name = name
	}

	if f := cmd.Flags().Lookup("private"); f != nil && f.Changed {
		b, _ := strconv.ParseBool(f.Value.String())
		params.Private = b
	} else {
		b, err := prompt.CreateSurveyConfirm("Make repository private?", prompt.AskOpts{
			Default: false,
		})
		if err != nil && prompt.IsTTY() {
			return params, err
		}
		if err == nil {
			params.Private = b
		}
	}

	if v, _ := cmd.Flags().GetString("remote"); v != "" {
		params.Remote = v
	} else {
		if prompt.IsTTY() {
			r, err := prompt.CreateSurveySelect("Remote URL type", []string{"ssh", "https"}, prompt.AskOpts{
				Default:  "ssh",
				PageSize: 2,
			})
			if err != nil {
				return params, err
			}
			params.Remote = r
		}
	}

	if v, _ := cmd.Flags().GetString("description"); v != "" {
		params.Description = v
	} else {
		// optional field; allow empty in non-TTY
		if prompt.IsTTY() {
			desc, err := prompt.CreateSurveyInput("Repository description", prompt.AskOpts{
				Default: "",
				Help:    "you can leave this empty",
			})
			if err != nil {
				return params, err
			}
			params.Description = desc
		}
	}

	if f := cmd.Flags().Lookup("github"); f != nil && f.Changed {
		b, _ := strconv.ParseBool(f.Value.String())
		params.UseGitHub = b
	} else {
		if prompt.IsTTY() {
			useGH, err := prompt.CreateSurveyConfirm(
				"Create GitHub repository and push initial commit?",
				prompt.AskOpts{
					Default: true,
					Help:    "If yes, taco will create a repo on your GitHub account and push the scaffolded code.",
				},
			)
			if err != nil {
				return params, err
			}
			params.UseGitHub = useGH
		}
	}

	return params, nil
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Create repo and scaffold",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			rootCtx := cmd.Context()
			stack := map[string]string{}
			params, err := gatherInitParams(cmd, args)
			if err != nil {
				return err
			}

			projectRoot := filepath.Join("..", params.Name)
			if err := os.MkdirAll(projectRoot, 0o755); err != nil {
				return fmt.Errorf("mkdir project root: %w", err)
			}

			//TODO: We're gonna have to refactor this into a "dependency-style" selection, so only db's supported by chosen backend are seen

			stack["frontend"], _ = prompt.CreateSurveySelect("Choose a Frontend Stack:\n", []string{"NextJS", "None"}, prompt.AskOpts{})
			stack["frontend"] = strings.ToLower(stack["frontend"])
			frontend, err := GetFactory(stack["frontend"])
			if err != nil {
				return err
			}

			stack["backend"], _ = prompt.CreateSurveySelect("Choose a Backend Stack:\n", []string{"Express", "None"}, prompt.AskOpts{})
			stack["backend"] = strings.ToLower(stack["backend"])
			backend, err := GetFactory(stack["backend"])
			if err != nil {
				return err
			}
			stack["database"], _ = prompt.CreateSurveySelect("Choose a Database Stack:\n", []string{"MongoDB", "None"}, prompt.AskOpts{})
			stack["database"] = strings.ToLower(stack["database"])
			database, err := GetFactory(stack["database"])
			if err != nil {
				return err
			}

			stack["auth"], _ = prompt.CreateSurveySelect("Choose an Auth Stack:\n", []string{"Firebase", "None"}, prompt.AskOpts{})
			stack["auth"] = strings.ToLower(stack["auth"])
			auth, err := GetFactory(stack["auth"])
			if err != nil {
				return err
			}

			opts := &stacks.Options{
				ProjectRoot: projectRoot,
				AppName:     params.Name,
				Frontend:    "http://localhost:3000",
				BackendURL:  "http://localhost:4000",
				Port:        4000,
			}

			// This is core core

			g, ctx := errgroup.WithContext(rootCtx)

			g.Go(func() error { return runSelected(ctx, "Frontend", frontend, opts, []string{"init", "generate"}) })
			g.Go(func() error { return runSelected(ctx, "Backend", backend, opts, []string{"init", "generate"}) })
			g.Go(func() error { return runSelected(ctx, "Database", database, opts, []string{"init", "seed"}) })
			g.Go(func() error { return runSelected(ctx, "Auth", auth, opts, []string{"init"}) })

			if err := g.Wait(); err != nil {
				return err
			}

			if err := runSelected(rootCtx, "Auth", auth, opts, []string{"generate"}); err != nil {
				return err
			}

			if err := runSelected(rootCtx, "Database", database, opts, []string{"generate"}); err != nil {
				return err
			}

			g, ctx = errgroup.WithContext(cmd.Context())
			g.Go(func() error {
				if err := runSelected(ctx, "Frontend", frontend, opts, []string{"post"}); err != nil {
					return err
				}
				if err := runSelected(ctx, "Auth", auth, opts, []string{"post"}); err != nil {
					return err
				}
				return nil
			})
			g.Go(func() error {
				if err := runSelected(ctx, "Backend", backend, opts, []string{"post"}); err != nil {
					return err
				}
				if err := runSelected(ctx, "Database", database, opts, []string{"post"}); err != nil {
					return err
				}
				return nil
			})

			if err := g.Wait(); err != nil {
				return err
			}

			// This is additional templates
			if params.UseGitHub {
				log.Println("Starting gh command")
				client := gh.MustFromContext(cmd.Context())
				log.Println("GitHub client initialized")
				ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
				defer cancel()

				newRepo := &github.Repository{
					Name:        github.String(params.Name),
					Private:     github.Bool(params.Private),
					Description: github.String(params.Description),
				}

				repo, _, err := client.Repositories.Create(ctx, "", newRepo)
				if err != nil {
					return fmt.Errorf("create repo: %w", err)
				}

				log.Println(cmd.OutOrStdout(), "Created:", repo.GetHTMLURL())
				remoteURL := repo.GetSSHURL()
				if params.Remote == "https" {
					remoteURL = repo.GetCloneURL()
				}
				log.Println("Committing and Pushing to Github...")
				if err := git.InitAndPush(ctx, projectRoot, remoteURL, "initial-commit"); err != nil {
					owner := ""
					if repo.GetOwner() != nil {
						owner = repo.GetOwner().GetLogin()
					}

					// Fallback
					if owner == "" {
						parts := strings.Split(repo.GetFullName(), "/")
						if len(parts) == 2 {
							owner = parts[0]
						}
					}

					if owner != "" {
						if _, delErr := client.Repositories.Delete(ctx, owner, repo.GetName()); delErr != nil {
							log.Printf("warning: failed to delete repo after push failure: %v", delErr)
						}
					} else {
						log.Printf("warning: could not determine owner for cleanup of repo %q", repo.GetFullName())
					}

					return fmt.Errorf("git init/push failed: %w", err)
				}
				log.Println("Pushed:", repo.GetHTMLURL())
			} else {
				log.Println("Skipping GitHub repo creation")
			}

			log.Println("Time Taken:", time.Since(start))
			return nil
		},
	}
	// Flags that feed into gatherInitParams
	cmd.Flags().Bool("private", false, "Make the repository private")
	cmd.Flags().String("remote", "", "Remote URL type ssh or https")
	cmd.Flags().String("description", "", "Repository description")
	cmd.Flags().Bool("github", false, "Create and push to a GitHub repository")
	return cmd
}

// TODO: We'll prob have to add this to like a middleware/logging package helper lol
func timedStep(name string, fn func() error) error {
	start := time.Now()
	err := fn()
	dur := time.Since(start)
	prompt.TermLock.Lock()
	defer prompt.TermLock.Unlock()
	if err != nil {
		log.Printf("%s failed in %s: %v", name, dur, err)
		return err
	}
	log.Printf("%s finished in %s", name, dur)
	return nil
}

func stackSteps(
	ctx context.Context,
	label string,
	s stacks.Stack,
	opts *stacks.Options,
	funcs []string,
) ([]Step, error) {
	if s == nil {
		return nil, nil
	}

	steps := make([]Step, 0, len(funcs))
	// optional capability
	seeder, hasSeed := any(s).(stacks.Seeder)

	for _, name := range funcs {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "init":
			steps = append(steps, Step{
				Name: label + " Init",
				Fn:   func() error { return s.Init(ctx, opts) },
			})
		case "generate":
			steps = append(steps, Step{
				Name: label + " Generate",
				Fn:   func() error { return s.Generate(ctx, opts) },
			})
		case "post":
			steps = append(steps, Step{
				Name: label + " Post",
				Fn:   func() error { return s.Post(ctx, opts) },
			})
		case "seed":
			if !hasSeed {
				return nil, fmt.Errorf("%s does not support seed", label)
			}
			// seeding is optional but explicit
			steps = append(steps, Step{
				Name: label + " Seed",
				Fn:   func() error { return seeder.Seed(ctx, opts) },
			})
		default:
			return nil, fmt.Errorf("unknown step %q (allowed: init, generate, post, seed)", name)
		}
	}
	return steps, nil
}

func runSteps(label string, steps []Step) error {
	return timedStep(label+" total", func() error {
		for _, s := range steps {
			if err := timedStep(s.Name, s.Fn); err != nil {
				return err
			}
		}
		return nil
	})
}

func runSelected(ctx context.Context, label string, s stacks.Stack, opts *stacks.Options, funcs []string) error {
	if s == nil {
		return nil
	}
	steps, err := stackSteps(ctx, label, s, opts, funcs)
	if err != nil {
		return err
	}
	return runSteps(label, steps)
}
