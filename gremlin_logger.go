package gremlin

import (
	"context"
	"fmt"
	"time"
)

type GremlinLogger struct {
	next   Gremlin_i
	logger Logger
}

func NewGremlinLogger(next Gremlin_i, logger Logger) GremlinLogger {
	return GremlinLogger{
		next:   next,
		logger: logger,
	}
}

func (g GremlinLogger) ExecQueryF(ctx context.Context, query string, args ...interface{}) (response string, err error) {
	method := CoalesceStrings(OpNameFromContext(ctx), "Gremlin.ExecQueryF")
	defer func(begin time.Time) {
		if err != nil {
			g.logger.Error("message", fmt.Sprintf("%s completed in %v with error: %v, with query: %s", method, time.Since(begin), err, fmt.Sprintf(query, args...)))
		} else {
			g.logger.Debug("message", fmt.Sprintf("%s completed in %v with query: %s", method, time.Since(begin), fmt.Sprintf(query, args...)))
		}
	}(time.Now())
	return g.next.ExecQueryF(ctx, query, args...)
}

func (g GremlinLogger) PingDatabase(ctx context.Context) (err error) {
	method := CoalesceStrings(OpNameFromContext(ctx), "Gremlin.PingDatabase")
	defer func(begin time.Time) {
		if err != nil {
			g.logger.Error("message", fmt.Sprintf("%s completed in %v with error: %v", method, time.Since(begin), err))
		} else {
			g.logger.Debug("message", fmt.Sprintf("%s completed in %v.", method, time.Since(begin)))
		}
	}(time.Now())
	return g.next.PingDatabase(ctx)
}
