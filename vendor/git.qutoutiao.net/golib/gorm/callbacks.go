package gorm

import (
	"errors"
	"fmt"
	"strings"
	"time"

	gormio "gorm.io/gorm"

	"git.qutoutiao.net/golib/gorm/metrics"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	issuedAtKey = "__issued_at"
)

const (
	GormSuceess = "200"
	GormFailed  = "400"
)

func registerTraceCallbacks(c *Client) {
	db, ok := c.value.Load().(*gormio.DB)
	if !ok {
		return
	}

	dbCallback := db.Callback()
	dbInstance := c.config.mycfg.Addr + "/" + c.config.mycfg.DBName
	traceIncludeNotFound := c.config.TraceIncludeNotFound

	// for creation
	createCB := dbCallback.Create()
	createCB.Before("gorm:before_create").Register("pedestal:trace_before_create", func(db *gormio.DB) {
		if db.Error != nil || db.Statement.Schema == nil || db.Statement.Context == nil || db.DryRun {
			return
		}

		span, ctx := opentracing.StartSpanFromContext(db.Statement.Context, "CREATE", opentracing.StartTime(time.Now()))

		ext.SpanKindRPCClient.Set(span)
		ext.Component.Set(span, ComponentName)

		ext.DBType.Set(span, "sql")
		ext.DBInstance.Set(span, db.Statement.Table)

		ext.PeerService.Set(span, "mysql")
		ext.PeerAddress.Set(span, dbInstance)

		db.Statement.Context = ctx
	})
	createCB.After("gorm:after_create").Register("pedestal:trace_after_create", func(db *gormio.DB) {
		if db.Statement.Context == nil || db.DryRun {
			return
		}

		span := opentracing.SpanFromContext(db.Statement.Context)
		if span == nil {
			return
		}
		defer span.Finish()

		if db.Error != nil && (traceIncludeNotFound || !errors.Is(db.Error, gormio.ErrRecordNotFound)) {
			ext.Error.Set(span, true)

			ext.DBStatement.Set(span, fmt.Sprintf("%s{%+v}", db.Statement.SQL.String(), db.Statement.Vars))

			span.LogKV("event", "error", "message", db.Error.Error())
		} else {
			ext.DBStatement.Set(span, db.Statement.SQL.String())
		}
	})

	// for update
	updateCB := dbCallback.Update()
	updateCB.Before("gorm:before_update").Register("pedestal:trace_before_update", func(db *gormio.DB) {
		if db.Error != nil || db.Statement.Schema == nil || db.Statement.Context == nil || db.DryRun {
			return
		}

		span, ctx := opentracing.StartSpanFromContext(db.Statement.Context, "UPDATE", opentracing.StartTime(time.Now()))

		ext.SpanKindRPCClient.Set(span)
		ext.Component.Set(span, ComponentName)

		ext.DBType.Set(span, "sql")
		ext.DBInstance.Set(span, db.Statement.Table)

		ext.PeerService.Set(span, "mysql")
		ext.PeerAddress.Set(span, dbInstance)

		db.Statement.Context = ctx
	})
	updateCB.After("gorm:after_update").Register("pedestal:trace_after_update", func(db *gormio.DB) {
		if db.Statement.Context == nil || db.DryRun {
			return
		}

		span := opentracing.SpanFromContext(db.Statement.Context)
		if span == nil {
			return
		}
		defer span.Finish()

		if db.Error != nil && (traceIncludeNotFound || !errors.Is(db.Error, gormio.ErrRecordNotFound)) {
			ext.Error.Set(span, true)

			ext.DBStatement.Set(span, fmt.Sprintf("%s{%+v}", db.Statement.SQL.String(), db.Statement.Vars))

			span.LogKV("event", "error", "message", db.Error.Error())
		} else {
			ext.DBStatement.Set(span, db.Statement.SQL.String())
		}
	})

	// for query
	queryCB := dbCallback.Query()
	queryCB.Before("gorm:query").Register("pedestal:trace_before_query", func(db *gormio.DB) {
		if db.Error != nil || db.Statement.Schema == nil || db.Statement.Context == nil || db.DryRun {
			return
		}

		span, ctx := opentracing.StartSpanFromContext(db.Statement.Context, "QUERY", opentracing.StartTime(time.Now()))

		ext.SpanKindRPCClient.Set(span)
		ext.Component.Set(span, ComponentName)

		ext.DBType.Set(span, "sql")
		ext.DBInstance.Set(span, db.Statement.Table)

		ext.PeerService.Set(span, "mysql")
		ext.PeerAddress.Set(span, dbInstance)

		db.Statement.Context = ctx
	})
	queryCB.After("gorm:after_query").Register("pedestal:trace_after_query", func(db *gormio.DB) {
		if db.Statement.Context == nil || db.DryRun {
			return
		}

		span := opentracing.SpanFromContext(db.Statement.Context)
		if span == nil {
			return
		}
		defer span.Finish()

		if db.Error != nil && (traceIncludeNotFound || !errors.Is(db.Error, gormio.ErrRecordNotFound)) {
			ext.Error.Set(span, true)

			ext.DBStatement.Set(span, fmt.Sprintf("%s{%+v}", db.Statement.SQL.String(), db.Statement.Vars))

			span.LogKV("event", "error", "message", db.Error.Error())
		} else {
			ext.DBStatement.Set(span, db.Statement.SQL.String())
		}
	})

	// for delete
	deleteCB := dbCallback.Delete()
	deleteCB.Before("gorm:before_delete").Register("pedestal:trace_before_delete", func(db *gormio.DB) {
		if db.Error != nil || db.Statement.Schema == nil || db.Statement.Context == nil || db.DryRun {
			return
		}

		span, ctx := opentracing.StartSpanFromContext(db.Statement.Context, "DELETE", opentracing.StartTime(time.Now()))

		ext.SpanKindRPCClient.Set(span)
		ext.Component.Set(span, "gorm")

		ext.DBType.Set(span, "sql")
		ext.DBInstance.Set(span, db.Statement.Table)

		ext.PeerService.Set(span, "mysql")
		ext.PeerAddress.Set(span, dbInstance)

		db.Statement.Context = ctx
	})
	deleteCB.After("gorm:after_delete").Register("pedestal:trace_after_delete", func(db *gormio.DB) {
		if db.Statement.Context == nil || db.DryRun {
			return
		}

		span := opentracing.SpanFromContext(db.Statement.Context)
		if span == nil {
			return
		}
		defer span.Finish()

		if db.Error != nil && (traceIncludeNotFound || !errors.Is(db.Error, gormio.ErrRecordNotFound)) {
			ext.Error.Set(span, true)

			ext.DBStatement.Set(span, fmt.Sprintf("%s{%+v}", db.Statement.SQL.String(), db.Statement.Vars))

			span.LogKV("event", "error", "message", db.Error.Error())
		} else {
			ext.DBStatement.Set(span, db.Statement.SQL.String())
		}
	})

	// for raw
	rawCB := dbCallback.Raw()
	rawCB.Before("gorm:raw").Register("pedestal:trace_before_sql", func(db *gormio.DB) {
		if db.Error != nil || db.Statement.Schema == nil || db.Statement.Context == nil || db.DryRun {
			return
		}

		sql := strings.TrimSpace(db.Statement.SQL.String())
		name := "exec:" + strings.ToUpper(strings.SplitN(sql, " ", 2)[0])

		span, ctx := opentracing.StartSpanFromContext(db.Statement.Context, name, opentracing.StartTime(time.Now()))

		ext.SpanKindRPCClient.Set(span)
		ext.Component.Set(span, ComponentName)

		ext.DBType.Set(span, "sql")
		ext.DBInstance.Set(span, db.Statement.Table)

		ext.PeerService.Set(span, "mysql")
		ext.PeerAddress.Set(span, dbInstance)

		db.Statement.Context = ctx
	})
	rawCB.After("gorm:row").Register("pedestal:trace_after_sql", func(db *gormio.DB) {
		if db.Statement.Context == nil || db.DryRun {
			return
		}

		span := opentracing.SpanFromContext(db.Statement.Context)
		if span == nil {
			return
		}
		defer span.Finish()

		if db.Error != nil && (traceIncludeNotFound || !errors.Is(db.Error, gormio.ErrRecordNotFound)) {
			ext.Error.Set(span, true)

			ext.DBStatement.Set(span, fmt.Sprintf("%s{%+v}", db.Statement.SQL.String(), db.Statement.Vars))

			span.LogKV("event", "error", "message", db.Error.Error())
		} else {
			ext.DBStatement.Set(span, db.Statement.SQL.String())
		}
	})
}

func registerMetricsCallbacks(c *Client) {
	db, ok := c.value.Load().(*gormio.DB)
	if !ok {
		return
	}

	dbCallback := db.Callback()
	dbDriver := c.config.Driver
	dbInstance := c.config.mycfg.Addr + "/" + c.config.mycfg.DBName
	metricsIncludeNotFound := c.config.MetricsIncludeNotFound

	// for creation
	createCB := dbCallback.Create()
	createCB.Before("gorm:before_create").Register("pedestal:metrics_before_create", func(db *gormio.DB) {
		if db.Error != nil || db.DryRun {
			return
		}

		db.InstanceSet(issuedAtKey, time.Now())
	})
	createCB.After("gorm:after_create").Register("pedestal:metrics_after_create", func(db *gormio.DB) {
		if db.DryRun {
			return
		}

		labels := prometheus.Labels{
			"client": dbDriver,
			"cmd":    "create",
			"to":     dbInstance,
			"status": GormSuceess,
		}
		if db.Error != nil && (metricsIncludeNotFound || !errors.Is(db.Error, gormio.ErrRecordNotFound)) {
			labels["status"] = GormFailed
		}

		if iface, ok := db.InstanceGet(issuedAtKey); ok {
			if issuedAt, ok := iface.(time.Time); ok {
				metrics.ObserveOp(labels, issuedAt)
			}
		}
		metrics.IncOp(labels)
	})

	// for update
	updateCB := dbCallback.Update()
	updateCB.Before("gorm:before_update").Register("pedestal:metrics_before_update", func(db *gormio.DB) {
		if db.Error != nil || db.DryRun {
			return
		}

		db.InstanceSet(issuedAtKey, time.Now())
	})
	updateCB.After("gorm:after_update").Register("pedestal:metrics_after_update", func(db *gormio.DB) {
		if db.DryRun {
			return
		}

		labels := prometheus.Labels{
			"client": dbDriver,
			"cmd":    "update",
			"to":     dbInstance,
			"status": GormSuceess,
		}
		if db.Error != nil && (metricsIncludeNotFound || !errors.Is(db.Error, gormio.ErrRecordNotFound)) {
			labels["status"] = GormFailed
		}

		if iface, ok := db.InstanceGet(issuedAtKey); ok {
			if issuedAt, ok := iface.(time.Time); ok {
				metrics.ObserveOp(labels, issuedAt)
			}
		}
		metrics.IncOp(labels)
	})

	// for query
	queryCB := dbCallback.Query()
	queryCB.Before("gorm:query").Register("pedestal:metrics_before_query", func(db *gormio.DB) {
		if db.Error != nil || db.DryRun {
			return
		}

		db.InstanceSet(issuedAtKey, time.Now())
	})
	queryCB.After("gorm:after_query").Register("pedestal:metrics_after_query", func(db *gormio.DB) {
		if db.DryRun {
			return
		}

		labels := prometheus.Labels{
			"client": dbDriver,
			"cmd":    "query",
			"to":     dbInstance,
			"status": GormSuceess,
		}
		if db.Error != nil && (metricsIncludeNotFound || !errors.Is(db.Error, gormio.ErrRecordNotFound)) {
			labels["status"] = GormFailed
		}

		if iface, ok := db.InstanceGet(issuedAtKey); ok {
			if issuedAt, ok := iface.(time.Time); ok {
				metrics.ObserveOp(labels, issuedAt)
			}
		}
		metrics.IncOp(labels)
	})

	// for delete
	deleteCB := dbCallback.Delete()
	deleteCB.Before("gorm:before_delete").Register("pedestal:metrics_before_delete", func(db *gormio.DB) {
		if db.Error != nil || db.DryRun {
			return
		}

		db.InstanceSet(issuedAtKey, time.Now())
	})
	deleteCB.After("gorm:after_delete").Register("pedestal:metrics_after_delete", func(db *gormio.DB) {
		if db.DryRun {
			return
		}

		labels := prometheus.Labels{
			"client": dbDriver,
			"cmd":    "delete",
			"to":     dbInstance,
			"status": GormSuceess,
		}
		if db.Error != nil && (metricsIncludeNotFound || !errors.Is(db.Error, gormio.ErrRecordNotFound)) {
			labels["status"] = GormFailed
		}

		if iface, ok := db.InstanceGet(issuedAtKey); ok {
			if issuedAt, ok := iface.(time.Time); ok {
				metrics.ObserveOp(labels, issuedAt)
			}
		}
		metrics.IncOp(labels)
	})

	// for raw
	rawCB := dbCallback.Raw()
	rawCB.Before("gorm:raw").Register("pedestal:metrics_before_sql", func(db *gormio.DB) {
		if db.Error != nil || db.DryRun {
			return
		}

		db.InstanceSet(issuedAtKey, time.Now())
	})
	rawCB.After("gorm:row").Register("pedestal:metrics_after_sql", func(db *gormio.DB) {
		if db.DryRun {
			return
		}

		sql := strings.TrimSpace(db.Statement.SQL.String())
		name := "exec:" + strings.ToLower(strings.SplitN(sql, " ", 2)[0])

		labels := prometheus.Labels{
			"client": dbDriver,
			"cmd":    name,
			"to":     dbInstance,
			"status": GormSuceess,
		}
		if db.Error != nil && (metricsIncludeNotFound || !errors.Is(db.Error, gormio.ErrRecordNotFound)) {
			labels["status"] = GormFailed
		}

		if iface, ok := db.InstanceGet(issuedAtKey); ok {
			if issuedAt, ok := iface.(time.Time); ok {
				metrics.ObserveOp(labels, issuedAt)
			}
		}
		metrics.IncOp(labels)
	})
}
