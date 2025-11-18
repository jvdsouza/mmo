package stats

// CharacterArchetype defines the base stats and modifiers for a character type
type CharacterArchetype interface {
	GetName() string
	GetBaseStats() map[StatType]float64
	GetStatGrowth() map[StatType]float64  // Per-level growth
	GetPassiveModifiers() []StatModifier  // Passive abilities
}

// ApplyArchetype applies an archetype's stats to a stat container
func ApplyArchetype(container *StatContainer, archetype CharacterArchetype) {
	// Set base stats
	for statType, value := range archetype.GetBaseStats() {
		container.SetBaseStat(statType, value)
	}

	// Add passive modifiers
	// Note: Passive modifiers are defined but need to be applied per-archetype
	// See ApplyWarriorPassives(), ApplyMagePassives(), etc. for examples
	_ = archetype.GetPassiveModifiers() // Acknowledge we're aware of them
}

// LevelUpArchetype applies growth stats based on archetype
func LevelUpArchetype(container *StatContainer, archetype CharacterArchetype) {
	growth := archetype.GetStatGrowth()
	for statType, growthValue := range growth {
		currentBase := container.GetBaseStat(statType)
		container.SetBaseStat(statType, currentBase+growthValue)
	}
}
