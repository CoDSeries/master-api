package lobby

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type Client struct {
	conn  *websocket.Conn
	lobby *Lobby
	send  chan []byte // Canal pour envoyer des messages
}

type Lobby struct {
	ID         string           `json:"id"`
	clients    map[*Client]bool `json:"-"` // Le "-" signifie que ce champ ne sera pas sérialisé en JSON
	register   chan *Client     `json:"-"`
	unregister chan *Client     `json:"-"`
	broadcast  chan []byte      `json:"-"`
	mutex      sync.Mutex       `json:"-"`
}

type LobbyInfo struct {
	ID      string `json:"id"`
	Clients int    `json:"clients"`
}

func (manager *LobbyManager) ListLobbies() []LobbyInfo {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	var lobbies []LobbyInfo
	for id, lobby := range manager.lobbies {
		lobbies = append(lobbies, LobbyInfo{
			ID:      id,
			Clients: len(lobby.clients),
		})
	}

	return lobbies
}

func NewLobby(id string) *Lobby {
	return &Lobby{
		ID:         id,
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

func (l *Lobby) run() {
	for {
		select {
		case client := <-l.register:
			l.mutex.Lock()
			l.clients[client] = true
			l.mutex.Unlock()
		case client := <-l.unregister:
			l.mutex.Lock()
			if _, ok := l.clients[client]; ok {
				delete(l.clients, client)
				close(client.send)
			}
			l.mutex.Unlock()
		case message := <-l.broadcast:
			l.mutex.Lock()
			for client := range l.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(l.clients, client)
				}
			}
			l.mutex.Unlock()
		}
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request, lobby *Lobby) {
	ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		fmt.Println(err)
		return
	}
	client := &Client{conn: ws, lobby: lobby, send: make(chan []byte, 256)}
	lobby.register <- client

	defer func() { lobby.unregister <- client }()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		lobby.broadcast <- message
	}
}

type LobbyManager struct {
	lobbies map[string]*Lobby
	mutex   sync.RWMutex
}

func NewLobbyManager() *LobbyManager {
	return &LobbyManager{
		lobbies: make(map[string]*Lobby),
	}
}

func (manager *LobbyManager) CreateLobby(id string) *Lobby {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	lobby, exists := manager.lobbies[id]
	if !exists {
		lobby = NewLobby(id)
		manager.lobbies[id] = lobby
		go lobby.run()
	}
	return lobby
}

func (manager *LobbyManager) GetLobby(id string) (*Lobby, bool) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	lobby, exists := manager.lobbies[id]
	return lobby, exists
}

func (manager *LobbyManager) DeleteLobby(id string) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	if lobby, exists := manager.lobbies[id]; exists {
		close(lobby.register)
		close(lobby.unregister)
		close(lobby.broadcast)
		delete(manager.lobbies, id)
	}
}
