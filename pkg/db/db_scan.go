package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/censys/scan-takehome/pkg/scanning"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Scanning is an interface that defines the methods for a database that can store service scan
// results.
type Scanning interface {
	Upsert(ctx context.Context, scan scanning.Scan) error
}

// PostgresScanning contains a connection pool to a postgres database and implements the [Scanning]
// interface.
type PostgresScanning struct {
	pool *pgxpool.Pool
}

// NewDatabase creates a new [Scanning] instance based on the database URL.
// It currently only supports postgres databases, but this can be extended to support other types.
func NewDatabase(ctx context.Context, dbUrl string) (Scanning, error) {
	if strings.HasPrefix(dbUrl, "postgres://") {
		return NewPostgresScanning(ctx, dbUrl)
	}

	return nil, fmt.Errorf("unsupported database URL: %s", dbUrl)
}

// NewPostgresScanning creates a new [PostgresScanning] instance.
func NewPostgresScanning(ctx context.Context, dbUrl string) (Scanning, error) {
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresScanning{pool: pool}, nil
}

// Upsert inserts a scan into the database, or updates it if it already exists and this scan is more
// recent.
func (p *PostgresScanning) Upsert(ctx context.Context, scan scanning.Scan) error {
	query := `
INSERT INTO scans AS s (ipv4_addr, port, service, resp, updated_at)
VALUES (@ipv4_addr, @port, @service, @resp, to_timestamp(@updated_at))
ON CONFLICT (ipv4_addr, port, service) DO UPDATE SET
	resp = EXCLUDED.resp,
	updated_at = EXCLUDED.updated_at
WHERE EXCLUDED.updated_at > s.updated_at
	`

	_, err := p.pool.Exec(ctx, query, pgx.StrictNamedArgs{
		"ipv4_addr":  scan.Ip,
		"port":       scan.Port,
		"service":    scan.Service,
		"resp":       scan.Data,
		"updated_at": scan.Timestamp,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert scan: %w", err)
	}

	return nil
}
