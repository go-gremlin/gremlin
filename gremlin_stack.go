package gremlin

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/cbinsights/gremlin/lock"
)

type Gremlin_i interface {
	ExecQueryF(ctx context.Context, gremlinQuery GremlinQuery) (response string, err error)
	StartMonitor(ctx context.Context, interval time.Duration) (err error)
	Close(ctx context.Context) (err error)
}

type GremlinStackOptions struct {
	MaxCap         int
	MaxRetries     int
	VerboseLogging bool
	PingInterval   int
	Logger         Logger_i
	Tracer         opentracing.Tracer
	Instr          InstrumentationProvider_i
	LockClient     lock.LockClient_i
}

func NewGremlinStackSimple(urlStr string, maxCap int, maxRetries int, verboseLogging bool, pingInterval int, options ...OptAuth) (Gremlin_i, error) {
	var (
		err error
		g   Gremlin_i
	)

	var lockClient lock.LockClient_i
	lockClient = lock.NewLocalLockClient(5, 10)
	g, err = NewGremlinClient(urlStr, maxCap, maxRetries, verboseLogging, lockClient, options...)
	if err != nil {
		return nil, err
	}
	_ = g.StartMonitor(context.Background(), time.Duration(pingInterval))
	return g, nil
}

func NewGremlinStack(urlStr string, gremlinStackOptions GremlinStackOptions, authOptions ...OptAuth) (Gremlin_i, error) {
	var (
		err error
		g   Gremlin_i
	)
	maxCap := DEFAULT_MAX_CAP
	if gremlinStackOptions.MaxCap != 0 {
		maxCap = gremlinStackOptions.MaxCap
	}
	maxRetries := DEFAULT_MAX_GREMLIN_RETRIES
	if gremlinStackOptions.MaxRetries != 0 {
		maxRetries = gremlinStackOptions.MaxRetries
	}
	verboseLogging := DEFAULT_VERBOSE_LOGGING
	if gremlinStackOptions.VerboseLogging != false {
		verboseLogging = gremlinStackOptions.VerboseLogging
	}
	pingInterval := DEFAULT_PING_INTERVAL
	if gremlinStackOptions.PingInterval != 0 {
		pingInterval = gremlinStackOptions.PingInterval
	}
	var lockClient lock.LockClient_i
	if gremlinStackOptions.LockClient != nil {
		lockClient = gremlinStackOptions.LockClient
	} else {
		lockClient = lock.NewLocalLockClient(5, 10)
	}

	g, err = NewGremlinClient(urlStr, maxCap, maxRetries, verboseLogging, lockClient, authOptions...)
	if err != nil {
		return nil, err
	}
	if gremlinStackOptions.Logger != nil {
		g = NewGremlinLogger(g, gremlinStackOptions.Logger)
	}
	if gremlinStackOptions.Tracer != nil {
		g = NewGremlinTracer(g, gremlinStackOptions.Tracer)
	}
	if gremlinStackOptions.Instr != nil {
		g = NewGremlinInstr(g, gremlinStackOptions.Instr)
	}

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
