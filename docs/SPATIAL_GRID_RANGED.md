# Spatial Grid Optimization for Ranged Combat

## The Challenge

**Melee**: Check 2-3 meter radius → ~1 grid cell
**Ranged**: Check 20-50 meter radius → ~25-100 grid cells

**Plus**:
- Projectiles travel over time
- Need line-of-sight checks
- Target may move during flight
- Collision detection along path

---

## How Spatial Grid Handles Ranged

### 1. Initial Target Query (Same Pattern, Bigger Radius)

```go
// Melee: Check small radius
func (cm *CombatManager) FindMeleeTargets(attacker CombatEntity) []CombatEntity {
    range := 3.0  // 3 meters
    return cm.spatialGrid.GetNearbyEntities(attacker.GetPosition(), range)
    // Checks ~1-4 cells
}

// Ranged: Check larger radius
func (cm *CombatManager) FindRangedTargets(attacker CombatEntity, weapon *RangedWeapon) []CombatEntity {
    range := weapon.GetRange()  // 20-50 meters
    return cm.spatialGrid.GetNearbyEntities(attacker.GetPosition(), range)
    // Checks ~25-100 cells
}
```

**Key**: Grid still works! Just queries more cells.

**Performance**:
```
Melee:  Check 1-4 cells   = ~10-40 entities   (vs 1000 total) ✅
Ranged: Check 25-100 cells = ~100-500 entities (vs 1000 total) ✅

Still much better than checking ALL 1000 entities!
```

---

### 2. Line-of-Sight Check (Ray Casting Through Grid)

**Problem**: Can't shoot through walls/obstacles

**Solution**: Ray-cast through grid cells

```go
// Check if attacker can see target (no obstacles)
func (sg *SpatialGrid) HasLineOfSight(from, to Vector3) bool {
    // 1. Get all cells along the ray
    cells := sg.RaycastCells(from, to)

    // 2. Check each cell for obstacles
    for _, cell := range cells {
        for _, entity := range sg.cells[cell] {
            if entity.GetType() == EntityTypeObstacle {
                // Check if obstacle blocks ray
                if entity.IntersectsRay(from, to) {
                    return false  // Blocked!
                }
            }
        }
    }

    return true  // Clear shot!
}

// Get all grid cells along a ray (Bresenham's line algorithm)
func (sg *SpatialGrid) RaycastCells(from, to Vector3) []GridCell {
    cells := []GridCell{}

    // Convert world positions to grid coordinates
    x0 := int(from.X / sg.cellSize)
    z0 := int(from.Z / sg.cellSize)
    x1 := int(to.X / sg.cellSize)
    z1 := int(to.Z / sg.cellSize)

    // Bresenham's line algorithm in 2D
    dx := abs(x1 - x0)
    dz := abs(z1 - z0)
    sx := sign(x1 - x0)
    sz := sign(z1 - z0)
    err := dx - dz

    x, z := x0, z0

    for {
        cells = append(cells, GridCell{X: x, Z: z})

        if x == x1 && z == z1 {
            break
        }

        e2 := 2 * err
        if e2 > -dz {
            err -= dz
            x += sx
        }
        if e2 < dx {
            err += dx
            z += sz
        }
    }

    return cells
}
```

**Performance**: Only checks cells along the ray, not entire area!

```
Without grid: Check all 1000 entities for obstacles ❌
With grid:    Check ~5-20 cells along ray ✅
```

---

### 3. Projectile Movement (Moving Entity in Grid)

#### **Pattern A: Instant Hit (Hitscan)**

Used for: Guns, lasers, instant spells

```go
type HitscanAbility struct {
    BaseAbility
    Range float64
}

func (h *HitscanAbility) Execute(caster, target CombatEntity) *CombatResult {
    // 1. Validate range
    distance := caster.GetPosition().Distance(target.GetPosition())
    if distance > h.Range {
        return &CombatResult{Hit: false, Reason: "out_of_range"}
    }

    // 2. Check line of sight
    if !spatialGrid.HasLineOfSight(caster.GetPosition(), target.GetPosition()) {
        return &CombatResult{Hit: false, Reason: "blocked"}
    }

    // 3. Instant hit!
    damage := CalculateDamage(caster, target, h)
    return &CombatResult{
        Hit:    true,
        Damage: damage,
    }
}
```

**Network**: Send hit immediately (no travel time)

---

#### **Pattern B: Projectile (Physics-Based)**

Used for: Arrows, fireballs, thrown objects

```go
type Projectile struct {
    ID           string
    CasterID     string
    TargetID     string
    Position     Vector3
    Velocity     Vector3
    Speed        float64
    Range        float64
    Damage       int
    StartPos     Vector3
    StartTime    time.Time
    LastCell     GridCell
}

func (p *Projectile) Update(deltaTime float64, grid *SpatialGrid) *CombatResult {
    // 1. Move projectile
    oldPos := p.Position
    p.Position.X += p.Velocity.X * p.Speed * deltaTime
    p.Position.Y += p.Velocity.Y * p.Speed * deltaTime
    p.Position.Z += p.Velocity.Z * p.Speed * deltaTime

    // 2. Check if crossed into new cell
    newCell := grid.GetCell(p.Position)
    if newCell != p.LastCell {
        // 3. Update grid tracking
        grid.MoveEntity(p, p.LastCell, newCell)
        p.LastCell = newCell

        // 4. Check for hits in new cell
        if hit := p.CheckCollisions(grid, oldPos, p.Position); hit != nil {
            return hit
        }
    }

    // 5. Check range limit
    distanceTraveled := p.Position.Distance(p.StartPos)
    if distanceTraveled > p.Range {
        return &CombatResult{Hit: false, Reason: "max_range"}
    }

    return nil  // Still flying
}

func (p *Projectile) CheckCollisions(grid *SpatialGrid, from, to Vector3) *CombatResult {
    // Get entities in current cell
    cell := grid.GetCell(p.Position)
    entities := grid.cells[cell]

    for _, entity := range entities {
        // Skip caster
        if entity.GetID() == p.CasterID {
            continue
        }

        // Check if projectile hit entity
        if p.IntersectsEntity(entity, from, to) {
            // HIT!
            return &CombatResult{
                Hit:      true,
                TargetID: entity.GetID(),
                Damage:   p.Damage,
                Position: p.Position,
            }
        }
    }

    return nil  // No hit
}
```

**Key**: Projectile moves through grid cells, only checking current cell for hits!

**Performance**:
```
Without grid: Check all 1000 entities every frame ❌
With grid:    Check ~10-20 entities in current cell ✅
```

---

### 4. Moving Target Prediction

**Problem**: Target moves during projectile flight

**Solution**: Predictive targeting

```go
// Calculate where to aim based on target velocity
func PredictTargetPosition(target CombatEntity, projectileSpeed float64, shooterPos Vector3) Vector3 {
    targetPos := target.GetPosition()
    targetVel := target.GetVelocity()  // Moving target

    // Calculate time for projectile to reach target
    distance := shooterPos.Distance(targetPos)
    timeToHit := distance / projectileSpeed

    // Predict where target will be
    predictedPos := Vector3{
        X: targetPos.X + targetVel.X*timeToHit,
        Y: targetPos.Y + targetVel.Y*timeToHit,
        Z: targetPos.Z + targetVel.Z*timeToHit,
    }

    return predictedPos
}

// Use in ranged ability
func (r *RangedAbility) Execute(caster, target CombatEntity) *CombatResult {
    // Aim at predicted position
    predictedPos := PredictTargetPosition(target, r.ProjectileSpeed, caster.GetPosition())

    // Create projectile aiming at prediction
    projectile := &Projectile{
        Position: caster.GetPosition(),
        Velocity: predictedPos.Subtract(caster.GetPosition()).Normalize(),
        Speed:    r.ProjectileSpeed,
        TargetID: target.GetID(),
    }

    // Spawn projectile
    world.AddProjectile(projectile)

    return &CombatResult{Projectile: projectile}
}
```

**Network**: Client predicts locally, server confirms

---

### 5. Area-of-Effect (AoE) Ranged

**Example**: Fireball that explodes in an area

```go
type AoEProjectile struct {
    Projectile
    ExplosionRadius float64
}

func (aoe *AoEProjectile) OnHit(hitPosition Vector3, grid *SpatialGrid) []*CombatResult {
    // 1. Find all entities in explosion radius
    targets := grid.GetNearbyEntities(hitPosition, aoe.ExplosionRadius)

    results := []*CombatResult{}

    // 2. Damage each target
    for _, target := range targets {
        // Skip caster
        if target.GetID() == aoe.CasterID {
            continue
        }

        // Calculate distance-based damage
        distance := hitPosition.Distance(target.GetPosition())
        damageFalloff := 1.0 - (distance / aoe.ExplosionRadius)
        finalDamage := int(float64(aoe.Damage) * damageFalloff)

        result := &CombatResult{
            Hit:      true,
            TargetID: target.GetID(),
            Damage:   finalDamage,
            Position: hitPosition,
        }
        results = append(results, result)
    }

    return results
}
```

**Performance**: Grid finds AoE targets efficiently!

```
Without grid: Check all 1000 entities ❌
With grid:    Check ~20-50 in radius ✅
```

---

## Grid Cell Size Optimization

### **Rule of Thumb**

```
Cell Size = Max Melee Range * 2

Example: Melee = 3m → Cell Size = 6-10m
```

**Why?**
- Melee checks 1 cell + neighbors (9 total)
- Ranged checks multiple cells (but still fewer than all)

### **Adaptive Cell Size**

```go
type SpatialGrid struct {
    baseCellSize float64
    // Different cell sizes for different queries
}

// For melee queries
func (sg *SpatialGrid) GetMeleeTargets(pos Vector3) []CombatEntity {
    cellSize := sg.baseCellSize  // 10m cells
    return sg.query(pos, 3.0, cellSize)
}

// For ranged queries (larger cells = fewer checks)
func (sg *SpatialGrid) GetRangedTargets(pos Vector3, range float64) []CombatEntity {
    // Use hierarchical grid or just accept checking more cells
    cellSize := sg.baseCellSize
    return sg.query(pos, range, cellSize)
}
```

---

## Network Optimization for Ranged

### **Pattern 1: Hitscan (Immediate)**

```json
// Client → Server
{
  "type": "use_ability",
  "ability_id": "sniper_shot",
  "target_id": "monster_5",
  "timestamp": 12345
}

// Server → Clients (immediate)
{
  "type": "combat_event",
  "caster_id": "player_1",
  "target_id": "monster_5",
  "ability_id": "sniper_shot",
  "hit": true,
  "damage": 50
}
```

**Latency**: ~100ms round trip (same as melee)

---

### **Pattern 2: Projectile (Simulated)**

#### **Server-Side Simulation**

```json
// Client → Server
{
  "type": "use_ability",
  "ability_id": "fireball",
  "target_position": {"x": 20, "y": 0, "z": 15}
}

// Server → Clients (projectile spawned)
{
  "type": "projectile_spawned",
  "projectile_id": "proj_123",
  "caster_id": "player_1",
  "start_pos": {"x": 0, "y": 1, "z": 0},
  "velocity": {"x": 0.8, "y": 0.1, "z": 0.6},
  "speed": 20.0
}

// Server → Clients (projectile hit, 1 second later)
{
  "type": "projectile_hit",
  "projectile_id": "proj_123",
  "target_id": "monster_5",
  "hit_position": {"x": 19.5, "y": 0.5, "z": 14.8},
  "damage": 40
}
```

**Key**: Server simulates projectile authoritatively, clients visualize

---

### **Pattern 3: Client Prediction (Optimistic)**

```json
// Client predicts projectile locally (instant feedback)
// Client spawns visual projectile immediately

// Server confirms (100ms later)
{
  "type": "projectile_confirmed",
  "projectile_id": "proj_123",
  "correction": {"x": 0.2, "y": 0.0, "z": -0.1}  // Small correction
}

// If server rejects (rare):
{
  "type": "projectile_rejected",
  "projectile_id": "proj_123",
  "reason": "target_out_of_range"
}
// Client removes visual projectile
```

---

## Performance Comparison

### **Melee vs Ranged Grid Performance**

| Operation | Melee | Ranged | Without Grid |
|-----------|-------|--------|--------------|
| **Target Query** | 1-4 cells | 25-100 cells | ALL entities |
| **Entities Checked** | ~10-40 | ~100-500 | ~1000 |
| **Line of Sight** | Not needed | 5-20 cells | ALL entities |
| **Projectile Update** | N/A | 1 cell/frame | ALL entities/frame |

**Result**: Even ranged is **10-100x faster** than no grid!

---

## Example: Complete Ranged Combat Flow

### **Archer Shooting Monster**

```go
// 1. Player aims at monster
archer := GetPlayer("player_1")
target := GetMonster("monster_5")

// 2. Find target using spatial grid
range := archer.GetWeapon().GetRange()  // 30m
possibleTargets := spatialGrid.GetNearbyEntities(archer.GetPosition(), range)
// Returns ~150 entities (in 30m radius) instead of 1000 ✅

// 3. Validate target is in list
if !contains(possibleTargets, target) {
    return errors.New("target_out_of_range")
}

// 4. Check line of sight using grid ray-cast
if !spatialGrid.HasLineOfSight(archer.GetPosition(), target.GetPosition()) {
    return errors.New("blocked_by_obstacle")
}
// Only checks ~10 cells along ray ✅

// 5. Create arrow projectile
arrow := &Projectile{
    Position: archer.GetPosition(),
    Velocity: CalculateVelocity(archer, target),
    Speed:    25.0,  // 25 m/s
}

// 6. Add to world (projectile is entity in grid)
spatialGrid.AddEntity(arrow)

// 7. Update arrow each frame (60 FPS)
func UpdateArrow(arrow *Projectile) {
    arrow.Update(1.0/60.0, spatialGrid)
    // Only checks current cell (~10 entities) ✅

    if hit := arrow.CheckCollisions(spatialGrid); hit != nil {
        // Arrow hit target!
        target.TakeDamage(arrow.Damage, archer)
        spatialGrid.RemoveEntity(arrow)
    }
}
```

**Performance**:
- Target finding: 150 entities checked (vs 1000) = **6.6x faster**
- Line of sight: 10 cells checked (vs 1000 entities) = **100x faster**
- Collision: 10 entities/frame (vs 1000) = **100x faster**

---

## Code Example: Complete Ranged Ability

```go
type BowAbility struct {
    BaseAbility
    Range           float64  // 30m
    ProjectileSpeed float64  // 25 m/s
    ArrowDamage     int
}

func (b *BowAbility) Execute(caster, target CombatEntity) *CombatResult {
    // 1. Validate range using spatial grid
    distance := caster.GetPosition().Distance(target.GetPosition())
    if distance > b.Range {
        return &CombatResult{Hit: false, Reason: "out_of_range"}
    }

    // 2. Check line of sight
    if !spatialGrid.HasLineOfSight(caster.GetPosition(), target.GetPosition()) {
        return &CombatResult{Hit: false, Reason: "blocked"}
    }

    // 3. Predict target movement
    predictedPos := PredictTargetPosition(target, b.ProjectileSpeed, caster.GetPosition())

    // 4. Create arrow projectile
    arrow := &Projectile{
        ID:       generateID(),
        CasterID: caster.GetID(),
        TargetID: target.GetID(),
        Position: caster.GetPosition(),
        Velocity: predictedPos.Subtract(caster.GetPosition()).Normalize(),
        Speed:    b.ProjectileSpeed,
        Range:    b.Range,
        Damage:   b.ArrowDamage,
    }

    // 5. Add to spatial grid
    spatialGrid.AddEntity(arrow)

    // 6. Return result (projectile spawned)
    return &CombatResult{
        Projectile: arrow,
        Hit:        false,  // Not hit yet, projectile in flight
    }
}

// Server updates all projectiles
func (world *World) UpdateProjectiles(deltaTime float64) {
    for _, projectile := range world.projectiles {
        result := projectile.Update(deltaTime, world.spatialGrid)

        if result != nil {
            if result.Hit {
                // Projectile hit something!
                target := world.GetEntity(result.TargetID)
                target.TakeDamage(projectile.Damage, projectile.Caster)

                // Broadcast hit event
                world.BroadcastCombatEvent(result)

                // Remove projectile
                world.spatialGrid.RemoveEntity(projectile)
                world.RemoveProjectile(projectile.ID)
            } else {
                // Projectile missed or out of range
                world.RemoveProjectile(projectile.ID)
            }
        }
    }
}
```

---

## Summary: Grid Optimization for Ranged

### ✅ **Works Great for Ranged**

| Aspect | How Grid Helps |
|--------|----------------|
| **Target Finding** | Query larger radius, still O(k) not O(n) |
| **Line of Sight** | Ray-cast through cells, not all entities |
| **Projectile Movement** | Check current cell only per frame |
| **AoE Explosions** | Query radius efficiently |
| **Prediction** | Fast nearby entity queries |

### 🎯 **Key Patterns**

1. **Hitscan**: Instant hit, use grid for LoS check
2. **Projectile**: Moving entity, check current cell only
3. **AoE**: Explosion uses grid radius query
4. **Prediction**: Calculate ahead for moving targets

### 📊 **Performance**

```
Ranged without grid: O(n) = 1000 checks ❌
Ranged with grid:    O(k) = 100-500 checks ✅

Still 2-10x faster than no optimization!
```

### 🔑 **Grid Cell Size**

```
Recommended: 6-10 meters

- Small enough for melee (1-4 cells)
- Large enough for ranged (not too many cells)
- Balance between granularity and overhead
```

---

## Gotchas & Solutions

### **Gotcha 1**: Very long range (100m+)

**Problem**: Would check hundreds of cells

**Solution**: Use hierarchical grid or limit max range
```go
const MaxCombatRange = 50.0  // Reasonable MMO range
```

### **Gotcha 2**: Fast projectiles skip cells

**Problem**: Projectile moves 2+ cells in one frame

**Solution**: Use continuous collision detection
```go
func (p *Projectile) Update(deltaTime float64) {
    // Check all cells along path, not just destination
    cellsAlongPath := grid.RaycastCells(p.Position, newPosition)
    for _, cell := range cellsAlongPath {
        // Check for hits
    }
}
```

### **Gotcha 3**: Projectiles need cleanup

**Problem**: Missed projectiles keep flying forever

**Solution**: Max range and lifetime limits
```go
type Projectile struct {
    MaxRange  float64
    MaxLifetime time.Duration
    StartTime time.Time
}

func (p *Projectile) ShouldRemove() bool {
    return p.DistanceTraveled() > p.MaxRange ||
           time.Since(p.StartTime) > p.MaxLifetime
}
```

---

**Grid optimization works excellently for ranged combat!** Just query a larger radius - still way better than checking all entities.

Want me to implement the projectile system with spatial grid integration?
