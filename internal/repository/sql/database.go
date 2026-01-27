package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"wechat-service/pkg/logger"
)

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host            string
	Port            int
	Username        string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewDatabaseConfig creates a new DatabaseConfig with defaults
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}
}

// ConnectionString returns the PostgreSQL connection string
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Name, c.SSLMode,
	)
}

// Database wraps sql.DB with additional functionality
type Database struct {
	db          *sql.DB
	cfg         *DatabaseConfig
	logger      *logger.Logger
	tablePrefix string
}

// NewDatabase creates a new Database instance
func NewDatabase(cfg *DatabaseConfig, tablePrefix string, l *logger.Logger) (*Database, error) {
	db, err := sql.Open("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if l == nil {
		l = logger.New()
	}

	return &Database{
		db:          db,
		cfg:         cfg,
		logger:      l,
		tablePrefix: tablePrefix,
	}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// DB returns the underlying sql.DB
func (d *Database) DB() *sql.DB {
	return d.db
}

// TablePrefix returns the table prefix
func (d *Database) TablePrefix() string {
	return d.tablePrefix
}

// Ping checks database connectivity
func (d *Database) Ping(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// Stats returns database connection statistics
func (d *Database) Stats() sql.DBStats {
	return d.db.Stats()
}

// NewUnitOfWork creates a new UnitOfWork
func (d *Database) NewUnitOfWork() *UnitOfWork {
	return NewUnitOfWork(d.db)
}

// User returns a UserSQLExecutor
func (d *Database) User() *UserSQLExecutor {
	return NewUserSQLExecutor(d.db, d.tablePrefix)
}

// Message returns a MessageSQLExecutor
func (d *Database) Message() *MessageSQLExecutor {
	return NewMessageSQLExecutor(d.db, d.tablePrefix)
}

// WithTransaction executes a function within a transaction
func (d *Database) WithTransaction(ctx context.Context, opts ...TxOption, fn func(uow *UnitOfWork) error) error {
	uow := d.NewUnitOfWork()
	if err := uow.Begin(ctx, opts...); err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = uow.Rollback()
			panic(p)
		}
	}()

	if err := fn(uow); err != nil {
		if rbErr := uow.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %w", err, rbErr)
		}
		return err
	}

	return uow.Commit()
}

// HealthCheck performs a database health check
func (d *Database) HealthCheck(ctx context.Context) error {
	if err := d.Ping(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	stats := d.Stats()
	if stats.MaxOpenConnections > 0 {
		openRatio := float64(stats.OpenConnections) / float64(stats.MaxOpenConnections)
		if openRatio > 0.9 {
			d.logger.Warnf("Database connection pool near capacity: %d/%d",
				stats.OpenConnections, stats.MaxOpenConnections)
		}
	}

	return nil
}

// Exec executes a query without returning results
func (d *Database) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (d *Database) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (d *Database) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}
