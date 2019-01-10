package gremlin

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
)

type Gremlin_i interface {
	ExecQueryF(ctx context.Context, query string, args ...interface{}) (response string, err error)
	StartMonitor(ctx context.Context, interval time.Duration) (err error)
	Close(ctx context.Context) (err error)
}

func NewGremlinStackSimple(urlStr string, maxCap int, maxRetries int, verboseLogging bool, pingInterval int, options ...OptAuth) (Gremlin_i, error) {
	var (
		err error
		g   Gremlin_i
	)
	g, err = NewGremlinClient(urlStr, maxCap, maxRetries, verboseLogging, options...)
	if err != nil {
		return nil, err
	}
	_ = g.StartMonitor(context.Background(), time.Duration(pingInterval))
	return g, nil
}

func NewGremlinStack(urlStr string, maxCap int, maxRetries int, verboseLogging bool, pingInterval int, logger Logger_i, tracer opentracing.Tracer, instr InstrumentationProvider_i, options ...OptAuth) (Gremlin_i, error) {
	var (
		err error
		g   Gremlin_i
	)
	g, err = NewGremlinClient(urlStr, maxCap, maxRetries, verboseLogging, options...)
	if err != nil {
		return nil, err
	}

	g = NewGremlinLogger(g, logger)
	g = NewGremlinTracer(g, tracer)
	g = NewGremlinInstr(g, instr)

	_ = g.StartMonitor(context.Background(), time.Duration(pingInterval))
	return g, nil
}

// HELPERS
type contextKey struct{}

var operationNameKey = contextKey{}

func ContextWithOpName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, operationNameKey, name)
}

func OpNameFromContext(ctx context.Context) string {
	val := ctx.Value(operationNameKey)
	name, _ := val.(string)
	return name
}
