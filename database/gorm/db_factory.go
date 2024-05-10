package gorm

import (
	cfg "github.com/loongkirin/gdk/database"
	pg "github.com/loongkirin/gdk/database/gorm/postgres"
)

func CreateDbContext(cfg cfg.DbConfig) DbContext {
	var dbcontext DbContext
	switch cfg.DbType {
	case "postgres":
		dbcontext = pg.NewPostgresDbContext(&cfg)
	}

	return dbcontext
}
