// Package database provides MongoDB connection management and lifecycle operations.
package database

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"go-auth/config"
	"go-auth/pkg/logger"
)

// NewDatabase creates a new Database instance.
func NewDatabase(ctx context.Context, cfg config.Database) (Database, error) {
	retrier := newRetrier(retryConfig{
		MaxRetries:        cfg.Resilience.MaxRetries,
		InitialBackoff:    cfg.Resilience.RetryBackoff,
		MaxBackoffTime:    defaultMaxBackoff,
		BackoffMultiplier: defaultBackoffMultiplier,
		InitialRetryQuota: defaultInitialRetryQuota,
	})

	hostPort := net.JoinHostPort(cfg.Connection.Host, strconv.FormatUint(cfg.Connection.Port, 10))
	uri := fmt.Sprintf("mongodb://%s/%s", hostPort, cfg.Connection.Name)

	opts := options.Client().
		ApplyURI(uri).
		SetAppName(cfg.Connection.Name).
		SetMaxPoolSize(cfg.Pool.MaxOpen).
		SetMinPoolSize(cfg.Pool.MinOpen).
		SetMaxConnIdleTime(cfg.Pool.MaxIdleTime).
		SetConnectTimeout(cfg.Timeout.Connect)

	var client *mongo.Client

	connection := createConnection(cfg, opts, &client)

	if err := retrier.Do(ctx, connection); err != nil {
		if client != nil {
			dcCtx, dcCancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.Timeout.Close)
			defer dcCancel()

			_ = client.Disconnect(dcCtx)
		}

		return nil, wrapDbError(operationRetry, err, map[string]any{
			"database": cfg.Connection.Name,
		})
	}

	db := &database{
		client:  client,
		config:  cfg,
		retrier: retrier,
	}

	version, err := fetchVersion(ctx, client, retrier, cfg.Connection.Name)
	if err != nil {
		logger.S().Warnw("Database connected but failed to fetch version info",
			"db_name", cfg.Connection.Name,
			"error", err,
		)

		return db, nil
	}

	logger.S().Infow("Database connected",
		"db_name", cfg.Connection.Name,
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
	closeCtx, cancel := context.WithTimeout(context.Background(), db.config.Timeout.Close)
	defer cancel()

	if err := db.client.Disconnect(closeCtx); err != nil &&
		!errors.Is(err, mongo.ErrClientDisconnected) {
		return wrapDbError(operationDisconnect, err, map[string]any{
			"db_name": db.config.Connection.Name,
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
		connectCtx, connectCancel := context.WithTimeout(ctx, cfg.Timeout.Connect)
		defer connectCancel()

		clientLocal, err := mongo.Connect(connectCtx, opts)
		if err != nil {
			return wrapDbError(operationConnect, err, map[string]any{
				"db_name": cfg.Connection.Name,
			})
		}

		pingCtx, pingCancel := context.WithTimeout(ctx, cfg.Timeout.Ping)
		defer pingCancel()

		if err := clientLocal.Ping(pingCtx, readpref.Primary()); err != nil {
			dcCtx, dcCancel := context.WithTimeout(context.WithoutCancel(ctx), cfg.Timeout.Close)
			defer dcCancel()

			_ = clientLocal.Disconnect(dcCtx)

			return wrapDbError(operationPing, err, map[string]any{
				"db_name": cfg.Connection.Name,
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
				"db_name": dbName,
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
