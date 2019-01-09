package gremlin

import (
	"encoding/json"
)

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
