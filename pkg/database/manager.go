// Package database provides MongoDB connection management and lifecycle operations.
package database

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"go-auth/config"
	"go-auth/pkg/logger"
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
			disconnectCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.CloseTO)
			defer cancel()

			_ = client.Disconnect(disconnectCtx)
		}

		return nil, wrapDbError(operationRetry, err, map[string]any{
			"database": cfg.Name,
		})
	}

	db := &database{
		client:  client,
		config:  cfg,
		retrier: retrier,
	}

	version, err := fetchVersion(ctx, client, retrier, cfg.Name)
	if err != nil {
		logger.S().Warnw("Database connected but failed to fetch version info",
			"database", cfg.Name,
			"error", err,
		)

		return db, nil
	}

	logger.S().Infow("Database connected",
		"database", cfg.Name,
		"db_version", version,
	)

	return db, nil
}

// Ping checks database connectivity.
func (db *database) Ping(ctx context.Context) error {
	return db.retrier.Do(ctx, func(ctx context.Context) error {
		return db.client.Ping(ctx, readpref.Primary())
	})
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
		connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTO)
		defer cancel()

		clientLocal, err := mongo.Connect(connectCtx, opts)
		if err != nil {
			return wrapDbError(operationConnect, err, map[string]any{
				"database": cfg.Name,
			})
		}

		pingCtx, cancel := context.WithTimeout(ctx, cfg.PingTO)
		defer cancel()

		if err := clientLocal.Ping(pingCtx, readpref.Primary()); err != nil {
			disconnectCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.CloseTO)
			defer cancel()

			_ = clientLocal.Disconnect(disconnectCtx)

			return wrapDbError(operationPing, err, map[string]any{
				"database": cfg.Name,
			})
		}

		*client = clientLocal

		return nil
	}
}

// fetchVersion fetches the database version.
func fetchVersion(
	ctx context.Context,
	client *mongo.Client,
	retrier *retrier,
	dbName string,
) (string, error) {
	var version string

	versionOp := func(opCtx context.Context) error {
		var result struct {
			Version string `bson:"version"`
		}

		command := map[string]any{"buildInfo": 1}

		err := client.Database("admin").RunCommand(opCtx, command).Decode(&result)
		if err != nil {
			return wrapDbError(operationGetVersion, err, map[string]any{
				"database": dbName,
			})
		}

		version = result.Version

		return nil
	}

	if err := retrier.Do(ctx, versionOp); err != nil {
		return "", err
	}

	return version, nil
}
