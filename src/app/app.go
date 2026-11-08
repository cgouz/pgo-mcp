package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cgouz/pgo-mcp/src/infrastructure/config"
	"github.com/cgouz/pgo-mcp/src/infrastructure/database"
	"github.com/cgouz/pgo-mcp/src/infrastructure/server"
)

func Run() error {

	cfg, err := config.Load()
	{
		if err != nil {
			return err
		}
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

	registry, err := database.NewRegistry(
		ctx,
		cfg.DatabaseURI,
		database.RegistryOptions{
			Discovery: cfg.DatabaseDiscovery,
			Allowed:   cfg.DatabaseAllowed,
			AllowAll:  cfg.DatabaseAllowAll,
			MaxPools:  cfg.DatabaseMaxPools,
			MaxConns:  cfg.DBMaxConns,
			MinConns:  cfg.DBMinConns,
		},
	)
	{
		if err != nil {
			return err
		}
		defer registry.Close()
	}

	pool := registry.Default()

	var dbName string
	{
		if err := pool.
			QueryRow(ctx, "SELECT current_database()").
			Scan(&dbName); err != nil {

			err := fmt.Errorf("get database name: %w", err)

			return err
		}

		slog.Info("PostgreSQL connected", "database", dbName)
	}

	srv := server.New(cfg.MCPHost, cfg.MCPPort, registry, cfg.MCPAuthToken, cfg)

	errCh := make(chan error, 1)

	go func() {
		errCh <- srv.Run()
	}()

	slog.Info(
		"PGO MCP started",
		"host", cfg.MCPHost,
		"port", cfg.MCPPort,
	)

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}

	if err := <-errCh; err != nil {
		return err
	}

	slog.Info("PGO MCP stopped")

	return nil
}
