# Combat Entity System - Extensible Architecture

## The Problem

Building combat for an MMO requires handling:
- ✅ Players fighting players
- ✅ Players fighting monsters
- ✅ Players destroying objects (crates, grass, walls)
- ✅ Monsters fighting players
- ✅ Monsters fighting monsters (rare, but possible)
- ✅ Environmental damage (lava, traps)
- ✅ Hundreds of entities in combat simultaneously
- ✅ Network efficiency (minimize messages)
- ✅ Rules: "Who can hit what?"

**Challenge**: Keep code clean, extensible, and performant.

---

## Solution: Combat Entity Interface

### Core Abstraction

**Any entity that can participate in combat implements `CombatEntity`:**

```go
// CombatEntity - ANYTHING that can be attacked or attack
type CombatEntity interface {
    // Identity
    GetID() string
    GetType() EntityType
    GetPosition() Vector3

    // Combat State
    IsAlive() bool
    CanTakeDamage() bool
    CanDealDamage() bool

    // Faction/Targeting
    GetFaction() Faction
    IsHostileTo(other CombatEntity) bool
    IsValidTarget(attacker CombatEntity) bool

    // Combat Actions
    TakeDamage(amount int, source CombatEntity) *DamageResult
    OnDeath(killer CombatEntity)

    // Network
    GetInterestRadius() float64  // How far can this entity be seen?
    ShouldBroadcastCombat() bool // Should combat be broadcast?
}
```

**Benefits**:
- ✅ Players, monsters, objects all use same interface
- ✅ Combat code doesn't care WHAT you're hitting
- ✅ Easy to add new entity types
- ✅ Type-safe with Go interfaces

---

## Entity Types

### EntityType Enum

```go
type EntityType int

const (
    EntityTypePlayer EntityType = iota
    EntityTypeMonster
    EntityTypeBoss
    EntityTypeNPC
    EntityTypeDestructible  // Crates, barrels
    EntityTypeHarvestable   // Trees, rocks, grass
    EntityTypeEnvironmental // Traps, lava
    EntityTypeProjectile    // Arrows, fireballs
)
```

### Type-Specific Implementations

#### 1. Player (CombatEntity)
```go
type Player struct {
    // ... existing fields
}

func (p *Player) GetType() EntityType { return EntityTypePlayer }
func (p *Player) GetFaction() Faction { return FactionPlayer }
func (p *Player) CanTakeDamage() bool { return p.IsAlive() }
func (p *Player) CanDealDamage() bool { return p.IsAlive() && !p.IsStunned() }
func (p *Player) IsHostileTo(other CombatEntity) bool {
    // Players can attack monsters, bosses, and hostile players
    switch other.GetType() {
    case EntityTypeMonster, EntityTypeBoss:
        return true
    case EntityTypePlayer:
        return p.IsPvPEnabled() && other.(*Player).IsPvPEnabled()
    default:
        return false
    }
}
func (p *Player) GetInterestRadius() float64 { return 50.0 }
func (p *Player) ShouldBroadcastCombat() bool { return true }
```

#### 2. Monster (CombatEntity)
```go
type Monster struct {
    ID           string
    Type         MonsterType  // Goblin, Dragon, etc.
    Position     Vector3
    Stats        *stats.StatContainer
    Faction      Faction
    AIController *AIController
}

func (m *Monster) GetType() EntityType { return EntityTypeMonster }
func (m *Monster) GetFaction() Faction { return m.Faction }
func (m *Monster) IsHostileTo(other CombatEntity) bool {
    // Monsters attack players and different-faction monsters
    switch other.GetType() {
    case EntityTypePlayer:
        return true
    case EntityTypeMonster:
        return m.Faction != other.GetFaction()
    default:
        return false
    }
}
func (m *Monster) TakeDamage(amount int, source CombatEntity) *DamageResult {
    // Monster-specific logic: drop loot, aggro, etc.
    m.Stats.ModifyCurrentValue(stats.StatHealth, float64(-amount))

    if m.Stats.GetStat(stats.StatHealth) <= 0 {
        m.OnDeath(source)
    }

    return &DamageResult{
        Target:    m.ID,
        Damage:    amount,
        Remaining: int(m.Stats.GetStat(stats.StatHealth)),
    }
}
```

#### 3. Destructible Object (CombatEntity)
```go
type DestructibleObject struct {
    ID         string
    ObjectType DestructibleType  // Crate, Barrel, Door
    Position   Vector3
    Health     int
    MaxHealth  int
    LootTable  *LootTable
}

func (d *DestructibleObject) GetType() EntityType { return EntityTypeDestructible }
func (d *DestructibleObject) GetFaction() Faction { return FactionNeutral }
func (d *DestructibleObject) CanTakeDamage() bool { return d.Health > 0 }
func (d *DestructibleObject) CanDealDamage() bool { return false }
func (d *DestructibleObject) IsHostileTo(other CombatEntity) bool { return false }
func (d *DestructibleObject) IsValidTarget(attacker CombatEntity) bool {
    // Only players can attack objects
    return attacker.GetType() == EntityTypePlayer
}
func (d *DestructibleObject) TakeDamage(amount int, source CombatEntity) *DamageResult {
    d.Health -= amount
    if d.Health <= 0 {
        d.Health = 0
        d.OnDeath(source)  // Drop loot, play break animation
    }
    return &DamageResult{
        Target:    d.ID,
        Damage:    amount,
        Remaining: d.Health,
    }
}
func (d *DestructibleObject) GetInterestRadius() float64 { return 30.0 }
func (d *DestructibleObject) ShouldBroadcastCombat() bool { return true }
```

#### 4. Harvestable Resource (CombatEntity)
```go
type HarvestableResource struct {
    ID           string
    ResourceType ResourceType  // Tree, Rock, Grass, Bush
    Position     Vector3
    Durability   int
    RequiredTool ToolType      // Axe, Pickaxe, None
}

func (h *HarvestableResource) GetType() EntityType { return EntityTypeHarvestable }
func (h *HarvestableResource) IsValidTarget(attacker CombatEntity) bool {
    // Only players with correct tool can harvest
    if attacker.GetType() != EntityTypePlayer {
        return false
    }
    player := attacker.(*Player)
    return player.HasTool(h.RequiredTool)
}
func (h *HarvestableResource) TakeDamage(amount int, source CombatEntity) *DamageResult {
    h.Durability -= amount
    if h.Durability <= 0 {
        h.OnDeath(source)  // Despawn, give resources
    }
    return &DamageResult{
        Target:    h.ID,
        Damage:    amount,
        Remaining: h.Durability,
    }
}
```

---

## Faction System - "Who Can Hit What?"

### Faction Definitions

```go
type Faction int

const (
    FactionPlayer Faction = iota
    FactionHostileMob
    FactionPassiveMob
    FactionNeutral
    FactionAllied
    FactionGuild1  // Guild-specific factions
    FactionGuild2
    // ... more factions
)

// FactionRelations defines hostility rules
var FactionRelations = map[Faction]map[Faction]bool{
    FactionPlayer: {
        FactionHostileMob: true,   // Players can attack hostile mobs
        FactionPassiveMob: false,  // Players can't attack passive mobs
        FactionNeutral:    false,  // Can't attack neutral
        FactionPlayer:     false,  // Can't attack other players (unless PvP)
    },
    FactionHostileMob: {
        FactionPlayer:     true,   // Hostile mobs attack players
        FactionPassiveMob: false,
        FactionNeutral:    false,
    },
    FactionPassiveMob: {
        FactionPlayer:     false,  // Passive mobs don't attack anyone
        FactionHostileMob: false,
    },
    FactionNeutral: {
        FactionPlayer:     false,  // Neutral doesn't attack
        FactionHostileMob: false,
    },
}
```

### Hostility Check

```go
func IsHostile(attacker, target CombatEntity) bool {
    // Check faction relations
    attackerFaction := attacker.GetFaction()
    targetFaction := target.GetFaction()

    if relations, ok := FactionRelations[attackerFaction]; ok {
        if hostile, ok := relations[targetFaction]; ok && hostile {
            return true
        }
    }

    // Special case: PvP
    if attacker.GetType() == EntityTypePlayer && target.GetType() == EntityTypePlayer {
        p1 := attacker.(*Player)
        p2 := target.(*Player)
        return p1.IsPvPEnabled() && p2.IsPvPEnabled()
    }

    // Special case: Guild wars
    if attacker.GetFaction() >= FactionGuild1 && target.GetFaction() >= FactionGuild1 {
        return IsGuildWar(attacker.GetFaction(), target.GetFaction())
    }

    return false
}
```

### Target Validation

```go
func CanAttack(attacker CombatEntity, target CombatEntity) error {
    // 1. Basic state checks
    if !attacker.CanDealDamage() {
        return errors.New("attacker_cannot_attack")
    }
    if !target.CanTakeDamage() {
        return errors.New("target_cannot_take_damage")
    }
    if !target.IsAlive() {
        return errors.New("target_already_dead")
    }

    // 2. Target validation (entity-specific)
    if !target.IsValidTarget(attacker) {
        return errors.New("invalid_target")
    }

    // 3. Hostility check
    if !IsHostile(attacker, target) {
        return errors.New("not_hostile")
    }

    return nil
}
```

---

## Spatial Partitioning - Handling Many Entities

### Problem
- 1000 entities in world
- Each attack needs to check nearby entities
- O(n²) = 1,000,000 checks per frame = **BAD**

### Solution: Spatial Grid

```go
// SpatialGrid divides world into cells for fast queries
type SpatialGrid struct {
    cellSize float64
    cells    map[GridCell][]*CombatEntity
    mu       sync.RWMutex
}

type GridCell struct {
    X int
    Z int
}

func (sg *SpatialGrid) GetCell(pos Vector3) GridCell {
    return GridCell{
        X: int(pos.X / sg.cellSize),
        Z: int(pos.Z / sg.cellSize),
    }
}

func (sg *SpatialGrid) AddEntity(entity CombatEntity) {
    sg.mu.Lock()
    defer sg.mu.Unlock()

    cell := sg.GetCell(entity.GetPosition())
    sg.cells[cell] = append(sg.cells[cell], &entity)
}

func (sg *SpatialGrid) GetNearbyEntities(pos Vector3, radius float64) []CombatEntity {
    sg.mu.RLock()
    defer sg.mu.RUnlock()

    // Calculate which cells to check
    minCell := sg.GetCell(Vector3{pos.X - radius, 0, pos.Z - radius})
    maxCell := sg.GetCell(Vector3{pos.X + radius, 0, pos.Z + radius})

    nearby := []CombatEntity{}

    // Check only relevant cells (not entire world!)
    for x := minCell.X; x <= maxCell.X; x++ {
        for z := minCell.Z; z <= maxCell.Z; z++ {
            cell := GridCell{X: x, Z: z}
            for _, entity := range sg.cells[cell] {
                if pos.Distance((*entity).GetPosition()) <= radius {
                    nearby = append(nearby, *entity)
                }
            }
        }
    }

    return nearby
}

// Performance: O(k) where k = entities in nearby cells (~10-50)
// Instead of O(n) where n = all entities in world (~1000+)
```

### Usage in Combat

```go
func (cm *CombatManager) FindTargetsInRadius(attacker CombatEntity, radius float64) []CombatEntity {
    // Fast: Only checks nearby cells
    nearby := cm.spatialGrid.GetNearbyEntities(attacker.GetPosition(), radius)

    targets := []CombatEntity{}
    for _, entity := range nearby {
        if CanAttack(attacker, entity) == nil {
            targets = append(targets, entity)
        }
    }

    return targets
}
```

---

## Network Optimization

### Problem: Message Explosion

**Scenario**: 100 players fighting 200 monsters
- Combat events per second: 300+ (lots of attacks)
- Each event broadcast to 100 players = 30,000 messages/sec
- **Unacceptable**

### Solution 1: Interest Management

**Only send combat events to players who can see them**

```go
type InterestManager struct {
    playerInterests map[string]*InterestSet  // player_id -> entities they care about
}

type InterestSet struct {
    NearbyEntities map[string]CombatEntity
    Radius         float64
}

func (im *InterestManager) GetInterestedPlayers(event *CombatEvent) []*Player {
    eventPos := event.Position

    interested := []*Player{}
    for playerID, interest := range im.playerInterests {
        // Is combat event within player's interest radius?
        playerPos := interest.GetPlayerPosition()
        if eventPos.Distance(playerPos) <= interest.Radius {
            interested = append(interested, GetPlayer(playerID))
        }
    }

    return interested
}

func (cm *CombatManager) BroadcastCombatEvent(event *CombatEvent) {
    // Don't broadcast to everyone!
    interestedPlayers := cm.interestManager.GetInterestedPlayers(event)

    // Send only to nearby players
    for _, player := range interestedPlayers {
        player.SendCombatEvent(event)
    }
}
```

**Result**: 30,000 messages/sec → ~1,500 messages/sec (20x reduction!)

### Solution 2: Event Aggregation

**Batch multiple events into one message**

```go
type CombatEventBatch struct {
    Events    []*CombatEvent
    Timestamp time.Time
}

type CombatEventAggregator struct {
    pendingEvents map[string][]*CombatEvent  // player_id -> events
    ticker        *time.Ticker
}

func (cea *CombatEventAggregator) AddEvent(event *CombatEvent, playerID string) {
    cea.pendingEvents[playerID] = append(cea.pendingEvents[playerID], event)
}

func (cea *CombatEventAggregator) Run() {
    cea.ticker = time.NewTicker(50 * time.Millisecond)  // Batch every 50ms

    for range cea.ticker.C {
        cea.flush()
    }
}

func (cea *CombatEventAggregator) flush() {
    for playerID, events := range cea.pendingEvents {
        if len(events) == 0 {
            continue
        }

        // Send batch
        batch := &CombatEventBatch{
            Events:    events,
            Timestamp: time.Now(),
        }

        SendToPlayer(playerID, batch)

        // Clear
        cea.pendingEvents[playerID] = nil
    }
}
```

**Result**: 1,500 messages/sec → ~30 batches/sec (50x reduction!)

### Solution 3: Priority System

**Send important events immediately, batch less important ones**

```go
type CombatEventPriority int

const (
    PriorityImmediate CombatEventPriority = iota  // Local player damaged
    PriorityHigh                                   // Nearby player damaged
    PriorityMedium                                 // Monster damaged
    PriorityLow                                    // Distant combat
)

func (cm *CombatManager) SendCombatEvent(event *CombatEvent, player *Player) {
    priority := cm.CalculatePriority(event, player)

    switch priority {
    case PriorityImmediate:
        // Send immediately
        player.SendCombatEvent(event)
    case PriorityHigh:
        // Send in next batch (within 50ms)
        cm.aggregator.AddEvent(event, player.ID)
    case PriorityMedium, PriorityLow:
        // Send in slower batch (within 200ms)
        cm.slowAggregator.AddEvent(event, player.ID)
    }
}

func (cm *CombatManager) CalculatePriority(event *CombatEvent, player *Player) CombatEventPriority {
    // Is it the local player?
    if event.TargetID == player.ID || event.CasterID == player.ID {
        return PriorityImmediate
    }

    // Is it nearby?
    distance := event.Position.Distance(player.GetPosition())
    if distance < 10.0 {
        return PriorityHigh
    } else if distance < 30.0 {
        return PriorityMedium
    }

    return PriorityLow
}
```

---

## Clean Message Protocol

### Problem: Message Type Explosion

As we add features:
- Player attacks
- Monster attacks
- Environmental damage
- DoT (damage over time)
- Healing
- Buffs
- Debuffs
- ...message types explode!

### Solution: Unified Combat Event

```go
// Single message type for ALL combat events
type CombatEvent struct {
    Type      CombatEventType
    EventID   string
    Timestamp time.Time

    // Participants (generic entity IDs)
    CasterID  string
    TargetID  string

    // Location
    Position Vector3

    // Combat Data
    AbilityID string
    Damage    int
    Healing   int

    // Result
    TargetHealth    int
    TargetMaxHealth int
    TargetDied      bool

    // Effects
    BuffsApplied   []string
    DebuffsApplied []string

    // Metadata
    IsCritical bool
    WasDodged  bool
    WasBlocked bool
}

type CombatEventType int

const (
    CombatEventDamage CombatEventType = iota
    CombatEventHeal
    CombatEventBuff
    CombatEventDebuff
    CombatEventDeath
    CombatEventDodge
    CombatEventBlock
    CombatEventMiss
)
```

**Benefits**:
- ✅ One message type handles everything
- ✅ Easy to extend (add fields, not types)
- ✅ Client code simpler
- ✅ Network serialization efficient

### Message Examples

```json
// Player attacks monster
{
  "type": "damage",
  "event_id": "evt_123",
  "caster_id": "player_456",
  "target_id": "monster_789",
  "position": {"x": 10, "y": 0, "z": 5},
  "ability_id": "basic_attack",
  "damage": 25,
  "target_health": 75,
  "target_max_health": 100,
  "is_critical": false
}

// Player breaks crate
{
  "type": "damage",
  "caster_id": "player_456",
  "target_id": "crate_101",
  "position": {"x": 12, "y": 0, "z": 6},
  "damage": 50,
  "target_health": 0,
  "target_died": true
}

// Monster attacks player
{
  "type": "damage",
  "caster_id": "monster_789",
  "target_id": "player_456",
  "ability_id": "monster_bite",
  "damage": 15,
  "target_health": 85,
  "was_dodged": false
}

// Environmental damage (lava)
{
  "type": "damage",
  "caster_id": "environment",
  "target_id": "player_456",
  "damage": 10,
  "debuffs_applied": ["burning"]
}
```

---

## Combat Manager Architecture

### High-Level Structure

```go
type CombatManager struct {
    // Entity Management
    entities      map[string]CombatEntity
    spatialGrid   *SpatialGrid

    // Ability System
    abilityRegistry *AbilityRegistry
    validatorChain  *ValidatorChain

    // Targeting
    factionRelations map[Faction]map[Faction]bool

    // Network Optimization
    interestManager *InterestManager
    eventAggregator *CombatEventAggregator

    // Configuration
    maxEntitiesPerCell int
    interestRadius     float64

    mu sync.RWMutex
}
```

### Core Methods

```go
// RegisterEntity adds an entity to combat tracking
func (cm *CombatManager) RegisterEntity(entity CombatEntity) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    cm.entities[entity.GetID()] = entity
    cm.spatialGrid.AddEntity(entity)
}

// UnregisterEntity removes an entity
func (cm *CombatManager) UnregisterEntity(entityID string) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    delete(cm.entities, entityID)
    cm.spatialGrid.RemoveEntity(entityID)
}

// ExecuteAbility processes an ability use
func (cm *CombatManager) ExecuteAbility(
    casterID string,
    targetID string,
    abilityID string,
) (*CombatEvent, error) {

    // 1. Get entities
    caster := cm.entities[casterID]
    target := cm.entities[targetID]

    // 2. Get ability
    ability := cm.abilityRegistry.Get(abilityID)

    // 3. Validate (range, cooldown, resources, etc.)
    if err := cm.validatorChain.Validate(caster, target, ability); err != nil {
        return nil, err
    }

    // 4. Check targeting rules
    if err := CanAttack(caster, target); err != nil {
        return nil, err
    }

    // 5. Execute ability
    result := ability.Execute(caster, target)

    // 6. Apply damage
    damageResult := target.TakeDamage(result.Damage, caster)

    // 7. Create combat event
    event := &CombatEvent{
        Type:            CombatEventDamage,
        CasterID:        casterID,
        TargetID:        targetID,
        AbilityID:       abilityID,
        Damage:          damageResult.Damage,
        TargetHealth:    damageResult.Remaining,
        TargetMaxHealth: damageResult.Max,
        TargetDied:      damageResult.Remaining <= 0,
        IsCritical:      result.CriticalHit,
        Position:        target.GetPosition(),
        Timestamp:       time.Now(),
    }

    // 8. Broadcast to interested players
    cm.BroadcastCombatEvent(event)

    return event, nil
}

// FindTargetsInRadius for AoE abilities
func (cm *CombatManager) FindTargetsInRadius(
    casterID string,
    position Vector3,
    radius float64,
) []CombatEntity {
    caster := cm.entities[casterID]

    // Fast spatial query
    nearby := cm.spatialGrid.GetNearbyEntities(position, radius)

    // Filter to valid targets
    targets := []CombatEntity{}
    for _, entity := range nearby {
        if CanAttack(caster, entity) == nil {
            targets = append(targets, entity)
        }
    }

    return targets
}
```

---

## Configuration System

### Combat Rules Configuration

```go
type CombatConfig struct {
    // Entity Limits
    MaxEntitiesPerRegion int
    MaxCombatantsPerCell int

    // Network
    InterestRadius        float64
    EventBatchInterval    time.Duration
    MaxEventsPerBatch     int

    // Spatial
    SpatialCellSize       float64

    // Combat Rules
    FriendlyFireEnabled   bool
    PvPEnabled            bool
    MonsterInfighting     bool

    // Performance
    MaxCombatEventsPerSec int
}

var DefaultCombatConfig = &CombatConfig{
    MaxEntitiesPerRegion:  1000,
    MaxCombatantsPerCell:  50,
    InterestRadius:        50.0,
    EventBatchInterval:    50 * time.Millisecond,
    MaxEventsPerBatch:     20,
    SpatialCellSize:       10.0,
    FriendlyFireEnabled:   false,
    PvPEnabled:            true,
    MonsterInfighting:     false,
    MaxCombatEventsPerSec: 100,
}
```

---

## Summary: Solving All Your Concerns

| Concern | Solution |
|---------|----------|
| **Abstraction** | `CombatEntity` interface - works for ANY entity |
| **Network Efficiency** | Interest management + Event aggregation + Priority |
| **Many Entities** | Spatial grid partitioning (O(k) instead of O(n)) |
| **Hit Rules** | Faction system + `IsHostileTo()` + `IsValidTarget()` |
| **Message Complexity** | Single `CombatEvent` type, extensible with fields |
| **Extensibility** | OCP: Add entity types without modifying core |

---

## Next Steps

1. Implement `CombatEntity` interface
2. Create entity implementations (Player, Monster, Destructible)
3. Implement `SpatialGrid`
4. Implement `CombatManager`
5. Add network optimization layers
6. Test with 100+ entities

**Ready to implement?**
