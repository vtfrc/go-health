package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type ConnectionManager struct {
	connections map[string]*websocket.Conn // Map of username to their WebSocket connection
	lock        sync.RWMutex               // Mutex to ensure safe concurrent access
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*websocket.Conn),
	}
}

func (manager *ConnectionManager) AddConnection(username string, conn *websocket.Conn) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	manager.connections[username] = conn
}

func (manager *ConnectionManager) RemoveConnection(username string) {
	manager.lock.Lock()
	defer manager.lock.Unlock()

	if conn, ok := manager.connections[username]; ok {
		conn.Close()
		delete(manager.connections, username)
	}
}

func (manager *ConnectionManager) GetConnection(username string) (*websocket.Conn, bool) {
	manager.lock.RLock()
	defer manager.lock.RUnlock()

	conn, ok := manager.connections[username]
	return conn, ok
}
