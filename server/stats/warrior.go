package stats

// WarriorArchetype defines the Warrior character class
type WarriorArchetype struct {
	level int
}

// NewWarriorArchetype creates a new warrior archetype
func NewWarriorArchetype(level int) *WarriorArchetype {
	return &WarriorArchetype{
		level: level,
	}
}

func (w *WarriorArchetype) GetName() string {
	return "Warrior"
}

func (w *WarriorArchetype) GetBaseStats() map[StatType]float64 {
	return map[StatType]float64{
		// Resources - Warriors have high HP, low mana
		StatMaxHealth:  150,
		StatHealth:     150,
		StatMaxMana:    50,
		StatMana:       50,
		StatMaxStamina: 100,
		StatStamina:    100,

		// Combat - Warriors are strong attackers and defenders
		StatAttackPower:    20,
		StatDefense:        15,
		StatCritChance:     0.10, // 10% crit chance
		StatCritMultiplier: 1.5,  // 1.5x damage on crit
	}
}

func (w *WarriorArchetype) GetStatGrowth() map[StatType]float64 {
	return map[StatType]float64{
		StatMaxHealth:   10, // +10 HP per level
		StatMaxMana:     2,  // +2 mana per level
		StatMaxStamina:  5,  // +5 stamina per level
		StatAttackPower: 2,  // +2 attack per level
		StatDefense:     1,  // +1 defense per level
	}
}

func (w *WarriorArchetype) GetPassiveModifiers() []StatModifier {
	// Berserker Rage: +20% attack when health < 50%
	berserkerRage := NewConditionalModifier(
		"berserker_rage",
		NewPercentModifier("berserker_rage_effect", 0.20, 0), // Permanent
		func(stats *StatContainer) bool {
			health := stats.GetStat(StatHealth)
			maxHealth := stats.GetStat(StatMaxHealth)
			return health < maxHealth*0.5
		},
	)

	return []StatModifier{
		berserkerRage,
	}
}

// ApplyWarriorPassives applies warrior-specific passive modifiers
func ApplyWarriorPassives(container *StatContainer, archetype *WarriorArchetype) {
	for _, modifier := range archetype.GetPassiveModifiers() {
		// Berserker Rage affects attack power
		container.AddModifier(StatAttackPower, modifier)
	}
}
