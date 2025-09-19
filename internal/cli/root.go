package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v55/github"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	// "github.com/AlecAivazis/survey/v2"
)

type ctxKey int

const ghClientKey ctxKey = iota

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
		if ctx.Value(ghClientKey) != nil {
			return nil
		}

		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			return fmt.Errorf("set GITHUB_TOKEN")
		}

		gh := ghClient(ctx, token)
		gh.UserAgent = "taco-cli"
		gh_ctx := context.WithValue(ctx, ghClientKey, gh)
		cmd.SetContext(gh_ctx)
		return nil
	}
	cmd.AddCommand(newGHWhoAmICmd())
	cmd.AddCommand(makeNewGHRepo())
	return cmd
}

func ghClient(context context.Context, token string) *github.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return github.NewClient(oauth2.NewClient(context, tokenSource))
}

func ghFromContext(cmd *cobra.Command) *github.Client {
	v := cmd.Context().Value(ghClientKey)
	if v == nil {
		// Root PersistentPreRunE should have set it
		panic("github client missing in context")
	}
	return v.(*github.Client)
}

func newGHWhoAmICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gh",
		Short: "Show the authenticated GitHub user",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println("Starting gh command")
			client := ghFromContext(cmd)
			log.Println("GitHub client initialized")

			// Per-call timeout that still reuses the long-lived client
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			user, _, err := client.Users.Get(ctx, "")
			if err != nil {
				log.Printf("Error fetching user: %v", err)
				return err
			}

			log.Printf("Successfully fetched user: %s", user.GetLogin())
			fmt.Fprintln(cmd.OutOrStdout(), user.GetLogin())
			return nil
		},
	}
}

func makeNewGHRepo() *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "Make new Github Repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println("Starting gh command")
			client := ghFromContext(cmd)
			log.Println("GitHub client initialized")
			newRepo := &github.Repository{
				Name:    github.String(args[0]),
				Private: github.Bool(false),
			}
			repo, _, err := client.Repositories.Create(cmd.Context(), "", newRepo)
			if err != nil {
				fmt.Println("Repo:", repo)
			}

			return nil
		},
	}
}
