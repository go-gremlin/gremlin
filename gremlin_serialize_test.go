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
			expectedVertexString := fmt.Sprintf("%v", expectedVertex)
			resultVertexString := fmt.Sprintf("%v", resultVertex)
			if expectedVertexString != resultVertexString {
				t.Error("given", given, "expected", expected, "result", result)

			}
		}
	}
}
