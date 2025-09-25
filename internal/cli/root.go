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
	"github.com/b-jonathan/taco/internal/prompt"
	"github.com/b-jonathan/taco/internal/stacks"
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

	return params, nil
}

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Create repo and scaffold",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			ctx := cmd.Context()
			stack := map[string]string{
				"frontend": "",
				"backend":  "",
			}
			params, err := gatherInitParams(cmd, args)
			if err != nil {
				return err
			}

			projectRoot := filepath.Join("..", params.Name)
			if err := os.MkdirAll(projectRoot, 0o755); err != nil {
				return fmt.Errorf("mkdir project root: %w", err)
			}

			stack["frontend"], _ = prompt.CreateSurveySelect("Choose a Frontend Stack:\n", []string{"NextJS", "None"}, prompt.AskOpts{})
			stack["frontend"] = strings.ToLower(stack["frontend"])
			stack["backend"], _ = prompt.CreateSurveySelect("Choose a Backend Stack:\n", []string{"Express", "None"}, prompt.AskOpts{})
			stack["backend"] = strings.ToLower(stack["backend"])

			frontend, err := GetFactory(stack["frontend"])
			if err != nil {
				return err
			}
			backend, err := GetFactory(stack["backend"])
			if err != nil {
				return err
			}

			opts := stacks.Options{
				ProjectRoot:    projectRoot,
				AppName:        params.Name,
				FrontendOrigin: "http://localhost:3000",
				BackendURL:     "http://localhost:4000",
				Port:           4000,
			}

			g, ctx := errgroup.WithContext(cmd.Context())

			g.Go(func() error { return runStackSequenceNoPost(ctx, "Frontend", frontend, opts) })
			g.Go(func() error { return runStackSequenceNoPost(ctx, "Backend", backend, opts) })

			if err := g.Wait(); err != nil {
				return err
			}

			// log.Println("Starting gh command")
			// client := gh.MustFromContext(cmd.Context())
			// log.Println("GitHub client initialized")
			// ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			// defer cancel()

			// newRepo := &github.Repository{
			// 	Name:        github.String(params.Name),
			// 	Private:     github.Bool(params.Private),
			// 	Description: github.String(params.Description),
			// }

			// repo, _, err := client.Repositories.Create(ctx, "", newRepo)
			// if err != nil {
			// 	return fmt.Errorf("create repo: %w", err)
			// }

			// log.Println(cmd.OutOrStdout(), "Created:", repo.GetHTMLURL())
			// remoteURL := repo.GetSSHURL()
			// if params.Remote == "https" {
			// 	remoteURL = repo.GetCloneURL()
			// }
			// log.Println("Committing and Pushing to Github...")
			// if err := git.InitAndPush(ctx, projectRoot, remoteURL, "chore: initial commit"); err != nil {
			// 	_, err := client.Repositories.Delete(ctx, "", *newRepo.Name)
			// 	return err
			// }
			// log.Println("Pushed:", repo.GetHTMLURL())
			log.Println("Time Taken:", time.Since(start))
			return nil
		},
	}
	// Flags that feed into gatherInitParams
	cmd.Flags().Bool("private", false, "Make the repository private")
	cmd.Flags().String("remote", "ssh", "Remote URL type ssh or https")
	cmd.Flags().String("description", "", "Repository description")
	return cmd
}

func timedStep(name string, fn func() error) error {
	start := time.Now()
	err := fn()
	dur := time.Since(start)
	if err != nil {
		log.Printf("%s failed in %s: %v", name, dur, err)
		return err
	}
	log.Printf("%s finished in %s", name, dur)
	return nil
}

func runStackSequenceNoPost(ctx context.Context, label string, s stacks.Stack, opts stacks.Options) error {
	if s == nil {
		return nil
	}
	return timedStep(label+" total", func() error {
		if err := timedStep(label+" Init", func() error { return s.Init(ctx, opts) }); err != nil {
			return err
		}
		if err := timedStep(label+" Generate", func() error { return s.Generate(ctx, opts) }); err != nil {
			return err
		}
		return timedStep(label+" Post", func() error { return s.Post(ctx, opts) })

	})
}
