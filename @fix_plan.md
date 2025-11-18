# Combat System Development Plan - Phase 1: Basic Melee

## Priority: HIGH - Implement Basic Melee Combat System

### Context
We are building an MMO with Godot client and Go server. We need to implement a combat system with the following design principles:
- Interface-based architecture (extensible for ranged, AoE, etc.)
- Server-authoritative (anti-cheat)
- Client prediction (responsive gameplay)
- Network-aware (designed for 50-150ms latency)

**Architecture Document**: See `docs/COMBAT_ARCHITECTURE.md` for full design.

---

## Tasks - Execute in Order

### ✅ Task 0: Architecture Design (COMPLETED)
- [x] Combat architecture documented
- [x] Interfaces defined
- [x] Message protocol specified

### 🔲 Task 1: Server - Combat Interfaces & Base Types
**File**: `server/combat/ability.go`

Implement:
```go
- Ability interface
- AbilityType enum
- CombatResult struct
- ResourceCost struct
- BaseAbility struct (default implementations)
```

**File**: `server/combat/validators.go`

Implement validator chain:
```go
- AbilityValidator interface
- ValidatorChain struct
- AliveValidator
- CooldownValidator
- RangeValidator
- ResourceValidator
```

**Success Criteria**:
- Code compiles
- Interfaces are clear and extensible
- Unit tests pass

---

### 🔲 Task 2: Server - Player Stats & Resources
**File**: `server/player/stats.go`

Add to Player struct:
```go
type PlayerStats struct {
    Health      int
    MaxHealth   int
    Mana        int
    MaxMana     int
    Stamina     int
    MaxStamina  int
    AttackPower int
    Defense     int
    IsDead      bool
}
```

**File**: `server/player/cooldown_manager.go`

Implement cooldown tracking:
```go
type CooldownManager struct {
    cooldowns map[string]time.Time
}
- IsReady(abilityID string) bool
- SetCooldown(abilityID string, duration time.Duration)
- GetRemaining(abilityID string) time.Duration
```

**Success Criteria**:
- Player has combat stats
- Cooldowns track per-ability
- Thread-safe with mutex

---

### 🔲 Task 3: Server - Melee Ability Implementation
**File**: `server/combat/melee_ability.go`

Implement:
```go
type MeleeAbility struct {
    BaseAbility
    Range float64
}

func (m *MeleeAbility) Validate(caster, target *Player, world *World) error
func (m *MeleeAbility) Execute(caster, target *Player, world *World) *CombatResult
```

**File**: `server/combat/damage_calculator.go`

Implement:
```go
func CalculateMeleeDamage(caster, target *Player, ability *MeleeAbility) int {
    // Base damage = AttackPower - Defense
    // Add randomness (90-110%)
}
```

**File**: `server/combat/ability_registry.go`

Create registry:
```go
type AbilityRegistry struct {
    abilities map[string]Ability
}
- Register(id string, ability Ability)
- Get(id string) Ability
```

Register 2 abilities:
- "basic_attack": 0s cooldown, 2m range, 10 base damage
- "heavy_attack": 3s cooldown, 2.5m range, 25 base damage, costs 20 stamina

**Success Criteria**:
- Two melee abilities registered
- Damage calculation works
- Range validation works

---

### 🔲 Task 4: Server - Combat Manager & Message Handling
**File**: `server/combat/combat_manager.go`

Implement:
```go
type CombatManager struct {
    abilityRegistry *AbilityRegistry
    validatorChain  *ValidatorChain
}

func (cm *CombatManager) HandleAbilityRequest(req *AbilityRequest) (*CombatResult, error)
func (cm *CombatManager) BroadcastCombatEvent(event *CombatEvent, nearbyPlayers []*Player)
```

**File**: `server/main.go`

Add message handler:
```go
case "use_ability":
    handleUseAbility(client, msg)
```

Update world state to include player health.

**Success Criteria**:
- Server receives ability requests
- Server validates using validator chain
- Server calculates damage
- Server broadcasts combat events
- Health updates sync to clients

---

### 🔲 Task 5: Client - Godot Combat Interfaces
**File**: `client/scripts/combat/ability.gd`

Base class:
```gdscript
class_name Ability
extends Resource

var id: String
var ability_type: int  # MELEE = 0, RANGED = 1
var cooldown: float
var range: float
var cost: Dictionary

func can_use(caster: Player, target: Node3D) -> bool
func predict_execute(caster: Player, target: Node3D)
func on_confirmed(result: Dictionary)
func on_rejected(reason: String)
```

**File**: `client/scripts/combat/melee_ability.gd`

```gdscript
class_name MeleeAbility
extends Ability

func can_use(caster: Player, target: Node3D) -> bool:
    # Check cooldown, range, resources

func predict_execute(caster: Player, target: Node3D):
    # Play attack animation
    # Show predicted effect
```

**Success Criteria**:
- Base Ability class exists
- MeleeAbility extends it
- Client can check if ability is usable

---

### 🔲 Task 6: Client - Player Stats & UI
**File**: `client/scripts/player/player_stats.gd`

```gdscript
class_name PlayerStats
extends Node

var health: int = 100
var max_health: int = 100
var mana: int = 100
var max_mana: int = 100
var stamina: int = 100
var max_stamina: int = 100

signal health_changed(new_health, max_health)
signal mana_changed(new_mana, max_mana)
```

**File**: `client/scripts/ui/health_bar.gd`

Create health bar UI:
- Shows current/max health
- Updates on health_changed signal
- Green → Yellow → Red color gradient

**File**: `client/scripts/ui/damage_numbers.gd`

Floating damage numbers:
- Spawn at hit position
- Float upward and fade
- Different color for crits

**Success Criteria**:
- Health bars visible above players
- Damage numbers spawn on hit
- Stats update from network

---

### 🔲 Task 7: Client - Ability Manager & Input
**File**: `client/scripts/combat/ability_manager.gd`

```gdscript
class_name AbilityManager
extends Node

var abilities: Array[Ability] = []
var cooldowns: Dictionary = {}

func use_ability(ability_id: String, target_id: String):
    # Check can_use
    # Call predict_execute
    # Send to server
    # Start cooldown timer
```

**File**: `client/scripts/player/player.gd`

Add to player:
```gdscript
@onready var ability_manager = $AbilityManager
var target: Player = null

func _input(event):
    if event.is_action_pressed("attack_basic"):
        ability_manager.use_ability("basic_attack", target)

    if event.is_action_pressed("attack_heavy"):
        ability_manager.use_ability("heavy_attack", target)
```

Add input actions to `project.godot`:
- attack_basic: Left Mouse Button
- attack_heavy: Right Mouse Button

**Success Criteria**:
- Player can target other players
- Abilities trigger on input
- Cooldowns prevent spam

---

### 🔲 Task 8: Client - Network Integration
**File**: `client/scripts/network_manager.gd`

Add methods:
```gdscript
func send_use_ability(ability_id: String, target_id: String):
    var message = {
        "type": "use_ability",
        "ability_id": ability_id,
        "target_id": target_id,
        "caster_position": player_position,
        "timestamp": Time.get_ticks_msec()
    }
    socket.send_text(JSON.stringify(message))

func handle_combat_event(data: Dictionary):
    # Update target health
    # Show damage numbers
    # Play hit effects

func handle_ability_rejected(data: Dictionary):
    # Show error message
    # Reset cooldown UI
```

**Success Criteria**:
- Client sends ability requests
- Client receives combat events
- Client handles rejections gracefully

---

### 🔲 Task 9: Integration & Testing
**Test Plan**:

1. **Server Tests**:
   - Start server
   - Verify abilities registered
   - Test damage calculation
   - Test validator chain

2. **Client Tests**:
   - Open Godot project
   - Verify UI elements load
   - Test ability inputs
   - Check cooldown timers

3. **Integration Tests**:
   - Run 2 clients + server
   - Player 1 attacks Player 2
   - Verify:
     - Attack animation plays
     - Damage numbers appear
     - Health bars update
     - Both clients see the combat

4. **Latency Tests**:
   - Add artificial 100ms delay
   - Verify combat still feels responsive
   - Check for desyncs

**Success Criteria**:
- Combat works end-to-end
- Feels responsive with latency
- No crashes or errors
- Both players see consistent state

---

### 🔲 Task 10: Documentation & Cleanup
**Files to Update**:
- `README.md`: Add combat system section
- `docs/QUICKSTART.md`: Add combat controls
- Add code comments
- Create `docs/COMBAT_TESTING.md`: Testing guide

**Success Criteria**:
- Documentation complete
- Code is commented
- Testing guide exists
- Ready for next phase

---

## Definition of Done

Phase 1 is complete when:
- [x] Architecture designed
- [ ] Server implements interface-based combat
- [ ] Client implements interface-based combat
- [ ] Two melee abilities work (basic, heavy)
- [ ] Health bars show above players
- [ ] Damage numbers appear on hit
- [ ] Combat works with 100ms simulated latency
- [ ] No crashes or errors
- [ ] Code committed and pushed
- [ ] Documentation updated

---

## Next Phase Preview

**Phase 2** will add:
- Ranged abilities (projectiles)
- Status effects (stun, poison)
- More abilities (3-4 more)
- Better animations
- Sound effects

---

## Notes

- **Keep it simple**: Don't add features not in this plan
- **Test frequently**: After each task, verify it works
- **Commit often**: Commit after each completed task
- **Server authority**: Always validate on server first
- **Network-aware**: Consider latency in all decisions

---

## Exit Signals for Ralph

Ralph should stop iterating when:
1. All tasks marked as complete (✅)
2. Integration tests pass
3. Documentation updated
4. Git status clean (all committed)
5. File contains: `PHASE_1_COMPLETE: true`

**PHASE_1_COMPLETE**: false
