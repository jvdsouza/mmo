# Network Architecture

## Overview

The MMO uses a **WebSocket-based client-server architecture** with a Go backend and Godot client. This document describes the current network implementation, message protocols, and data flow patterns.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT (Godot)                          │
├─────────────────────────────────────────────────────────────────┤
│  Player Script                                                  │
│    │                                                            │
│    ├──> NetworkManager (WebSocketPeer)                         │
│    │       │                                                    │
│    │       ├──> send_position(pos, rot)                        │
│    │       ├──> send_attack(target_id, ability_id)             │
│    │       ├──> send_ability(target_id, ability_id)            │
│    │       └──> send_chat(message)                             │
│    │                                                            │
│    └──> CombatEventManager                                     │
│            └──> Handles incoming combat events from server     │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
                              ↑↓ WebSocket (ws://localhost:8080/ws)
┌─────────────────────────────────────────────────────────────────┐
│                        SERVER (Go)                              │
├─────────────────────────────────────────────────────────────────┤
│  GameServer                                                     │
│    │                                                            │
│    ├──> Client Management                                      │
│    │      ├─ register chan (new connections)                   │
│    │      ├─ unregister chan (disconnections)                  │
│    │      └─ clients map[string]*Client                        │
│    │                                                            │
│    ├──> Player State                                           │
│    │      └─ players map[string]*Player                        │
│    │                                                            │
│    ├──> CombatManager                                          │
│    │      ├─ Entities map (all combat entities)               │
│    │      ├─ SpatialGrid (spatial partitioning)               │
│    │      ├─ AbilityRegistry (all abilities)                  │
│    │      ├─ ValidatorChain (combat validation)               │
│    │      └─ EventBroadcaster (combat events)                 │
│    │                                                            │
│    └──> World Loop (50ms tick = 20 tps)                        │
│           └─ Broadcasts world_state to all clients            │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Technology Stack

### Client (Godot 4)
- **Language:** GDScript
- **Networking:** WebSocketPeer (Godot built-in)
- **Protocol:** JSON over WebSocket
- **Connection:** Persistent bi-directional connection

### Server (Go)
- **Language:** Go 1.21+
- **WebSocket Library:** gorilla/websocket
- **Concurrency:** Go channels and goroutines
- **Protocol:** JSON over WebSocket
- **Tick Rate:** 20 TPS (50ms per tick)

## Message Protocol

All messages are JSON-formatted and sent over WebSocket connections.

### Client → Server Messages

#### Move Message
```json
{
  "type": "move",
  "player_id": "player_123",
  "data": {
    "position": {
      "x": 10.5,
      "y": 1.0,
      "z": 5.2
    },
    "rotation": {
      "x": 0.0,
      "y": 1.57,
      "z": 0.0
    }
  }
}
```

**Frequency:** ~60 Hz when player is moving, 0 Hz when stationary

#### Attack Message
```json
{
  "type": "attack",
  "player_id": "player_123",
  "data": {
    "target_id": "monster_456",
    "ability_id": "basic_attack"
  }
}
```

**Frequency:** Limited by ability cooldowns (~1 Hz)

#### Ability Message
```json
{
  "type": "ability",
  "player_id": "player_123",
  "data": {
    "target_id": "player_789",
    "ability_id": "fireball"
  }
}
```

**Frequency:** Limited by ability cooldowns

#### Chat Message
```json
{
  "type": "chat",
  "player_id": "player_123",
  "data": {
    "message": "Hello world!"
  }
}
```

**Frequency:** Infrequent (user-initiated)

### Server → Client Messages

#### Welcome Message
Sent immediately after connection is established.

```json
{
  "type": "welcome",
  "data": {
    "your_id": "player_123"
  }
}
```

**Frequency:** Once per connection

#### World State Message
Broadcast to all connected clients on every server tick.

```json
{
  "type": "world_state",
  "data": {
    "players": [
      {
        "id": "player_123",
        "position": { "x": 10.5, "y": 1.0, "z": 5.2 },
        "rotation": { "x": 0.0, "y": 1.57, "z": 0.0 }
      },
      {
        "id": "player_456",
        "position": { "x": 15.2, "y": 1.0, "z": 8.1 },
        "rotation": { "x": 0.0, "y": 0.0, "z": 0.0 }
      }
    ]
  }
}
```

**Frequency:** 20 Hz (every 50ms)

#### Combat Event Message
Sent when combat actions occur (attacks, damage, deaths).

```json
{
  "type": "combat_event",
  "data": {
    "event_type": "damage",
    "attacker_id": "player_123",
    "target_id": "monster_456",
    "damage": 25,
    "is_critical": false,
    "animation_hint": "swing_right",
    "target_position": { "x": 12.0, "y": 1.0, "z": 6.0 }
  }
}
```

**Frequency:** Variable (depends on combat activity)

#### Combat Events Batch
Multiple combat events sent together for efficiency.

```json
{
  "type": "combat_events_batch",
  "data": {
    "events": [
      { "event_type": "damage", "attacker_id": "...", ... },
      { "event_type": "death", "entity_id": "...", ... }
    ]
  }
}
```

**Frequency:** Variable

## Connection Flow

### 1. Initial Connection

```
Client                              Server
  │                                    │
  ├──── WebSocket Connect ────────────>│
  │     (ws://localhost:8080/ws)       │
  │                                    │
  │                         Generate unique player_id
  │                         Register client in clients map
  │                         Create Player entity
  │                                    │
  │<──── "welcome" message ─────────────┤
  │     { your_id: "player_123" }      │
  │                                    │
  └─── Store player_id locally         │
       Emit connection_established     │
```

### 2. Position Synchronization

```
Client (every frame if moved)         Server (20 tps)
  │                                      │
  ├──── "move" message ─────────────────>│
  │     { position, rotation }           │
  │                                      │
  │                         Update player state
  │                         Update spatial grid
  │                                      │
  │<──── "world_state" broadcast ────────┤
  │     { all_players[] }                │
  │                                      │
  └─── Spawn/update remote players       │
```

**Client sends:** Position updates when player moves
**Server sends:** Full world state 20 times per second
**Client updates:** Remote player positions from world state

### 3. Combat Flow

```
Client                                   Server
  │                                         │
  ├─ Player clicks attack                   │
  ├─ CombatValidator.validate_attack()      │
  │  (client-side prediction)               │
  ├─ weapon.attack(target)                  │
  ├─ play_attack() animation                │
  │  (OPTIMISTIC CLIENT PREDICTION)         │
  │                                         │
  ├──── "attack" message ──────────────────>│
  │     { target_id, ability_id }           │
  │                                         │
  │              CombatManager.ExecuteAbility()
  │              ├─ Get caster entity
  │              ├─ Get target entity
  │              ├─ Get ability from registry
  │              ├─ Validate (range, cooldown, resources)
  │              ├─ Calculate damage (with crits)
  │              ├─ Apply damage to target
  │              └─ Create CombatEvent
  │                                         │
  │<──── "combat_event" broadcast ──────────┤
  │     { damage, is_critical, ... }        │
  │                                         │
  └─── CombatEventManager handles event     │
       ├─ Play attack animation (server-confirmed)
       ├─ Play hit reaction on target
       ├─ Show damage numbers
       └─ Handle death if needed
```

**Flow:**
1. Client validates and predicts attack locally
2. Client sends attack message to server
3. Server validates (authoritative)
4. Server calculates damage
5. Server broadcasts event to ALL clients
6. All clients play synchronized animations

## Client-Side Implementation

### NetworkManager

**Location:** `client/scripts/network_manager.gd`

**Purpose:** Manages WebSocket connection and message routing

**Key Components:**

```gdscript
extends Node

# WebSocket connection
var socket = WebSocketPeer.new()
var server_url = "ws://localhost:8080/ws"
var connected = false
var player_id = ""

# Signals
signal connection_established(id: String)
signal connection_failed()
signal connection_closed()
signal world_state_received(players: Array)

# Main methods
func connect_to_server()
func disconnect_from_server()
func send_position(position: Vector3, rotation: Vector3)
func send_attack(target_id: String, ability_id: String)
func send_ability(target_id: String, ability_id: String)
func send_chat(message_text: String)
func handle_server_message(packet: PackedByteArray)
```

**Message Handling:**
- `handle_connect(data)` - Process welcome message, store player_id
- `handle_world_state(data)` - Emit signal for world manager
- `handle_combat_event(data)` - Forward to CombatEventManager
- `handle_combat_batch(data)` - Forward batch to CombatEventManager

**Update Loop:**
```gdscript
func _process(_delta):
    socket.poll()
    var state = socket.get_ready_state()

    # Check connection state
    if state == WebSocketPeer.STATE_OPEN and not connected:
        connected = true

    # Process incoming messages
    if state == WebSocketPeer.STATE_OPEN:
        while socket.get_available_packet_count() > 0:
            var packet = socket.get_packet()
            handle_server_message(packet)
```

### CombatEventManager

**Location:** `client/scripts/combat_event_manager.gd`

**Purpose:** Receives combat events from server and triggers animations

**Key Responsibilities:**
- Register/unregister combat entities
- Route combat events to correct entities
- Trigger attack animations
- Trigger hit reactions
- Trigger death animations
- Spawn VFX effects

## Server-Side Implementation

### GameServer

**Location:** `server/main.go`

**Purpose:** Main game server managing connections and world state

**Key Components:**

```go
type GameServer struct {
    clients    map[string]*Client
    players    map[string]*Player
    register   chan *Client
    unregister chan *Client
    broadcast  chan []byte
    mu         sync.RWMutex
}
```

**Main Loop:**
```go
func (gs *GameServer) Run() {
    ticker := time.NewTicker(50 * time.Millisecond)  // 20 tps
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
```

**Connection Management:**
- `handleRegister(client)` - Add new client, send welcome message
- `handleUnregister(client)` - Remove client, cleanup resources
- `handleBroadcast(message)` - Send message to all clients
- `sendWorldState()` - Broadcast world state (20 Hz)

### Client Goroutines

Each client connection runs two goroutines:

**Read Pump:**
```go
func (c *Client) readPump() {
    defer func() {
        c.Server.unregister <- c
        c.Conn.Close()
    }()

    for {
        _, message, err := c.Conn.ReadMessage()
        if err != nil {
            break
        }
        c.Server.handlePlayerMessage(c, message)
    }
}
```

**Write Pump:**
```go
func (c *Client) writePump() {
    ticker := time.NewTicker(54 * time.Second)  // Ping interval

    for {
        select {
        case message, ok := <-c.Send:
            if !ok {
                return
            }
            c.Conn.WriteMessage(websocket.TextMessage, message)

        case <-ticker.C:
            c.Conn.WriteMessage(websocket.PingMessage, nil)
        }
    }
}
```

### CombatManager

**Location:** `server/combat/combat_manager.go`

**Purpose:** Handles all combat logic and validation

**Key Components:**

```go
type CombatManager struct {
    entities         map[string]CombatEntity
    spatialGrid      *SpatialGrid
    abilityRegistry  *AbilityRegistry
    validatorChain   *ValidatorChain
    eventBroadcaster *EventBroadcaster
    config           *CombatConfig
}
```

**Main Method:**
```go
func (cm *CombatManager) ExecuteAbility(
    casterID string,
    targetID string,
    abilityID string,
) (*CombatEvent, error) {
    // 1. Get entities from entity map
    caster := cm.entities[casterID]
    target := cm.entities[targetID]

    // 2. Get ability from registry
    ability := cm.abilityRegistry.Get(abilityID)

    // 3. Validate targeting rules
    if err := CanAttack(caster, target); err != nil {
        return nil, err
    }

    // 4. Run validation chain
    if err := cm.validatorChain.Validate(caster, target, ability); err != nil {
        return nil, err
    }

    // 5. Execute ability (calculate damage, apply effects)
    event := ability.Execute(caster, target)

    // 6. Broadcast event to all clients
    cm.eventBroadcaster.Broadcast(event)

    return event, nil
}
```

**Validation Chain:**
- Range check
- Cooldown check
- Resource check (mana, stamina)
- Line of sight check (future)
- State check (stunned, silenced)
- Friendly fire check

**Spatial Grid:**
- Efficient range queries
- O(1) entity lookup by region
- Used for AoE abilities and proximity checks

### Player Entity

**Location:** `server/player/player.go`

**Purpose:** Represents a player character with stats and state

**Structure:**

```go
type Player struct {
    ID              string
    Name            string
    Position        combat.Vector3
    Rotation        combat.Vector3
    Stats           *stats.StatContainer
    Archetype       stats.CharacterArchetype
    Level           int
    CooldownManager *combat.CooldownManager
    PvPEnabled      bool
    Stunned         bool
}
```

**Implements Interfaces:**
- `combat.CombatEntity` - Core combat interface
- `combat.Player` - Player-specific interface

## Network Performance

### Bandwidth Estimates

**Upstream (Client → Server):**
- Position updates: ~200 bytes/msg × 60 Hz = ~12 KB/s (when moving)
- Combat actions: ~150 bytes/msg × 1 Hz = ~150 B/s
- Chat messages: Variable, infrequent

**Downstream (Server → Client):**
- World state (20 players): ~2 KB/msg × 20 Hz = ~40 KB/s
- World state (100 players): ~10 KB/msg × 20 Hz = ~200 KB/s
- Combat events: ~500 bytes/event × variable

**Total per client (100 players):**
- Upload: ~12 KB/s
- Download: ~210 KB/s
- **Total: ~222 KB/s (~1.8 Mbps)**

### Latency Characteristics

**Position Update Latency:**
```
Client moves → Send to server → Server processes →
Broadcast world_state → Other clients receive
```
- Best case: ~50ms (1 tick)
- Typical: ~100ms (2 ticks + network)
- Worst case: ~150ms (3 ticks + network)

**Combat Event Latency:**
```
Attack click → Validate → Send to server →
Server validates → Server broadcasts →
Clients receive → Play animation
```
- Best case: ~25ms (server + network RTT)
- Typical: ~50ms
- Worst case: ~100ms (with packet loss retry)

### Tick Rate Analysis

**Server Tick Rate:** 20 TPS (50ms per tick)

**Pros:**
- Predictable server load
- Consistent world state broadcasts
- Good balance of responsiveness vs bandwidth

**Cons:**
- Maximum 50ms position update delay
- Choppy movement on remote players without interpolation

## Current Architecture Strengths

### 1. WebSocket-Based (Low Latency)
✅ Persistent bi-directional connection
✅ No HTTP overhead on each message
✅ Real-time position updates
✅ Server can push data without polling

### 2. Gorilla WebSocket (Production-Ready)
✅ Battle-tested library
✅ Handles ping/pong automatically
✅ Proper connection cleanup
✅ Read/write deadlines built-in

### 3. Channel-Based Server Architecture
✅ Go channels for safe concurrency
✅ Separate read/write pumps per client
✅ Non-blocking broadcast
✅ Clean goroutine lifecycle

### 4. 20 TPS World State
✅ Balance between responsiveness and bandwidth
✅ Consistent tick rate
✅ Predictable network traffic

### 5. Combat System Integration
✅ Separate CombatManager for combat logic
✅ SpatialGrid for efficient range queries
✅ ValidatorChain for extensible validation
✅ EventBroadcaster for synchronized animations

### 6. Clean Separation of Concerns
✅ NetworkManager handles networking
✅ CombatEventManager handles animations
✅ Player handles game logic
✅ No tight coupling

## Current Limitations

### 1. Simple State Sync
⚠️ Full world state broadcast every 50ms
⚠️ No delta compression (sends all data every tick)
⚠️ No interest management (all players get all state)
⚠️ Bandwidth scales linearly with player count

**Impact:** With 100 players, each client receives ~10 KB per tick

### 2. No Client Prediction
⚠️ Movement feels laggy on high ping
⚠️ No local prediction of own movement
⚠️ No server reconciliation
⚠️ Input delay equals network RTT + tick time

**Impact:** Players see their own movement delayed by ~50-100ms

### 3. No Entity Interpolation
⚠️ Remote players may appear choppy
⚠️ No smoothing between position updates
⚠️ Teleporting effect on missed packets

**Impact:** Remote players move in 50ms jumps, not smoothly

### 4. JSON Serialization
⚠️ Human-readable but verbose
⚠️ More bandwidth than binary protocols
⚠️ Slower to parse than binary

**Impact:** 2-3x more bandwidth than MessagePack/Protobuf

### 5. Single Server Instance
⚠️ No horizontal scaling
⚠️ No load balancing
⚠️ Limited to one region
⚠️ Single point of failure

**Impact:** Limited to ~1000 concurrent players per server

### 6. No Interest Management
⚠️ All players receive updates for ALL entities
⚠️ No spatial culling on server
⚠️ Unnecessary network traffic

**Impact:** Player in zone A receives updates for players in zone Z

## Potential Improvements

### Short-Term (Low Hanging Fruit)

**1. Delta Compression**
Only send changed positions instead of full world state.

**2. Client Interpolation**
Smooth remote player movement between updates.

**3. Interest Management**
Only send nearby entity updates (Area of Interest).

### Medium-Term

**4. Binary Protocol**
Switch from JSON to MessagePack or Protobuf (50-70% bandwidth reduction).

**5. Client Prediction**
Predict local player movement immediately, reconcile with server.

**6. Entity State Quantization**
Reduce precision for distant entities (less bandwidth).

### Long-Term

**7. Horizontal Scaling**
Multiple game server instances with load balancing.

**8. Zone Sharding**
Split world into zones, players only connect to relevant servers.

**9. State Synchronization Framework**
Replace custom sync with library like Nakama or Colyseus.

## Security Considerations

### Current State

**Authentication:** ❌ None - Auto-generated player IDs
**Validation:** ✅ Server-side combat validation (authoritative)
**Input Sanitization:** ⚠️ Basic JSON parsing
**Rate Limiting:** ❌ No client rate limiting
**Encryption:** ❌ Plain WebSocket (ws://), not secure (wss://)

### Recommended Additions

1. **Authentication System**
   - Login with credentials
   - Session tokens
   - Account management

2. **Input Validation**
   - Rate limiting on client messages
   - Movement speed validation
   - Ability cooldown enforcement

3. **Encryption**
   - Switch to wss:// (WebSocket Secure)
   - TLS certificates
   - Encrypted player data

4. **Anti-Cheat**
   - Server-authoritative movement validation
   - Anomaly detection
   - Replay analysis

## Testing & Monitoring

### Current Testing

**Manual Testing:**
- Connect multiple clients
- Test position synchronization
- Test combat interactions

**Recommended Additions:**

1. **Load Testing**
   - Simulate 100+ concurrent connections
   - Measure bandwidth and CPU usage
   - Test combat event spam

2. **Network Simulation**
   - Test with artificial latency (100ms, 200ms, 500ms)
   - Test with packet loss (1%, 5%, 10%)
   - Test reconnection logic

3. **Monitoring**
   - Connection count metrics
   - Message throughput
   - Combat event frequency
   - Server tick time

## Summary

The current network architecture is:

✅ **Solid foundation** - WebSocket + Go channels is proven
✅ **Real-time capable** - 20 TPS is good for MMO gameplay
✅ **Clean architecture** - Separation of concerns, extensible
✅ **Combat-ready** - Server-authoritative with event broadcasting

⚠️ **Needs optimization** - Delta compression, interpolation, interest management
⚠️ **Needs security** - Authentication, encryption, anti-cheat
⚠️ **Needs scaling** - Load balancing, zone sharding for large player counts

For an early-stage MMO with <100 concurrent players, this architecture is **excellent**. For scaling to thousands of players, you'll need the improvements listed above.

---

**Last Updated:** 2025-01-18
**Version:** 1.0
**Related Documents:**
- COMBAT_REFACTOR_PROGRESS.md
- UI_SYSTEM_ARCHITECTURE.md
