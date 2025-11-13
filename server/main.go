package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message types
const (
	MsgTypeConnect    = "connect"
	MsgTypeDisconnect = "disconnect"
	MsgTypeMove       = "move"
	MsgTypeWorldState = "world_state"
)

// Vector3 represents a 3D position
type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// Player represents a connected player
type Player struct {
	ID       string   `json:"id"`
	Position Vector3  `json:"position"`
	Rotation Vector3  `json:"rotation"`
	LastSeen time.Time `json:"-"`
}

// Message represents a network message
type Message struct {
	Type     string                 `json:"type"`
	PlayerID string                 `json:"player_id,omitempty"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// Client represents a connected WebSocket client
type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	Server *GameServer
}

// GameServer manages all game state and connections
type GameServer struct {
	clients    map[string]*Client
	players    map[string]*Player
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// NewGameServer creates a new game server instance
func NewGameServer() *GameServer {
	return &GameServer{
		clients:    make(map[string]*Client),
		players:    make(map[string]*Player),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
	}
}

// Run starts the main game server loop
func (gs *GameServer) Run() {
	ticker := time.NewTicker(50 * time.Millisecond) // 20 ticks per second
	defer ticker.Stop()

	for {
		select {
		case client := <-gs.register:
			gs.handleRegister(client)

		case client := <-gs.unregister:
			gs.handleUnregister(client)

		case message := <-gs.broadcast:
			gs.handleBroadcast(message)

		case <-ticker.C:
			gs.sendWorldState()
		}
	}
}

func (gs *GameServer) handleRegister(client *Client) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.clients[client.ID] = client
	gs.players[client.ID] = &Player{
		ID:       client.ID,
		Position: Vector3{X: 0, Y: 1, Z: 0},
		Rotation: Vector3{X: 0, Y: 0, Z: 0},
		LastSeen: time.Now(),
	}

	log.Printf("Client registered: %s (Total: %d)", client.ID, len(gs.clients))

	// Send welcome message
	msg := Message{
		Type:     MsgTypeConnect,
		PlayerID: client.ID,
		Data: map[string]interface{}{
			"your_id": client.ID,
		},
	}
	data, _ := json.Marshal(msg)
	client.Send <- data
}

func (gs *GameServer) handleUnregister(client *Client) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if _, ok := gs.clients[client.ID]; ok {
		delete(gs.clients, client.ID)
		delete(gs.players, client.ID)
		close(client.Send)
		log.Printf("Client unregistered: %s (Total: %d)", client.ID, len(gs.clients))
	}
}

func (gs *GameServer) handleBroadcast(message []byte) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	for _, client := range gs.clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(gs.clients, client.ID)
		}
	}
}

func (gs *GameServer) sendWorldState() {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if len(gs.players) == 0 {
		return
	}

	// Build world state
	playerStates := make([]Player, 0, len(gs.players))
	for _, player := range gs.players {
		playerStates = append(playerStates, *player)
	}

	msg := Message{
		Type: MsgTypeWorldState,
		Data: map[string]interface{}{
			"players": playerStates,
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling world state: %v", err)
		return
	}

	// Send to all clients
	for _, client := range gs.clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

func (gs *GameServer) handlePlayerMessage(client *Client, messageData []byte) {
	var msg Message
	if err := json.Unmarshal(messageData, &msg); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return
	}

	switch msg.Type {
	case MsgTypeMove:
		gs.handlePlayerMove(client, msg)
	}
}

func (gs *GameServer) handlePlayerMove(client *Client, msg Message) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	player, ok := gs.players[client.ID]
	if !ok {
		return
	}

	// Update player position
	if pos, ok := msg.Data["position"].(map[string]interface{}); ok {
		if x, ok := pos["x"].(float64); ok {
			player.Position.X = x
		}
		if y, ok := pos["y"].(float64); ok {
			player.Position.Y = y
		}
		if z, ok := pos["z"].(float64); ok {
			player.Position.Z = z
		}
	}

	// Update player rotation
	if rot, ok := msg.Data["rotation"].(map[string]interface{}); ok {
		if x, ok := rot["x"].(float64); ok {
			player.Rotation.X = x
		}
		if y, ok := rot["y"].(float64); ok {
			player.Rotation.Y = y
		}
		if z, ok := rot["z"].(float64); ok {
			player.Rotation.Z = z
		}
	}

	player.LastSeen = time.Now()
}

// Read pump handles incoming messages from the client
func (c *Client) readPump() {
	defer func() {
		c.Server.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.Server.handlePlayerMessage(c, message)
	}
}

// Write pump handles outgoing messages to the client
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func handleWebSocket(server *GameServer, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Generate unique client ID
	clientID := generateClientID()

	client := &Client{
		ID:     clientID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Server: server,
	}

	server.register <- client

	// Start read and write pumps
	go client.writePump()
	go client.readPump()
}

func generateClientID() string {
	return fmt.Sprintf("player_%d", time.Now().UnixNano())
}

func main() {
	server := NewGameServer()
	go server.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(server, w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("MMO Server Running"))
	})

	addr := ":8080"
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe error:", err)
	}
}
