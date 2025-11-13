# MMO Game Prototype - Godot + Go

A basic MMO game prototype with:
- **Godot 4.2** client (3D third-person)
- **Go** server backend (WebSocket-based)

## Project Structure

```
mmo/
├── server/          # Go backend server
│   ├── go.mod
│   └── main.go
├── client/          # Godot game client
│   ├── project.godot
│   ├── scripts/
│   │   ├── network_manager.gd
│   │   ├── player.gd
│   │   └── multiplayer_controller.gd
│   └── scenes/
│       ├── main.tscn
│       └── player.tscn
└── docs/            # Documentation
```

## Features

### Current Implementation

✅ **Server (Go)**
- WebSocket server on port 8080
- Player connection management
- World state broadcasting (20 ticks/sec)
- Position synchronization
- Concurrent player handling with goroutines

✅ **Client (Godot)**
- 3D character controller (WASD movement)
- Mouse look camera
- WebSocket connection to server
- Real-time position updates
- Multiplayer player rendering
- Smooth interpolation for remote players

## Quick Start

### 1. Run the Go Server

```bash
cd server
go mod download
go run main.go
```

Server will start on `http://localhost:8080`

You should see:
```
Server starting on :8080
```

### 2. Run the Godot Client

**Option A: Using Godot Editor**
1. Download Godot 4.2+ from https://godotengine.org/
2. Open the project: `File > Open Project` → select `client/project.godot`
3. Press `F5` or click the Play button

**Option B: Export and Run**
1. Export the project as a standalone executable
2. Run the exported game

### 3. Test Multiplayer

To test multiplayer functionality:
1. Run multiple instances of the Godot client
2. Each client connects to the same server
3. You should see other players moving around

## Controls

- **W/A/S/D** - Move forward/left/back/right
- **Mouse** - Look around
- **Space** - Jump
- **ESC** - Release/capture mouse cursor

## Network Protocol

### Message Types

**Client → Server:**
```json
{
  "type": "move",
  "player_id": "player_12345",
  "data": {
    "position": {"x": 0.0, "y": 1.0, "z": 0.0},
    "rotation": {"x": 0.0, "y": 0.0, "z": 0.0}
  }
}
```

**Server → Client:**
```json
{
  "type": "world_state",
  "data": {
    "players": [
      {
        "id": "player_12345",
        "position": {"x": 0.0, "y": 1.0, "z": 0.0},
        "rotation": {"x": 0.0, "y": 0.0, "z": 0.0}
      }
    ]
  }
}
```

## Architecture

### Server Architecture
```
Client 1 ──┐
           ├──> WebSocket ──> Go Game Server ──> Game Loop (20 Hz)
Client 2 ──┘                        │
                                    └──> Player State Management
```

### Client Architecture
```
Input ──> Player Controller ──> NetworkManager ──> WebSocket ──> Server
                │                      │
                └──> Local Prediction  │
                                       └──> World State Updates ──> Render Remote Players
```

## Next Steps

### Essential Features to Add

1. **Server-Side Validation**
   - Movement validation (speed checks)
   - Anti-cheat measures
   - Position sanity checks

2. **Combat System**
   - Attack actions
   - Health system
   - Damage calculation (server-authoritative)

3. **Persistence**
   - Database integration (PostgreSQL)
   - Player accounts
   - Character data saving

4. **Interest Management**
   - Only send nearby player data
   - Spatial partitioning
   - Reduce network traffic

5. **Game Features**
   - Chat system
   - Inventory
   - NPCs
   - Quests

6. **Optimization**
   - Client-side prediction
   - Better interpolation
   - Lag compensation
   - Network compression

## Development Tips

### Server Development
```bash
# Install dependencies
cd server
go mod download

# Run with auto-reload (install air)
go install github.com/cosmtrek/air@latest
air

# Build for production
go build -o mmo-server main.go
```

### Client Development
- Use Godot's debugger for network issues
- Check console output for connection logs
- Test with localhost before deploying

### Testing Multiplayer Locally
```bash
# Terminal 1: Run server
cd server && go run main.go

# Terminal 2-N: Run multiple Godot instances
# Open Godot editor and press F5 multiple times
```

## Configuration

### Server Configuration
Edit `server/main.go`:
```go
addr := ":8080"  // Change port
```

### Client Configuration
Edit `client/scripts/network_manager.gd`:
```gdscript
var server_url = "ws://localhost:8080/ws"  // Change server address
```

## Troubleshooting

**Connection Failed**
- Ensure server is running
- Check server URL in `network_manager.gd`
- Check firewall settings

**Players Not Visible**
- Check console for errors
- Verify world state messages are being received
- Ensure player scene is assigned in main.tscn

**Laggy Movement**
- Server may be overloaded
- Network latency
- Implement client-side prediction

## Requirements

- **Go**: 1.21+
- **Godot**: 4.2+
- **OS**: Windows, Linux, or macOS

## License

MIT License - Feel free to use for learning and development

## Resources

- [Godot Networking Docs](https://docs.godotengine.org/en/stable/tutorials/networking/)
- [Go WebSocket Tutorial](https://github.com/gorilla/websocket)
- [MMO Architecture Patterns](https://www.gabrielgambetta.com/client-server-game-architecture.html)
