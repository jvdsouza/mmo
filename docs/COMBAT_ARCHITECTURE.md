# Combat System Architecture

## Overview
Decoupled, interface-based combat system supporting multiple attack types (melee, ranged, AoE, etc.) with network synchronization.

## Design Principles

1. **Interface-based**: All abilities implement common interface
2. **Server-authoritative**: Server validates and calculates all combat
3. **Client prediction**: Instant feedback for responsive gameplay
4. **Extensible**: Easy to add new ability types
5. **Network-aware**: Built for 50-150ms latency from day 1

---

## Core Interfaces

### Server Side (Go)

```go
// Ability - Core interface all abilities implement
type Ability interface {
    GetID() string
    GetType() AbilityType
    GetCooldown() time.Duration
    GetRange() float64
    GetCost() ResourceCost

    // Validate checks if ability can be used
    Validate(caster *Player, target *Player, world *World) error

    // Execute performs the ability effect
    Execute(caster *Player, target *Player, world *World) *CombatResult
}

// AbilityType enum
type AbilityType int
const (
    AbilityTypeMelee AbilityType = iota
    AbilityTypeRanged
    AbilityTypeAoE
    AbilityTypeBuff
    AbilityTypeHeal
)

// CombatResult - What happened
type CombatResult struct {
    AbilityID    string
    CasterID     string
    TargetID     string
    Damage       int
    Hit          bool
    CriticalHit  bool
    Effects      []StatusEffect
    Timestamp    time.Time
}

// Validator chain for ability validation
type AbilityValidator interface {
    Validate(caster *Player, target *Player, ability Ability) error
}

// Validators: RangeValidator, CooldownValidator, ResourceValidator, etc.
```

### Client Side (Godot)

```gdscript
# Base ability class (interface-like)
class_name Ability
extends Resource

var id: String
var ability_type: AbilityType
var cooldown: float
var range: float
var cost: Dictionary  # {mana: 50, stamina: 10}

# Virtual methods - override in subclasses
func can_use(caster: Player, target: Node3D) -> bool:
    return false

func predict_execute(caster: Player, target: Node3D):
    pass  # Client-side prediction

func on_confirmed(result: Dictionary):
    pass  # Server confirmed

func on_rejected(reason: String):
    pass  # Server rejected
```

---

## Combat Flow

```
┌─────────────────────────────────────────────────────────────┐
│ CLIENT 1 (Attacker)                                          │
│                                                              │
│ 1. Player clicks attack                                     │
│    ↓                                                         │
│ 2. ability.can_use() → check cooldown, range, resources    │
│    ↓                                                         │
│ 3. ability.predict_execute() → play animation instantly     │
│    ↓                                                         │
│ 4. Send to server: {                                        │
│      type: "use_ability",                                   │
│      ability_id: "basic_attack",                            │
│      target_id: "player_2",                                 │
│      timestamp: 12345,                                      │
│      position: {x, y, z}  // for validation                │
│    }                                                         │
│                                                              │
└──────────────────┬───────────────────────────────────────────┘
                   │ WebSocket (50ms)
                   ↓
┌──────────────────────────────────────────────────────────────┐
│ SERVER (Go)                                                  │
│                                                              │
│ 1. Receive ability request                                  │
│    ↓                                                         │
│ 2. Load ability: ability := AbilityRegistry.Get("basic_attack") │
│    ↓                                                         │
│ 3. Run validator chain:                                     │
│    - AliveValidator: Is caster alive?                       │
│    - CooldownValidator: Is ability ready?                   │
│    - RangeValidator: Is target in range?                    │
│    - ResourceValidator: Has mana/stamina?                   │
│    ↓                                                         │
│ 4. If valid: result := ability.Execute(caster, target, world) │
│    ↓                                                         │
│ 5. Update game state:                                       │
│    - target.health -= result.Damage                         │
│    - caster.cooldowns[ability_id] = time.Now()              │
│    - caster.mana -= ability.Cost.Mana                       │
│    ↓                                                         │
│ 6. Broadcast to nearby players: {                           │
│      type: "combat_event",                                  │
│      caster_id: "player_1",                                 │
│      target_id: "player_2",                                 │
│      ability_id: "basic_attack",                            │
│      result: {damage: 25, hit: true, crit: false},          │
│      target_health: 75                                      │
│    }                                                         │
│                                                              │
└──────────────────┬───────────────────────────────────────────┘
                   │ WebSocket (50ms)
                   ↓
┌──────────────────────────────────────────────────────────────┐
│ ALL NEARBY CLIENTS                                           │
│                                                              │
│ CLIENT 1 (Attacker):                                         │
│   - ability.on_confirmed(result)                            │
│   - Update cooldown UI                                      │
│   - Show damage numbers                                     │
│                                                              │
│ CLIENT 2 (Target):                                           │
│   - Play hit animation                                      │
│   - Update health bar                                       │
│   - Show damage taken                                       │
│                                                              │
│ CLIENT 3 (Observer):                                         │
│   - Play combat VFX between players                         │
│   - Show damage numbers                                     │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## Ability Types

### 1. Melee Abilities

```go
type MeleeAbility struct {
    BaseAbility
    Range float64  // 2-3 meters
}

func (m *MeleeAbility) Validate(caster, target *Player, world *World) error {
    // Check distance
    distance := caster.Position.Distance(target.Position)
    if distance > m.Range {
        return errors.New("target_out_of_range")
    }
    return nil
}

func (m *MeleeAbility) Execute(caster, target *Player, world *World) *CombatResult {
    damage := calculateMeleeDamage(caster, target, m)
    return &CombatResult{
        Damage: damage,
        Hit:    true,
        // ...
    }
}
```

### 2. Ranged Abilities (Future)

```go
type RangedAbility struct {
    BaseAbility
    ProjectileSpeed float64
    Range           float64  // 20-50 meters
}

func (r *RangedAbility) Execute(caster, target *Player, world *World) *CombatResult {
    // Spawn projectile
    projectile := NewProjectile(caster.Position, target.Position, r.ProjectileSpeed)
    world.AddProjectile(projectile)

    // Server simulates projectile flight
    // Hit detection happens when projectile reaches target
    return nil  // Result comes later when projectile hits
}
```

### 3. AoE Abilities (Future)

```go
type AoEAbility struct {
    BaseAbility
    Radius float64
}

func (a *AoEAbility) Execute(caster, target *Player, world *World) *CombatResult {
    // Find all players in radius
    affected := world.GetPlayersInRadius(target.Position, a.Radius)

    // Apply damage to all
    results := []CombatResult{}
    for _, player := range affected {
        result := calculateDamage(caster, player, a)
        results = append(results, result)
    }
    return &CombatResult{MultiTarget: results}
}
```

---

## Component Structure

### Server (Go)

```
server/
├── main.go
├── combat/
│   ├── ability.go           # Ability interface
│   ├── ability_registry.go  # Registry of all abilities
│   ├── melee_ability.go     # Melee implementation
│   ├── ranged_ability.go    # Ranged implementation
│   ├── combat_result.go     # Result structures
│   ├── validators.go        # Validation chain
│   └── damage_calculator.go # Damage formulas
├── player/
│   ├── player.go            # Player with combat stats
│   ├── stats.go             # HP, mana, stamina
│   └── cooldown_manager.go  # Track cooldowns
└── world/
    └── combat_manager.go    # Orchestrates combat
```

### Client (Godot)

```
client/
├── scripts/
│   ├── combat/
│   │   ├── ability.gd              # Base ability class
│   │   ├── ability_manager.gd      # Manages player abilities
│   │   ├── melee_ability.gd        # Melee implementation
│   │   ├── ranged_ability.gd       # Ranged implementation
│   │   ├── combat_predictor.gd     # Client prediction
│   │   └── combat_visualizer.gd    # VFX, animations
│   ├── player/
│   │   ├── player.gd               # Extended with combat
│   │   ├── player_stats.gd         # HP, mana display
│   │   └── cooldown_ui.gd          # Cooldown indicators
│   └── ui/
│       ├── health_bar.gd
│       ├── ability_bar.gd
│       └── damage_numbers.gd
```

---

## Message Protocol

### Client → Server

```json
{
  "type": "use_ability",
  "ability_id": "basic_attack",
  "target_id": "player_123",
  "target_position": {"x": 10, "y": 0, "z": 5},
  "caster_position": {"x": 8, "y": 0, "z": 3},
  "timestamp": 1699123456789
}
```

### Server → Clients

```json
{
  "type": "combat_event",
  "event_id": "evt_12345",
  "caster_id": "player_456",
  "target_id": "player_123",
  "ability_id": "basic_attack",
  "result": {
    "damage": 25,
    "hit": true,
    "critical": false,
    "target_health": 75,
    "target_max_health": 100
  },
  "timestamp": 1699123456890
}
```

### Server → Client (Rejection)

```json
{
  "type": "ability_rejected",
  "ability_id": "basic_attack",
  "reason": "target_out_of_range",
  "cooldown_remaining": 0.5
}
```

---

## Validation Chain

```go
type ValidatorChain struct {
    validators []AbilityValidator
}

func (vc *ValidatorChain) Validate(caster, target *Player, ability Ability) error {
    for _, validator := range vc.validators {
        if err := validator.Validate(caster, target, ability); err != nil {
            return err
        }
    }
    return nil
}

// Validators
type AliveValidator struct{}
type CooldownValidator struct{}
type RangeValidator struct{}
type ResourceValidator struct{}
type LineOfSightValidator struct{}  // Future: check walls
```

---

## Stats System

```go
type PlayerStats struct {
    // Resources
    Health       int
    MaxHealth    int
    Mana         int
    MaxMana      int
    Stamina      int
    MaxStamina   int

    // Combat Stats
    AttackPower  int
    Defense      int
    CritChance   float64
    CritMultiplier float64

    // Status
    IsDead       bool
    StatusEffects []StatusEffect
}

type StatusEffect struct {
    Type     string  // "stun", "poison", "slow"
    Duration time.Duration
    Value    int
}
```

---

## Extensibility Examples

### Adding a New Ability

**Server:**
```go
// 1. Create new ability
type PowerStrike struct {
    MeleeAbility
}

func (p *PowerStrike) Execute(caster, target *Player, world *World) *CombatResult {
    damage := calculateMeleeDamage(caster, target, p) * 2  // Double damage
    return &CombatResult{Damage: damage, Hit: true}
}

// 2. Register it
func init() {
    AbilityRegistry.Register("power_strike", &PowerStrike{
        BaseAbility: BaseAbility{
            ID: "power_strike",
            Cooldown: 5 * time.Second,
            Cost: ResourceCost{Stamina: 20},
        },
        Range: 3.0,
    })
}
```

**Client:**
```gdscript
# 1. Create new ability class
class_name PowerStrike
extends MeleeAbility

func _init():
    id = "power_strike"
    cooldown = 5.0
    range = 3.0

func predict_execute(caster: Player, target: Node3D):
    # Play special animation
    caster.play_animation("power_strike")
    # Bigger effect
    spawn_vfx("power_strike_effect")
```

### Supporting Ranged Combat

Just add `RangedAbility` class that implements the same `Ability` interface:
- Validation checks longer range
- Execute spawns projectile instead of instant hit
- Client predicts projectile trajectory

**No changes needed to core combat system!**

---

## Anti-Cheat Considerations

### Server Validation
```go
func (cv *CombatValidator) ValidateAttack(req *AbilityRequest) error {
    // 1. Range check with caster position
    distance := req.CasterPosition.Distance(req.TargetPosition)
    if distance > ability.GetRange() * 1.1 {  // 10% tolerance for latency
        return errors.New("impossible_range")
    }

    // 2. Timestamp check (prevent replay attacks)
    if time.Now().Sub(req.Timestamp) > 500*time.Millisecond {
        return errors.New("timestamp_too_old")
    }

    // 3. Cooldown check (server-side)
    if caster.Cooldowns[ability.GetID()].After(time.Now()) {
        return errors.New("ability_on_cooldown")
    }

    // 4. Rate limiting
    if caster.AbilityUsageRate() > maxRate {
        return errors.New("rate_limit_exceeded")
    }

    return nil
}
```

---

## Performance Considerations

### Interest Management
Only broadcast combat events to nearby players:

```go
func (cm *CombatManager) BroadcastCombatEvent(event *CombatEvent) {
    // Get players within view distance
    nearbyPlayers := world.GetPlayersInRadius(event.Position, 50.0)

    for _, player := range nearbyPlayers {
        player.Send(event)
    }
}
```

### Cooldown Tracking
```go
// Server tracks authoritative cooldowns
type CooldownManager struct {
    cooldowns map[string]time.Time
    mu        sync.RWMutex
}

func (cm *CooldownManager) IsReady(abilityID string) bool {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    expiry, exists := cm.cooldowns[abilityID]
    if !exists {
        return true
    }
    return time.Now().After(expiry)
}
```

---

## Testing Strategy

### Unit Tests
- Each ability validates correctly
- Damage calculations are correct
- Cooldowns work properly

### Integration Tests
- Full combat flow with mock network
- Validator chain works
- Combat events broadcast correctly

### Latency Tests
- Test with simulated 50ms, 100ms, 150ms latency
- Ensure combat feels responsive
- Verify no desyncs

---

## Phase 1: Basic Melee (Current Focus)

**Abilities to Implement:**
1. **Basic Attack** - Fast, low damage, no cooldown
2. **Heavy Attack** - Slow, high damage, 3s cooldown

**Features:**
- ✅ Server-authoritative damage
- ✅ Client prediction for instant feedback
- ✅ Health bars
- ✅ Damage numbers
- ✅ Cooldown UI
- ✅ Range validation

**NOT in Phase 1:**
- Lag compensation (later)
- Status effects (later)
- Combos (later)
- Advanced animations (later)

---

## Future Phases

**Phase 2: Ranged Combat**
- Projectile abilities
- Hitscan abilities
- Lead prediction

**Phase 3: Advanced Combat**
- AoE abilities
- Status effects (stun, poison, slow)
- Lag compensation
- Combo system

**Phase 4: Polish**
- Advanced animations
- VFX system
- Sound effects
- Hit reactions

---

## Summary

This architecture provides:
- ✅ **Extensibility**: Easy to add new ability types
- ✅ **Decoupling**: Melee and ranged are independent
- ✅ **Network-aware**: Built for latency from day 1
- ✅ **Server authority**: Cheat-resistant
- ✅ **Client prediction**: Responsive gameplay
- ✅ **Testable**: Clear interfaces for unit tests

Ready to implement Phase 1: Basic Melee!
