package postgresql

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"
)

// 	pgxLog "github.com/jackc/pgx-zap"

type Secrets struct {
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	MigrateUser     string `mapstructure:"migrateUser"`
	MigratePassword string `mapstructure:"migratePassword"`
}

type PostgresCfg struct {
	Host               string `mapstructure:"host" dsn:"host" validate:"required"`
	Port               string `mapstructure:"port" dsn:"port" validate:"required"`
	User               string `mapstructure:"user" dsn:"user"`
	Password           string `mapstructure:"password" dsn:"password"`
	DBName             string `mapstructure:"dbName" dsn:"dbname" validate:"required"`
	StatementCacheMode string `mapstructure:"cacheMode" dsn:"default_query_exec_mode,omitempty"`

	SSLMode     string `mapstructure:"sslMode" dsn:"sslmode,omitempty"`
	SSLCert     string `mapstructure:"sslcert" dsn:"sslcert,omitempty"`
	SSLKey      string `mapstructure:"sslkey" dsn:"sslkey,omitempty"`
	SSLPassword string `mapstructure:"sslpassword" dsn:"sslpassword,omitempty"`
	SSLRootCert string `mapstructure:"sslrootcert" dsn:"sslrootcert,omitempty"`

	MigrateTable     string `mapstructure:"migrateTable"`
	MigrateSchema    string `mapstructure:"migrateSchema"`
	MigrateUser      string `mapstructure:"migrateUser"`
	MigratePassword  string `mapstructure:"migratePassword"`
	MigratePreScript string `mapstructure:"migratePreScript"`

	MaxConn                  int32 `mapstructure:"maxConn" validate:"required,gte=0"`
	MinConn                  int32 `mapstructure:"minConn" validate:"required,gte=0"`
	HealthCheckPeriodSeconds int   `mapstructure:"healthCheckPeriodSeconds" validate:"required,gte=0"`
	MaxConnIdleTimeSeconds   int   `mapstructure:"maxConnIdleTimeSeconds" validate:"required,gte=0"`
	MaxConnLifetimeSeconds   int   `mapstructure:"maxConnLifetimeSeconds" validate:"required,gte=0"`
}

// NewPgxPool pool
func NewPgxPool(cfg *PostgresCfg) (*pgxpool.Pool, error) {
	ctx := context.Background()

	dsn, err := prepareDSNString(cfg)
	if err != nil {
		return nil, err
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = cfg.MaxConn
	poolCfg.HealthCheckPeriod = time.Duration(cfg.HealthCheckPeriodSeconds) * time.Second
	poolCfg.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTimeSeconds) * time.Second
	poolCfg.MaxConnLifetime = time.Duration(cfg.MaxConnLifetimeSeconds) * time.Second
	poolCfg.MinConns = cfg.MinConn

	connPool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("%w: NewPgxPool", err)
	}

	return connPool, nil
}

func newPgxPool(cfg *PostgresCfg, logger *zap.Logger) (*pgxpool.Pool, error) {
	fmt.Println("cfg.user====", cfg.User)
	if cfg.User == "" {
		return nil, errors.New("empty user")
	}

	if cfg.Password == "" {
		return nil, errors.New("empty password")
	}

	ctx := context.Background()

	dsn, err := prepareDSNString(cfg)
	if err != nil {
		return nil, err
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = cfg.MaxConn
	poolCfg.HealthCheckPeriod = time.Duration(cfg.HealthCheckPeriodSeconds) * time.Second
	poolCfg.MaxConnIdleTime = time.Duration(cfg.MaxConnIdleTimeSeconds) * time.Second
	poolCfg.MaxConnLifetime = time.Duration(cfg.MaxConnLifetimeSeconds) * time.Second
	poolCfg.MinConns = cfg.MinConn

	connPool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("%w: newPgxPool", err)
	}

	return connPool, nil
}

func NewPgxConnConfig(cfg *PostgresCfg) (*pgx.ConnConfig, error) {

	dsn, err := prepareDSNString(cfg)
	if err != nil {
		return nil, err
	}

	connCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	return connCfg, nil
}

// Prepare libpq conn string
// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
//
//	# Example DSN
//	user=jack password=secret host=pg.example.com port=5432 dbname=mydb sslmode=verify-ca
func prepareDSNString(cfg *PostgresCfg) (string, error) {

	params := new(map[string]interface{})

	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:              "dsn",
		IgnoreUntaggedFields: true,
		Result:               params,
	})
	if err != nil {
		return "", err
	}

	if err := d.Decode(cfg); err != nil {
		return "", err
	}
	// //fix for pgbouncer https://github.com/jackc/pgx/issues/650?ysclid=lio5ajydwr669313171
	// (*params)["default_query_exec_mode"] = "simple_protocol"

	if _, ok := (*params)["default_query_exec_mode"]; !ok {
		(*params)["default_query_exec_mode"] = "simple_protocol"
	}

	b := new(bytes.Buffer)
	for k, v := range *params {
		fmt.Fprintf(b, "%s=%v ", k, v)
	}
	return b.String(), nil
}
