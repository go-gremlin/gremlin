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

func makeDummyGenericValue(gvType string, value interface{}) GenericValue {
	return GenericValue{
		Type:  gvType,
		Value: value,
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
	if len(edge1.Value.Properties) != len(edge2.Value.Properties) {
		return false
	}
	for label, edge1Props := range edge1.Value.Properties {
		edge2Props := edge2.Value.Properties[label]
		if edge1Props.Type != edge2Props.Type {
			return false
		}
		if edge1Props.Value.Label != edge2Props.Value.Label ||
			fmt.Sprintf("%v", edge1Props.Value.Label) != fmt.Sprintf("%v", edge2Props.Value.Label) {
			return false
		}
	}
	return true
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
	if len(vertex1.Value.Properties) != len(vertex2.Value.Properties) {
		return false
	}
	for label, vertex1Props := range vertex1.Value.Properties {
		vertex2Props := vertex2.Value.Properties[label]
		if len(vertex1Props) != len(vertex2Props) {
			return false
		}
		for i, vertex1PropsElement := range vertex1Props {
			vertex2PropsElement := vertex2Props[i]
			if vertex1PropsElement.Type != vertex2PropsElement.Type {
				return false
			}
			if vertex1PropsElement.Value.ID.Type != vertex2PropsElement.Value.ID.Type ||
				fmt.Sprintf("%v", vertex1PropsElement.Value.ID.Value) != fmt.Sprintf("%v", vertex2PropsElement.Value.ID.Value) {
				return false
			}
			if vertex1PropsElement.Value.Label != vertex2PropsElement.Value.Label {
				return false
			}
			if fmt.Sprintf("%v", vertex1PropsElement.Value.Value) != fmt.Sprintf("%v", vertex2PropsElement.Value.Value) {
				return false
			}
		}
	}
	return true
}

func genericValuesMatch(gv1, gv2 GenericValue) bool {
	if gv1.Type != gv2.Type {
		return false
	}
	gv1ValueString := fmt.Sprintf("%v", gv1.Value)
	gv2ValueString := fmt.Sprintf("%v", gv2.Value)
	return gv1ValueString == gv2ValueString
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
				t.Error("given", given, "expected", expectedVertex.Value.Properties, "result", resultVertex.Value.Properties)
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
			if !edgesMatch(resultEdge, expectedEdge) {
				expectedEdgeString := fmt.Sprintf("%v", expectedEdge)
				resultEdgeString := fmt.Sprintf("%v", resultEdge)
				t.Error("given", given, "expected", expectedEdgeString, "result", resultEdgeString)
			}
		}
	}
}

func TestSerializeGenericValue(t *testing.T) {
	givens := []string{
		// test empty response
		`[]`,
		// test single gv, core return type
		`[{"@type":"g:Edge", "@value": 1}]`,
		// test 2 gv, core return type
		`[{"@type":"g:Edge", "@value": 1}, {"@type":"g:Edge2", "@value": "test"}]`,
		// test single gv, map return type
		`[{"@type":"g:Edge", "@value": {"test": "test"}}]`,
		// test single gv, nested map return type
		`[{"@type":"g:Edge", "@value": {"test": {"test": "test"}}}]`,
	}
	expecteds := [][]GenericValue{
		{},
		{makeDummyGenericValue("g:Edge", 1)},
		{makeDummyGenericValue("g:Edge", 1), makeDummyGenericValue("g:Edge2", "test")},
		{makeDummyGenericValue("g:Edge", map[string]interface{}{"test": "test"})},
		{makeDummyGenericValue("g:Edge", map[string]interface{}{"test": map[string]interface{}{"test": "test"}})},
	}

	for i, given := range givens {
		expected := expecteds[i]
		result, err := SerializeGenericValues(given)

		if err != nil || len(result) != len(expected) {
			t.Error("given", given, "expected", expected, "result", result, "err", err)
		}

		for j, resultGenericValue := range result {
			expectedGenericValue := expected[j]
			if !genericValuesMatch(resultGenericValue, expectedGenericValue) {
				t.Error("given", given, "expected", expectedGenericValue, "result", resultGenericValue)
			}
		}
	}
}

func TestConvertToCleanVertexes(t *testing.T) {
	givens := [][]Vertex{
		{},
		{makeDummyVertex("test-id", "label", map[string]interface{}{"health": 1})},
		{makeDummyVertex("test-id", "label", map[string]interface{}{"health": 1}), makeDummyVertex("test-id2", "label", map[string]interface{}{"health": 1})},
	}
	expecteds := [][]CleanVertex{
		{},
		{CleanVertex{Id: "test-id", Label: "label"}},
		{CleanVertex{Id: "test-id", Label: "label"}, CleanVertex{Id: "test-id2", Label: "label"}},
	}

	for i, given := range givens {
		expected := expecteds[i]
		result := ConvertToCleanVertexes(given)

		if len(result) != len(expected) {
			t.Error("given", given, "expected", expected, "result", result)
		}

		for j, resultCleanVertex := range result {
			expectedCleanVertex := expected[j]
			if expectedCleanVertex.Id != resultCleanVertex.Id || expectedCleanVertex.Label != expectedCleanVertex.Label {
				t.Error("given", given, "expected", expected, "result", result)
			}
		}
	}
}

func TestConvertToCleanEdges(t *testing.T) {
	givens := [][]Edge{
		{},
		{makeDummyEdge("test-id", "label", "inVLabel", "outVLabel", "inV", "outV", map[string]interface{}{"test": "test"})},
		{makeDummyEdge("test-id", "label", "inVLabel", "outVLabel", "inV", "outV", map[string]interface{}{"test": "test"}), makeDummyEdge("test-id2", "label", "inVLabel", "outVLabel", "inV2", "outV2", map[string]interface{}{"test": "test"})},
	}
	expecteds := [][]CleanEdge{
		{},
		{CleanEdge{Source: "inV", Target: "outV"}},
		{CleanEdge{Source: "inV", Target: "outV"}, CleanEdge{Source: "inV2", Target: "outV2"}},
	}

	for i, given := range givens {
		expected := expecteds[i]
		result := ConvertToCleanEdges(given)

		if len(result) != len(expected) {
			t.Error("given", given, "expected", expected, "result", result)
		}

		for j, resultCleanEdges := range result {
			expectedCleanEdges := expected[j]
			if expectedCleanEdges.Source != resultCleanEdges.Source || expectedCleanEdges.Target != expectedCleanEdges.Target {
				t.Error("given", given, "expected", expected, "result", result)
			}
		}
	}
}
