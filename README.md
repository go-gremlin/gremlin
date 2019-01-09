# Gremlin Server Client for Go

This library will allow you to connect to any graph database that supports TinkerPop3 using `Go`. This includes databases like JanusGraph and Neo4J. TinkerPop3 uses Gremlin Server to communicate with clients using either WebSockets or REST API. This library talks to Gremlin Server using WebSockets.

This library developed by the engineering department at [CB Insights](https://www.cbinsights.com/)


Installation
==========
```
go get github.com/cbinsights/gremlin
```

Usage
======

Gremlin Stack
---------------

Using the Gremlin Stack allows you to include a number of features on top of the Websocket that connects to your Gremlin server:
	1. Connection Pooling & maintenance to keep connections alive.
	2. Query re-trying.
	3. Optional tracing, logging and instrumentation features to your Gremlin queries.

The stack sits on top of a pool of connections to Gremlin, which wrap the Websocket connections defined by ["github.com/gorilla/websocket"]("github.com/gorilla/websocket")

### Creating the Gremlin Stack

You can instantiate it using NewGremlinStack or NewGremlinStackSimple (which excludes tracing, logging and instrumentation):
```go
	gremlinServer := "ws://server1:8182/gremlin"
	maxPoolCapacity := 10
	maxRetries := 3
	verboseLogging := true

	myStack, err := gremlin.NewGremlinStackSimple(gremlinServer, maxPoolCapacity, maxRetries, verboseLogging)
	if err != nil {
		// handle error here
	}
```

### Querying the database

To query the database, using your stack, call `ExecQueryF()`, which requires a Context.context object:
```go

	ctx, _ := context.WithCancel(context.Background())
	query := g.V("vertexID")
	response, err := myStack.ExecQueryF(ctx, query)
	if err != nil {
		// handle error here
	}
```

Optionally, you may parameterize the query and pass arguments into `ExecQueryF()`:
```go
	var args []interface{
		"vertexID",
		"10"
	}
	query := "g.V().hasLabel("%s").limit(%d)"
	response, err := myStack.ExecQueryF(ctx, query)
	if err != nil {
		// handle error here
	}

```
The gremlin client will handle the interpolation of the query, and escaping any characters that might be necessary.

### Reading the response

`ExecQueryF()` will return the string response returned by your Gremlin instance, which may look something like this:
```json
	[
	  {
	    "@type": "g:Vertex",
	    "@value": {
	      "id": "test-id",
	      "label": "label",
	      "properties": {
	        "health": [
	          {
	            "@type": "g:VertexProperty",
	            "@value": {
	              "id": {
	                "@type": "Type",
	                "@value": 1
	              },
	              "value": "1",
	              "label": "health"
	            }
	          }
	        ]
	      }
	    }
	  }
	]
```

This library provides a few helpers to serialize this response into Go structs. There are specific helpers for serializing a list of vertexes and for a list of edges, and there is additionally a generic vertex that serializes the response into a struct with a type & a value interface{}.

`SerializeVertexes()` would turn the response above into the following, which is considerably easier to manipulate:
```go
	Vertex{
		Type: "g:Vertex",
		Value: VertexValue{
			ID:    "test-id",
			Label: "label",
			Properties: map[string][]VertexProperty{
				"health": []VertexProperty{
					VertexProperty{
						Type: "g:VertexProperty",
						Value: VertexPropertyValue{
							ID: GenericValue{
								Type:  "Type",
								Value: 1,
							},
							Value: "1",
							Label: "health",
						},
					},
				},
			},
		},
	}
```

In addition, there are provided utilities to convert to "CleanVertexes" and "CleanEdges."

`ConvertToCleanVertexes()` converts a `Vertex` object as seen above into a much simpler `CleanVertex` with only an ID and a Label
`ConvertToCleanEdges()` converts an `Edge` object as seen above into a much simpler `CleanEdge` with only an ID and a Label


### Limitations

The Gremlin client forces some restraints on the characters allowed in a gremlin query to avoid issues with the query syntax. Currently the allowed characters are:

1. All alpha-numeric characters

2. All whitespace characters

3. The following punctuation: \, ;, ., :, /, -, ?, !, \*, (, ), &, \_, =, ,, #, ?, !, "


Go-Gremlin Usage Notes
---------
Note: This is the usage defined by the library from which this is forked ([github.com/go-gremlin/gremlin](github.com/go-gremlin/gremlin))
Export the list of databases you want to connect to as `GREMLIN_SERVERS` like so:-
```bash
export GREMLIN_SERVERS="ws://server1:8182/gremlin, ws://server2:8182/gremlin"
```

Import the library eg `import "github.com/go-gremlin/gremlin"`.

Parse and save your cluster of services. You only need to do this once before submitting any queries (Perhaps in `main()`):-
```go
	if err := gremlin.NewCluster(); err != nil {
		// handle error here
	}
```

Instead of using an environment variable, you can also pass the servers directly to `NewCluster()`. This is more convenient in development. For example:-
```go
	if err := gremlin.NewCluster("ws://dev.local:8182/gremlin", "ws://staging.local:8182/gremlin"); err != nil {
		// handle error
	}
```

To actually run queries against the database, make sure the package is imported and issue a gremlin query like this:-
```go
	data, err := gremlin.Query(`g.V()`).Exec()
	if err != nil  {
		// handle error
	}
```
`data` is a `JSON` array in bytes `[]byte` if any data is returned otherwise it is `nil`. For example you can print it using:-
```go
	fmt.Println(string(data))
```
or unmarshal it as desired.

You can also execute a query with bindings like this:-
```go
	data, err := gremlin.Query(`g.V().has("name", userName).valueMap()`).Bindings(gremlin.Bind{"userName": "john"}).Exec()
```

You can also execute a query with Session, Transaction and Aliases
```go
	aliases := make(map[string]string)
	aliases["g"] = fmt.Sprintf("%s.g", "demo-graph")
	session := "de131c80-84c0-417f-abdf-29ad781a7d04"  //use UUID generator
	data, err := gremlin.Query(`g.V().has("name", userName).valueMap()`).Bindings(gremlin.Bind{"userName": "john"}).Session(session).ManageTransaction(true).SetProcessor("session").Aliases(aliases).Exec()
```

Authentication
===
For authentication, you can set environment variables `GREMLIN_USER` and `GREMLIN_PASS` and create a `Client`, passing functional parameter `OptAuthEnv`

```go
	auth := gremlin.OptAuthEnv()
	myStack, err := gremlin.NewGremlinStackSimple("ws://server1:8182/gremlin", maxPoolCapacity, maxRetries, verboseLogging, auth)
	data, err = client.ExecQueryF(`g.V()`)
	if err != nil {
		panic(err)
	}
	doStuffWith(data)
```

If you don't like environment variables you can authenticate passing username and password string in the following way:
```go
	auth := gremlin.OptAuthUserPass("myusername", "mypass")
	myStack, err := gremlin.NewGremlinStackSimple("ws://server1:8182/gremlin", maxPoolCapacity, maxRetries, verboseLogging, auth)
	data, err = client.ExecQueryF(`g.V()`)
	if err != nil {
		panic(err)
	}
	doStuffWith(data)
```
