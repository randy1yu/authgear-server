// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package analytic

import (
	"context"
	"github.com/authgear/authgear-server/pkg/lib/analytic"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/auditdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/analyticredis"
)

// Injectors from wire.go:

func NewUserWeeklyReport(ctx context.Context, pool *db.Pool, databaseCredentials *config.DatabaseCredentials) *analytic.UserWeeklyReport {
	databaseConfig := NewDatabaseConfig()
	databaseEnvironmentConfig := NewDatabaseEnvironmentConfig(databaseCredentials, databaseConfig)
	factory := NewLoggerFactory()
	handle := globaldb.NewHandle(ctx, pool, databaseEnvironmentConfig, factory)
	sqlBuilder := globaldb.NewSQLBuilder(databaseEnvironmentConfig)
	sqlExecutor := globaldb.NewSQLExecutor(ctx, handle)
	globalDBStore := &analytic.GlobalDBStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	appdbHandle := appdb.NewHandle(ctx, pool, databaseConfig, databaseCredentials, factory)
	appID := NewEmptyAppID()
	appdbSQLBuilder := appdb.NewSQLBuilder(databaseCredentials, appID)
	appdbSQLExecutor := appdb.NewSQLExecutor(ctx, appdbHandle)
	appDBStore := &analytic.AppDBStore{
		SQLBuilder:  appdbSQLBuilder,
		SQLExecutor: appdbSQLExecutor,
	}
	userWeeklyReport := &analytic.UserWeeklyReport{
		GlobalHandle:  handle,
		GlobalDBStore: globalDBStore,
		AppDBHandle:   appdbHandle,
		AppDBStore:    appDBStore,
	}
	return userWeeklyReport
}

func NewProjectWeeklyReport(ctx context.Context, pool *db.Pool, databaseCredentials *config.DatabaseCredentials, auditDatabaseCredentials *config.AuditDatabaseCredentials) *analytic.ProjectWeeklyReport {
	databaseConfig := NewDatabaseConfig()
	databaseEnvironmentConfig := NewDatabaseEnvironmentConfig(databaseCredentials, databaseConfig)
	factory := NewLoggerFactory()
	handle := globaldb.NewHandle(ctx, pool, databaseEnvironmentConfig, factory)
	sqlBuilder := globaldb.NewSQLBuilder(databaseEnvironmentConfig)
	sqlExecutor := globaldb.NewSQLExecutor(ctx, handle)
	globalDBStore := &analytic.GlobalDBStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	appdbHandle := appdb.NewHandle(ctx, pool, databaseConfig, databaseCredentials, factory)
	appID := NewEmptyAppID()
	appdbSQLBuilder := appdb.NewSQLBuilder(databaseCredentials, appID)
	appdbSQLExecutor := appdb.NewSQLExecutor(ctx, appdbHandle)
	appDBStore := &analytic.AppDBStore{
		SQLBuilder:  appdbSQLBuilder,
		SQLExecutor: appdbSQLExecutor,
	}
	readHandle := auditdb.NewReadHandle(ctx, pool, databaseConfig, auditDatabaseCredentials, factory)
	auditdbSQLBuilder := auditdb.NewSQLBuilder(auditDatabaseCredentials, appID)
	writeHandle := auditdb.NewWriteHandle(ctx, pool, databaseConfig, auditDatabaseCredentials, factory)
	writeSQLExecutor := auditdb.NewWriteSQLExecutor(ctx, writeHandle)
	auditDBStore := &analytic.AuditDBStore{
		SQLBuilder:  auditdbSQLBuilder,
		SQLExecutor: writeSQLExecutor,
	}
	projectWeeklyReport := &analytic.ProjectWeeklyReport{
		GlobalHandle:  handle,
		GlobalDBStore: globalDBStore,
		AppDBHandle:   appdbHandle,
		AppDBStore:    appDBStore,
		AuditDBHandle: readHandle,
		AuditDBStore:  auditDBStore,
	}
	return projectWeeklyReport
}

func NewCountCollector(ctx context.Context, pool *db.Pool, databaseCredentials *config.DatabaseCredentials, auditDatabaseCredentials *config.AuditDatabaseCredentials, redisPool *redis.Pool, credentials *config.AnalyticRedisCredentials) *analytic.CountCollector {
	databaseConfig := NewDatabaseConfig()
	databaseEnvironmentConfig := NewDatabaseEnvironmentConfig(databaseCredentials, databaseConfig)
	factory := NewLoggerFactory()
	handle := globaldb.NewHandle(ctx, pool, databaseEnvironmentConfig, factory)
	sqlBuilder := globaldb.NewSQLBuilder(databaseEnvironmentConfig)
	sqlExecutor := globaldb.NewSQLExecutor(ctx, handle)
	globalDBStore := &analytic.GlobalDBStore{
		SQLBuilder:  sqlBuilder,
		SQLExecutor: sqlExecutor,
	}
	appdbHandle := appdb.NewHandle(ctx, pool, databaseConfig, databaseCredentials, factory)
	appID := NewEmptyAppID()
	appdbSQLBuilder := appdb.NewSQLBuilder(databaseCredentials, appID)
	appdbSQLExecutor := appdb.NewSQLExecutor(ctx, appdbHandle)
	appDBStore := &analytic.AppDBStore{
		SQLBuilder:  appdbSQLBuilder,
		SQLExecutor: appdbSQLExecutor,
	}
	writeHandle := auditdb.NewWriteHandle(ctx, pool, databaseConfig, auditDatabaseCredentials, factory)
	auditdbSQLBuilder := auditdb.NewSQLBuilder(auditDatabaseCredentials, appID)
	writeSQLExecutor := auditdb.NewWriteSQLExecutor(ctx, writeHandle)
	auditDBStore := &analytic.AuditDBStore{
		SQLBuilder:  auditdbSQLBuilder,
		SQLExecutor: writeSQLExecutor,
	}
	redisConfig := NewRedisConfig()
	analyticredisHandle := analyticredis.NewHandle(redisPool, redisConfig, credentials, factory)
	readStoreRedis := &analytic.ReadStoreRedis{
		Context: ctx,
		Redis:   analyticredisHandle,
	}
	countCollector := &analytic.CountCollector{
		GlobalHandle:  handle,
		GlobalDBStore: globalDBStore,
		AppDBHandle:   appdbHandle,
		AppDBStore:    appDBStore,
		AuditDBHandle: writeHandle,
		AuditDBStore:  auditDBStore,
		CounterStore:  readStoreRedis,
	}
	return countCollector
}

// wire.go:

func NewEmptyAppID() config.AppID {

	return config.AppID("")
}
