package combat

// EntityType represents different types of combat entities
type EntityType int

const (
	EntityTypePlayer EntityType = iota
	EntityTypeMonster
	EntityTypeBoss
	EntityTypeNPC
	EntityTypeDestructible  // Crates, barrels, doors
	EntityTypeHarvestable   // Trees, rocks, grass
	EntityTypeEnvironmental // Traps, lava, poison zones
	EntityTypeProjectile    // Arrows, fireballs
)

// Faction represents entity alignment for targeting
type Faction int

const (
	FactionPlayer Faction = iota
	FactionHostileMob
	FactionPassiveMob
	FactionNeutral
	FactionAllied
	FactionEnvironment
	FactionGuild1
	FactionGuild2
	// More factions can be added
)

// CombatEntity is the universal interface for anything that can participate in combat
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

	// Network/Optimization
	GetInterestRadius() float64
	ShouldBroadcastCombat() bool
}

// DamageResult represents the outcome of damage dealt
type DamageResult struct {
	TargetID       string
	Damage         int
	RemainingHP    int
	MaxHP          int
	Died           bool
	WasBlocked     bool
	WasDodged      bool
	WasCritical    bool
	DamageType     DamageType
	EffectsApplied []string
}

// DamageType for different damage types
type DamageType int

const (
	DamageTypePhysical DamageType = iota
	DamageTypeMagical
	DamageTypeFire
	DamageTypeIce
	DamageTypePoison
	DamageTypeLightning
	DamageTypeHoly
	DamageTypeEnvironmental
)

// FactionRelations defines which factions are hostile to each other
var FactionRelations = map[Faction]map[Faction]bool{
	FactionPlayer: {
		FactionHostileMob: true,  // Players can attack hostile mobs
		FactionPassiveMob: false, // Can't attack passive mobs
		FactionNeutral:    false,
		FactionPlayer:     false, // Can't attack other players (unless PvP)
		FactionAllied:     false,
	},
	FactionHostileMob: {
		FactionPlayer:     true, // Hostile mobs attack players
		FactionPassiveMob: false,
		FactionNeutral:    false,
		FactionHostileMob: false,
		FactionAllied:     false,
	},
	FactionPassiveMob: {
		FactionPlayer:     false, // Passive mobs don't attack
		FactionHostileMob: false,
		FactionPassiveMob: false,
	},
	FactionNeutral: {
		FactionPlayer:     false, // Neutral doesn't attack anyone
		FactionHostileMob: false,
	},
	FactionEnvironment: {
		FactionPlayer:     true,  // Environment damages players
		FactionHostileMob: true,  // And mobs
		FactionPassiveMob: true,
		FactionNeutral:    true,
		FactionAllied:     true,
	},
}

// IsHostile checks if two entities are hostile based on faction
func IsHostile(attacker, target CombatEntity) bool {
	attackerFaction := attacker.GetFaction()
	targetFaction := target.GetFaction()

	// Check faction relations
	if relations, ok := FactionRelations[attackerFaction]; ok {
		if hostile, ok := relations[targetFaction]; ok && hostile {
			return true
		}
	}

	// Delegate to entity-specific logic
	return attacker.IsHostileTo(target)
}

// CanAttack validates if an attacker can target an entity
func CanAttack(attacker, target CombatEntity) error {
	// 1. State checks
	if !attacker.CanDealDamage() {
		return ErrCannotAttack
	}
	if !target.CanTakeDamage() {
		return ErrCannotBeDamaged
	}
	if !target.IsAlive() {
		return ErrTargetDead
	}

	// 2. Self-targeting
	if attacker.GetID() == target.GetID() {
		return ErrCannotTargetSelf
	}

	// 3. Target validation (entity-specific)
	if !target.IsValidTarget(attacker) {
		return ErrInvalidTarget
	}

	// 4. Hostility check
	if !IsHostile(attacker, target) {
		return ErrNotHostile
	}

	return nil
}

// Combat errors
var (
	ErrCannotAttack      = &CombatError{Code: "cannot_attack", Message: "Attacker cannot deal damage"}
	ErrCannotBeDamaged   = &CombatError{Code: "cannot_be_damaged", Message: "Target cannot take damage"}
	ErrTargetDead        = &CombatError{Code: "target_dead", Message: "Target is already dead"}
	ErrCannotTargetSelf  = &CombatError{Code: "cannot_target_self", Message: "Cannot target self"}
	ErrInvalidTarget     = &CombatError{Code: "invalid_target", Message: "Invalid target for this entity"}
	ErrNotHostile        = &CombatError{Code: "not_hostile", Message: "Entities are not hostile"}
	ErrOutOfRange        = &CombatError{Code: "out_of_range", Message: "Target out of range"}
	ErrNoLineOfSight     = &CombatError{Code: "no_line_of_sight", Message: "No line of sight to target"}
	ErrAbilityOnCooldown = &CombatError{Code: "ability_on_cooldown", Message: "Ability is on cooldown"}
)

// CombatError represents combat-specific errors
type CombatError struct {
	Code    string
	Message string
}

func (e *CombatError) Error() string {
	return e.Message
}
