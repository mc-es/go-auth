// Package database provides MongoDB connection management and lifecycle operations.
package database

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go-auth/config"
	"go-auth/pkg/logger"
)

// Database wraps MongoDB client and database instances with connection management.
type Database struct {
	Client   *mongo.Client         // The MongoDB client instance.
	Database *mongo.Database       // The MongoDB database instance.
	config   config.DatabaseConfig // The database configuration.
	health   *HealthMonitor        // The health monitor instance.
}

// NewDatabase creates a new Database instance.
func NewDatabase(cfg config.DatabaseConfig) (*Database, error) {
	start := time.Now()

	logger.S().Infow("Initializing database connection",
		"database", cfg.Name,
		"max_connections", cfg.MaxConns,
		"min_connections", cfg.MinConns,
		"max_idle_time", cfg.MaxIdleTime,
		"connect_timeout", cfg.ConnectTO,
	)

	opts := options.Client().
		ApplyURI(cfg.URL).
		SetAppName(cfg.Name).
		SetMaxPoolSize(cfg.MaxConns).
		SetMinPoolSize(cfg.MinConns).
		SetMaxConnIdleTime(cfg.MaxIdleTime).
		SetConnectTimeout(cfg.ConnectTO).
		SetServerSelectionTimeout(cfg.SelectionTO)

	connectCtx, connectCancel := context.WithTimeout(context.Background(), cfg.ConnectTO)
	defer connectCancel()

	client, err := mongo.Connect(connectCtx, opts)
	if err != nil {
		return nil, WrapErrorWithMetadata(operationConnect, errConnectionFailed, map[string]any{
			"error":    err.Error(),
			"duration": time.Since(start),
			"database": cfg.Name,
		})
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), cfg.PingTO)
	defer pingCancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())

		return nil, WrapErrorWithMetadata(operationPing, errPingFailed, map[string]any{
			"error":    err.Error(),
			"duration": time.Since(start),
			"database": cfg.Name,
		})
	}

	logger.S().Debugw("Database connected successfully",
		"database", cfg.Name,
		"url", maskConnectionString(cfg.URL),
		"duration", time.Since(start),
	)

	database := &Database{
		Client:   client,
		Database: client.Database(cfg.Name),
		config:   cfg,
	}
	database.health = NewHealthMonitor(database.Ping)

	return database, nil
}

// Ping checks database connectivity.
func (db *Database) Ping(ctx context.Context) error {
	start := time.Now()

	pingCtx, cancel := context.WithTimeout(ctx, db.config.PingTO)
	defer cancel()

	if err := db.Client.Ping(pingCtx, nil); err != nil {
		return WrapErrorWithMetadata(operationPing, errPingFailed, map[string]any{
			"error":    err.Error(),
			"duration": time.Since(start),
			"database": db.config.Name,
		})
	}

	return nil
}

// Close disconnects the database and stops the health check goroutine.
func (db *Database) Close(ctx context.Context) error {
	start := time.Now()

	if db.health != nil {
		db.health.StopHealthCheck()
	}

	ctx, cancel := context.WithTimeout(ctx, db.config.CloseTO)
	defer cancel()

	if err := db.Client.Disconnect(ctx); err != nil &&
		!errors.Is(err, mongo.ErrClientDisconnected) {
		return WrapErrorWithMetadata(operationDisconnect, errDisconnectFailed, map[string]any{
			"error":    err.Error(),
			"duration": time.Since(start),
			"database": db.config.Name,
		})
	}

	return nil
}

// StartHealthCheck starts the health check goroutine.
func (db *Database) StartHealthCheck(ctx context.Context) {
	if db.health != nil {
		cfg := HealthCheckConfig{
			Interval: db.config.HealthCheckIT,
			Timeout:  db.config.HealthCheckTO,
		}
		db.health.StartHealthCheck(ctx, cfg)
	}
}

// maskConnectionString masks the password in MongoDB connection strings.
func maskConnectionString(rawURL string) string {
	if rawURL == "" {
		return rawURL
	}

	if !strings.HasPrefix(rawURL, "mongodb://") && !strings.HasPrefix(rawURL, "mongodb+srv://") {
		return rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if parsedURL.User != nil {
		if _, hasPassword := parsedURL.User.Password(); hasPassword {
			parsedURL.User = url.UserPassword(parsedURL.User.Username(), "***")
		} else {
			parsedURL.User = url.User(parsedURL.User.Username())
		}
	}

	return parsedURL.String()
}
