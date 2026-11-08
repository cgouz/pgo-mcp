package database

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RegistryOptions struct {
	Discovery bool
	Allowed   []string
	AllowAll  bool
	MaxPools  int
	MaxConns  int32
	MinConns  int32
}

type Database struct {
	Name    string
	Default bool
}

type Registry struct {
	config      *pgxpool.Config
	defaultName string
	allowed     map[string]bool
	allowAll    bool
	maxPools    int

	mu     sync.Mutex
	pools  map[string]*pgxpool.Pool
	closed bool
}

func NewRegistry(ctx context.Context, uri string, options RegistryOptions) (*Registry, error) {
	if !options.Discovery &&
		(options.AllowAll || len(options.Allowed) > 0) {
		return nil, fmt.Errorf(
			"database discovery must be enabled for additional databases",
		)
	}

	cfg, err := pgxpool.ParseConfig(uri)
	{
		if err != nil {
			return nil, fmt.Errorf("parse database config: %w", err)
		}

		if cfg.ConnConfig.Database == "" {
			return nil, fmt.Errorf("DATABASE_URI must include a database name")
		}

		cfg.MaxConns = options.MaxConns
		cfg.MinConns = options.MinConns
	}

	if options.MaxPools < 1 {
		return nil, fmt.Errorf("database max pools must be positive")
	}

	pool, err := connectConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	registry := &Registry{
		config:      cfg,
		defaultName: cfg.ConnConfig.Database,
		allowed:     make(map[string]bool),
		allowAll:    options.AllowAll,
		maxPools:    options.MaxPools,
		pools: map[string]*pgxpool.Pool{
			cfg.ConnConfig.Database: pool,
		},
	}

	for _, name := range options.Allowed {
		registry.allowed[name] = true
	}

	if options.Discovery && len(registry.allowed) > 0 {
		if err := registry.validateAllowed(ctx); err != nil {
			registry.Close()
			return nil, err
		}
	}

	return registry, nil
}

func (r *Registry) Default() *pgxpool.Pool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.pools[r.defaultName]
}

func (r *Registry) DefaultName() string {
	return r.defaultName
}

func (r *Registry) Pool(ctx context.Context, name string) (*pgxpool.Pool, error) {

	if name == "" {
		name = r.defaultName
	}

	if name != r.defaultName {
		if !r.allowAll && !r.allowed[name] {
			return nil, fmt.Errorf("database is not permitted")
		}

		eligible, err := r.eligible(ctx, name)
		if err != nil {
			return nil, err
		}
		if !eligible {
			return nil, fmt.Errorf("database not found or CONNECT privilege denied")
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, fmt.Errorf("database registry is closed")
	}

	if pool := r.pools[name]; pool != nil {
		return pool, nil
	}

	if len(r.pools) >= r.maxPools {
		return nil, fmt.Errorf("database pool limit reached")
	}

	cfg := r.config.Copy()
	cfg.ConnConfig.Database = name

	pool, err := connectConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	r.pools[name] = pool

	return pool, nil
}

func (r *Registry) Databases(ctx context.Context) ([]Database, error) {

	rows, err := r.Default().Query(
		ctx,
		`SELECT
			datname
		FROM
			pg_catalog.pg_database
		WHERE
			datallowconn = true
			AND datistemplate = false
			AND has_database_privilege(oid, 'CONNECT')
		ORDER BY
			datname`,
	)
	{
		if err != nil {
			return nil, fmt.Errorf("list databases: %w", err)
		}

		defer rows.Close()
	}

	databases := make([]Database, 0)
	foundDefault := false

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan database: %w", err)
		}

		if name != r.defaultName &&
			!r.allowAll && !r.allowed[name] {
			continue
		}

		databases = append(databases, Database{
			Name:    name,
			Default: name == r.defaultName,
		})

		if name == r.defaultName {
			foundDefault = true
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read databases: %w", err)
	}

	if !foundDefault {
		databases = append(databases, Database{
			Name:    r.defaultName,
			Default: true,
		})
		sort.Slice(databases, func(i, j int) bool {
			return databases[i].Name < databases[j].Name
		})
	}

	return databases, nil
}

func (r *Registry) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return
	}

	r.closed = true

	for _, pool := range r.pools {
		pool.Close()
	}
}

func (r *Registry) validateAllowed(ctx context.Context) error {
	for name := range r.allowed {
		if name == r.defaultName {
			continue
		}

		eligible, err := r.eligible(ctx, name)
		if err != nil {
			return err
		}
		if !eligible {
			return fmt.Errorf(
				"database %q not found or CONNECT privilege denied", name,
			)
		}
	}

	return nil
}

func (r *Registry) eligible(ctx context.Context, name string) (bool, error) {
	var eligible bool

	if err := r.Default().QueryRow(
		ctx,
		`SELECT EXISTS (
			SELECT 1
			FROM pg_catalog.pg_database
			WHERE datname = $1
				AND datallowconn = true
				AND datistemplate = false
				AND has_database_privilege(oid, 'CONNECT')
		)`, name,
	).Scan(&eligible); err != nil {
		return false, fmt.Errorf("check database access: %w", err)
	}

	return eligible, nil
}

func connectConfig(ctx context.Context, cfg *pgxpool.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.NewWithConfig(ctx, cfg.Copy())
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return pool, nil
}
