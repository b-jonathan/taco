package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/b-jonathan/taco/internal/gh"
	"github.com/google/go-github/v55/github"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
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
		if !isTTY() {
			return params, fmt.Errorf("name required in non-interactive mode")
		}
		name, err := createSurveyInput("Repository Name:", AskOpts{Help: "lowercase letters, numbers, dash, and underscore only", Validator: survey.Required})
		if err != nil {
			return params, err
		}
		params.Name = name
	}

	if f := cmd.Flags().Lookup("private"); f != nil && f.Changed {
		b, _ := strconv.ParseBool(f.Value.String())
		params.Private = b
	} else {
		b, err := createSurveyConfirm("Make repository private?", AskOpts{
			Default: false,
		})
		if err != nil && isTTY() {
			return params, err
		}
		if err == nil {
			params.Private = b
		}
	}

	if v, _ := cmd.Flags().GetString("remote"); v != "" {
		params.Remote = v
	} else {
		if isTTY() {
			r, err := createSurveySelect("Remote URL type", []string{"ssh", "https"}, AskOpts{
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
		if isTTY() {
			desc, err := createSurveyInput("Repository description", AskOpts{
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
			params, err := gatherInitParams(cmd, args)
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Name=%s Private=%t Remote=%s Desc = %q\n", params.Name, params.Private, params.Remote, params.Description)

			log.Println("Starting gh command")
			gh := gh.MustFromContext(cmd.Context())
			log.Println("GitHub client initialized")
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			newRepo := &github.Repository{
				Name:        github.String(params.Name),
				Private:     github.Bool(params.Private),
				Description: github.String(params.Description),
			}

			repo, _, err := gh.Repositories.Create(ctx, "", newRepo)
			if err != nil {
				return fmt.Errorf("create repo: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Created:", repo.GetHTMLURL())
			// Continue with local scaffold, git init, push, etc., using p.Remote to choose SSH or HTTPS
			return nil
		},
	}
	// Flags that feed into gatherInitParams
	cmd.Flags().Bool("private", false, "Make the repository private")
	cmd.Flags().String("remote", "ssh", "Remote URL type ssh or https")
	cmd.Flags().String("description", "", "Repository description")
	return cmd
}

func isTTY() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
