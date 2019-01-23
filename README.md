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

4. Locking features designed to handle concurrent modification

The stack sits on top of a pool of connections to Gremlin, which wrap the Websocket connections defined by ["github.com/gorilla/websocket"]("github.com/gorilla/websocket")

### Creating the Gremlin Stack

You can instantiate it using NewGremlinStack or NewGremlinStackSimple (which excludes tracing, logging, instrumentation and the non-local locking mechanism):
```go
	gremlinServer := "ws://server1:8182/gremlin"
	maxPoolCapacity := 10
	maxRetries := 3
	verboseLogging := true
	pingInterval := 5

	myStack, err := gremlin.NewGremlinStackSimple(gremlinServer, maxPoolCapacity, maxRetries, verboseLogging, pingInterval)
	if err != nil {
		// handle error here
	}
```

Note:
The arguments for NewGremlinStack perform the following functions:

* maxPoolCapacity sets the number connections to keep in the pool

* maxRetries sets the number of retries to use on a connection before trying another

* verboseLogging denotes the verbosity of logs received from your gremlin server

* pingInterval denotes how often (in seconds) the gremlin pool should refresh itself
	* This keeps the Websocket connections alive; otherwise, if left inactive, they will close

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


### Locking

The Gremlin client is designed to accept a `LockClient` interface, with a client that implements a `LockKey(key string)` function to retrieve a `Lock` from the client.

The `Lock` interface implements 3 functions:

* `Lock()` which locks the corresponding key. If another `Lock` has locked the key, then the `Lock` will wait until the key is free.
* `Unlock()` which unlocks the corresponding key.
* `Destroy()` which removes the corresponding key from the client store.

In addition to the generic interface, two implementations have been provided, a local implementation and one for Consul.

#### Local Lock

The local lock uses a Map of Mutexes, so that for a given LocalKey passed in by the query, the Mutex will lock writes on that key. For instance, if we are operating on
a Vertex and expect simultaneous writes to that Vertex's properties, we can lock the Mutex corresponding to the Vertex ID to avoid a ConcurrentModificationException.

However, this system only works for a single client, so a distributed system, or a system with multiple writer clients will still be at risk of concurrency exceptions.


#### Consul Lock

The Consul lock uses the Consul API to prevent ConcurrentModificationExceptions. Similar to the Local Lock, we can pass in a Vertex ID to avoid a ConcurrentModificationException.
But, instead of writing to a local Map, it writes to Consul's KV configs, acquiring and releasing the corresponding configs as necessary, with a timeout to ensure that operations running too long, or operations that have been cancelled, do not retain their lock on the key.

For more information about Consul Distributed Key implementations, check the [Consul API docs](https://godoc.org/github.com/hashicorp/consul/api#Lock) and this [quick implementation guide](https://medium.com/@mthenw/distributed-locks-with-consul-and-golang-c4eccc217dd5)



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
