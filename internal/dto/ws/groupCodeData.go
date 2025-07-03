package wsdto

import (
	"sync"

	"github.com/gorilla/websocket"
)

type GroupCodeDataClient struct {
	Conn *websocket.Conn
	Sub  string // user id
}

type GroupCodeDataClientManagerMap struct {
	Map map[string][]*GroupCodeDataClient // groupCode -> client
	mu  sync.Mutex
}

func NewGroupCodeDataClientManagerMap() *GroupCodeDataClientManagerMap {
	return &GroupCodeDataClientManagerMap{
		Map: make(map[string][]*GroupCodeDataClient),
	}
}

func (m *GroupCodeDataClientManagerMap) Add(key string, client *GroupCodeDataClient) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Map[key] = append(m.Map[key], client)
}

func (m *GroupCodeDataClientManagerMap) Remove(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Map, key)
}

func (m *GroupCodeDataClientManagerMap) Get(key string) ([]*GroupCodeDataClient, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	client, ok := m.Map[key]
	return client, ok
}
