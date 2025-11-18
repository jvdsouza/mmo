package entities

import (
	"mmo-server/combat"
	"sync"
)

// DestructibleType represents different categories of destructible objects
type DestructibleType int

const (
	DestructibleTypeCrate DestructibleType = iota
	DestructibleTypeBarrel
	DestructibleTypeGrass
	DestructibleTypeBush
	DestructibleTypeRock
	DestructibleTypePot
	DestructibleTypeChest
)

// DestructibleTemplate defines a destructible object's configuration
type DestructibleTemplate struct {
	ID              string
	Name            string
	Type            DestructibleType
	MaxHealth       float64
	RequiresTool    bool   // Some objects require specific tools
	RequiredToolID  string // e.g., "sword" for grass, "pickaxe" for rocks
	LootTable       *LootTable
	Respawnable     bool
	InterestRadius  float64
	CanTakeDamage   bool // Some objects are invulnerable
}

// Destructible represents a destructible object in the game world
type Destructible struct {
	// Core identity
	ID         string
	TemplateID string
	Name       string
	Type       DestructibleType

	// Position
	Position combat.Vector3
	Rotation combat.Vector3

	// Health (destructibles don't have mana/stamina)
	MaxHealth     float64
	CurrentHealth float64

	// Behavior
	RequiresTool   bool
	RequiredToolID string
	IsDamageable   bool

	// Loot
	LootTable *LootTable
	Destroyed bool

	// Respawn
	Respawnable bool

	mu sync.RWMutex
}

// NewDestructible creates a new destructible object from a template
func NewDestructible(id string, template *DestructibleTemplate, position combat.Vector3) *Destructible {
	return &Destructible{
		ID:             id,
		TemplateID:     template.ID,
		Name:           template.Name,
		Type:           template.Type,
		Position:       position,
		Rotation:       combat.Vector3{X: 0, Y: 0, Z: 0},
		MaxHealth:      template.MaxHealth,
		CurrentHealth:  template.MaxHealth,
		RequiresTool:   template.RequiresTool,
		RequiredToolID: template.RequiredToolID,
		IsDamageable:   template.CanTakeDamage,
		LootTable:      template.LootTable,
		Respawnable:    template.Respawnable,
		Destroyed:      false,
	}
}

// =============================================================================
// CombatEntity Interface Implementation
// =============================================================================

// GetID returns the destructible's unique ID
func (d *Destructible) GetID() string {
	return d.ID
}

// GetType returns the entity type
func (d *Destructible) GetType() combat.EntityType {
	return combat.EntityTypeDestructible
}

// GetPosition returns the destructible's position
func (d *Destructible) GetPosition() combat.Vector3 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Position
}

// IsAlive checks if destructible is still intact
func (d *Destructible) IsAlive() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.CurrentHealth > 0 && !d.Destroyed
}

// CanTakeDamage checks if destructible can receive damage
func (d *Destructible) CanTakeDamage() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.IsDamageable && d.CurrentHealth > 0
}

// CanDealDamage always returns false (destructibles don't attack)
func (d *Destructible) CanDealDamage() bool {
	return false
}

// GetFaction returns neutral faction
func (d *Destructible) GetFaction() combat.Faction {
	return combat.FactionNeutral
}

// IsHostileTo always returns false (destructibles are not hostile)
func (d *Destructible) IsHostileTo(other combat.CombatEntity) bool {
	return false
}

// IsValidTarget checks if this destructible can be targeted by attacker
func (d *Destructible) IsValidTarget(attacker combat.CombatEntity) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Must be alive and damageable
	if !d.IsAlive() || !d.IsDamageable {
		return false
	}

	// Only players can attack destructibles
	if attacker.GetType() != combat.EntityTypePlayer {
		return false
	}

	// If requires a tool, check if player has it
	if d.RequiresTool {
		// TODO: Check player's equipped tool when inventory system exists
		// For now, assume player has the required tool
		return true
	}

	return true
}

// TakeDamage applies damage to the destructible
func (d *Destructible) TakeDamage(amount int, source combat.CombatEntity) *combat.DamageResult {
	d.mu.Lock()

	// Apply damage
	d.CurrentHealth -= float64(amount)

	// Clamp to 0
	if d.CurrentHealth < 0 {
		d.CurrentHealth = 0
	}

	newHealth := int(d.CurrentHealth)
	maxHealth := int(d.MaxHealth)

	// Check destruction
	destroyed := d.CurrentHealth <= 0

	d.mu.Unlock()

	if destroyed {
		d.OnDeath(source)
	}

	return &combat.DamageResult{
		TargetID:    d.ID,
		Damage:      amount,
		RemainingHP: newHealth,
		MaxHP:       maxHealth,
		Died:        destroyed,
	}
}

// OnDeath handles destructible destruction
func (d *Destructible) OnDeath(destroyer combat.CombatEntity) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.Destroyed = true

	// TODO: Drop loot when inventory system exists
	// if d.LootTable != nil {
	//     drops := d.LootTable.GenerateLoot()
	//     // Spawn loot items at position
	// }

	// TODO: Implement respawn logic if Respawnable is true
}

// GetInterestRadius returns how far this entity can be seen
func (d *Destructible) GetInterestRadius() float64 {
	switch d.Type {
	case DestructibleTypeChest:
		return 30.0 // Chests visible from farther
	default:
		return 20.0 // Normal destructibles
	}
}

// ShouldBroadcastCombat returns whether combat involving this entity should be broadcast
func (d *Destructible) ShouldBroadcastCombat() bool {
	// Only broadcast important destructibles like chests
	return d.Type == DestructibleTypeChest
}

// =============================================================================
// Destructible Template Library
// =============================================================================

// GetWoodenCrateTemplate returns a basic wooden crate template
func GetWoodenCrateTemplate() *DestructibleTemplate {
	return &DestructibleTemplate{
		ID:        "wooden_crate",
		Name:      "Wooden Crate",
		Type:      DestructibleTypeCrate,
		MaxHealth: 50,
		LootTable: &LootTable{
			Items: []LootItem{
				{ItemID: "wood_plank", DropChance: 0.8, MinQuantity: 1, MaxQuantity: 3},
				{ItemID: "small_coin", DropChance: 0.3, MinQuantity: 1, MaxQuantity: 5},
				{ItemID: "potion_health", DropChance: 0.1, MinQuantity: 1, MaxQuantity: 1},
			},
		},
		Respawnable:    true,
		InterestRadius: 20.0,
		CanTakeDamage:  true,
	}
}

// GetBarrelTemplate returns a barrel template
func GetBarrelTemplate() *DestructibleTemplate {
	return &DestructibleTemplate{
		ID:        "barrel",
		Name:      "Barrel",
		Type:      DestructibleTypeBarrel,
		MaxHealth: 60,
		LootTable: &LootTable{
			Items: []LootItem{
				{ItemID: "water", DropChance: 0.5, MinQuantity: 1, MaxQuantity: 1},
				{ItemID: "fish", DropChance: 0.2, MinQuantity: 1, MaxQuantity: 2},
				{ItemID: "small_coin", DropChance: 0.4, MinQuantity: 1, MaxQuantity: 8},
			},
		},
		Respawnable:    true,
		InterestRadius: 20.0,
		CanTakeDamage:  true,
	}
}

// GetGrassTemplate returns grass that can be cut
func GetGrassTemplate() *DestructibleTemplate {
	return &DestructibleTemplate{
		ID:             "tall_grass",
		Name:           "Tall Grass",
		Type:           DestructibleTypeGrass,
		MaxHealth:      10,
		RequiresTool:   true,
		RequiredToolID: "sword", // Requires sword to cut
		LootTable: &LootTable{
			Items: []LootItem{
				{ItemID: "grass_blade", DropChance: 0.6, MinQuantity: 1, MaxQuantity: 3},
				{ItemID: "seed", DropChance: 0.3, MinQuantity: 1, MaxQuantity: 2},
			},
		},
		Respawnable:    true,
		InterestRadius: 15.0,
		CanTakeDamage:  true,
	}
}

// GetBushTemplate returns a bush that can be destroyed
func GetBushTemplate() *DestructibleTemplate {
	return &DestructibleTemplate{
		ID:        "bush",
		Name:      "Bush",
		Type:      DestructibleTypeBush,
		MaxHealth: 20,
		LootTable: &LootTable{
			Items: []LootItem{
				{ItemID: "berry", DropChance: 0.5, MinQuantity: 1, MaxQuantity: 4},
				{ItemID: "stick", DropChance: 0.7, MinQuantity: 1, MaxQuantity: 2},
			},
		},
		Respawnable:    true,
		InterestRadius: 15.0,
		CanTakeDamage:  true,
	}
}

// GetSmallRockTemplate returns a small rock that can be broken
func GetSmallRockTemplate() *DestructibleTemplate {
	return &DestructibleTemplate{
		ID:             "small_rock",
		Name:           "Small Rock",
		Type:           DestructibleTypeRock,
		MaxHealth:      80,
		RequiresTool:   true,
		RequiredToolID: "pickaxe", // Requires pickaxe
		LootTable: &LootTable{
			Items: []LootItem{
				{ItemID: "stone", DropChance: 0.9, MinQuantity: 1, MaxQuantity: 3},
				{ItemID: "iron_ore", DropChance: 0.2, MinQuantity: 1, MaxQuantity: 1},
			},
		},
		Respawnable:    true,
		InterestRadius: 20.0,
		CanTakeDamage:  true,
	}
}

// GetPotTemplate returns a decorative pot
func GetPotTemplate() *DestructibleTemplate {
	return &DestructibleTemplate{
		ID:        "ceramic_pot",
		Name:      "Ceramic Pot",
		Type:      DestructibleTypePot,
		MaxHealth: 15,
		LootTable: &LootTable{
			Items: []LootItem{
				{ItemID: "clay_shard", DropChance: 0.8, MinQuantity: 1, MaxQuantity: 2},
				{ItemID: "small_coin", DropChance: 0.6, MinQuantity: 1, MaxQuantity: 10},
				{ItemID: "key", DropChance: 0.05, MinQuantity: 1, MaxQuantity: 1},
			},
		},
		Respawnable:    false, // Pots don't respawn
		InterestRadius: 15.0,
		CanTakeDamage:  true,
	}
}

// GetTreasureChestTemplate returns a valuable chest
func GetTreasureChestTemplate() *DestructibleTemplate {
	return &DestructibleTemplate{
		ID:        "treasure_chest",
		Name:      "Treasure Chest",
		Type:      DestructibleTypeChest,
		MaxHealth: 200,
		LootTable: &LootTable{
			Items: []LootItem{
				{ItemID: "large_coin", DropChance: 1.0, MinQuantity: 50, MaxQuantity: 200},
				{ItemID: "rare_gem", DropChance: 0.3, MinQuantity: 1, MaxQuantity: 3},
				{ItemID: "epic_weapon", DropChance: 0.1, MinQuantity: 1, MaxQuantity: 1},
				{ItemID: "potion_health", DropChance: 0.8, MinQuantity: 2, MaxQuantity: 5},
			},
		},
		Respawnable:    false, // Chests don't respawn
		InterestRadius: 30.0,
		CanTakeDamage:  true,
	}
}

// GetInvulnerableObjectTemplate returns an indestructible object
func GetInvulnerableObjectTemplate() *DestructibleTemplate {
	return &DestructibleTemplate{
		ID:             "stone_pillar",
		Name:           "Stone Pillar",
		Type:           DestructibleTypeRock,
		MaxHealth:      99999,
		CanTakeDamage:  false, // Cannot be damaged
		Respawnable:    false,
		InterestRadius: 25.0,
	}
}
