package gremlin

// cbi made up, not a real graphson or gremlin thing
type GremlinResponse struct {
	V Vertexes
	E Edges
}

type Vertexes []Vertex

type Vertex struct {
	Type  string      `json:"@type"`
	Value VertexValue `json:"@value"`
}

type VertexValue struct {
	ID         string                      `json:"id"`
	Label      string                      `json:"label"`
	Properties map[string][]VertexProperty `json:"properties"`
}

type VertexProperty struct {
	Type  string              `json:"@type"`
	Value VertexPropertyValue `json:"@value"`
}

type EdgeProperty struct {
	Type  string            `json:"@type"`
	Value EdgePropertyValue `json:"@value"`
}

type VertexPropertyValue struct {
	ID    GenericValue `json:"id"`
	Label string       `json:"label"`
	Value interface{}  `json:"value"`
}

type EdgePropertyValue struct {
	Label string      `json:"key"`
	Value interface{} `json:"value"`
}

type GenericValues []GenericValue

type GenericValue struct {
	Type  string      `json:"@type"`
	Value interface{} `json:"@value"`
}

type Edges []Edge

type Edge struct {
	Type  string    `json:"@type"`
	Value EdgeValue `json:"@value"`
}

type EdgeValue struct {
	ID         string // TODO: does this need to be a GenericValue? interface{}?
	Label      string
	InVLabel   string
	OutVLabel  string
	InV        string
	OutV       string
	Properties map[string]EdgeProperty
}

type CleanResponse struct {
	V []CleanVertex
	E []CleanEdge
}

type CleanEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type CleanVertex struct {
	Id    string `json:"id"`
	Label string `json:"label"`
}

// type TypeID int

// const (
// 	TypeString TypeID = iota
// 	TypeBoolean
// 	TypeMap
// 	TypeCollection
// 	TypeClass
// 	TypeDate
// 	TypeDouble
// 	TypeFloat
// 	TypeInteger
// 	TypeLong
// 	TypeTimestamp
// 	TypeUUID
// 	TypeVertex
// 	TypeVertexProperty
// )

// const (
// 	TypeStrDate           = "g:Date"
// 	TypeStrDouble         = "g:Double"
// 	TypeStrFloat          = "g:Float"
// 	TypeStrInteger        = "g:Int32"
// 	TypeStrLong           = "g:Int64"
// 	TypeStrTimestamp      = "g:Timestamp"
// 	TypeStrUUID           = "g:UUID"
// 	TypeStrVertex         = "g:Vertex"
// 	TypeStrVertexProperty = "g:VertexProperty"
// 	TypeStrProperty       = "g:Property"
// 	TypeStrEdge           = "g:Edge"
// )
