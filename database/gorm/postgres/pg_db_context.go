package postgres

import (
	"fmt"
	"strings"
	"sync"
	"time"

	cfg "github.com/loongkirin/gdk/database"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresDbContext struct {
	DbConfig *cfg.DbConfig
	db       *gorm.DB
	lock     sync.Mutex
}

func NewPostgresDbContext(dbconfig *cfg.DbConfig) *PostgresDbContext {
	dbcontext := PostgresDbContext{DbConfig: dbconfig}
	return &dbcontext
}

func (dc *PostgresDbContext) DSN() string {
	var sb strings.Builder
	sb.WriteString("host=")
	sb.WriteString(dc.DbConfig.Host)

	sb.WriteString(" user=")
	sb.WriteString(dc.DbConfig.User)

	sb.WriteString(" password=")
	sb.WriteString(dc.DbConfig.Password)

	sb.WriteString(" dbname=")
	sb.WriteString(dc.DbConfig.DbName)

	sb.WriteString(" port=")
	sb.WriteString(dc.DbConfig.Port)

	sb.WriteString(" ")
	sb.WriteString(dc.DbConfig.Config)
	return sb.String()
}

func (dc *PostgresDbContext) GetDb() *gorm.DB {
	if dc.db != nil {
		return dc.db
	}
	dc.lock.Lock()
	defer dc.lock.Unlock()
	pgsqlconfig := gormpostgres.Config{
		DSN:                  dc.DSN(),
		PreferSimpleProtocol: false,
	}

	if db, err := gorm.Open(gormpostgres.New(pgsqlconfig), &gorm.Config{}); err != nil {
		fmt.Println("open db error")
		fmt.Println(err)
		return nil
	} else {
		fmt.Println("open db success")
		sqlDB, _ := db.DB()
		sqlDB.SetMaxIdleConns(dc.DbConfig.MaxIdleConns)
		sqlDB.SetMaxOpenConns(dc.DbConfig.MaxOpenConns)
		if duration, err := time.ParseDuration(dc.DbConfig.ConnMaxLifetime); err == nil {
			sqlDB.SetConnMaxLifetime(duration)
		}
		dc.db = db
		return db
	}
}
