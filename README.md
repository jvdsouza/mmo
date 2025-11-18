# MMO Game - Godot Client + Go Server

A multiplayer MMO game with Godot 4.2 client and Go backend featuring real-time combat, spatial optimization, and extensible systems.

## 🚀 Quick Start

### Option 1: Docker (Easiest)

```bash
# Start everything with one command
make docker-up

# Or without make:
docker-compose up

# Access:
# - Game Server: http://localhost:8080
# - Web Client: http://localhost:3000
# - Test UI: http://localhost:3000/test/websocket_test.html
# - Database UI: http://localhost:8081
```

### Option 2: Local Development (Fastest)

```bash
# Terminal 1 - Server
cd server
go run ./cmd/server

# Terminal 2 - Godot Client
cd client
godot project.godot  # Press F5 to run
```

### Option 3: Using Makefile

```bash
# Complete setup
make quick-start

# Or step by step:
make setup-dev    # Setup environment
make deps         # Install dependencies
make dev          # Run server
```

## 📋 Requirements

- **Go 1.21+** (server)
- **Godot 4.2+** (client)
- **Docker** (optional, for containerized dev)
- **Make** (optional, for convenience commands)

## 🎮 Features

### Combat System
- ✅ Server-authoritative combat
- ✅ Multiple entity types (players, monsters, objects, environment)
- ✅ Real-time animations synchronized via WebSocket
- ✅ Spatial grid optimization (O(k) instead of O(n²))
- ✅ Faction system for targeting rules
- ✅ AI with aggro management
- ✅ Loot tables and respawn system

### Stats System (OCP Compliant)
- ✅ Extensible stat system (Open-Closed Principle)
- ✅ Character archetypes (Warrior, Mage, etc.)
- ✅ Stat modifiers (Flat, Percent, Multiplier, Conditional)
- ✅ Passive abilities
- ✅ Easy to add new stats without modifying core

### Network
- ✅ WebSocket real-time communication
- ✅ Event batching (50ms intervals)
- ✅ Interest management (only broadcast to nearby players)
- ✅ Combat event system for animations
- ✅ 20Hz tick rate (configurable)

## 🏗️ Architecture

```
┌─────────────────┐         WebSocket (JSON)         ┌─────────────────┐
│  Godot Client   │◄─────────────────────────────────►│   Go Server     │
│                 │                                   │                 │
│  - Animations   │      Combat Events (50ms)         │  - Combat Mgr   │
│  - VFX          │◄──────────────────────────────────│  - Spatial Grid │
│  - Input        │                                   │  - Stats System │
│  - UI           │      World State (20Hz)           │  - Entity Mgr   │
└─────────────────┘◄──────────────────────────────────└─────────────────┘
```

## 📁 Project Structure

```
mmo/
├── server/                 # Go backend
│   ├── cmd/
│   │   ├── server/        # Main game server ⭐
│   │   └── combat_demo/   # Combat system demo
│   ├── combat/            # Combat system
│   ├── entities/          # Monsters, objects, hazards
│   ├── player/            # Player management
│   ├── stats/             # Stats & archetypes
│   └── go.mod
│
├── client/                # Godot 4.2 client
│   ├── scripts/           # GDScript
│   │   ├── combat_event_manager.gd  # Combat events ⭐
│   │   ├── combat_entity_example.gd # Animation example
│   │   └── vfx_manager.gd           # Visual effects
│   ├── test/
│   │   └── websocket_test.html      # Browser test UI ⭐
│   └── project.godot
│
├── docs/                  # Documentation
│   ├── ANIMATION_INTEGRATION.md
│   ├── COMBAT_ENTITY_SYSTEM.md
│   ├── DEVELOPMENT_GUIDE.md        # Detailed guide ⭐
│   └── STATS_OCP_DESIGN.md
│
├── .env.development       # Dev config
├── .env.staging          # Staging config
├── .env.production       # Prod config
├── docker-compose.yml    # Local stack
├── Dockerfile            # Server container
├── Makefile              # Dev commands ⭐
└── README.md             # This file
```

## 🧪 Testing

### 1. Test Server Health

```bash
curl http://localhost:8080/health

# Expected:
{
  "status": "healthy",
  "environment": "development",
  "uptime": 45.2,
  "players": 0
}
```

### 2. Test WebSocket (Browser)

Open `http://localhost:3000/test/websocket_test.html`

- Visual test UI
- Send movement, attacks, chat
- View real-time combat events
- Monitor server stats

### 3. Test Combat System

```bash
# Run combat demo
make test-combat

# Or directly:
cd server
go run ./cmd/combat_demo
```

### 4. Test in Godot

1. Open `client/project.godot` in Godot
2. Ensure server is running
3. Press F5
4. Use WASD to move
5. Click monsters to attack

## 🔧 Development Commands

```bash
# Quick Commands
make help           # Show all commands
make dev            # Start dev server
make dev-client     # Open Godot
make docker-up      # Start all services
make test-websocket # Open test UI

# Build
make build          # Build server
make build-demo     # Build combat demo
make build-all      # Build everything

# Testing
make test           # Run tests
make test-combat    # Run combat demo

# Docker
make docker-build   # Build image
make docker-logs    # View logs
make docker-down    # Stop services

# Cleanup
make clean          # Clean builds
make clean-docker   # Clean Docker

# See Makefile for all commands
```

## 🌍 Environments

### Development
```bash
ENVIRONMENT=development make dev
# Features: CORS enabled, test monsters, debug logging
```

### Staging
```bash
ENVIRONMENT=staging make dev
# Features: Production-like, limited test data
```

### Production
```bash
ENVIRONMENT=production ./bin/mmo-server
# Features: Strict security, minimal logging, metrics
```

## 📡 Network Protocol

### Client → Server

**Movement:**
```json
{
  "type": "move",
  "data": {"position": {"x": 10, "y": 0, "z": 5}}
}
```

**Attack:**
```json
{
  "type": "attack",
  "data": {"target_id": "monster_123", "ability_id": "basic_attack"}
}
```

### Server → Client

**Combat Event:**
```json
{
  "type": "combat_event",
  "data": {
    "damage": 25,
    "is_critical": true,
    "animation_hint": "swing_right",
    "target_health": 75
  }
}
```

See `docs/DEVELOPMENT_GUIDE.md` for complete protocol.

## 🎯 Key Systems

### Combat Manager
- Central orchestration for all combat
- Spatial grid for O(k) queries
- Event broadcasting to clients
- Server-authoritative validation

### Stats System
- Open-Closed Principle design
- Add new stats without modifying core
- Stat modifiers (flat, percent, multiplier, conditional)
- Character archetypes with passive abilities

### Entity System
- Universal CombatEntity interface
- Players, monsters, destructibles, environmental hazards
- Faction-based targeting
- AI states for monsters

### Animation Integration
- Server events → Client animations
- VFX pooling for performance
- Damage numbers (color-coded)
- Hit reactions, death animations

## 📚 Documentation

- **[Development Guide](docs/DEVELOPMENT_GUIDE.md)** - Complete dev workflow
- **[Animation Integration](docs/ANIMATION_INTEGRATION.md)** - Client-server animation sync
- **[Combat System](docs/COMBAT_ENTITY_SYSTEM.md)** - Combat architecture
- **[Stats System](docs/STATS_OCP_DESIGN.md)** - OCP-compliant stats

## 🐛 Troubleshooting

**Server won't start:**
```bash
lsof -i :8080          # Check port
make clean             # Clean builds
make docker-restart    # Restart Docker
```

**Client can't connect:**
```bash
curl http://localhost:8080/health  # Check server
# Verify firewall settings
# Check WebSocket URL in client
```

**Combat events not received:**
- Check CombatEventManager is autoload in Godot
- Verify entities are registered
- Check browser network tab for WebSocket messages

## 🚢 Deployment

### Docker (Recommended)
```bash
docker build -t mmo-server:latest .
docker run -p 8080:8080 --env-file .env.production mmo-server
```

### Binary
```bash
make build-prod
./bin/mmo-server-prod
```

### Kubernetes
```bash
kubectl apply -f k8s/
# (k8s configs not included yet)
```

## 🤝 Contributing

1. Create feature branch
2. Make changes
3. Test locally: `make test`
4. Build: `make build`
5. Test in Docker: `make docker-up`
6. Submit PR

## 📊 Performance

- **Tick Rate:** 20 Hz (50ms)
- **Combat Events:** Batched every 50ms
- **Spatial Grid:** O(k) queries (k = nearby entities)
- **Target Players:** 1000+ concurrent
- **Tested Load:** 100 players + 500 monsters

## 🛣️ Roadmap

- [ ] Persistence (PostgreSQL)
- [ ] User authentication
- [ ] Inventory system
- [ ] Ranged combat & projectiles
- [ ] Crafting system
- [ ] Quest system
- [ ] Guild system
- [ ] Prometheus metrics
- [ ] Kubernetes deployment configs

## 📝 License

MIT License - See LICENSE file

## 🙏 Acknowledgments

- Built with [Godot Engine](https://godotengine.org/)
- WebSocket via [Gorilla WebSocket](https://github.com/gorilla/websocket)
- Inspired by classic MMOs

---

**Ready to build your MMO?**

```bash
make quick-start
```

Open `http://localhost:3000/test/websocket_test.html` and start testing! 🎮
