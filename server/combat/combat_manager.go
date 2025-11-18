package combat

import (
	"fmt"
	"sync"
	"time"
)

// CombatManager orchestrates all combat in the game
type CombatManager struct {
	// Entity Management
	entities    map[string]CombatEntity
	spatialGrid *SpatialGrid

	// Ability System
	abilityRegistry *AbilityRegistry
	validatorChain  *ValidatorChain

	// Event Broadcasting (for animations)
	eventBroadcaster *EventBroadcaster

	// Configuration
	config *CombatConfig

	mu sync.RWMutex
}

// CombatConfig holds combat system configuration
type CombatConfig struct {
	MaxEntitiesPerRegion int
	InterestRadius       float64
	SpatialCellSize      float64
	PvPEnabled           bool
	FriendlyFireEnabled  bool
}

// DefaultCombatConfig returns default configuration
func DefaultCombatConfig() *CombatConfig {
	return &CombatConfig{
		MaxEntitiesPerRegion: 1000,
		InterestRadius:       50.0,
		SpatialCellSize:      10.0,
		PvPEnabled:           true,
		FriendlyFireEnabled:  false,
	}
}

// NewCombatManager creates a new combat manager
func NewCombatManager(config *CombatConfig) *CombatManager {
	if config == nil {
		config = DefaultCombatConfig()
	}

	cm := &CombatManager{
		entities:        make(map[string]CombatEntity),
		spatialGrid:     NewSpatialGrid(config.SpatialCellSize),
		abilityRegistry: NewAbilityRegistry(),
		validatorChain:  CreateDefaultValidatorChain(),
		config:          config,
	}

	// Event broadcaster will be set via SetBroadcastCallback
	// (allows dependency injection from main server)

	// Register default abilities
	RegisterDefaultMeleeAbilities(cm.abilityRegistry)

	return cm
}

// SetBroadcastCallback configures event broadcasting for animations
func (cm *CombatManager) SetBroadcastCallback(sendFunc func(string, []byte)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.eventBroadcaster = NewEventBroadcaster(sendFunc)
}

// RegisterEntity adds an entity to combat tracking
func (cm *CombatManager) RegisterEntity(entity CombatEntity) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.entities[entity.GetID()] = entity
	cm.spatialGrid.AddEntity(entity)
}

// UnregisterEntity removes an entity from combat tracking
func (cm *CombatManager) UnregisterEntity(entityID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if entity, ok := cm.entities[entityID]; ok {
		cm.spatialGrid.RemoveEntity(entity)
		delete(cm.entities, entityID)
	}
}

// UpdateEntityPosition updates entity position in spatial grid
func (cm *CombatManager) UpdateEntityPosition(entity CombatEntity, oldPos, newPos Vector3) {
	cm.spatialGrid.UpdateEntity(entity, oldPos, newPos)
}

// GetEntity retrieves an entity by ID
func (cm *CombatManager) GetEntity(entityID string) CombatEntity {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.entities[entityID]
}

// ExecuteAbility processes an ability use
func (cm *CombatManager) ExecuteAbility(
	casterID string,
	targetID string,
	abilityID string,
) (*CombatEvent, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// 1. Get entities
	caster, ok := cm.entities[casterID]
	if !ok {
		return nil, fmt.Errorf("caster not found: %s", casterID)
	}

	target, ok := cm.entities[targetID]
	if !ok {
		return nil, fmt.Errorf("target not found: %s", targetID)
	}

	// 2. Get ability
	ability := cm.abilityRegistry.Get(abilityID)
	if ability == nil {
		return nil, fmt.Errorf("ability not found: %s", abilityID)
	}

	// 3. Validate targeting rules
	if err := CanAttack(caster, target); err != nil {
		return nil, err
	}

	// 4. Cast to Player interface for validation
	casterPlayer, casterOK := caster.(Player)
	targetPlayer, targetOK := target.(Player)

	// Only validate if both are players (ability validation requires Player interface)
	if casterOK && targetOK {
		// Full validation chain for players
		if err := cm.validatorChain.Validate(casterPlayer, targetPlayer, ability); err != nil {
			return nil, err
		}
	}
	// For non-player entities, skip ability-specific validation
	// CanAttack() already handled basic validation above

	// 5. Execute ability
	var result *CombatResult
	if casterOK && targetOK {
		result = ability.Execute(casterPlayer, targetPlayer)
	} else {
		// Simplified execution for non-player entities
		result = &CombatResult{
			AbilityID:   abilityID,
			CasterID:    casterID,
			TargetID:    targetID,
			Damage:      CalculateMeleeDamage(caster, target),
			Hit:         true,
			CriticalHit: false,
			Timestamp:   time.Now(),
		}
	}

	// 6. Apply damage to target
	damageResult := target.TakeDamage(result.Damage, caster)

	// 7. Set cooldown for caster (if player)
	if player, ok := caster.(Player); ok {
		player.GetCooldownManager().SetCooldown(abilityID, ability.GetCooldown())
	}

	// 8. Create combat event
	event := &CombatEvent{
		Type:            CombatEventDamage,
		EventID:         generateEventID(),
		Timestamp:       time.Now(),
		CasterID:        casterID,
		TargetID:        targetID,
		Position:        target.GetPosition(),
		AbilityID:       abilityID,
		Damage:          damageResult.Damage,
		TargetHealth:    damageResult.RemainingHP,
		TargetMaxHealth: damageResult.MaxHP,
		TargetDied:      damageResult.Died,
		IsCritical:      result.CriticalHit,
	}

	return event, nil
}

// FindTargetsInRadius finds all valid targets within a radius
func (cm *CombatManager) FindTargetsInRadius(
	casterID string,
	position Vector3,
	radius float64,
) []CombatEntity {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	caster, ok := cm.entities[casterID]
	if !ok {
		return []CombatEntity{}
	}

	// Get nearby entities from spatial grid
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

// GetNearbyEntities returns all entities within radius (no filtering)
func (cm *CombatManager) GetNearbyEntities(position Vector3, radius float64) []CombatEntity {
	return cm.spatialGrid.GetNearbyEntities(position, radius)
}

// HasLineOfSight checks if two positions have clear line of sight
func (cm *CombatManager) HasLineOfSight(from, to Vector3) bool {
	return cm.spatialGrid.HasLineOfSight(from, to)
}

// ApplyEnvironmentalDamage applies damage from environment (lava, traps, etc.)
func (cm *CombatManager) ApplyEnvironmentalDamage(
	sourceID string,
	targetID string,
	damage int,
	damageType DamageType,
) *CombatEvent {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	target, ok := cm.entities[targetID]
	if !ok {
		return nil
	}

	// Create fake environmental source
	source := cm.entities[sourceID]

	// Apply damage
	damageResult := target.TakeDamage(damage, source)

	// Create event
	return &CombatEvent{
		Type:            CombatEventDamage,
		EventID:         generateEventID(),
		Timestamp:       time.Now(),
		CasterID:        sourceID,
		TargetID:        targetID,
		Position:        target.GetPosition(),
		Damage:          damageResult.Damage,
		TargetHealth:    damageResult.RemainingHP,
		TargetMaxHealth: damageResult.MaxHP,
		TargetDied:      damageResult.Died,
	}
}

// GetEntityCount returns total number of tracked entities
func (cm *CombatManager) GetEntityCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.entities)
}

// CombatEvent represents a combat event to broadcast
type CombatEvent struct {
	Type      CombatEventType `json:"type"`
	EventID   string          `json:"event_id"`
	Timestamp time.Time       `json:"timestamp"`

	// Participants
	CasterID string `json:"caster_id"`
	TargetID string `json:"target_id"`

	// Location
	Position  Vector3 `json:"position"`
	Direction Vector3 `json:"direction,omitempty"`

	// Combat Data
	AbilityID  string `json:"ability_id,omitempty"`
	Damage     int    `json:"damage"`
	Healing    int    `json:"healing,omitempty"`
	DamageType string `json:"damage_type,omitempty"`

	// Result
	TargetHealth    int  `json:"target_health"`
	TargetMaxHealth int  `json:"target_max_health"`
	TargetDied      bool `json:"target_died"`

	// Effects
	BuffsApplied   []string `json:"buffs_applied,omitempty"`
	DebuffsApplied []string `json:"debuffs_applied,omitempty"`

	// Metadata
	IsCritical bool `json:"is_critical"`
	WasDodged  bool `json:"was_dodged"`
	WasBlocked bool `json:"was_blocked"`

	// Animation Hints (for client)
	AnimationHint string `json:"animation_hint,omitempty"` // "swing_right", "thrust", "overhead"
}

// CombatEventType represents types of combat events
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

// Event ID generator
var eventCounter uint64
var eventMutex sync.Mutex

func generateEventID() string {
	eventMutex.Lock()
	defer eventMutex.Unlock()

	eventCounter++
	return fmt.Sprintf("evt_%d_%d", time.Now().Unix(), eventCounter)
}
