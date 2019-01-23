package gremlin

import (
	"context"
	"fmt"
	"time"
)

type Logger_i interface {
	Log(keyvals ...interface{}) error
}

type GremlinLogger struct {
	next   Gremlin_i
	logger Logger_i
}

func NewGremlinLogger(next Gremlin_i, logger Logger_i) GremlinLogger {
	return GremlinLogger{
		next:   next,
		logger: logger,
	}
}

func (g GremlinLogger) ExecQueryF(ctx context.Context, gremlinQuery GremlinQuery) (response string, err error) {
	method := CoalesceStrings(OpNameFromContext(ctx), "Gremlin.ExecQueryF")
	defer func(begin time.Time) {
		logQuery := fmt.Sprintf(gremlinQuery.Query, gremlinQuery.Args...)
		if err != nil {
			g.logger.Log("level", "error", "message", fmt.Sprintf("%s completed in %v with error: %v, with query: %s", method, time.Since(begin), err, logQuery))
		} else {
			g.logger.Log("level", "debug", "message", fmt.Sprintf("%s completed in %v with query: %s", method, time.Since(begin), logQuery))
		}
	}(time.Now())
	return g.next.ExecQueryF(ctx, gremlinQuery)
}

func (g GremlinLogger) StartMonitor(ctx context.Context, interval time.Duration) (err error) {
	method := CoalesceStrings(OpNameFromContext(ctx), "Gremlin.StartMonitor")
	defer func(begin time.Time) {
		if err != nil {
			g.logger.Log("level", "error", fmt.Sprintf("%s completed in %v with error: %v", method, time.Since(begin), err))
		} else {
			g.logger.Log("level", "debug", "message", fmt.Sprintf("%s completed in %v.", method, time.Since(begin)))
		}
	}(time.Now())
	return g.next.StartMonitor(ctx, interval)
}

func (g GremlinLogger) Close(ctx context.Context) (err error) {
	method := CoalesceStrings(OpNameFromContext(ctx), "Gremlin.Close")
	defer func(begin time.Time) {
		if err != nil {
			g.logger.Log("level", "error", fmt.Sprintf("%s completed in %v with error: %v", method, time.Since(begin), err))
		} else {
			g.logger.Log("level", "debug", "message", fmt.Sprintf("%s completed in %v.", method, time.Since(begin)))
		}
	}(time.Now())
	return g.next.Close(ctx)
}
