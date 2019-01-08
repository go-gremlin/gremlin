package gremlin

import (
	"context"
	"encoding/json"

	"github.com/opentracing/opentracing-go"
)

type Gremlin_i interface {
	ExecQueryF(ctx context.Context, query string, args ...interface{}) (response string, err error)
	PingDatabase(ctx context.Context) (err error)
}

func NewGremlinStackSimple(urlStr string, maxCap int, verboseLogging bool, options ...OptAuth) (Gremlin_i, error) {
	var (
		err error
		g   Gremlin_i
	)
	g, err = NewGremlinClient(urlStr, maxCap, verboseLogging, options...)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func NewGremlinStack(urlStr string, maxCap int, verboseLogging bool, logger Logger, tracer opentracing.Tracer, instr InstrumentationProvider_i, options ...OptAuth) (Gremlin_i, error) {
	var (
		err error
		g   Gremlin_i
	)
	g, err = NewGremlinClient(urlStr, maxCap, verboseLogging, options...)
	if err != nil {
		return nil, err
	}

	g = NewGremlinLogger(g, logger)
	g = NewGremlinTracer(g, tracer)
	g = NewGremlinInstr(g, instr)
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

func SerializeVertexes(rawResponse string) (Vertexes, error) {
	// TODO: empty strings for property values will cause invalid json
	// make so it can handle that case
	var response Vertexes
	if rawResponse == "" {
		return response, nil
	}
	err := json.Unmarshal([]byte(rawResponse), &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func SerializeGremlinCount(rawResponse string) ([]GremlinCount, error) {
	// TODO: empty strings for property values will cause invalid json
	// make so it can handle that case
	var response []GremlinCount
	err := json.Unmarshal([]byte(rawResponse), &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func SerializeEdges(rawResponse string) (Edges, error) {
	var response Edges
	if rawResponse == "" {
		return response, nil
	}
	err := json.Unmarshal([]byte(rawResponse), &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func ConvertToCleanVertexes(vertexes Vertexes) []CleanVertex {
	var responseVertexes []CleanVertex
	for _, vertex := range vertexes {
		responseVertexes = append(responseVertexes, CleanVertex{
			Id:    vertex.Value.ID,
			Label: vertex.Value.Label,
		})
	}
	return responseVertexes
}

func ConvertToCleanEdges(edges Edges) []CleanEdge {
	var responseEdges []CleanEdge
	for _, edge := range edges {
		responseEdges = append(responseEdges, CleanEdge{
			Source: edge.Value.InV,
			Target: edge.Value.OutV,
		})
	}
	return responseEdges
}
