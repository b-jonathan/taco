package cli

import (
	"fmt"
	"context"
	"log"
	"os"
	"github.com/spf13/cobra"
    "github.com/google/go-github/v55/github"
    "golang.org/x/oauth2"
	"github.com/joho/godotenv"
	
	// "github.com/AlecAivazis/survey/v2"
)

func Execute() error { 
	_ = godotenv.Load()
	return newRootCmd().Execute() 
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "taco",
		Short: "Project Initializer",
		Long:  `taco is a CLI tool for initializing new projects that's language-agnostic.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newGHWhoAmICmd())
	return cmd
}

func ghClient(context context.Context, token string) *github.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return github.NewClient(oauth2.NewClient(context, tokenSource))
}

func newGHWhoAmICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gh",
		Short: "Show the authenticated GitHub user",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println("Starting gh command")

			token := os.Getenv("GITHUB_TOKEN")
			if token == "" {
				log.Println("No GITHUB_TOKEN found")
				return fmt.Errorf("set GITHUB_TOKEN")
			}
			log.Println("GITHUB_TOKEN found")

			ctx := cmd.Context()
			client := ghClient(ctx, token)
			log.Println("GitHub client initialized")

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