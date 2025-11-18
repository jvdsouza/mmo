package combat

import (
	"time"
)

// AbilityType represents different categories of abilities
type AbilityType int

const (
	AbilityTypeMelee AbilityType = iota
	AbilityTypeRanged
	AbilityTypeAoE
	AbilityTypeBuff
	AbilityTypeHeal
)

// ResourceCost represents the resources required to use an ability
type ResourceCost struct {
	Mana    int
	Stamina int
}

// CombatResult represents the outcome of using an ability
type CombatResult struct {
	AbilityID    string
	CasterID     string
	TargetID     string
	Damage       int
	Hit          bool
	CriticalHit  bool
	TargetHealth int
	TargetMaxHealth int
	Timestamp    time.Time
}

// Ability is the core interface that all abilities must implement
type Ability interface {
	GetID() string
	GetType() AbilityType
	GetCooldown() time.Duration
	GetRange() float64
	GetCost() ResourceCost

	// Validate checks if the ability can be used
	Validate(caster Player, target Player) error

	// Execute performs the ability effect
	Execute(caster Player, target Player) *CombatResult
}

// Player interface - abilities need access to player data
type Player interface {
	GetID() string
	GetPosition() Vector3
	GetStats() *PlayerStats
	GetCooldownManager() *CooldownManager
	TakeDamage(amount int)
	IsAlive() bool
}

// PlayerStats represents a player's combat statistics
type PlayerStats struct {
	Health       int
	MaxHealth    int
	Mana         int
	MaxMana      int
	Stamina      int
	MaxStamina   int
	AttackPower  int
	Defense      int
	CritChance   float64
	CritMultiplier float64
	IsDead       bool
}

// Vector3 represents a 3D position
type Vector3 struct {
	X float64
	Y float64
	Z float64
}

// Distance calculates the distance between two vectors
func (v Vector3) Distance(other Vector3) float64 {
	dx := v.X - other.X
	dy := v.Y - other.Y
	dz := v.Z - other.Z
	return sqrt(dx*dx + dy*dy + dz*dz)
}

// Normalize returns a unit vector in the same direction
func (v Vector3) Normalize() Vector3 {
	length := sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
	if length == 0 {
		return Vector3{X: 0, Y: 0, Z: 0}
	}
	return Vector3{
		X: v.X / length,
		Y: v.Y / length,
		Z: v.Z / length,
	}
}

func sqrt(x float64) float64 {
	// Simple approximation - in production use math.Sqrt
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

// BaseAbility provides default implementations for common ability functionality
type BaseAbility struct {
	ID       string
	Type     AbilityType
	Cooldown time.Duration
	Range    float64
	Cost     ResourceCost
}

func (b *BaseAbility) GetID() string {
	return b.ID
}

func (b *BaseAbility) GetType() AbilityType {
	return b.Type
}

func (b *BaseAbility) GetCooldown() time.Duration {
	return b.Cooldown
}

func (b *BaseAbility) GetRange() float64 {
	return b.Range
}

func (b *BaseAbility) GetCost() ResourceCost {
	return b.Cost
}

// Validate provides basic validation (can be overridden)
func (b *BaseAbility) Validate(caster Player, target Player) error {
	return nil // Override in specific ability types
}

// Execute should be overridden by specific ability implementations
func (b *BaseAbility) Execute(caster Player, target Player) *CombatResult {
	return nil // Override in specific ability types
}

// CooldownManager tracks ability cooldowns for a player
type CooldownManager struct {
	cooldowns map[string]time.Time
}

// NewCooldownManager creates a new cooldown manager
func NewCooldownManager() *CooldownManager {
	return &CooldownManager{
		cooldowns: make(map[string]time.Time),
	}
}

// IsReady checks if an ability is off cooldown
func (cm *CooldownManager) IsReady(abilityID string) bool {
	expiry, exists := cm.cooldowns[abilityID]
	if !exists {
		return true
	}
	return time.Now().After(expiry)
}

// IsOnCooldown checks if an ability is currently on cooldown
func (cm *CooldownManager) IsOnCooldown(abilityID string) bool {
	return !cm.IsReady(abilityID)
}

// SetCooldown sets the cooldown for an ability
func (cm *CooldownManager) SetCooldown(abilityID string, duration time.Duration) {
	cm.cooldowns[abilityID] = time.Now().Add(duration)
}

// GetRemaining returns the remaining cooldown time
func (cm *CooldownManager) GetRemaining(abilityID string) time.Duration {
	expiry, exists := cm.cooldowns[abilityID]
	if !exists {
		return 0
	}
	remaining := time.Until(expiry)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// AbilityRegistry stores all registered abilities
type AbilityRegistry struct {
	abilities map[string]Ability
}

// NewAbilityRegistry creates a new ability registry
func NewAbilityRegistry() *AbilityRegistry {
	return &AbilityRegistry{
		abilities: make(map[string]Ability),
	}
}

// Register adds an ability to the registry
func (ar *AbilityRegistry) Register(id string, ability Ability) {
	ar.abilities[id] = ability
}

// Get retrieves an ability by ID
func (ar *AbilityRegistry) Get(id string) Ability {
	return ar.abilities[id]
}

// Has checks if an ability exists
func (ar *AbilityRegistry) Has(id string) bool {
	_, exists := ar.abilities[id]
	return exists
}

// GetAll returns all registered abilities
func (ar *AbilityRegistry) GetAll() map[string]Ability {
	return ar.abilities
}
