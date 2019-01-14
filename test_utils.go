package gremlin

func MakeDummyVertexProperty(label string, value interface{}) VertexProperty {
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

func MakeDummyVertex(vertexID, vertexLabel string, params map[string]interface{}) Vertex {
	properties := make(map[string][]VertexProperty)
	for label, value := range params {
		var vp []VertexProperty
		vSlice, err := value.([]interface{})
		if err {
			for _, p := range vSlice {
				vertexProperty := MakeDummyVertexProperty(label, p)
				vp = append(vp, vertexProperty)
			}
		} else {
			vertexProperty := MakeDummyVertexProperty(label, value)
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

func MakeDummyProperty(label string, value interface{}) EdgeProperty {
	return EdgeProperty{
		Type: "g:Property",
		Value: EdgePropertyValue{
			Value: value,
			Label: label,
		},
	}
}

func MakeDummyEdge(edgeID, edgeLabel, inVLabel, outVLabel, inV, outV string, params map[string]interface{}) Edge {
	properties := make(map[string]EdgeProperty)
	for label, value := range params {
		properties[label] = MakeDummyProperty(label, value)
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

func MakeDummyGenericValue(gvType string, value interface{}) GenericValue {
	return GenericValue{
		Type:  gvType,
		Value: value,
	}
}
