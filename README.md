# Gremlin Server Client for Go

This library will allow you to connect to any graph database that supports TinkerPop3 using `Go`. This includes databases like Titan and Neo4J. TinkerPop3 uses Gremlin Server to communicate with clients using either WebSockets or REST API. This library talks to Gremlin Server using WebSockets.


Installation
==========
```
go get github.com/go-gremlin/gremlin
```

Usage
======
Export the list of databases you want to connect to as `GREMLIN_SERVERS` like so:-
```bash
export GREMLIN_SERVERS="ws://server1:8182|http://server1:8182,ws://server2:8182|http://server2:8182"
```

Import the library in `main.go` eg `import "github.com/go-gremlin/gremlin"`.

At the top of your `main()` function of your program, include the following code to open a connection to your database and maintain it:-
```go
	// Create a connection
	if err := gremlin.Connect(""); err != nil {
		// handle error here
	}
	// Disconnect when main() exits
	defer gremlin.Disconnect()
	// Maintain the connection in the background
	go gremlin.MaintainConnection()
```
At this point, whenever you launch your program it will open a websocket connection to one of your databases. It will also actively check in the background if the connection is still up, falling back to other databases when a connection fails.

Instead of using an environment variable, you can also pass the connection string directly to `Connect()`. This is more convenient in development. For example:-
```go
 	// Create a connection
	if err := gremlin.Connect("ws://localhost:8182|http://localhost:8182"); err != nil {
		// handle error
	}
```

To actually run queries against the database, make sure the package is imported and issue a gremlin query like this:-
```go
	data, err := gremlin.Query("g.V()").Exec()
	if err != nil  {
		// handle error
	}
```
`data` is a `JSON` slice of bytes `[]byte`. For example you can print it using:-
```go
	fmt.Println(string(data))
```
or unmarshal it as desired.

You can also execute a query with bindings like this:-
```go
	data, err := gremlin.Query("g.V().has('name', userName).valueMap()").Bindings(gremlin.Bind{"userName": "john"}).Exec()
```
