package gremlin

import (
	"fmt"
	"testing"
)

func makeDummyVertexProperty(label string, value interface{}) VertexProperty {
	return VertexProperty{
		Type: "g:VertexProperty",
		Value: VertexPropertyValue{
			ID: GenericValue{
				Type:  "Type",
				Value: 1,
			},
			Value: value,
			Label: label,
		},
	}
}

func makeDummyVertex(vertexID, vertexLabel string, params map[string]interface{}) Vertex {
	properties := make(map[string][]VertexProperty)
	for label, value := range params {
		var vp []VertexProperty
		vSlice, err := value.([]interface{})
		if err {
			for _, p := range vSlice {
				vertexProperty := makeDummyVertexProperty(label, p)
				vp = append(vp, vertexProperty)
			}
		} else {
			vertexProperty := makeDummyVertexProperty(label, value)
			vp = append(vp, vertexProperty)
		}
		properties[label] = vp
	}
	vertexValue := VertexValue{
		ID:         vertexID,
		Label:      vertexLabel,
		Properties: properties,
	}
	return Vertex{
		Type:  "g:Vertex",
		Value: vertexValue,
	}
}

func makeDummyProperty(label string, value interface{}) EdgeProperty {
	return EdgeProperty{
		Type: "g:Property",
		Value: EdgePropertyValue{
			Value: value,
			Label: label,
		},
	}
}

func makeDummyEdge(edgeID, edgeLabel, inVLabel, outVLabel, inV, outV string, params map[string]interface{}) Edge {
	properties := make(map[string]EdgeProperty)
	for label, value := range params {
		properties[label] = makeDummyProperty(label, value)
	}
	edgeValue := EdgeValue{
		ID:         edgeID,
		Label:      edgeLabel,
		InVLabel:   inVLabel,
		OutVLabel:  outVLabel,
		InV:        inV,
		OutV:       outV,
		Properties: properties,
	}
	return Edge{
		Type:  "g:Edge",
		Value: edgeValue,
	}
}

func edgesMatch(edge1, edge2 Edge) bool {
	if edge1.Type != edge2.Type {
		return false
	}
	if edge1.Value.ID != edge2.Value.ID {
		return false
	}
	if edge1.Value.Label != edge2.Value.Label {
		return false
	}
	if edge1.Value.InV != edge2.Value.InV || edge1.Value.InVLabel != edge2.Value.InVLabel {
		return false
	}
	if edge1.Value.OutV != edge2.Value.OutV || edge1.Value.OutVLabel != edge2.Value.OutVLabel {
		return false
	}
	edge1PropsString := fmt.Sprintf("%v", edge1.Value.Properties)
	edge2PropsString := fmt.Sprintf("%v", edge2.Value.Properties)
	return edge1PropsString == edge2PropsString
}

func vertexesMatch(vertex1, vertex2 Vertex) bool {
	if vertex1.Type != vertex2.Type {
		return false
	}
	if vertex1.Value.ID != vertex2.Value.ID {
		return false
	}
	if vertex1.Value.Label != vertex2.Value.Label {
		return false
	}
	vertex1PropsString := fmt.Sprintf("%v", vertex1.Value.Properties)
	vertex2PropsString := fmt.Sprintf("%v", vertex2.Value.Properties)
	return vertex1PropsString == vertex2PropsString
}

func TestSerializeVertexes(t *testing.T) {
	givens := []string{
		// test empty response
		`[]`,
		// test single vertex, single property
		`[{"@type":"g:Vertex","@value":{"id":"test-id","label":"label","properties":{"health":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"Type","@value": 1},"value":"1","label":"health"}}]}}}]`,
		// test two vertexes, single property
		`[{"@type":"g:Vertex","@value":{"id":"test-id","label":"label","properties":{"health":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"Type","@value": 1},"value":"1","label":"health"}}]}}}, {"@type":"g:Vertex","@value":{"id":"test-id2","label":"label","properties":{"health":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"Type","@value": 1},"value":"1","label":"health"}}]}}}]`,
		// test single vertex, two properties
		`[{"@type":"g:Vertex","@value":{"id":"test-id","label":"label","properties":{"health":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"Type","@value": 1},"value":"1","label":"health"}}], "health2":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"Type","@value": 1},"value":"2","label":"health2"}}]}}}]`,
		// test single vertex, single property - but property has multiple values
		`[{"@type":"g:Vertex","@value":{"id":"test-id","label":"label","properties":{"health":[{"@type":"g:VertexProperty","@value":{"id":{"@type":"Type","@value": 1},"value":"1","label":"health"}}, {"@type":"g:VertexProperty","@value":{"id":{"@type":"Type","@value": 1},"value":"2","label":"health"}}]}}}]`,
	}
	expecteds := [][]Vertex{
		{},
		{makeDummyVertex("test-id", "label", map[string]interface{}{"health": 1})},
		{makeDummyVertex("test-id", "label", map[string]interface{}{"health": 1}), makeDummyVertex("test-id2", "label", map[string]interface{}{"health": 1})},
		{makeDummyVertex("test-id", "label", map[string]interface{}{"health": 1, "health2": 2})},
		{makeDummyVertex("test-id", "label", map[string]interface{}{"health": []interface{}{1, 2}})},
	}
	for i, given := range givens {
		expected := expecteds[i]
		result, err := SerializeVertexes(given)

		if err != nil || len(result) != len(expected) {
			t.Error("given", given, "expected", expected, "result", result, "err", err)
		}
		for j, resultVertex := range result {
			expectedVertex := expected[j]
			if !vertexesMatch(resultVertex, expectedVertex) {
				t.Error("given", given, "expected", expected, "result", result)
			}
		}
	}
}

func TestSerializeEdges(t *testing.T) {
	givens := []string{
		// test empty response
		`[]`,
		// test single edge, single property
		`[{"@type":"g:Edge","@value":{"id":"test-id","label":"label","inVLabel":"inVLabel","outVLabel":"outVLabel","inV":"inV","outV":"outV","properties":{"test":{"@type":"g:Property","@value":{"key":"test","value":"test"}}}}}]`,
		// test two edges, single property
		`[{"@type":"g:Edge","@value":{"id":"test-id","label":"label","inVLabel":"inVLabel","outVLabel":"outVLabel","inV":"inV","outV":"outV","properties":{"test":{"@type":"g:Property","@value":{"key":"test","value":"test"}}}}}, {"@type":"g:Edge","@value":{"id":"test-id2","label":"label","inVLabel":"inVLabel","outVLabel":"outVLabel","inV":"inV","outV":"outV","properties":{"test":{"@type":"g:Property","@value":{"key":"test","value":"test"}}}}}]`,
		// test single edge, multiple properties
		`[{"@type":"g:Edge","@value":{"id":"test-id","label":"label","inVLabel":"inVLabel","outVLabel":"outVLabel","inV":"inV","outV":"outV","properties":{"test":{"@type":"g:Property","@value":{"key":"test","value":"test"}}, "test2":{"@type":"g:Property","@value":{"key":"test2","value":1}}}}}]`,
	}
	expecteds := [][]Edge{
		{},
		{makeDummyEdge("test-id", "label", "inVLabel", "outVLabel", "inV", "outV", map[string]interface{}{"test": "test"})},
		{makeDummyEdge("test-id", "label", "inVLabel", "outVLabel", "inV", "outV", map[string]interface{}{"test": "test"}), makeDummyEdge("test-id2", "label", "inVLabel", "outVLabel", "inV", "outV", map[string]interface{}{"test": "test"})},
		{makeDummyEdge("test-id", "label", "inVLabel", "outVLabel", "inV", "outV", map[string]interface{}{"test": "test", "test2": 1})},
	}

	for i, given := range givens {
		expected := expecteds[i]
		result, err := SerializeEdges(given)

		if err != nil || len(result) != len(expected) {
			t.Error("given", given, "expected", expected, "result", result, "err", err)
		}

		for j, resultEdge := range result {
			expectedEdge := expected[j]
			expectedEdgeString := fmt.Sprintf("%v", expectedEdge)
			resultEdgeString := fmt.Sprintf("%v", resultEdge)
			if !edgesMatch(resultEdge, expectedEdge) {
				t.Error("given", given, "expected", expectedEdgeString, "result", resultEdgeString)
			}
		}
	}
}
