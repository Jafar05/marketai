package postgresql

import (
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	migrateDriver "github.com/golang-migrate/migrate/v4/database/pgx"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
)

const (
	defaultStatementTimeout = 30 * time.Second
)

type Migrate struct {
	config *PostgresCfg
	logger *zap.Logger
	source *bindata.AssetSource
}

func newMigrate(
	config *PostgresCfg,
	l *zap.Logger,
	source *bindata.AssetSource,
) *Migrate {

	return &Migrate{config: config, logger: l, source: source}
}

func (m *Migrate) Run() (err error) {
	d, err := bindata.WithInstance(m.source)
	if err != nil {
		return fmt.Errorf("cannot init driver: %w", err)
	}

	cfg, err := NewPgxConnConfig(m.config)

	if err != nil {
		return fmt.Errorf("cannot init pg config: %w", err)
	}

	m.logger.Debug(
		"statementMode",
		zap.Stringer("val", cfg.DefaultQueryExecMode),
	)

	if m.config.MigrateUser != "" && m.config.MigratePassword != "" {
		cfg.User = m.config.MigrateUser
		cfg.Password = m.config.MigratePassword
	}

	pgxDb := stdlib.OpenDB(*cfg)
	defer func() {
		_ = pgxDb.Close()
	}()

	if m.config.MigratePreScript != "" {
		_, err := pgxDb.Exec(m.config.MigratePreScript)
		if err != nil {
			m.logger.Warn("cannot run migrate prescript", zap.Error(err))
		}
	}

	driver, err := migrateDriver.WithInstance(pgxDb, &migrateDriver.Config{
		SchemaName:       m.config.MigrateSchema,
		MigrationsTable:  m.config.MigrateTable,
		StatementTimeout: defaultStatementTimeout,
	})
	if err != nil {
		return fmt.Errorf("cannot init driver: %w", err)
	}

	defer func() {
		_ = driver.Close()
	}()

	mgr, err := migrate.NewWithInstance("go-bindata", d, m.config.DBName, driver)
	if err != nil {
		return err
	}

	defer func() {
		srcErr, dbErr := mgr.Close()

		if srcErr != nil {
			m.logger.Info("migration source close error", zap.Error(srcErr))
		}

		if dbErr != nil {
			m.logger.Info("database close error", zap.Error(dbErr))
		}
	}()

	if err = mgr.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return err
	}

	return nil
}
