package stats

import (
	"sync"
)

// StatContainer manages a collection of stats and their modifiers
type StatContainer struct {
	stats     map[StatType]*BaseStat
	modifiers map[StatType][]StatModifier
	mu        sync.RWMutex
}

// NewStatContainer creates a new stat container
func NewStatContainer() *StatContainer {
	return &StatContainer{
		stats:     make(map[StatType]*BaseStat),
		modifiers: make(map[StatType][]StatModifier),
	}
}

// SetBaseStat sets the base value for a stat
func (sc *StatContainer) SetBaseStat(statType StatType, value float64) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if stat, exists := sc.stats[statType]; exists {
		stat.SetBaseValue(value)
	} else {
		sc.stats[statType] = NewBaseStat(statType, value)
	}
	sc.recalculateStat(statType)
}

// GetBaseStat returns the base value of a stat (without modifiers)
func (sc *StatContainer) GetBaseStat(statType StatType) float64 {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if stat, exists := sc.stats[statType]; exists {
		return stat.GetBaseValue()
	}
	return 0
}

// GetStat returns the current calculated value of a stat (with modifiers)
func (sc *StatContainer) GetStat(statType StatType) float64 {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if stat, exists := sc.stats[statType]; exists {
		return stat.GetCurrentValue()
	}
	return 0
}

// AddModifier adds a modifier to a stat
func (sc *StatContainer) AddModifier(statType StatType, modifier StatModifier) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Set container reference for conditional modifiers
	if condMod, ok := modifier.(*ConditionalModifier); ok {
		condMod.SetContainer(sc)
	}

	sc.modifiers[statType] = append(sc.modifiers[statType], modifier)
	sc.recalculateStat(statType)
}

// RemoveModifier removes a modifier by ID
func (sc *StatContainer) RemoveModifier(modifierID string) bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for statType, mods := range sc.modifiers {
		for i, mod := range mods {
			if mod.GetID() == modifierID {
				// Remove modifier from slice
				sc.modifiers[statType] = append(mods[:i], mods[i+1:]...)
				sc.recalculateStat(statType)
				return true
			}
		}
	}
	return false
}

// RemoveExpiredModifiers removes all expired modifiers
func (sc *StatContainer) RemoveExpiredModifiers() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	recalculate := make(map[StatType]bool)

	for statType, mods := range sc.modifiers {
		newMods := make([]StatModifier, 0, len(mods))
		for _, mod := range mods {
			if !mod.IsExpired() {
				newMods = append(newMods, mod)
			} else {
				recalculate[statType] = true
			}
		}
		sc.modifiers[statType] = newMods
	}

	// Recalculate affected stats
	for statType := range recalculate {
		sc.recalculateStat(statType)
	}
}

// GetModifiers returns all modifiers for a stat
func (sc *StatContainer) GetModifiers(statType StatType) []StatModifier {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return sc.modifiers[statType]
}

// RecalculateAll recalculates all stats
func (sc *StatContainer) RecalculateAll() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for statType := range sc.stats {
		sc.recalculateStat(statType)
	}
}

// recalculateStat recalculates a single stat (must be called with lock held)
func (sc *StatContainer) recalculateStat(statType StatType) {
	stat, exists := sc.stats[statType]
	if !exists {
		return
	}

	baseValue := stat.GetBaseValue()
	currentValue := baseValue

	// Apply modifiers in order: Flat -> Percent -> Multiplier
	modifiers := sc.modifiers[statType]

	// Phase 1: Flat modifiers
	for _, mod := range modifiers {
		if !mod.IsExpired() && mod.GetType() == ModifierTypeFlat {
			currentValue = mod.Apply(currentValue)
		}
	}

	// Phase 2: Percent modifiers (additive with each other)
	percentBonus := 0.0
	for _, mod := range modifiers {
		if !mod.IsExpired() && mod.GetType() == ModifierTypePercent {
			percentBonus += mod.GetValue()
		}
	}
	if percentBonus > 0 {
		currentValue = currentValue * (1.0 + percentBonus)
	}

	// Phase 3: Multiplier modifiers (multiplicative with each other)
	for _, mod := range modifiers {
		if !mod.IsExpired() && mod.GetType() == ModifierTypeMultiplier {
			currentValue = mod.Apply(currentValue)
		}
	}

	stat.SetCurrentValue(currentValue)
}

// GetAllStats returns a map of all stats with their current values
func (sc *StatContainer) GetAllStats() map[StatType]float64 {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	result := make(map[StatType]float64)
	for statType, stat := range sc.stats {
		result[statType] = stat.GetCurrentValue()
	}
	return result
}

// HasStat checks if a stat exists in the container
func (sc *StatContainer) HasStat(statType StatType) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	_, exists := sc.stats[statType]
	return exists
}

// ModifyCurrentValue directly modifies the current value (for things like damage/healing)
// This bypasses the base value and modifiers
func (sc *StatContainer) ModifyCurrentValue(statType StatType, delta float64) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if stat, exists := sc.stats[statType]; exists {
		newValue := stat.GetCurrentValue() + delta
		stat.SetCurrentValue(newValue)
	}
}

// RestoreStat restores a stat to its calculated maximum
func (sc *StatContainer) RestoreStat(statType StatType) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.recalculateStat(statType)
}
