package gremlin

import (
	"golang.org/x/net/websocket"
	"strings"
	"os"
	"errors"
	"time"
)

var conn *websocket.Conn

var servers Servers

var userDisconnected bool

var CheckConnection time.Duration = 5

type Server struct {
	Ws string
	Origin string
}

type Servers []Server

//Connect("ws://localhost:8182|http://localhost:8182,ws://host.local:8182|http://host.local:8182")
//Connect("")
func Connect(connString string) (err error) {
	if strings.TrimSpace(connString) == "" {
		connString = os.Getenv("GREMLIN_SERVERS")
	}
	if connString == "" {
		err = errors.New("No servers set. Configure servers to connect to using the GREMLIN_SERVERS environment variable.")
		return
	}
	if servers, err = SplitServers(connString); err != nil {
		return
	}
	err = CreateConnection(servers)
	return
}

func SplitServers(connString string) (servers Servers, err error) {
	serverStrings := strings.Split(connString, ",")
	formatError := errors.New("Connection string is not in expected format. An example of the expected format is 'ws://localhost:8182|http://localhost:8182,ws://host.local:8182|http://host.local:8182'.")
	if len(serverStrings) < 1 {
		err = formatError
		return
	}
	for _, serverString := range serverStrings {
		serverArray := strings.Split(serverString, "|")
		if len(serverArray) != 2 {
			err = formatError
			return
		}
		server := Server{Ws: strings.TrimSpace(serverArray[0]), Origin: strings.TrimSpace(serverArray[1])}
		servers = append(servers, server)
		return
	}
	return
}

func CreateConnection(servers Servers) (err error) {
	// Disconnect from current connection if any
	Disconnect()
	userDisconnected = false
	// Create WebSocket config
	connEstablished := false
	for _, server := range servers {
		var config *websocket.Config
		var connErr error
		if config, connErr = websocket.NewConfig(server.Ws, server.Origin); connErr != nil {
			continue
		}
		// Set Mime Type
		config.Header.Set("Mime-Type", MimeType)
		// Connect to the database
		if conn, connErr = websocket.DialConfig(config); connErr != nil {
			continue
		}
		// Verify connection
		if connectionIsBad() {
			continue
		}
		connEstablished = true
		break
	}
	if !connEstablished {
		err = errors.New("Could not establish connection. Please check your connection string and ensure at least one server is up.")
	}
	return
}

func MaintainConnection() {
	if servers == nil {	// there is nothing to maintain
		return
	}
	for {
		if userDisconnected { // our job here is done
			return
		}
		time.Sleep(CheckConnection * time.Second)
		if connectionIsBad() {
			CreateConnection(servers)
		}
	}
}

func Disconnect() error {
	userDisconnected = true
	if conn == nil {
		return nil
	}
	return conn.Close()
}

func connectionIsBad() bool {
	if data, err := NewRequest("graph.features()").Exec().Json(); err != nil || data == nil {
		return true
	}
	return false
}
