# MMO Development Guide

## Quick Start

### Prerequisites

- **Go 1.21+** - Server development
- **Godot 4.2+** - Client development
- **Docker & Docker Compose** - For full stack development
- **Git** - Version control

### Option 1: Docker (Recommended for Testing)

Start entire stack with one command:

```bash
# Start all services
docker-compose up

# Services available:
# - Game Server: http://localhost:8080
# - Client: http://localhost:3000
# - PostgreSQL: localhost:5432
# - Redis: localhost:6379
# - Adminer (DB UI): http://localhost:8081
```

### Option 2: Local Development (Fastest for Development)

**Terminal 1 - Start Server:**
```bash
cd server
cp ../.env.development .env
go run ./cmd/server
```

**Terminal 2 - Start Godot Client:**
```bash
cd client
godot # Open project, press F5 to run
```

## Environment Setup

### Development
```bash
export ENVIRONMENT=development
# or on Windows:
set ENVIRONMENT=development

# Server runs on :8080 with:
# - CORS enabled
# - Debug logging
# - Test monsters spawned
# - All features enabled
```

### Staging
```bash
export ENVIRONMENT=staging
# Load staging config
# - Stricter CORS
# - Info logging
# - No test monsters
```

### Production
```bash
export ENVIRONMENT=production
# Load production config
# - Strict security
# - Minimal logging
# - Metrics enabled
```

## Testing the Combat System

### Test 1: Server Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Expected response:
{
  "status": "healthy",
  "environment": "development",
  "uptime": 45.2,
  "players": 0
}

# Metrics
curl http://localhost:8080/metrics

# Expected response:
{
  "players": 2,
  "monsters": 5,
  "tick_rate": 20,
  "environment": "development"
}
```

### Test 2: WebSocket Connection

Using `websocat` (install: `cargo install websocat`):

```bash
# Connect to server
websocat ws://localhost:8080/ws

# You'll receive welcome message:
{
  "type": "welcome",
  "data": {
    "player_id": "player_20251118_1",
    "server_time": "2025-11-18T10:30:00Z",
    "tick_rate": 20
  }
}

# Send movement
{"type": "move", "data": {"position": {"x": 10, "y": 0, "z": 5}}}

# Attack a monster
{"type": "attack", "data": {"target_id": "monster_123", "ability_id": "basic_attack"}}
```

### Test 3: Browser Testing

Open `client/test/websocket_test.html` in browser:

```html
<!DOCTYPE html>
<html>
<head>
    <title>MMO WebSocket Test</title>
</head>
<body>
    <h1>MMO Server Test</h1>
    <div id="status">Connecting...</div>
    <div id="messages"></div>

    <button onclick="move()">Move</button>
    <button onclick="attack()">Attack Monster</button>

    <script>
        const ws = new WebSocket('ws://localhost:8080/ws');

        ws.onopen = () => {
            document.getElementById('status').textContent = 'Connected';
        };

        ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            console.log('Received:', msg);

            const div = document.getElementById('messages');
            div.innerHTML += '<pre>' + JSON.stringify(msg, null, 2) + '</pre>';
        };

        function move() {
            ws.send(JSON.stringify({
                type: 'move',
                data: { position: { x: 10, y: 0, z: 5 } }
            }));
        }

        function attack() {
            ws.send(JSON.stringify({
                type: 'attack',
                data: { target_id: 'monster_1', ability_id: 'basic_attack' }
            }));
        }
    </script>
</body>
</html>
```

### Test 4: Godot Client

1. Open Godot project: `godot client/project.godot`
2. Update `network_manager.gd`:
   ```gdscript
   var server_url = "ws://localhost:8080/ws"
   ```
3. Press F5 to run
4. Check console for connection messages
5. Move player with WASD
6. Click on monsters to attack

## Project Structure

```
mmo/
├── server/               # Go server
│   ├── cmd/
│   │   ├── server/      # Main server binary
│   │   └── combat_demo/ # Combat system demo
│   ├── combat/          # Combat system
│   ├── entities/        # Game entities (monsters, etc.)
│   ├── player/          # Player management
│   ├── stats/           # Stats system
│   └── go.mod
│
├── client/              # Godot client
│   ├── scripts/         # GDScript files
│   ├── scenes/          # Godot scenes
│   ├── assets/          # Art, sounds, etc.
│   └── project.godot
│
├── docs/                # Documentation
│   ├── ANIMATION_INTEGRATION.md
│   ├── COMBAT_ENTITY_SYSTEM.md
│   └── DEVELOPMENT_GUIDE.md
│
├── .env.development     # Dev environment config
├── .env.staging         # Staging config
├── .env.production      # Production config
├── docker-compose.yml   # Local stack
└── Dockerfile           # Server container
```

## Network Protocol

### Client → Server Messages

**Move:**
```json
{
  "type": "move",
  "data": {
    "position": {"x": 10.5, "y": 0, "z": 5.2},
    "rotation": {"x": 0, "y": 45, "z": 0}
  }
}
```

**Attack:**
```json
{
  "type": "attack",
  "data": {
    "target_id": "monster_123",
    "ability_id": "basic_attack"
  }
}
```

**Chat:**
```json
{
  "type": "chat",
  "data": {
    "message": "Hello world!"
  }
}
```

### Server → Client Messages

**Welcome:**
```json
{
  "type": "welcome",
  "data": {
    "player_id": "player_123",
    "server_time": "2025-11-18T10:30:00Z",
    "tick_rate": 20
  }
}
```

**World State (20 Hz):**
```json
{
  "type": "world_state",
  "data": {
    "players": [
      {
        "id": "player_1",
        "name": "Player_1",
        "position": {"x": 0, "y": 0, "z": 0},
        "health": 100,
        "max_health": 150
      }
    ],
    "monsters": [
      {
        "id": "monster_1",
        "name": "Goblin",
        "position": {"x": 10, "y": 0, "z": 5},
        "health": 75,
        "max_health": 100,
        "ai_state": 2
      }
    ],
    "timestamp": 1700308800000
  }
}
```

**Combat Event:**
```json
{
  "type": "combat_event",
  "data": {
    "event_id": "evt_123",
    "type": 0,
    "caster_id": "player_1",
    "target_id": "monster_1",
    "damage": 25,
    "is_critical": true,
    "animation_hint": "swing_right",
    "target_health": 50,
    "target_max_health": 100,
    "target_died": false
  }
}
```

**Combat Events Batch (every 50ms):**
```json
{
  "type": "combat_events_batch",
  "data": {
    "events": [
      { /* combat event 1 */ },
      { /* combat event 2 */ },
      { /* combat event 3 */ }
    ],
    "timestamp": "2025-11-18T10:30:00.050Z"
  }
}
```

## Debugging

### Server Logs

```bash
# View all logs
docker-compose logs -f server

# View specific number of lines
docker-compose logs --tail=100 server

# Filter for errors
docker-compose logs server | grep ERROR
```

### Client Debugging (Godot)

1. Enable debug console: **Project → Project Settings → Debug → Print**
2. Add debug prints:
   ```gdscript
   print("Position: ", position)
   print("Combat event: ", event)
   ```
3. View in Godot console (bottom panel)

### Network Debugging

**Chrome DevTools:**
1. Open DevTools (F12)
2. Network tab → Filter: WS
3. Click WebSocket connection
4. View Messages tab

**Wireshark:**
```bash
# Capture WebSocket traffic
sudo wireshark -i lo -f "port 8080"

# Filter: websocket
```

## Performance Testing

### Load Testing (k6)

```bash
# Install k6
brew install k6  # macOS
# or download from https://k6.io

# Run load test
k6 run tests/load_test.js
```

**tests/load_test.js:**
```javascript
import ws from 'k6/ws';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '1m', target: 100 },  // Ramp to 100 users
    { duration: '3m', target: 100 },  // Stay at 100
    { duration: '1m', target: 0 },    // Ramp down
  ],
};

export default function () {
  const url = 'ws://localhost:8080/ws';

  const res = ws.connect(url, function (socket) {
    socket.on('open', () => console.log('connected'));
    socket.on('message', (data) => console.log('Message:', data));

    socket.setTimeout(function () {
      socket.send(JSON.stringify({
        type: 'move',
        data: { position: { x: 10, y: 0, z: 5 } }
      }));
    }, 1000);

    socket.setTimeout(function () {
      socket.close();
    }, 60000);
  });

  check(res, { 'status is 101': (r) => r && r.status === 101 });
}
```

## Continuous Integration

### GitHub Actions

`.github/workflows/test.yml`:
```yaml
name: Test

on: [push, pull_request]

jobs:
  test-server:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: |
          cd server
          go test ./...

      - name: Build
        run: |
          cd server
          go build ./cmd/server

  test-client:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Godot tests
        run: |
          # Install Godot headless
          # Run Godot tests
          echo "TODO: Godot test runner"
```

## Deployment

### Development
```bash
docker-compose up
```

### Staging
```bash
docker-compose -f docker-compose.staging.yml up -d
```

### Production (Kubernetes)
```bash
kubectl apply -f k8s/
```

## Troubleshooting

### Server won't start
```bash
# Check port availability
lsof -i :8080

# Check logs
docker-compose logs server

# Rebuild
docker-compose build --no-cache server
```

### Client can't connect
```bash
# Check server is running
curl http://localhost:8080/health

# Check WebSocket endpoint
websocat ws://localhost:8080/ws

# Check firewall
sudo ufw status
```

### Combat events not received
1. Check server logs for broadcast errors
2. Verify client CombatEventManager is autoload
3. Check network tab in browser/Godot debugger
4. Ensure entities are registered with CombatEventManager

## Next Steps

1. **Add Persistence:**
   - Player accounts
   - Save player position
   - Inventory system

2. **Add More Features:**
   - Ranged combat
   - Abilities with cooldowns
   - Quests
   - Crafting

3. **Optimize:**
   - Interest management
   - Entity culling
   - Network compression

4. **Security:**
   - Authentication
   - Rate limiting
   - Anti-cheat server validation

5. **Monitoring:**
   - Prometheus metrics
   - Grafana dashboards
   - Error tracking (Sentry)
