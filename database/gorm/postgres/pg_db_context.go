package postgres

import (
	"fmt"
	"sync"

	database "github.com/loongkirin/gdk/database"
	"github.com/loongkirin/gdk/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/metrics"
	"gorm.io/plugin/opentelemetry/tracing"
)

type PostgresDbContext struct {
	DbConfig *database.DbConfig
	master   *gorm.DB
	slaves   []*gorm.DB
	lock     sync.RWMutex
	current  int
}

func NewPostgresDbContext(cfg *database.DbConfig) (*PostgresDbContext, error) {
	master, err := connectDB(cfg.Master)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master: %w", err)
	}

	tracingPlugin := tracing.NewPlugin(
		tracing.WithDBName(cfg.Master.DbName),
		tracing.WithRecordStackTrace(),
		tracing.WithQueryFormatter(func(query string) string {
			return query
		}),
		tracing.WithAttributes(
			attribute.String("db.type", "postgres"),
			attribute.String("db.host", cfg.Master.Host),
			attribute.Int("db.port", cfg.Master.Port),
		),
	)

	if err := master.Use(tracingPlugin); err != nil {
		return nil, fmt.Errorf("failed to enable tracing: %w", err)
	}

	sqlDB, err := master.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql db: %w", err)
	}

	opts := []metric.ObserveOption{
		metric.WithAttributes(
			attribute.String("db_name", cfg.Master.DbName),
			attribute.String("db_type", "postgres"),
			attribute.String("db_host", cfg.Master.Host),
			attribute.Int("db_port", cfg.Master.Port),
			attribute.String("db_role", "master"),
		),
	}

	masterOpts := append(opts, metric.WithAttributes(
		attribute.String("db_role", "master"),
	))
	metrics.ReportDBStatsMetrics(sqlDB, masterOpts...)

	var slaves []*gorm.DB
	for i, slaveCfg := range cfg.Slaves {
		slave, err := connectDB(slaveCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to slave: %w", err)
		}

		tracingPlugin := tracing.NewPlugin(
			tracing.WithDBName(slaveCfg.DbName),
			tracing.WithRecordStackTrace(),
			tracing.WithQueryFormatter(func(query string) string {
				return query
			}),
			tracing.WithAttributes(
				attribute.String("db_type", "postgres"),
				attribute.String("db_host", slaveCfg.Host),
				attribute.Int("db_port", slaveCfg.Port),
				attribute.String("db_role", fmt.Sprintf("slave_%d", i)),
			),
		)

		if err := slave.Use(tracingPlugin); err != nil {
			return nil, fmt.Errorf("failed to enable tracing: %w", err)
		}
		slaveOpts := append(opts, metric.WithAttributes(
			attribute.String("db_role", fmt.Sprintf("slave_%d", i)),
		))
		sqlDB, err := slave.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get sql db: %w", err)
		}
		metrics.ReportDBStatsMetrics(sqlDB, slaveOpts...)
		slaves = append(slaves, slave)
	}

	return &PostgresDbContext{
		DbConfig: cfg,
		master:   master,
		slaves:   slaves,
	}, nil
}

func connectDB(cfg database.DBConnection) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable %s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DbName, cfg.Config,
	)
	pgsqlconfig := gormpostgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: false,
	}
	db, err := gorm.Open(gormpostgres.New(pgsqlconfig), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	if duration, err := util.ParseDuration(cfg.ConnMaxLifetime); err == nil {
		sqlDB.SetConnMaxLifetime(duration)
	}
	return db, nil
}

func (db *PostgresDbContext) GetMasterDb() *gorm.DB {
	return db.master
}

func (db *PostgresDbContext) GetSlaveDb() *gorm.DB {
	if len(db.slaves) == 0 {
		return db.master
	}

	db.lock.Lock()
	defer db.lock.Unlock()

	db.current = (db.current + 1) % len(db.slaves)
	return db.slaves[db.current]
}

func (db *PostgresDbContext) HealthCheck() error {
	// 检查 master
	if err := db.master.Exec("SELECT 1").Error; err != nil {
		return fmt.Errorf("master health check failed: %w", err)
	}

	// 检查 slaves
	for i, slave := range db.slaves {
		if err := slave.Exec("SELECT 1").Error; err != nil {
			return fmt.Errorf("slave_%d health check failed: %w", i, err)
		}
	}
	return nil
}
