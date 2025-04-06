package opentelemetry

import (
	database "github.com/loongkirin/gdk/database"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
)

func NewTracingPlugin(dbType, dbRole string, dbConnection database.DBConnection) gorm.Plugin {
	tracingPlugin := tracing.NewPlugin(
		tracing.WithDBName(dbConnection.DbName),
		tracing.WithRecordStackTrace(),
		tracing.WithQueryFormatter(func(query string) string {
			return query
		}),
		tracing.WithAttributes(
			attribute.String("db.type", dbType),
			attribute.String("db.role", dbRole),
			attribute.String("db.name", dbConnection.DbName),
			attribute.String("db.host", dbConnection.Host),
			attribute.Int("db.port", dbConnection.Port),
		),
	)
	return tracingPlugin
}

func NewMetricsObserverOptions(dbType, role string, dbConnection database.DBConnection) []metric.ObserveOption {
	return []metric.ObserveOption{
		metric.WithAttributes(
			attribute.String("db.type", dbType),
			attribute.String("db.role", role),
			attribute.String("db.name", dbConnection.DbName),
			attribute.String("db.host", dbConnection.Host),
			attribute.Int("db.port", dbConnection.Port),
		),
	}
}
