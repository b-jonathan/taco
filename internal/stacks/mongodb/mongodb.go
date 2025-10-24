package mongodb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/b-jonathan/taco/internal/execx"
	"github.com/b-jonathan/taco/internal/fsutil"
	"github.com/b-jonathan/taco/internal/prompt"
	"github.com/b-jonathan/taco/internal/stacks"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Stack = stacks.Stack
type Options = stacks.Options

type mongodb struct{}

func New() Stack { return &mongodb{} }

func (mongodb) Type() string { return "database" }
func (mongodb) Name() string { return "mongodb" }

func (mongodb) Seed(ctx context.Context, opts *Options) error {
	if opts.DatabaseURI == "" {
		return fmt.Errorf("DatabaseURI is empty — did Init() run?")
	}
	fmt.Println(" Starting MongoDB seeding with URI:", opts.DatabaseURI)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(opts.DatabaseURI))
	if err != nil {
		return fmt.Errorf("connect mongo: %w", err)
	}
	defer func() {
		_ = client.Disconnect(ctx)
	}()

	// Ping to confirm connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		return fmt.Errorf("ping mongo: %w", err)
	}

	// Use project name as DB name
	dbName := opts.AppName
	db := client.Database(dbName)
	// Create a test collection
	col := db.Collection("seed_test")

	// Insert a dummy row
	res, err := col.InsertOne(ctx, map[string]int{"value": 1})
	if err != nil {
		return fmt.Errorf("insert dummy row: %w", err)
	}

	fmt.Printf("Seed successful!\nDatabase: %s\nCollection: seed_test\nInserted _id: %v\n", dbName, res.InsertedID)
	return nil
}

func (mongodb) Init(ctx context.Context, opts *Options) error {
	var mongoURI string
	// Step 1: Ask local vs auth
	var choice string
	if prompt.IsTTY() {
		c, _ := prompt.CreateSurveySelect(
			"How do you want to connect to MongoDB?",
			[]string{"Local (default localhost:27017)", "Auth (Atlas or custom URI)"},
			prompt.AskOpts{
				Default:  "Local (default localhost:27017)",
				PageSize: 2,
			},
		)
		choice = c
	}

	// Step 2: Set URI if local
	if strings.HasPrefix(choice, "Local") {
		mongoURI = "mongodb://localhost:27017"
		fmt.Println("Using default local MongoDB URI:", mongoURI)
	} else {
		// Step 3: Interactive loop for authenticated URI with "undo" option
		for {
			if prompt.IsTTY() {
				uri, _ := prompt.CreateSurveyInput(
					"Enter your MongoDB connection URI (type 'undo' to go back):",
					prompt.AskOpts{
						Help:      "Example: mongodb+srv://username:password@cluster.mongodb.net/db",
						Validator: survey.Required,
					},
				)
				mongoURI = strings.TrimSpace(uri)
			}

			// Allow undo: re-ask local vs auth
			if mongoURI == "undo" {
				c, _ := prompt.CreateSurveySelect(
					"How do you want to connect to MongoDB?",
					[]string{"Local (default localhost:27017)", "Auth (Atlas or custom URI)"},
					prompt.AskOpts{
						Default:  "Local (default localhost:27017)",
						PageSize: 2,
					},
				)
				choice = c
				if strings.HasPrefix(choice, "Local") {
					mongoURI = "mongodb://localhost:27017"
					fmt.Println("Using default local MongoDB URI:", mongoURI)
					break
				}
				continue // go back to asking for URI
			}

			if err := EnsureMongoURI(mongoURI); err != nil {
				fmt.Println("Invalid MongoDB URI:", err)
				continue
			}
			break
		}
	}

	opts.DatabaseURI = mongoURI
	fmt.Println("✅ Final MongoDB URI set:", opts.DatabaseURI)

	return nil
}

func (mongodb) Generate(ctx context.Context, opts *Options) error {
	backendDir := filepath.Join(opts.ProjectRoot, "backend")
	if err := execx.RunCmd(ctx, backendDir, "npm install mongodb"); err != nil {
		return fmt.Errorf("npm install mongodb: %w", err)
	}
	if err := execx.RunCmd(ctx, backendDir, "npm install -D @types/mongodb"); err != nil {
		return fmt.Errorf("npm install dev: %w", err)
	}
	clientPath := filepath.Join(backendDir, "src", "db", "client.ts")
	clientContent, err := fsutil.RenderTemplate("mongodb/db/client.ts.tmpl")
	if err != nil {
		return err
	}

	client := fsutil.FileInfo{
		Path:    clientPath,
		Content: clientContent,
	}

	if err := fsutil.WriteFile(client); err != nil {
		return err
	}

	indexPath := filepath.Join(backendDir, "src", "index.ts")
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("read index.ts: %w", err)
	}

	src := string(indexContent)

	// Inject DB import
	if !strings.Contains(src, "connectDB") {
		src = strings.Replace(src, "// [DATABASE IMPORT]", `
		import { connectDB } from "./db/client";`, 1)
	}

	// Inject route
	if !strings.Contains(src, "/seed") {
		route, err := fsutil.RenderTemplate("mongodb/seed.tmpl")
		if err != nil {
			return fmt.Errorf("render seed route template: %w", err)
		}
		src = strings.Replace(src, "// [DATABASE ROUTE]", string(route), 1)
	}
	index := fsutil.FileInfo{
		Path:    indexPath,
		Content: []byte(src),
	}
	return fsutil.WriteFile(index)
}

func (mongodb) Post(ctx context.Context, opts *Options) error {
	// gitignorePath := filepath.Join(opts.ProjectRoot, ".gitignore")
	// if err := fsutil.EnsureFile(gitignorePath); err != nil {
	// 	return fmt.Errorf("ensure gitignore file: %w", err)
	// }

	// _ = fsutil.AppendUniqueLines(gitignorePath,
	// 	[]string{"backend/node_modules/", "backend/dist/", "backend/.env*"})
	path := filepath.Join(opts.ProjectRoot, "backend", ".env")
	// dir := filepath.Dir(path)
	// if err := os.MkdirAll(dir, 0o755); err != nil {
	// 	return fmt.Errorf("mkdir %s: %w", dir, err)
	// }
	// TODO: Make this not as scuffed lol
	content := fmt.Sprintf(`
	MONGODB_URI=%s/%s`, opts.DatabaseURI, opts.AppName)
	_ = fsutil.AppendUniqueLines(path, []string{content})
	return nil
}
