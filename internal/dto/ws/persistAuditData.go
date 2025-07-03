package wsdto

import (
	"sync"

	"github.com/gorilla/websocket"
)

type PersistAuditDataClient struct {
	Conn *websocket.Conn
	Sub  string // user id
}

type PersistAuditDataManagerMap struct {
	Map map[string]*PersistAuditDataClient // sub -> client
	mu  sync.Mutex
}

func NewPersistAuditDataManagerMap() *PersistAuditDataManagerMap {
	return &PersistAuditDataManagerMap{
		Map: make(map[string]*PersistAuditDataClient),
	}
}

func (m *PersistAuditDataManagerMap) Add(key string, client *PersistAuditDataClient) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Map[key] = client
}

func (m *PersistAuditDataManagerMap) Remove(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Map, key)
}

func (m *PersistAuditDataManagerMap) Get(key string) (*PersistAuditDataClient, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	client, ok := m.Map[key]
	return client, ok
}
