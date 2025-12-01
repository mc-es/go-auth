// Package database provides MongoDB connection management and lifecycle operations.
package database

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"go-auth/config"
)

// NewDatabase creates a new Database instance.
func NewDatabase(ctx context.Context, cfg config.Database) (Database, error) {
	retrier := newRetrier(retryConfig{
		MaxRetries:        cfg.MaxRetries,
		InitialBackoff:    cfg.RetryBackoff,
		MaxBackoffTime:    defaultMaxBackoff,
		BackoffMultiplier: defaultBackoffMultiplier,
		InitialRetryQuota: defaultInitialRetryQuota,
	})

	opts := options.Client().
		ApplyURI(cfg.URL).
		SetAppName(cfg.Name).
		SetMaxPoolSize(cfg.MaxConns).
		SetMinPoolSize(cfg.MinConns).
		SetMaxConnIdleTime(cfg.MaxIdleTime).
		SetConnectTimeout(cfg.ConnectTO).
		SetServerSelectionTimeout(cfg.SelectionTO)

	var client *mongo.Client

	connection := createConnection(cfg, opts, &client)

	if err := retrier.Do(ctx, connection); err != nil {
		if client != nil {
			disconnectCtx, disconnectCancel := context.WithTimeout(
				context.WithoutCancel(ctx),
				cfg.CloseTO,
			)
			_ = client.Disconnect(disconnectCtx)

			disconnectCancel()
		}

		return nil, wrapDbError(operationRetry, err, map[string]any{
			"database": cfg.Name,
		})
	}

	return &database{
		client:  client,
		config:  cfg,
		retrier: retrier,
	}, nil
}

// Ping checks database connectivity.
func (db *database) Ping(ctx context.Context) error {
	return db.retrier.Do(ctx, func(ctx context.Context) error {
		return db.client.Ping(ctx, readpref.Primary())
	})
}

// GetDBVersion gets the database version.
func (db *database) GetDBVersion(ctx context.Context) (string, error) {
	var version string

	versionOp := func(opCtx context.Context) error {
		var result struct {
			Version string `bson:"version"`
		}

		command := map[string]any{"buildInfo": 1}

		err := db.client.Database("admin").RunCommand(opCtx, command).Decode(&result)
		if err != nil {
			return wrapDbError(operationGetVersion, err, map[string]any{
				"database": db.config.Name,
			})
		}

		version = result.Version

		return nil
	}

	if err := db.retrier.Do(ctx, versionOp); err != nil {
		return "", wrapDbError(operationGetVersion, err, map[string]any{
			"database": db.config.Name,
		})
	}

	return version, nil
}

// Close disconnects the database.
func (db *database) Close() error {
	closeCtx, cancel := context.WithTimeout(context.Background(), db.config.CloseTO)
	defer cancel()

	if err := db.client.Disconnect(closeCtx); err != nil &&
		!errors.Is(err, mongo.ErrClientDisconnected) {
		return wrapDbError(operationDisconnect, err, map[string]any{
			"database": db.config.Name,
		})
	}

	return nil
}

// createConnection creates a connection function that connects and pings the database.
func createConnection(
	cfg config.Database,
	opts *options.ClientOptions,
	client **mongo.Client,
) func(context.Context) error {
	return func(ctx context.Context) error {
		connectCtx, connectCancel := context.WithTimeout(ctx, cfg.ConnectTO)
		defer connectCancel()

		clientLocal, err := mongo.Connect(connectCtx, opts)
		if err != nil {
			return wrapDbError(operationConnect, err, map[string]any{
				"database": cfg.Name,
			})
		}

		pingCtx, pingCancel := context.WithTimeout(ctx, cfg.PingTO)
		defer pingCancel()

		if err := clientLocal.Ping(pingCtx, readpref.Primary()); err != nil {
			disconnectCtx, disconnectCancel := context.WithTimeout(
				context.WithoutCancel(ctx),
				cfg.CloseTO,
			)
			_ = clientLocal.Disconnect(disconnectCtx)

			disconnectCancel()

			return wrapDbError(operationPing, err, map[string]any{
				"database": cfg.Name,
			})
		}

		*client = clientLocal

		return nil
	}
}
