# Quick Start Guide

Get your MMO running in 5 minutes!

## Prerequisites

1. **Install Go** (1.21 or later)
   - Download: https://go.dev/dl/
   - Verify: `go version`

2. **Install Godot** (4.2 or later)
   - Download: https://godotengine.org/download
   - Use the "Godot Engine - .NET" version if you want C# support (optional)

## Step 1: Start the Server

### Linux/Mac:
```bash
cd server
./run.sh
```

### Windows:
```bash
cd server
go run main.go
```

You should see:
```
Server starting on :8080
```

✅ Server is running!

## Step 2: Open the Godot Client

1. Launch Godot Engine
2. Click "Import"
3. Navigate to the `client` folder
4. Select `project.godot`
5. Click "Import & Edit"

## Step 3: Run the Game

1. In Godot, press **F5** (or click the Play button ▶)
2. You should see a 3D world with a capsule (your player)
3. Use **WASD** to move, **Mouse** to look around

✅ You're in the game!

## Step 4: Test Multiplayer

1. With the game running, press **F5** in Godot again
2. This opens a second game instance
3. Move around in one window
4. Watch the other player move in the second window!

✅ Multiplayer is working!

## Troubleshooting

### "Connection failed"
- Make sure the server is running (Step 1)
- Check that port 8080 is not blocked
- Verify server URL in `client/scripts/network_manager.gd`

### "Can't see other players"
- Check the Godot console (bottom panel) for errors
- Make sure both clients are connected to the same server
- Try moving around - positions update on movement

### "Mouse stuck"
- Press **ESC** to release the mouse cursor
- Press in the game window to recapture

## Next Steps

Now that you have the basic prototype running:

1. **Read the Architecture**: See `docs/ARCHITECTURE.md`
2. **Add Features**: Combat, inventory, NPCs
3. **Implement Security**: Server-side validation, anti-cheat
4. **Add Persistence**: Database for player data
5. **Optimize**: Interest management, lag compensation

## Development Workflow

### Typical Development Session

**Terminal 1** (Server):
```bash
cd server
go run main.go
```

**Godot Editor**:
1. Make changes to scripts/scenes
2. Press F5 to test
3. Check console for errors
4. Iterate!

### Making Changes

**Server-Side**:
1. Edit `server/main.go`
2. Stop server (Ctrl+C)
3. Restart: `go run main.go`
4. Test in Godot client

**Client-Side**:
1. Edit scripts in `client/scripts/`
2. Save file (Ctrl+S)
3. Godot auto-reloads scripts
4. Press F5 to test

## Common Tasks

### Change Server Port
**File**: `server/main.go`
```go
addr := ":8080"  // Change to ":3000" for example
```

**File**: `client/scripts/network_manager.gd`
```gdscript
var server_url = "ws://localhost:8080/ws"  // Update port here too
```

### Change Player Speed
**File**: `client/scripts/player.gd`
```gdscript
@export var speed = 5.0  // Increase for faster movement
```

### Adjust Network Tick Rate
**File**: `server/main.go`
```go
ticker := time.NewTicker(50 * time.Millisecond)  // 50ms = 20 Hz
```

## Server Commands

```bash
# Run server (development)
go run main.go

# Build server (production)
go build -o mmo-server main.go

# Run built server
./mmo-server

# Download dependencies
go mod download

# Update dependencies
go mod tidy
```

## Godot Commands

| Action | Shortcut |
|--------|----------|
| Run game | F5 |
| Run current scene | F6 |
| Stop game | F8 |
| Reload scripts | Ctrl+R |
| Open script | Ctrl+Shift+O |

## Directory Structure

```
mmo/
├── server/
│   ├── main.go          # Server code
│   ├── go.mod           # Go dependencies
│   └── run.sh           # Quick start script
├── client/
│   ├── project.godot    # Godot project file
│   ├── scripts/         # GDScript files
│   └── scenes/          # Scene files (.tscn)
├── docs/
│   ├── ARCHITECTURE.md  # Technical details
│   └── QUICKSTART.md    # This file
└── README.md            # Main documentation
```

## Resources

- **Godot Docs**: https://docs.godotengine.org/
- **Go Docs**: https://go.dev/doc/
- **WebSocket RFC**: https://tools.ietf.org/html/rfc6455
- **Gorilla WebSocket**: https://github.com/gorilla/websocket

## Getting Help

1. Check console output for errors
2. Read error messages carefully
3. Review `README.md` for detailed info
4. Check `ARCHITECTURE.md` for technical details

## What's Next?

See the **Next Steps** section in the main README.md for ideas on features to add:
- Combat system
- Chat
- Database persistence
- Anti-cheat
- Interest management
- And much more!

Happy coding! 🚀
