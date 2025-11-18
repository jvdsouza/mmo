package combat

import (
	"math/rand"
	"time"
)

// MeleeAbility represents a close-range physical attack
type MeleeAbility struct {
	BaseAbility
	Range float64
}

// NewMeleeAbility creates a basic melee ability
func NewMeleeAbility(id string, cooldown time.Duration, attackRange float64, cost ResourceCost) *MeleeAbility {
	return &MeleeAbility{
		BaseAbility: BaseAbility{
			ID:       id,
			Type:     AbilityTypeMelee,
			Cooldown: cooldown,
			Cost:     cost,
		},
		Range: attackRange,
	}
}

func (m *MeleeAbility) GetRange() float64 {
	return m.Range
}

// Validate checks if melee attack can be used
func (m *MeleeAbility) Validate(caster, target Player) error {
	// Check distance
	distance := caster.GetPosition().Distance(target.GetPosition())
	if distance > m.Range {
		return ErrOutOfRange
	}

	return nil
}

// Execute performs the melee attack
func (m *MeleeAbility) Execute(caster, target Player) *CombatResult {
	// Cast to CombatEntity
	casterEntity := caster.(CombatEntity)
	targetEntity := target.(CombatEntity)

	// Calculate damage
	damage := CalculateMeleeDamage(casterEntity, targetEntity)

	// Check for critical hit
	isCrit := CheckCriticalHit(casterEntity)
	if isCrit {
		critMultiplier := GetCritMultiplier(casterEntity)
		damage = int(float64(damage) * critMultiplier)
	}

	return &CombatResult{
		AbilityID:    m.ID,
		CasterID:     casterEntity.GetID(),
		TargetID:     targetEntity.GetID(),
		Damage:       damage,
		Hit:          true,
		CriticalHit:  isCrit,
		Timestamp:    time.Now(),
	}
}

// CalculateMeleeDamage calculates base melee damage
func CalculateMeleeDamage(caster, target CombatEntity) int {
	// Get stats through the interface
	casterStats := GetEntityStats(caster)
	targetStats := GetEntityStats(target)

	// Base damage = Attack - Defense
	baseDamage := casterStats.AttackPower - targetStats.Defense

	// Minimum damage
	if baseDamage < 1 {
		baseDamage = 1
	}

	// Add some randomness (90-110%)
	variance := 0.9 + rand.Float64()*0.2
	finalDamage := int(float64(baseDamage) * variance)

	return finalDamage
}

// CheckCriticalHit checks if attack is a critical hit
func CheckCriticalHit(entity CombatEntity) bool {
	stats := GetEntityStats(entity)
	return rand.Float64() < stats.CritChance
}

// GetCritMultiplier returns the critical hit multiplier
func GetCritMultiplier(entity CombatEntity) float64 {
	stats := GetEntityStats(entity)
	return stats.CritMultiplier
}

// GetEntityStats extracts combat stats from any entity
func GetEntityStats(entity CombatEntity) *PlayerStats {
	// This function bridges CombatEntity and stats
	// Different entity types will return their stats differently

	switch e := entity.(type) {
	case Player:
		return e.GetStats()
	default:
		// Default stats for entities without full stat system
		return &PlayerStats{
			AttackPower:    10,
			Defense:        5,
			CritChance:     0.05,
			CritMultiplier: 1.5,
		}
	}
}

// Predefined melee abilities
var (
	// BasicAttack - Fast, low damage, no cooldown
	BasicAttack = NewMeleeAbility(
		"basic_attack",
		0,    // No cooldown
		2.5,  // 2.5 meter range
		ResourceCost{Mana: 0, Stamina: 0},
	)

	// HeavyAttack - Slow, high damage, cooldown
	HeavyAttack = NewMeleeAbility(
		"heavy_attack",
		3*time.Second, // 3 second cooldown
		3.0,           // 3 meter range
		ResourceCost{Mana: 0, Stamina: 20},
	)

	// PowerStrike - Very high damage, long cooldown
	PowerStrike = NewMeleeAbility(
		"power_strike",
		10*time.Second, // 10 second cooldown
		2.5,            // 2.5 meter range
		ResourceCost{Mana: 0, Stamina: 40},
	)
)

// RegisterDefaultMeleeAbilities registers default melee abilities
func RegisterDefaultMeleeAbilities(registry *AbilityRegistry) {
	registry.Register(BasicAttack.ID, BasicAttack)
	registry.Register(HeavyAttack.ID, HeavyAttack)
	registry.Register(PowerStrike.ID, PowerStrike)
}
