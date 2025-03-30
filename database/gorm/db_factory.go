package gorm

import (
	database "github.com/loongkirin/gdk/database"
	gdkpostgres "github.com/loongkirin/gdk/database/gorm/postgres"
)

func CreateDbContext(cfg *database.DbConfig) DbContext {
	var dbcontext DbContext
	switch cfg.DbType {
	case "postgres":
		pgDbContext, err := gdkpostgres.NewPostgresDbContext(cfg)
		if err != nil {
			panic(err)
		}
		dbcontext = pgDbContext
	}

	return dbcontext
}
