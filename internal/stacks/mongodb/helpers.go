package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureMongoURI tries to connect to the MongoDB instance at the given URI
// and pings it to confirm the server is responding.
func EnsureMongoURI(uri string) error {
	if uri == "" {
		uri = "mongodb://127.0.0.1:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer dcancel()
		if derr := client.Disconnect(dctx); derr != nil {
			err = errors.Join(err, fmt.Errorf("disconnect failed: %w", derr))
		}
	}()

	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	return nil
}
