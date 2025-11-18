package entities

import (
	"math/rand"
	"mmo-server/combat"
	"mmo-server/stats"
	"sync"
	"time"
)

// MonsterType represents different categories of monsters
type MonsterType int

const (
	MonsterTypeNormal MonsterType = iota
	MonsterTypeElite
	MonsterTypeBoss
	MonsterTypeMinion
)

// AIState represents the monster's current AI state
type AIState int

const (
	AIStateIdle AIState = iota
	AIStatePatrolling
	AIStateChasing
	AIStateAttacking
	AIStateReturning
	AIStateFleeing
)

// AggroEntry tracks aggro for a specific entity
type AggroEntry struct {
	EntityID  string
	Threat    float64
	LastHitAt time.Time
}

// AggroTable manages monster threat/aggro
type AggroTable struct {
	entries map[string]*AggroEntry
	mu      sync.RWMutex
}

// NewAggroTable creates a new aggro table
func NewAggroTable() *AggroTable {
	return &AggroTable{
		entries: make(map[string]*AggroEntry),
	}
}

// AddThreat adds threat to an entity
func (at *AggroTable) AddThreat(entityID string, amount float64) {
	at.mu.Lock()
	defer at.mu.Unlock()

	if entry, exists := at.entries[entityID]; exists {
		entry.Threat += amount
		entry.LastHitAt = time.Now()
	} else {
		at.entries[entityID] = &AggroEntry{
			EntityID:  entityID,
			Threat:    amount,
			LastHitAt: time.Now(),
		}
	}
}

// GetTopThreat returns the entity with highest threat
func (at *AggroTable) GetTopThreat() string {
	at.mu.RLock()
	defer at.mu.RUnlock()

	var topID string
	var topThreat float64

	for id, entry := range at.entries {
		if entry.Threat > topThreat {
			topThreat = entry.Threat
			topID = id
		}
	}

	return topID
}

// Clear removes all aggro
func (at *AggroTable) Clear() {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.entries = make(map[string]*AggroEntry)
}

// DecayThreat reduces threat over time
func (at *AggroTable) DecayThreat(deltaTime float64) {
	at.mu.Lock()
	defer at.mu.Unlock()

	decayRate := 1.0 * deltaTime // 1 threat per second

	for id, entry := range at.entries {
		entry.Threat -= decayRate
		if entry.Threat <= 0 {
			delete(at.entries, id)
		}
	}
}

// LootItem represents a potential loot drop
type LootItem struct {
	ItemID      string
	DropChance  float64 // 0.0 to 1.0
	MinQuantity int
	MaxQuantity int
}

// LootTable defines what a monster can drop
type LootTable struct {
	Items         []LootItem
	GuaranteedXP  int
	BonusXPChance float64
	BonusXPAmount int
}

// GenerateLoot rolls for loot drops
func (lt *LootTable) GenerateLoot() []LootItem {
	drops := []LootItem{}

	for _, item := range lt.Items {
		if rand.Float64() <= item.DropChance {
			quantity := item.MinQuantity
			if item.MaxQuantity > item.MinQuantity {
				quantity += rand.Intn(item.MaxQuantity - item.MinQuantity + 1)
			}

			drops = append(drops, LootItem{
				ItemID:      item.ItemID,
				MinQuantity: quantity,
				MaxQuantity: quantity,
			})
		}
	}

	return drops
}

// MonsterTemplate defines a monster's base configuration
type MonsterTemplate struct {
	ID          string
	Name        string
	Type        MonsterType
	BaseStats   map[stats.StatType]float64
	LootTable   *LootTable
	AggroRadius float64
	LeashRadius float64 // Max distance from spawn before returning
	AttackRange float64
	MoveSpeed   float64
	RespawnTime time.Duration
}

// Monster represents a monster entity in the game
type Monster struct {
	// Core identity
	ID         string
	TemplateID string
	Name       string
	Type       MonsterType

	// Position and movement
	Position      combat.Vector3
	SpawnPosition combat.Vector3
	Rotation      combat.Vector3
	MoveSpeed     float64

	// Stats and combat
	Stats           *stats.StatContainer
	CooldownManager *combat.CooldownManager
	Faction         combat.Faction
	Stunned         bool

	// AI and behavior
	AIState     AIState
	AggroTable  *AggroTable
	CurrentTarget string
	AggroRadius float64
	LeashRadius float64
	AttackRange float64

	// Loot
	LootTable *LootTable
	KilledBy  string // Track who killed it for loot rights

	// Respawn
	RespawnTime   time.Duration
	DeathTime     time.Time
	ShouldRespawn bool

	mu sync.RWMutex
}

// NewMonster creates a new monster from a template
func NewMonster(id string, template *MonsterTemplate, spawnPosition combat.Vector3) *Monster {
	// Create stat container
	statContainer := stats.NewStatContainer()

	// Apply base stats from template
	for statType, value := range template.BaseStats {
		statContainer.SetBaseStat(statType, value)
	}

	// Set current resources to max
	statContainer.RestoreStat(stats.StatHealth)
	statContainer.RestoreStat(stats.StatMana)
	statContainer.RestoreStat(stats.StatStamina)

	monster := &Monster{
		ID:              id,
		TemplateID:      template.ID,
		Name:            template.Name,
		Type:            template.Type,
		Position:        spawnPosition,
		SpawnPosition:   spawnPosition,
		Rotation:        combat.Vector3{X: 0, Y: 0, Z: 0},
		MoveSpeed:       template.MoveSpeed,
		Stats:           statContainer,
		CooldownManager: combat.NewCooldownManager(),
		Faction:         combat.FactionHostileMob,
		AIState:         AIStateIdle,
		AggroTable:      NewAggroTable(),
		AggroRadius:     template.AggroRadius,
		LeashRadius:     template.LeashRadius,
		AttackRange:     template.AttackRange,
		LootTable:       template.LootTable,
		RespawnTime:     template.RespawnTime,
		ShouldRespawn:   true,
	}

	return monster
}

// =============================================================================
// CombatEntity Interface Implementation
// =============================================================================

// GetID returns the monster's unique ID
func (m *Monster) GetID() string {
	return m.ID
}

// GetType returns the entity type
func (m *Monster) GetType() combat.EntityType {
	switch m.Type {
	case MonsterTypeBoss:
		return combat.EntityTypeBoss
	default:
		return combat.EntityTypeMonster
	}
}

// GetPosition returns the monster's position
func (m *Monster) GetPosition() combat.Vector3 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Position
}

// IsAlive checks if monster is alive
func (m *Monster) IsAlive() bool {
	return m.Stats.GetStat(stats.StatHealth) > 0
}

// CanTakeDamage checks if monster can receive damage
func (m *Monster) CanTakeDamage() bool {
	return m.IsAlive()
}

// CanDealDamage checks if monster can attack
func (m *Monster) CanDealDamage() bool {
	return m.IsAlive() && !m.Stunned
}

// GetFaction returns the monster's faction
func (m *Monster) GetFaction() combat.Faction {
	return m.Faction
}

// IsHostileTo checks if monster is hostile to another entity
func (m *Monster) IsHostileTo(other combat.CombatEntity) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch other.GetType() {
	case combat.EntityTypePlayer:
		return true // Always hostile to players

	case combat.EntityTypeMonster, combat.EntityTypeBoss:
		// Check faction relations
		return m.Faction != other.GetFaction()

	case combat.EntityTypeDestructible, combat.EntityTypeHarvestable:
		return false // Monsters don't attack objects

	default:
		return false
	}
}

// IsValidTarget checks if this monster can be targeted by attacker
func (m *Monster) IsValidTarget(attacker combat.CombatEntity) bool {
	return m.IsAlive() && attacker.IsHostileTo(m)
}

// TakeDamage applies damage to the monster
func (m *Monster) TakeDamage(amount int, source combat.CombatEntity) *combat.DamageResult {
	m.mu.Lock()

	// Apply damage
	m.Stats.ModifyCurrentValue(stats.StatHealth, float64(-amount))

	// Clamp to 0
	if m.Stats.GetStat(stats.StatHealth) < 0 {
		m.Stats.ModifyCurrentValue(stats.StatHealth, -m.Stats.GetStat(stats.StatHealth))
	}

	newHealth := int(m.Stats.GetStat(stats.StatHealth))
	maxHealth := int(m.Stats.GetStat(stats.StatMaxHealth))

	// Add threat to aggro table
	if source != nil {
		threatAmount := float64(amount) * 1.0 // 1:1 damage to threat ratio
		m.AggroTable.AddThreat(source.GetID(), threatAmount)

		// Switch to combat if idle
		if m.AIState == AIStateIdle || m.AIState == AIStatePatrolling {
			m.AIState = AIStateChasing
			m.CurrentTarget = source.GetID()
		}
	}

	// Check death
	died := newHealth <= 0

	m.mu.Unlock()

	if died {
		m.OnDeath(source)
	}

	return &combat.DamageResult{
		TargetID:    m.ID,
		Damage:      amount,
		RemainingHP: newHealth,
		MaxHP:       maxHealth,
		Died:        died,
	}
}

// OnDeath handles monster death
func (m *Monster) OnDeath(killer combat.CombatEntity) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Track killer for loot rights
	if killer != nil {
		m.KilledBy = killer.GetID()
	}

	// Set death time for respawn
	m.DeathTime = time.Now()

	// Change state
	m.AIState = AIStateIdle

	// Clear aggro
	m.AggroTable.Clear()
	m.CurrentTarget = ""

	// TODO: Drop loot (implement when inventory system exists)
	// drops := m.LootTable.GenerateLoot()

	// TODO: Award XP to killer (implement when XP system exists)
}

// GetInterestRadius returns how far this entity can be seen
func (m *Monster) GetInterestRadius() float64 {
	switch m.Type {
	case MonsterTypeBoss:
		return 100.0 // Bosses visible from far away
	case MonsterTypeElite:
		return 60.0
	default:
		return 40.0 // Normal monsters
	}
}

// ShouldBroadcastCombat returns whether combat involving this entity should be broadcast
func (m *Monster) ShouldBroadcastCombat() bool {
	return true // Always broadcast monster combat
}

// =============================================================================
// AI and Behavior
// =============================================================================

// Update processes per-frame updates
func (m *Monster) Update(deltaTime float64, combatManager *combat.CombatManager) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if dead and should respawn
	if !m.IsAlive() && m.ShouldRespawn {
		if time.Since(m.DeathTime) >= m.RespawnTime {
			m.Respawn()
		}
		return
	}

	// Decay threat over time
	m.AggroTable.DecayThreat(deltaTime)

	// Update AI state machine
	switch m.AIState {
	case AIStateIdle:
		m.updateIdle(combatManager)

	case AIStatePatrolling:
		m.updatePatrolling(deltaTime)

	case AIStateChasing:
		m.updateChasing(deltaTime, combatManager)

	case AIStateAttacking:
		m.updateAttacking(deltaTime, combatManager)

	case AIStateReturning:
		m.updateReturning(deltaTime)

	case AIStateFleeing:
		m.updateFleeing(deltaTime)
	}
}

// updateIdle checks for nearby threats
func (m *Monster) updateIdle(combatManager *combat.CombatManager) {
	// Check for nearby players within aggro radius
	nearby := combatManager.GetNearbyEntities(m.Position, m.AggroRadius)

	for _, entity := range nearby {
		if m.IsHostileTo(entity) && entity.IsAlive() {
			// Found a target!
			m.AggroTable.AddThreat(entity.GetID(), 10.0)
			m.CurrentTarget = entity.GetID()
			m.AIState = AIStateChasing
			break
		}
	}
}

// updatePatrolling handles patrol movement
func (m *Monster) updatePatrolling(deltaTime float64) {
	// TODO: Implement patrol waypoints
	// For now, just switch to idle
	m.AIState = AIStateIdle
}

// updateChasing moves toward target
func (m *Monster) updateChasing(deltaTime float64, combatManager *combat.CombatManager) {
	// Get current target
	target := combatManager.GetEntity(m.CurrentTarget)

	if target == nil || !target.IsAlive() {
		// Target lost, get next from aggro table
		m.CurrentTarget = m.AggroTable.GetTopThreat()
		if m.CurrentTarget == "" {
			m.AIState = AIStateReturning
			return
		}
		target = combatManager.GetEntity(m.CurrentTarget)
	}

	if target == nil {
		m.AIState = AIStateReturning
		return
	}

	// Check leash distance
	distanceFromSpawn := m.Position.Distance(m.SpawnPosition)
	if distanceFromSpawn > m.LeashRadius {
		// Too far from spawn, return
		m.AIState = AIStateReturning
		m.AggroTable.Clear()
		m.CurrentTarget = ""
		return
	}

	// Calculate distance to target
	targetPos := target.GetPosition()
	distanceToTarget := m.Position.Distance(targetPos)

	// Check if in attack range
	if distanceToTarget <= m.AttackRange {
		m.AIState = AIStateAttacking
		return
	}

	// Move toward target
	direction := combat.Vector3{
		X: targetPos.X - m.Position.X,
		Y: targetPos.Y - m.Position.Y,
		Z: targetPos.Z - m.Position.Z,
	}.Normalize()

	moveAmount := m.MoveSpeed * deltaTime
	m.Position.X += direction.X * moveAmount
	m.Position.Y += direction.Y * moveAmount
	m.Position.Z += direction.Z * moveAmount
}

// updateAttacking performs attacks
func (m *Monster) updateAttacking(deltaTime float64, combatManager *combat.CombatManager) {
	target := combatManager.GetEntity(m.CurrentTarget)

	if target == nil || !target.IsAlive() {
		m.CurrentTarget = m.AggroTable.GetTopThreat()
		if m.CurrentTarget == "" {
			m.AIState = AIStateReturning
		} else {
			m.AIState = AIStateChasing
		}
		return
	}

	// Check if still in range
	targetPos := target.GetPosition()
	distanceToTarget := m.Position.Distance(targetPos)

	if distanceToTarget > m.AttackRange {
		m.AIState = AIStateChasing
		return
	}

	// Check cooldown for basic attack
	if !m.CooldownManager.IsOnCooldown("basic_attack") {
		// Execute attack through combat manager
		// Note: This would normally be done by the combat manager
		// For now, just set cooldown
		m.CooldownManager.SetCooldown("basic_attack", 1500*time.Millisecond)
	}
}

// updateReturning returns to spawn point
func (m *Monster) updateReturning(deltaTime float64) {
	distanceToSpawn := m.Position.Distance(m.SpawnPosition)

	if distanceToSpawn < 1.0 {
		// Reached spawn point
		m.Position = m.SpawnPosition
		m.AIState = AIStateIdle

		// Heal to full
		m.Stats.RestoreStat(stats.StatHealth)
		m.Stats.RestoreStat(stats.StatMana)
		return
	}

	// Move toward spawn
	direction := combat.Vector3{
		X: m.SpawnPosition.X - m.Position.X,
		Y: m.SpawnPosition.Y - m.Position.Y,
		Z: m.SpawnPosition.Z - m.Position.Z,
	}.Normalize()

	moveAmount := m.MoveSpeed * deltaTime
	m.Position.X += direction.X * moveAmount
	m.Position.Y += direction.Y * moveAmount
	m.Position.Z += direction.Z * moveAmount
}

// updateFleeing flees from threats
func (m *Monster) updateFleeing(deltaTime float64) {
	// TODO: Implement fleeing behavior (for low HP monsters)
	m.AIState = AIStateReturning
}

// Respawn resets the monster to spawn state
func (m *Monster) Respawn() {
	m.Position = m.SpawnPosition
	m.Rotation = combat.Vector3{X: 0, Y: 0, Z: 0}
	m.AIState = AIStateIdle
	m.CurrentTarget = ""
	m.KilledBy = ""

	// Restore stats
	m.Stats.RestoreStat(stats.StatHealth)
	m.Stats.RestoreStat(stats.StatMana)
	m.Stats.RestoreStat(stats.StatStamina)

	// Clear aggro
	m.AggroTable.Clear()
}

// =============================================================================
// Monster Template Library
// =============================================================================

// GetGoblinTemplate returns a basic goblin template
func GetGoblinTemplate() *MonsterTemplate {
	return &MonsterTemplate{
		ID:   "goblin",
		Name: "Goblin",
		Type: MonsterTypeNormal,
		BaseStats: map[stats.StatType]float64{
			stats.StatMaxHealth:     100,
			stats.StatHealth:        100,
			stats.StatMaxMana:       50,
			stats.StatMana:          50,
			stats.StatMaxStamina:    80,
			stats.StatStamina:       80,
			stats.StatAttackPower:   15,
			stats.StatDefense:       5,
			stats.StatCritChance:    0.05,
			stats.StatCritMultiplier: 1.5,
		},
		LootTable: &LootTable{
			Items: []LootItem{
				{ItemID: "goblin_ear", DropChance: 0.3, MinQuantity: 1, MaxQuantity: 2},
				{ItemID: "small_coin", DropChance: 0.8, MinQuantity: 1, MaxQuantity: 5},
			},
			GuaranteedXP:  10,
			BonusXPChance: 0.1,
			BonusXPAmount: 5,
		},
		AggroRadius: 15.0,
		LeashRadius: 40.0,
		AttackRange: 2.5,
		MoveSpeed:   4.0,
		RespawnTime: 30 * time.Second,
	}
}

// GetOrcTemplate returns an orc elite template
func GetOrcTemplate() *MonsterTemplate {
	return &MonsterTemplate{
		ID:   "orc_warrior",
		Name: "Orc Warrior",
		Type: MonsterTypeElite,
		BaseStats: map[stats.StatType]float64{
			stats.StatMaxHealth:     250,
			stats.StatHealth:        250,
			stats.StatMaxMana:       100,
			stats.StatMana:          100,
			stats.StatMaxStamina:    150,
			stats.StatStamina:       150,
			stats.StatAttackPower:   35,
			stats.StatDefense:       15,
			stats.StatCritChance:    0.10,
			stats.StatCritMultiplier: 2.0,
		},
		LootTable: &LootTable{
			Items: []LootItem{
				{ItemID: "orc_tooth", DropChance: 0.5, MinQuantity: 1, MaxQuantity: 1},
				{ItemID: "iron_ore", DropChance: 0.3, MinQuantity: 1, MaxQuantity: 3},
				{ItemID: "medium_coin", DropChance: 1.0, MinQuantity: 5, MaxQuantity: 15},
			},
			GuaranteedXP:  50,
			BonusXPChance: 0.2,
			BonusXPAmount: 25,
		},
		AggroRadius: 20.0,
		LeashRadius: 50.0,
		AttackRange: 3.0,
		MoveSpeed:   5.0,
		RespawnTime: 2 * time.Minute,
	}
}

// GetDragonTemplate returns a dragon boss template
func GetDragonTemplate() *MonsterTemplate {
	return &MonsterTemplate{
		ID:   "dragon_boss",
		Name: "Ancient Dragon",
		Type: MonsterTypeBoss,
		BaseStats: map[stats.StatType]float64{
			stats.StatMaxHealth:     5000,
			stats.StatHealth:        5000,
			stats.StatMaxMana:       1000,
			stats.StatMana:          1000,
			stats.StatMaxStamina:    500,
			stats.StatStamina:       500,
			stats.StatAttackPower:   150,
			stats.StatDefense:       50,
			stats.StatCritChance:    0.20,
			stats.StatCritMultiplier: 2.5,
		},
		LootTable: &LootTable{
			Items: []LootItem{
				{ItemID: "dragon_scale", DropChance: 1.0, MinQuantity: 3, MaxQuantity: 7},
				{ItemID: "dragon_heart", DropChance: 0.5, MinQuantity: 1, MaxQuantity: 1},
				{ItemID: "epic_weapon", DropChance: 0.1, MinQuantity: 1, MaxQuantity: 1},
				{ItemID: "large_coin", DropChance: 1.0, MinQuantity: 100, MaxQuantity: 500},
			},
			GuaranteedXP:  1000,
			BonusXPChance: 0.5,
			BonusXPAmount: 500,
		},
		AggroRadius: 30.0,
		LeashRadius: 100.0,
		AttackRange: 8.0,
		MoveSpeed:   6.0,
		RespawnTime: 24 * time.Hour,
	}
}
