package gremlin

import (
	"context"
	"fmt"
	"time"
)

type InstrumentationProvider_i interface {
	Incr(name string, tags []string, rate float64) error
}

func EmptyTags() []string {
	return emptyTags
}

var emptyTags []string

type GremlinInstr struct {
	next  Gremlin_i
	instr InstrumentationProvider_i
}

func NewGremlinInstr(next Gremlin_i, instr InstrumentationProvider_i) GremlinInstr {
	return GremlinInstr{
		next:  next,
		instr: instr,
	}
}

func (g GremlinInstr) ExecQueryF(ctx context.Context, gremlinQuery GremlinQuery) (response string, err error) {
	method := CoalesceStrings(OpNameFromContext(ctx), "Gremlin.ExecQueryF")
	defer func() {
		g.instr.Incr(method, EmptyTags(), 1)
		if err != nil {
			g.instr.Incr(fmt.Sprintf("%s.Error", method), EmptyTags(), 1)
		}
	}()
	return g.next.ExecQueryF(ctx, gremlinQuery)
}

func (g GremlinInstr) StartMonitor(ctx context.Context, interval time.Duration) (err error) {
	method := CoalesceStrings(OpNameFromContext(ctx), "Gremlin.StartMonitor")
	defer func() {
		g.instr.Incr(method, EmptyTags(), 1)
		if err != nil {
			g.instr.Incr(fmt.Sprintf("%s.Error", method), EmptyTags(), 1)
		}
	}()
	return g.next.StartMonitor(ctx, interval)
}

func (g GremlinInstr) Close(ctx context.Context) (err error) {
	method := CoalesceStrings(OpNameFromContext(ctx), "Gremlin.Close")
	defer func() {
		g.instr.Incr(method, EmptyTags(), 1)
		if err != nil {
			g.instr.Incr(fmt.Sprintf("%s.Error", method), EmptyTags(), 1)
		}
	}()
	return g.next.Close(ctx)
}
