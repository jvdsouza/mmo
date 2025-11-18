package entities

import (
	"mmo-server/combat"
	"sync"
	"time"
)

// EnvironmentalType represents different categories of environmental hazards
type EnvironmentalType int

const (
	EnvironmentalTypeLava EnvironmentalType = iota
	EnvironmentalTypePoison
	EnvironmentalTypeSpikeTrap
	EnvironmentalTypeFire
	EnvironmentalTypeIce
	EnvironmentalTypeLightning
	EnvironmentalTypeWindGust
	EnvironmentalTypeHealingSpring
)

// DamagePattern defines how damage is applied
type DamagePattern int

const (
	DamagePatternContinuous DamagePattern = iota // Constant DoT while in area
	DamagePatternPulse                           // Periodic bursts
	DamagePatternOneTime                         // Single damage on enter
)

// EnvironmentalTemplate defines an environmental hazard's configuration
type EnvironmentalTemplate struct {
	ID             string
	Name           string
	Type           EnvironmentalType
	DamageType     combat.DamageType
	DamageAmount   int
	DamagePattern  DamagePattern
	DamageInterval time.Duration // For continuous/pulse patterns
	Radius         float64       // Area of effect
	AffectsPlayers bool
	AffectsMonsters bool
	AffectsAll     bool // Affects everything
	IsHealing      bool // Healing zone instead of damage
	Duration       time.Duration // 0 = permanent
}

// Environmental represents an environmental hazard in the game world
type Environmental struct {
	// Core identity
	ID         string
	TemplateID string
	Name       string
	Type       EnvironmentalType

	// Position and area
	Position combat.Vector3
	Radius   float64

	// Damage properties
	DamageType     combat.DamageType
	DamageAmount   int
	DamagePattern  DamagePattern
	DamageInterval time.Duration
	IsHealing      bool

	// Targeting
	AffectsPlayers  bool
	AffectsMonsters bool
	AffectsAll      bool

	// State
	Active      bool
	Duration    time.Duration
	CreatedAt   time.Time
	LastDamage  time.Time
	NextPulse   time.Time

	// Tracking entities currently in area
	EntitiesInArea map[string]time.Time // entityID -> last damage time

	mu sync.RWMutex
}

// NewEnvironmental creates a new environmental hazard from a template
func NewEnvironmental(id string, template *EnvironmentalTemplate, position combat.Vector3) *Environmental {
	now := time.Now()

	return &Environmental{
		ID:              id,
		TemplateID:      template.ID,
		Name:            template.Name,
		Type:            template.Type,
		Position:        position,
		Radius:          template.Radius,
		DamageType:      template.DamageType,
		DamageAmount:    template.DamageAmount,
		DamagePattern:   template.DamagePattern,
		DamageInterval:  template.DamageInterval,
		IsHealing:       template.IsHealing,
		AffectsPlayers:  template.AffectsPlayers,
		AffectsMonsters: template.AffectsMonsters,
		AffectsAll:      template.AffectsAll,
		Active:          true,
		Duration:        template.Duration,
		CreatedAt:       now,
		LastDamage:      now,
		NextPulse:       now.Add(template.DamageInterval),
		EntitiesInArea:  make(map[string]time.Time),
	}
}

// =============================================================================
// CombatEntity Interface Implementation
// =============================================================================

// GetID returns the environmental hazard's unique ID
func (e *Environmental) GetID() string {
	return e.ID
}

// GetType returns the entity type
func (e *Environmental) GetType() combat.EntityType {
	return combat.EntityTypeEnvironmental
}

// GetPosition returns the hazard's center position
func (e *Environmental) GetPosition() combat.Vector3 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Position
}

// IsAlive always returns true (environmental hazards don't "die")
func (e *Environmental) IsAlive() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Active
}

// CanTakeDamage always returns false (can't damage environmental hazards)
func (e *Environmental) CanTakeDamage() bool {
	return false
}

// CanDealDamage checks if hazard can currently deal damage
func (e *Environmental) CanDealDamage() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Active && !e.IsHealing
}

// GetFaction returns neutral faction
func (e *Environmental) GetFaction() combat.Faction {
	return combat.FactionNeutral
}

// IsHostileTo checks if hazard affects this entity type
func (e *Environmental) IsHostileTo(other combat.CombatEntity) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if !e.Active || e.IsHealing {
		return false
	}

	if e.AffectsAll {
		return true
	}

	switch other.GetType() {
	case combat.EntityTypePlayer:
		return e.AffectsPlayers
	case combat.EntityTypeMonster, combat.EntityTypeBoss:
		return e.AffectsMonsters
	default:
		return false
	}
}

// IsValidTarget always returns false (can't target environmental hazards)
func (e *Environmental) IsValidTarget(attacker combat.CombatEntity) bool {
	return false
}

// TakeDamage does nothing (environmental hazards are invulnerable)
func (e *Environmental) TakeDamage(amount int, source combat.CombatEntity) *combat.DamageResult {
	return &combat.DamageResult{
		TargetID:    e.ID,
		Damage:      0,
		RemainingHP: 9999,
		MaxHP:       9999,
		Died:        false,
	}
}

// OnDeath does nothing (environmental hazards don't die from damage)
func (e *Environmental) OnDeath(killer combat.CombatEntity) {
	// Environmental hazards don't die from combat
}

// GetInterestRadius returns how far this hazard should be visible
func (e *Environmental) GetInterestRadius() float64 {
	// Visible from twice the damage radius
	return e.Radius * 2.0
}

// ShouldBroadcastCombat returns whether damage from this hazard should be broadcast
func (e *Environmental) ShouldBroadcastCombat() bool {
	// Always broadcast environmental damage
	return true
}

// =============================================================================
// Environmental Hazard Behavior
// =============================================================================

// Update processes per-frame updates for the hazard
func (e *Environmental) Update(deltaTime float64, combatManager *combat.CombatManager) []combat.CombatEntity {
	// Copy values we need while holding the lock
	e.mu.Lock()

	// Check if expired
	if e.Duration > 0 && time.Since(e.CreatedAt) >= e.Duration {
		e.Active = false
		e.mu.Unlock()
		return nil
	}

	if !e.Active {
		e.mu.Unlock()
		return nil
	}

	position := e.Position
	radius := e.Radius
	damagePattern := e.DamagePattern
	damageInterval := e.DamageInterval
	nextPulse := e.NextPulse

	e.mu.Unlock()

	now := time.Now()
	affectedEntities := []combat.CombatEntity{}

	// Get all entities in radius (lock released to avoid deadlock)
	nearby := combatManager.GetNearbyEntities(position, radius)

	// Lock to access EntitiesInArea map
	e.mu.Lock()
	entitiesInAreaCopy := make(map[string]time.Time)
	for k, v := range e.EntitiesInArea {
		entitiesInAreaCopy[k] = v
	}
	e.mu.Unlock()

	entitiesToUpdate := make(map[string]time.Time)

	for _, entity := range nearby {
		// Check if entity is affected by this hazard
		if !e.shouldAffect(entity) {
			continue
		}

		// Check distance
		distance := position.Distance(entity.GetPosition())
		if distance > radius {
			continue
		}

		// Apply damage based on pattern
		shouldDamage := false
		lastDamageTime, inArea := entitiesInAreaCopy[entity.GetID()]

		switch damagePattern {
		case DamagePatternOneTime:
			// Only damage once when entering
			shouldDamage = !inArea

		case DamagePatternContinuous:
			// Damage at intervals while in area
			if !inArea || now.Sub(lastDamageTime) >= damageInterval {
				shouldDamage = true
			}

		case DamagePatternPulse:
			// Damage all entities in area at fixed intervals
			if now.After(nextPulse) {
				shouldDamage = true
			}
		}

		if shouldDamage {
			affectedEntities = append(affectedEntities, entity)
			entitiesToUpdate[entity.GetID()] = now
		} else if !inArea {
			// Track entity entered area
			entitiesToUpdate[entity.GetID()] = now
		}
	}

	// Lock again to update state
	e.mu.Lock()
	defer e.mu.Unlock()

	// Update EntitiesInArea map
	for id, t := range entitiesToUpdate {
		e.EntitiesInArea[id] = t
	}

	// Update next pulse time
	if damagePattern == DamagePatternPulse && now.After(nextPulse) {
		e.NextPulse = now.Add(damageInterval)
	}

	// Clean up entities that left the area
	for entityID := range e.EntitiesInArea {
		found := false
		for _, entity := range nearby {
			if entity.GetID() == entityID {
				// Skip self to avoid deadlock
				if entity.GetID() == e.ID {
					found = true
					break
				}
				distance := position.Distance(entity.GetPosition())
				if distance <= radius {
					found = true
					break
				}
			}
		}
		if !found {
			delete(e.EntitiesInArea, entityID)
		}
	}

	return affectedEntities
}

// shouldAffect checks if this hazard should affect the given entity
func (e *Environmental) shouldAffect(entity combat.CombatEntity) bool {
	if !entity.IsAlive() {
		return false
	}

	if e.AffectsAll {
		return true
	}

	switch entity.GetType() {
	case combat.EntityTypePlayer:
		return e.AffectsPlayers
	case combat.EntityTypeMonster, combat.EntityTypeBoss:
		return e.AffectsMonsters
	default:
		return false
	}
}

// GetDamageAmount returns the damage this hazard deals
func (e *Environmental) GetDamageAmount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.DamageAmount
}

// Deactivate turns off the hazard
func (e *Environmental) Deactivate() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Active = false
}

// Activate turns on the hazard
func (e *Environmental) Activate() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Active = true
	e.CreatedAt = time.Now()
}

// =============================================================================
// Environmental Template Library
// =============================================================================

// GetLavaPoolTemplate returns a lava pool hazard
func GetLavaPoolTemplate() *EnvironmentalTemplate {
	return &EnvironmentalTemplate{
		ID:              "lava_pool",
		Name:            "Lava Pool",
		Type:            EnvironmentalTypeLava,
		DamageType:      combat.DamageTypeFire,
		DamageAmount:    50,
		DamagePattern:   DamagePatternContinuous,
		DamageInterval:  1 * time.Second,
		Radius:          5.0,
		AffectsPlayers:  true,
		AffectsMonsters: true,
		AffectsAll:      true,
		Duration:        0, // Permanent
	}
}

// GetPoisonCloudTemplate returns a poison cloud hazard
func GetPoisonCloudTemplate() *EnvironmentalTemplate {
	return &EnvironmentalTemplate{
		ID:              "poison_cloud",
		Name:            "Poison Cloud",
		Type:            EnvironmentalTypePoison,
		DamageType:      combat.DamageTypePoison,
		DamageAmount:    20,
		DamagePattern:   DamagePatternContinuous,
		DamageInterval:  2 * time.Second,
		Radius:          8.0,
		AffectsPlayers:  true,
		AffectsMonsters: true,
		AffectsAll:      true,
		Duration:        30 * time.Second, // Temporary
	}
}

// GetSpikeTrapTemplate returns a spike trap
func GetSpikeTrapTemplate() *EnvironmentalTemplate {
	return &EnvironmentalTemplate{
		ID:              "spike_trap",
		Name:            "Spike Trap",
		Type:            EnvironmentalTypeSpikeTrap,
		DamageType:      combat.DamageTypePhysical,
		DamageAmount:    100,
		DamagePattern:   DamagePatternPulse,
		DamageInterval:  5 * time.Second, // Activates every 5 seconds
		Radius:          3.0,
		AffectsPlayers:  true,
		AffectsMonsters: true,
		AffectsAll:      true,
		Duration:        0, // Permanent
	}
}

// GetFireWallTemplate returns a fire wall
func GetFireWallTemplate() *EnvironmentalTemplate {
	return &EnvironmentalTemplate{
		ID:              "fire_wall",
		Name:            "Wall of Fire",
		Type:            EnvironmentalTypeFire,
		DamageType:      combat.DamageTypeFire,
		DamageAmount:    75,
		DamagePattern:   DamagePatternContinuous,
		DamageInterval:  500 * time.Millisecond,
		Radius:          4.0,
		AffectsPlayers:  true,
		AffectsMonsters: true,
		AffectsAll:      true,
		Duration:        15 * time.Second, // Temporary spell
	}
}

// GetIceStormTemplate returns an ice storm area
func GetIceStormTemplate() *EnvironmentalTemplate {
	return &EnvironmentalTemplate{
		ID:              "ice_storm",
		Name:            "Ice Storm",
		Type:            EnvironmentalTypeIce,
		DamageType:      combat.DamageTypeIce,
		DamageAmount:    30,
		DamagePattern:   DamagePatternContinuous,
		DamageInterval:  1 * time.Second,
		Radius:          10.0,
		AffectsPlayers:  true,
		AffectsMonsters: true,
		AffectsAll:      true,
		Duration:        20 * time.Second,
	}
}

// GetLightningFieldTemplate returns a lightning field
func GetLightningFieldTemplate() *EnvironmentalTemplate {
	return &EnvironmentalTemplate{
		ID:              "lightning_field",
		Name:            "Lightning Field",
		Type:            EnvironmentalTypeLightning,
		DamageType:      combat.DamageTypeLightning,
		DamageAmount:    150,
		DamagePattern:   DamagePatternPulse,
		DamageInterval:  3 * time.Second, // Lightning strikes every 3 seconds
		Radius:          7.0,
		AffectsPlayers:  true,
		AffectsMonsters: true,
		AffectsAll:      true,
		Duration:        25 * time.Second,
	}
}

// GetHealingSpringTemplate returns a healing spring
func GetHealingSpringTemplate() *EnvironmentalTemplate {
	return &EnvironmentalTemplate{
		ID:              "healing_spring",
		Name:            "Healing Spring",
		Type:            EnvironmentalTypeHealingSpring,
		DamageType:      combat.DamageTypeHoly, // "Damage" type for healing
		DamageAmount:    25, // Actually healing amount
		DamagePattern:   DamagePatternContinuous,
		DamageInterval:  2 * time.Second,
		Radius:          6.0,
		AffectsPlayers:  true,
		AffectsMonsters: false, // Only heals players
		IsHealing:       true,
		Duration:        0, // Permanent
	}
}

// GetBossPoisonZoneTemplate returns a boss-created poison zone
func GetBossPoisonZoneTemplate() *EnvironmentalTemplate {
	return &EnvironmentalTemplate{
		ID:              "boss_poison_zone",
		Name:            "Corrupted Ground",
		Type:            EnvironmentalTypePoison,
		DamageType:      combat.DamageTypePoison,
		DamageAmount:    80,
		DamagePattern:   DamagePatternContinuous,
		DamageInterval:  1 * time.Second,
		Radius:          12.0,
		AffectsPlayers:  true,
		AffectsMonsters: false, // Only affects players
		Duration:        45 * time.Second,
	}
}

// GetWindGustTemplate returns a wind gust that pushes entities
func GetWindGustTemplate() *EnvironmentalTemplate {
	return &EnvironmentalTemplate{
		ID:              "wind_gust",
		Name:            "Powerful Wind",
		Type:            EnvironmentalTypeWindGust,
		DamageType:      combat.DamageTypePhysical,
		DamageAmount:    10, // Minor damage
		DamagePattern:   DamagePatternPulse,
		DamageInterval:  2 * time.Second,
		Radius:          15.0,
		AffectsPlayers:  true,
		AffectsMonsters: true,
		AffectsAll:      true,
		Duration:        10 * time.Second,
	}
}

// GetOneTimeExplosionTemplate returns a one-time explosion damage zone
func GetOneTimeExplosionTemplate() *EnvironmentalTemplate {
	return &EnvironmentalTemplate{
		ID:              "explosion",
		Name:            "Explosion",
		Type:            EnvironmentalTypeFire,
		DamageType:      combat.DamageTypeFire,
		DamageAmount:    200,
		DamagePattern:   DamagePatternOneTime, // Only damages once
		Radius:          8.0,
		AffectsPlayers:  true,
		AffectsMonsters: true,
		AffectsAll:      true,
		Duration:        100 * time.Millisecond, // Very short duration
	}
}
