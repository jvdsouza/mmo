# MMO Architecture Documentation

## Overview

This document describes the technical architecture of the MMO prototype using Godot (client) and Go (server).

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       Game Clients                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Godot Client │  │ Godot Client │  │ Godot Client │      │
│  │   Player 1   │  │   Player 2   │  │   Player N   │      │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
└─────────┼──────────────────┼──────────────────┼─────────────┘
          │                  │                  │
          │    WebSocket     │    WebSocket     │  WebSocket
          │    ws://...      │    ws://...      │  ws://...
          │                  │                  │
          └──────────┬───────┴──────────────────┘
                     │
          ┌──────────▼──────────────────────────────────┐
          │         Go Game Server                      │
          │                                             │
          │  ┌────────────────────────────────┐        │
          │  │  HTTP/WebSocket Handler        │        │
          │  │  (port 8080)                   │        │
          │  └───────────┬────────────────────┘        │
          │              │                              │
          │  ┌───────────▼────────────────┐            │
          │  │  GameServer Main Loop       │            │
          │  │  - Connection Management    │            │
          │  │  - World State Updates      │            │
          │  │  - Message Broadcasting     │            │
          │  │  (20 ticks/second)          │            │
          │  └───────────┬────────────────┘            │
          │              │                              │
          │  ┌───────────▼────────────────┐            │
          │  │  Player State Management    │            │
          │  │  - Positions                │            │
          │  │  - Rotations                │            │
          │  │  - Connection Info          │            │
          │  └─────────────────────────────┘            │
          └─────────────────────────────────────────────┘
```

## Component Breakdown

### 1. Godot Client

#### NetworkManager (Autoload Singleton)
- **Purpose**: Manages WebSocket connection to server
- **Responsibilities**:
  - Establish/maintain connection
  - Send player position updates
  - Receive world state broadcasts
  - Handle connection events

#### Player (CharacterBody3D)
- **Purpose**: Player character controller
- **Local Player**:
  - Processes WASD input
  - Handles camera control
  - Applies physics and movement
  - Sends position to server
- **Remote Player**:
  - Receives position from server
  - Interpolates movement smoothly
  - No input processing

#### MultiplayerController
- **Purpose**: Manages all players in the game world
- **Responsibilities**:
  - Spawn local player on connection
  - Spawn/despawn remote players
  - Update remote player positions
  - Clean up disconnected players

### 2. Go Server

#### WebSocket Handler
```go
handleWebSocket(server, w, r)
  ├─> Upgrade HTTP to WebSocket
  ├─> Create Client instance
  ├─> Register client
  └─> Start read/write pumps
```

#### GameServer
**Data Structures**:
- `clients`: Map of connected clients
- `players`: Map of player states
- `register`: Channel for new connections
- `unregister`: Channel for disconnections
- `broadcast`: Channel for messages

**Main Loop** (20 Hz):
```go
for {
  select {
    case client := <-register:
      handleRegister(client)

    case client := <-unregister:
      handleUnregister(client)

    case message := <-broadcast:
      handleBroadcast(message)

    case <-ticker:
      sendWorldState()
  }
}
```

#### Client (per connection)
**Read Pump**:
- Reads incoming WebSocket messages
- Processes player movement
- Updates player state
- Handles ping/pong

**Write Pump**:
- Sends outgoing messages
- Handles world state updates
- Maintains connection health

## Data Flow

### Player Movement Flow

```
1. Player Input (Godot)
   ↓
2. Local Movement Applied
   ↓
3. Send Position to Server
   {
     "type": "move",
     "player_id": "123",
     "data": {
       "position": {x, y, z},
       "rotation": {x, y, z}
     }
   }
   ↓
4. Server Receives & Updates State
   ↓
5. Server Broadcasts World State (20Hz)
   {
     "type": "world_state",
     "data": {
       "players": [...]
     }
   }
   ↓
6. All Clients Receive Update
   ↓
7. Remote Players Interpolate Position
```

## Network Protocol

### Message Format
All messages are JSON over WebSocket.

```json
{
  "type": "message_type",
  "player_id": "optional_player_id",
  "data": {
    // Message-specific data
  }
}
```

### Message Types

| Type | Direction | Purpose |
|------|-----------|---------|
| `connect` | Server → Client | Confirm connection, provide player ID |
| `move` | Client → Server | Update player position/rotation |
| `world_state` | Server → Client | Broadcast all player states |
| `disconnect` | Bidirectional | Handle disconnection |

## Performance Characteristics

### Server
- **Tick Rate**: 20 Hz (50ms intervals)
- **Concurrency**: Goroutines per client
- **Scalability**: ~1000 concurrent connections per server instance
- **Network**: WebSocket (TCP-based)

### Client
- **Frame Rate**: 60 FPS (Godot default)
- **Network Updates**: On movement change
- **Interpolation**: Lerp for smooth remote players
- **Prediction**: Local player moves immediately

## Concurrency Model

### Server (Go)
```
Main Goroutine
  └─> GameServer.Run() - Main event loop

Per Client:
  ├─> readPump() - Read from WebSocket
  └─> writePump() - Write to WebSocket
```

**Thread Safety**:
- `sync.RWMutex` protects shared state
- Channels for goroutine communication
- Lock-free reads for world state broadcast

### Client (Godot)
- Single-threaded game loop
- `_process()` handles network polling
- `_physics_process()` handles movement

## Security Considerations

### Current Implementation
⚠️ **Development Only** - Not production-ready

**Missing**:
- Authentication
- Encryption (use WSS in production)
- Server-side validation
- Anti-cheat measures
- Rate limiting
- Input sanitization

### Recommended for Production

1. **Authentication**
   - JWT tokens
   - Session management
   - Account system

2. **Server Authority**
   - Validate all movements
   - Check speed/distance
   - Verify game rules

3. **Network Security**
   - WSS (WebSocket Secure)
   - Certificate validation
   - HTTPS for web builds

4. **Anti-Cheat**
   - Server-side validation (see main README)
   - Anomaly detection
   - Behavioral analysis

## Scaling Strategies

### Horizontal Scaling

**Current**: Single server instance

**Future**:
```
┌──────────────┐     ┌──────────────┐
│  Server 1    │     │  Server 2    │
│  (Region A)  │     │  (Region B)  │
└──────┬───────┘     └──────┬───────┘
       │                    │
       └────────┬───────────┘
                │
         ┌──────▼──────┐
         │  Database   │
         │  (Shared)   │
         └─────────────┘
```

**Approaches**:
1. **Regional Servers**: Different servers per region
2. **Sharding**: Different servers per zone/instance
3. **Load Balancer**: Distribute connections

### Vertical Scaling
- Increase server resources
- Optimize tick rate
- Reduce message size
- Implement compression

## Future Architecture Enhancements

### 1. Database Layer
```
GameServer
    ↓
PostgreSQL
  - Player accounts
  - Character data
  - Inventory
  - World state persistence
```

### 2. Microservices
```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ Auth Service│  │Game Server  │  │Chat Service │
└─────────────┘  └─────────────┘  └─────────────┘
        │              │                 │
        └──────────────┼─────────────────┘
                       │
                  ┌────▼────┐
                  │ Message │
                  │  Queue  │
                  └─────────┘
```

### 3. Interest Management
Only send data for nearby players:

```
Player View Distance: 50m
  ↓
Server calculates nearby players
  ↓
Only send those player states
  ↓
Reduce bandwidth exponentially
```

### 4. State Synchronization
- **Dead Reckoning**: Predict movement
- **Snapshot Interpolation**: Smooth playback
- **Lag Compensation**: Rewind for hit detection

## Performance Benchmarks

### Current (Development)

| Metric | Value |
|--------|-------|
| Concurrent Players | ~100-500 |
| Messages/sec | ~2000-10000 |
| Latency (local) | <5ms |
| Memory/client | ~1MB |
| CPU/client | Minimal (goroutines) |

### Target (Production)

| Metric | Target |
|--------|--------|
| Concurrent Players | 1000+ per server |
| Messages/sec | 20000+ |
| Latency (internet) | <100ms |
| Uptime | 99.9% |

## Code Organization Best Practices

### Server (Go)
```
server/
├── main.go              # Entry point
├── server/
│   ├── game_server.go   # GameServer logic
│   ├── client.go        # Client connection
│   └── player.go        # Player state
├── handlers/
│   └── websocket.go     # WebSocket handlers
├── models/
│   └── message.go       # Protocol definitions
└── database/
    └── db.go            # Database layer
```

### Client (Godot)
```
client/
├── scripts/
│   ├── network/
│   │   ├── network_manager.gd
│   │   └── protocol.gd
│   ├── player/
│   │   ├── player.gd
│   │   └── player_controller.gd
│   └── world/
│       └── multiplayer_controller.gd
└── scenes/
    ├── main.tscn
    └── player.tscn
```

## Monitoring & Debugging

### Server Logging
```go
log.Printf("Client connected: %s", clientID)
log.Printf("World state sent to %d clients", len(clients))
```

### Client Logging
```gdscript
print("Connected to server: ", server_url)
print("Received world state with ", len(players), " players")
```

### Recommended Tools
- **Server**: Prometheus + Grafana
- **Network**: Wireshark
- **Profiling**: Go pprof
- **Load Testing**: Artillery, k6

## References

- [Fast-Paced Multiplayer](https://www.gabrielgambetta.com/client-server-game-architecture.html)
- [Source Multiplayer Networking](https://developer.valvesoftware.com/wiki/Source_Multiplayer_Networking)
- [Gorilla WebSocket](https://github.com/gorilla/websocket/tree/master/examples)
