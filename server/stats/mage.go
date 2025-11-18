package stats

// MageArchetype defines the Mage character class
type MageArchetype struct {
	level int
}

// NewMageArchetype creates a new mage archetype
func NewMageArchetype(level int) *MageArchetype {
	return &MageArchetype{
		level: level,
	}
}

func (m *MageArchetype) GetName() string {
	return "Mage"
}

func (m *MageArchetype) GetBaseStats() map[StatType]float64 {
	return map[StatType]float64{
		// Resources - Mages have low HP, high mana
		StatMaxHealth:  80,
		StatHealth:     80,
		StatMaxMana:    200,
		StatMana:       200,
		StatMaxStamina: 60,
		StatStamina:    60,

		// Combat - Mages are weak physically but have high crit
		StatAttackPower:    5,
		StatDefense:        5,
		StatCritChance:     0.25, // 25% crit chance
		StatCritMultiplier: 2.0,  // 2x damage on crit
	}
}

func (m *MageArchetype) GetStatGrowth() map[StatType]float64 {
	return map[StatType]float64{
		StatMaxHealth:   5,  // +5 HP per level
		StatMaxMana:     15, // +15 mana per level
		StatMaxStamina:  3,  // +3 stamina per level
		StatAttackPower: 1,  // +1 attack per level
		StatDefense:     0,  // No defense growth
	}
}

func (m *MageArchetype) GetPassiveModifiers() []StatModifier {
	// Arcane Mastery: +30% crit multiplier when mana > 80%
	arcaneMastery := NewConditionalModifier(
		"arcane_mastery",
		NewPercentModifier("arcane_mastery_effect", 0.30, 0), // Permanent
		func(stats *StatContainer) bool {
			mana := stats.GetStat(StatMana)
			maxMana := stats.GetStat(StatMaxMana)
			return mana > maxMana*0.8
		},
	)

	return []StatModifier{
		arcaneMastery,
	}
}

// ApplyMagePassives applies mage-specific passive modifiers
func ApplyMagePassives(container *StatContainer, archetype *MageArchetype) {
	for _, modifier := range archetype.GetPassiveModifiers() {
		// Arcane Mastery affects crit multiplier
		container.AddModifier(StatCritMultiplier, modifier)
	}
}
