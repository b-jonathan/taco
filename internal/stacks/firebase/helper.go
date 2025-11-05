package firebase

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"
)

func createCredentials(ctx context.Context, projectRoot string, appName string) error {
	projectID := fmt.Sprintf("%s-taco", appName)
	fmt.Printf("Fetching Firebase Web App credentials for project '%s'...\n", projectID)

	out, _, err := execx.RunCmdOutput(ctx, "", fmt.Sprintf("firebase apps:sdkconfig web --project %s --non-interactive", projectID))
	if err != nil {
		return fmt.Errorf("failed to fetch firebase sdk config: %w\nOutput: %s", err, out)
	}

	// Extract the JSON block from Firebase CLI output
	re := regexp.MustCompile(`(?s)\{.*\}`)
	match := re.FindString(out)
	if match == "" {
		return fmt.Errorf("failed to parse firebase sdkconfig output: %s", out)
	}

	var cfg map[string]string
	if err := json.Unmarshal([]byte(match), &cfg); err != nil {
		return fmt.Errorf("failed to unmarshal firebase config: %w", err)
	}

	lines := []string{
		"# --- Firebase Credentials ---",
		fmt.Sprintf("NEXT_PUBLIC_FIREBASE_API_KEY=%s", cfg["apiKey"]),
		fmt.Sprintf("NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=%s", cfg["authDomain"]),
		fmt.Sprintf("NEXT_PUBLIC_FIREBASE_PROJECT_ID=%s", cfg["projectId"]),
		fmt.Sprintf("NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET=%s", cfg["storageBucket"]),
		fmt.Sprintf("NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=%s", cfg["messagingSenderId"]),
		fmt.Sprintf("NEXT_PUBLIC_FIREBASE_APP_ID=%s", cfg["appId"]),
	}

	envPath := filepath.Join(projectRoot, "frontend", ".env.local")
	if err := fsutil.EnsureFile(envPath); err != nil {
		return fmt.Errorf("ensure .env.local: %w", err)
	}
	if err := fsutil.AppendUniqueLines(envPath, lines); err != nil {
		return fmt.Errorf("append firebase env vars: %w", err)
	}

	fmt.Printf("Firebase credentials appended to %s\n", envPath)
	return nil
}
