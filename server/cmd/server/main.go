package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"mmo-server/combat"
	"mmo-server/entities"
	"mmo-server/player"
	"mmo-server/stats"

	"github.com/gorilla/websocket"
)

// Configuration
type ServerConfig struct {
	Port            string
	Environment     string // "development", "staging", "production"
	TickRate        int
	MaxPlayers      int
	EnableCORS      bool
	LogLevel        string
}

func loadConfig() *ServerConfig {
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &ServerConfig{
		Port:        port,
		Environment: env,
		TickRate:    20, // 20 ticks per second
		MaxPlayers:  1000,
		EnableCORS:  env == "development",
		LogLevel:    "info",
	}
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins in development
		// In production, check against allowed origins
		if config.EnableCORS {
			return true
		}
		// TODO: Check against whitelist in production
		return true
	},
}

// Client represents a connected player
type Client struct {
	ID       string
	Player   *player.Player
	Conn     *websocket.Conn
	Send     chan []byte
	Server   *GameServer
	mu       sync.Mutex
}

// GameServer manages all game state
type GameServer struct {
	// Clients
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	// Game Systems
	combatManager *combat.CombatManager

	// World state
	monsters map[string]*entities.Monster

	mu sync.RWMutex
}

var (
	config *ServerConfig
	server *GameServer
)

func main() {
	// Load configuration
	config = loadConfig()

	log.Printf("Starting MMO Server in %s mode on port %s", config.Environment, config.Port)

	// Create game server
	server = NewGameServer()

	// Start game loop
	go server.Run()

	// HTTP handlers
	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/metrics", handleMetrics)

	// Static file serving for development
	if config.Environment == "development" {
		http.Handle("/", http.FileServer(http.Dir("../client")))
		log.Println("Serving client files from ../client")
	}

	// Start HTTP server
	addr := ":" + config.Port
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// NewGameServer creates a new game server
func NewGameServer() *GameServer {
	gs := &GameServer{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		monsters:   make(map[string]*entities.Monster),
	}

	// Create combat manager
	gs.combatManager = combat.NewCombatManager(combat.DefaultCombatConfig())

	// Set up combat event broadcasting
	gs.combatManager.SetBroadcastCallback(func(playerID string, message []byte) {
		gs.SendToClient(playerID, message)
	})

	// Spawn some test monsters in development
	if config.Environment == "development" {
		gs.spawnTestMonsters()
	}

	return gs
}

// Run starts the game server loop
func (gs *GameServer) Run() {
	ticker := time.NewTicker(time.Second / time.Duration(config.TickRate))
	defer ticker.Stop()

	log.Println("Game loop started")

	for {
		select {
		case client := <-gs.register:
			gs.handleRegister(client)

		case client := <-gs.unregister:
			gs.handleUnregister(client)

		case message := <-gs.broadcast:
			gs.handleBroadcast(message)

		case <-ticker.C:
			gs.Update(1.0 / float64(config.TickRate))
		}
	}
}

// Update game state
func (gs *GameServer) Update(deltaTime float64) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	// Update all monsters
	for _, monster := range gs.monsters {
		if monster.IsAlive() {
			monster.Update(deltaTime, gs.combatManager)
		}
	}

	// TODO: Update other game systems
	// - Environmental hazards
	// - Projectiles
	// - Buffs/debuffs expiration
	// - etc.

	// Broadcast world state to all clients
	gs.broadcastWorldState()
}

// Handle client registration
func (gs *GameServer) handleRegister(client *Client) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	gs.clients[client.ID] = client

	// Register player with combat manager
	gs.combatManager.RegisterEntity(client.Player)

	log.Printf("Client registered: %s (Total: %d)", client.ID, len(gs.clients))

	// Send welcome message
	welcome := map[string]interface{}{
		"type": "welcome",
		"data": map[string]interface{}{
			"player_id": client.ID,
			"server_time": time.Now(),
			"tick_rate": config.TickRate,
		},
	}
	client.SendJSON(welcome)

	// Send initial world state
	gs.sendWorldStateTo(client)
}

// Handle client unregistration
func (gs *GameServer) handleUnregister(client *Client) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	if _, ok := gs.clients[client.ID]; ok {
		delete(gs.clients, client.ID)

		// Unregister from combat manager
		gs.combatManager.UnregisterEntity(client.ID)

		close(client.Send)

		log.Printf("Client unregistered: %s (Total: %d)", client.ID, len(gs.clients))
	}
}

// Handle broadcast messages
func (gs *GameServer) handleBroadcast(message []byte) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	for _, client := range gs.clients {
		select {
		case client.Send <- message:
		default:
			// Client's send channel is full, disconnect them
			close(client.Send)
			delete(gs.clients, client.ID)
		}
	}
}

// Broadcast world state to all clients
func (gs *GameServer) broadcastWorldState() {
	// Only send every few ticks to save bandwidth
	// Full world state is expensive, use interest management in production

	players := make([]map[string]interface{}, 0, len(gs.clients))
	for _, client := range gs.clients {
		players = append(players, map[string]interface{}{
			"id":       client.Player.ID,
			"name":     client.Player.Name,
			"position": client.Player.Position,
			"health":   int(client.Player.Stats.GetStat(stats.StatHealth)),
			"max_health": int(client.Player.Stats.GetStat(stats.StatMaxHealth)),
		})
	}

	monsters := make([]map[string]interface{}, 0, len(gs.monsters))
	for _, monster := range gs.monsters {
		if monster.IsAlive() {
			monsters = append(monsters, map[string]interface{}{
				"id":       monster.ID,
				"name":     monster.Name,
				"position": monster.Position,
				"health":   int(monster.Stats.GetStat(stats.StatHealth)),
				"max_health": int(monster.Stats.GetStat(stats.StatMaxHealth)),
				"ai_state": monster.AIState,
			})
		}
	}

	state := map[string]interface{}{
		"type": "world_state",
		"data": map[string]interface{}{
			"players":  players,
			"monsters": monsters,
			"timestamp": time.Now().UnixMilli(),
		},
	}

	data, _ := json.Marshal(state)
	gs.broadcast <- data
}

// Send world state to specific client
func (gs *GameServer) sendWorldStateTo(client *Client) {
	// Similar to broadcastWorldState but for one client
	// In production, use interest management
	gs.broadcastWorldState()
}

// SendToClient sends a message to a specific client
func (gs *GameServer) SendToClient(clientID string, message []byte) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	if client, ok := gs.clients[clientID]; ok {
		select {
		case client.Send <- message:
		default:
			log.Printf("Failed to send to client %s (channel full)", clientID)
		}
	}
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	// Create player
	playerID := generatePlayerID()
	p := player.NewPlayerWithArchetype(
		playerID,
		"Player_"+playerID[:8],
		stats.NewWarriorArchetype(1),
	)
	p.Position = combat.Vector3{X: 0, Y: 0, Z: 0}

	// Create client
	client := &Client{
		ID:     playerID,
		Player: p,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Server: server,
	}

	// Register client
	server.register <- client

	// Start goroutines
	go client.writePump()
	go client.readPump()
}

// Client read pump
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

		c.handleMessage(message)
	}
}

// Client write pump
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

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
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

// Handle incoming messages from client
func (c *Client) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("JSON unmarshal error: %v", err)
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "move":
		c.handleMove(msg)
	case "attack":
		c.handleAttack(msg)
	case "ability":
		c.handleAbility(msg)
	case "chat":
		c.handleChat(msg)
	default:
		log.Printf("Unknown message type: %s", msgType)
	}
}

// Handle movement
func (c *Client) handleMove(msg map[string]interface{}) {
	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		return
	}

	// Extract position
	pos, ok := data["position"].(map[string]interface{})
	if !ok {
		return
	}

	x, _ := pos["x"].(float64)
	y, _ := pos["y"].(float64)
	z, _ := pos["z"].(float64)

	oldPos := c.Player.Position
	newPos := combat.Vector3{X: x, Y: y, Z: z}

	c.Player.Position = newPos
	c.Server.combatManager.UpdateEntityPosition(c.Player, oldPos, newPos)
}

// Handle attack
func (c *Client) handleAttack(msg map[string]interface{}) {
	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		return
	}

	targetID, ok := data["target_id"].(string)
	if !ok {
		return
	}

	abilityID, ok := data["ability_id"].(string)
	if !ok {
		abilityID = "basic_attack"
	}

	// Execute attack through combat manager
	event, err := c.Server.combatManager.ExecuteAbility(c.ID, targetID, abilityID)
	if err != nil {
		// Send error to client
		c.SendJSON(map[string]interface{}{
			"type": "error",
			"data": map[string]interface{}{
				"message": err.Error(),
			},
		})
		return
	}

	// Combat event automatically broadcast by combat manager
	_ = event
}

// Handle ability use
func (c *Client) handleAbility(msg map[string]interface{}) {
	// Same as attack for now
	c.handleAttack(msg)
}

// Handle chat
func (c *Client) handleChat(msg map[string]interface{}) {
	data, ok := msg["data"].(map[string]interface{})
	if !ok {
		return
	}

	message, ok := data["message"].(string)
	if !ok {
		return
	}

	// Broadcast chat message
	chat := map[string]interface{}{
		"type": "chat",
		"data": map[string]interface{}{
			"player_id": c.ID,
			"player_name": c.Player.Name,
			"message": message,
			"timestamp": time.Now(),
		},
	}

	chatData, _ := json.Marshal(chat)
	c.Server.broadcast <- chatData
}

// SendJSON helper
func (c *Client) SendJSON(data interface{}) {
	message, err := json.Marshal(data)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}

	select {
	case c.Send <- message:
	default:
		log.Printf("Client %s send channel full", c.ID)
	}
}

// Health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "healthy",
		"environment": config.Environment,
		"uptime": time.Since(startTime).Seconds(),
		"players": len(server.clients),
	})
}

// Metrics endpoint
func handleMetrics(w http.ResponseWriter, r *http.Request) {
	server.mu.RLock()
	defer server.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"players": len(server.clients),
		"monsters": len(server.monsters),
		"tick_rate": config.TickRate,
		"environment": config.Environment,
	})
}

// Spawn test monsters for development
func (gs *GameServer) spawnTestMonsters() {
	// Spawn some goblins
	for i := 0; i < 5; i++ {
		monster := entities.NewMonster(
			generateMonsterID(),
			entities.GetGoblinTemplate(),
			combat.Vector3{
				X: float64(i*10),
				Y: 0,
				Z: 10,
			},
		)
		gs.monsters[monster.ID] = monster
		gs.combatManager.RegisterEntity(monster)
	}

	log.Printf("Spawned %d test monsters", len(gs.monsters))
}

// Helper functions
var (
	playerIDCounter  uint64
	monsterIDCounter uint64
	idMutex          sync.Mutex
	startTime        = time.Now()
)

func generatePlayerID() string {
	idMutex.Lock()
	defer idMutex.Unlock()
	playerIDCounter++
	return "player_" + time.Now().Format("20060102150405") + "_" + string(rune(playerIDCounter))
}

func generateMonsterID() string {
	idMutex.Lock()
	defer idMutex.Unlock()
	monsterIDCounter++
	return "monster_" + time.Now().Format("20060102150405") + "_" + string(rune(monsterIDCounter))
}
