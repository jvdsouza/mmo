# Combat System Implementation Plan

## Your Questions Answered

### Q1: "How do we optimize with networks?"
**Answer**: 3-layer optimization strategy

1. **Interest Management** - Only send events to nearby players (20x reduction)
2. **Event Aggregation** - Batch events every 50ms (50x reduction)
3. **Priority System** - Immediate for local player, batched for distant

**Result**: 30,000 msgs/sec → 30 batches/sec = **1000x improvement**

---

### Q2: "Should this be abstract for ANY combatable entity?"
**Answer**: YES! Use `CombatEntity` interface

```go
type CombatEntity interface {
    GetID() string
    GetType() EntityType
    IsAlive() bool
    CanTakeDamage() bool
    CanDealDamage() bool
    GetFaction() Faction
    IsHostileTo(other CombatEntity) bool
    TakeDamage(amount int, source CombatEntity) *DamageResult
    // ...
}
```

**Supports**:
- ✅ Players
- ✅ Monsters
- ✅ Bosses
- ✅ Destructible objects (crates, barrels)
- ✅ Harvestable resources (trees, grass, rocks)
- ✅ Environmental damage (traps, lava)
- ✅ Future: Pets, summons, vehicles, etc.

---

### Q3: "Do we pin an interface to types in the combat scene?"
**Answer**: YES! `CombatEntity` interface is the contract

**Pattern**:
```go
// Combat code works with interface
func (cm *CombatManager) ExecuteAbility(
    caster CombatEntity,    // Could be ANY entity
    target CombatEntity,    // Could be ANY entity
    ability Ability,
) *CombatEvent {
    // Combat logic doesn't care WHAT caster/target are!
    // Just that they implement CombatEntity interface
}

// Type-specific logic stays in entity implementations
func (p *Player) TakeDamage(amount int, source CombatEntity) *DamageResult {
    // Player-specific: update stats, check death, etc.
}

func (m *Monster) TakeDamage(amount int, source CombatEntity) *DamageResult {
    // Monster-specific: drop loot, update AI aggro, etc.
}
```

---

### Q4: "How do we handle many entities (configurable amount)?"
**Answer**: Spatial Grid Partitioning

**Problem**: Checking 1000 entities = O(n²) = 1,000,000 checks ❌

**Solution**: Divide world into grid cells

```
World divided into 10x10 meter cells:

[Cell A] [Cell B] [Cell C]
[Cell D] [Cell E] [Cell F]
[Cell G] [Cell H] [Cell I]

Player attacks in Cell E:
- Only check Cell E and neighbors (D, F, B, H)
- ~10-50 entities instead of 1000
- O(k) where k = nearby entities ✅
```

**Configuration**:
```go
config := &CombatConfig{
    MaxEntitiesPerRegion:  1000,   // Total entities
    MaxCombatantsPerCell:  50,     // Per grid cell
    SpatialCellSize:       10.0,   // Cell size in meters
    InterestRadius:        50.0,   // Network interest
}
```

**Performance**: 1,000,000 checks → ~500 checks = **2000x faster**

---

### Q5: "How do we assign what entity can hit what?"
**Answer**: Faction System + Validation Methods

#### **Faction System**:
```go
type Faction int

const (
    FactionPlayer
    FactionHostileMob
    FactionPassiveMob
    FactionNeutral
    FactionGuild1
    // ...
)

// Define relations
FactionRelations[FactionPlayer][FactionHostileMob] = true  // Can attack
FactionRelations[FactionPlayer][FactionPassiveMob] = false // Can't attack
```

#### **Entity-Specific Validation**:
```go
// Player can attack monsters and hostile players
func (p *Player) IsHostileTo(other CombatEntity) bool {
    switch other.GetType() {
    case EntityTypeMonster:
        return true
    case EntityTypePlayer:
        return p.IsPvPEnabled() && other.(*Player).IsPvPEnabled()
    default:
        return false
    }
}

// Destructible objects can only be hit by players
func (d *DestructibleObject) IsValidTarget(attacker CombatEntity) bool {
    return attacker.GetType() == EntityTypePlayer
}

// Trees require axe
func (h *HarvestableResource) IsValidTarget(attacker CombatEntity) bool {
    if attacker.GetType() != EntityTypePlayer {
        return false
    }
    return attacker.(*Player).HasTool(h.RequiredTool)
}
```

**Rules are encoded IN the entity types**, not scattered across combat code!

---

### Q6: "How do we avoid message complexity?"
**Answer**: Single extensible `CombatEvent` message

**Instead of**:
```
❌ PlayerAttackMessage
❌ MonsterAttackMessage
❌ EnvironmentalDamageMessage
❌ DotDamageMessage
❌ HealMessage
❌ BuffMessage
❌ ... 50+ message types!
```

**Use one**:
```go
✅ CombatEvent {
    Type: CombatEventType  // damage/heal/buff/etc
    CasterID: string       // ANY entity
    TargetID: string       // ANY entity
    Damage: int
    Healing: int
    BuffsApplied: []string
    // ... extensible fields
}
```

**Handle ALL cases**:
- Player attacks monster
- Monster attacks player
- Player breaks crate
- Lava damages player
- Potion heals player
- Buff applied
- DoT tick

**One message type, clean client code!**

---

## Architecture Layers

### Layer 1: Entities
```
CombatEntity (interface)
    ↑
    ├── Player
    ├── Monster
    ├── Boss
    ├── DestructibleObject
    ├── HarvestableResource
    └── EnvironmentalHazard
```

### Layer 2: Combat Logic
```
CombatManager
    ├── AbilityRegistry (abilities)
    ├── ValidatorChain (validation)
    ├── SpatialGrid (fast queries)
    └── Faction System (targeting rules)
```

### Layer 3: Network Optimization
```
InterestManager (who sees what)
    ↓
EventAggregator (batch events)
    ↓
PrioritySystem (immediate vs batched)
    ↓
Clients (receive optimized updates)
```

---

## File Structure

```
server/
├── combat/
│   ├── entity.go              # CombatEntity interface
│   ├── faction.go             # Faction system
│   ├── combat_manager.go      # Core combat logic
│   ├── spatial_grid.go        # Spatial partitioning
│   ├── ability.go             # [existing]
│   ├── validators.go          # [existing]
│   └── melee_ability.go       # [next to implement]
├── entities/
│   ├── player.go              # Player as CombatEntity
│   ├── monster.go             # Monster implementation
│   ├── destructible.go        # Crates, barrels
│   └── harvestable.go         # Trees, rocks
├── network/
│   ├── interest_manager.go    # Interest management
│   ├── event_aggregator.go    # Event batching
│   └── combat_events.go       # Message protocol
└── world/
    └── region.go              # World regions
```

---

## Implementation Order

### Phase 1: Core Entities ✅ (Started)
- [x] Stats system (OCP)
- [x] Combat interfaces
- [ ] CombatEntity interface
- [ ] Player as CombatEntity
- [ ] Monster implementation

### Phase 2: Basic Combat
- [ ] Melee ability implementation
- [ ] Combat manager
- [ ] Faction system
- [ ] Target validation

### Phase 3: Optimization
- [ ] Spatial grid
- [ ] Interest manager
- [ ] Event aggregator

### Phase 4: Entity Variety
- [ ] Destructible objects
- [ ] Harvestable resources
- [ ] Environmental hazards

### Phase 5: Advanced
- [ ] AoE abilities
- [ ] DoT/HoT effects
- [ ] Projectiles
- [ ] Boss mechanics

---

## Design Principles Applied

### 1. Open-Closed Principle (OCP)
**Closed**: `CombatEntity` interface, `CombatManager` core
**Open**: Add new entity types by implementing interface

### 2. Interface Segregation
**Small, focused interfaces**:
- `CombatEntity` for combat
- `Movable` for movement
- `Lootable` for drops
- Entity can implement multiple!

### 3. Dependency Inversion
**Combat depends on abstractions**:
```go
// Combat code depends on interface, not concrete types
func Attack(attacker CombatEntity, target CombatEntity) {
    // Works with ANY CombatEntity
}
```

### 4. Single Responsibility
**Each component has ONE job**:
- `CombatEntity`: Define combat contract
- `SpatialGrid`: Fast spatial queries
- `InterestManager`: Determine who sees what
- `EventAggregator`: Batch messages

### 5. Composition Over Inheritance
**Go doesn't have inheritance**, so we compose:
```go
type Monster struct {
    Stats        *stats.StatContainer  // Compose stats
    AIController *AIController         // Compose AI
    Loot         *LootTable            // Compose loot
}
```

---

## Performance Targets

| Metric | Target | How |
|--------|--------|-----|
| **Entities per region** | 1000+ | Spatial grid |
| **Combat events/sec** | 300+ | Event aggregation |
| **Network msgs/sec** | <100 | Interest management |
| **Target query time** | <1ms | O(k) spatial lookup |
| **Combat latency** | <50ms | Client prediction |

---

## Configuration Example

```go
// Development: Small, intense combat
devConfig := &CombatConfig{
    MaxEntitiesPerRegion:  100,
    InterestRadius:        30.0,
    EventBatchInterval:    16 * time.Millisecond,  // 60Hz
    SpatialCellSize:       5.0,
    PvPEnabled:            true,
}

// Production: Large-scale MMO
prodConfig := &CombatConfig{
    MaxEntitiesPerRegion:  5000,
    InterestRadius:        50.0,
    EventBatchInterval:    50 * time.Millisecond,  // 20Hz
    SpatialCellSize:       10.0,
    PvPEnabled:            true,
}
```

---

## Key Takeaways

### ✅ **Abstraction**: `CombatEntity` interface handles everything
- Players, monsters, objects, environments
- Add new types without modifying combat code

### ✅ **Network Efficiency**: 3-layer optimization
- Interest management (20x)
- Event aggregation (50x)
- Priority system (selective batching)

### ✅ **Scalability**: Spatial partitioning
- O(k) instead of O(n²)
- Handles 1000+ entities easily

### ✅ **Targeting**: Faction + validation
- Faction system defines general rules
- Entity types define specific rules
- Clean, maintainable

### ✅ **Clean Messages**: Single event type
- Extensible with fields
- Not with message types
- Easy to serialize

---

## Next Action

**Ready to implement?**

**Option 1**: Start with `CombatEntity` interface + basic entities
**Option 2**: I can continue implementing while you review docs
**Option 3**: Discuss specific design decisions first

**Recommended**: Start with Phase 1 (Core Entities), then test with simple combat before adding optimizations.

What would you like to do?
