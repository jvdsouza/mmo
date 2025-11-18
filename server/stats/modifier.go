package stats

import (
	"time"
)

// ModifierType determines how a modifier is applied
type ModifierType int

const (
	ModifierTypeFlat ModifierType = iota // +10 Attack
	ModifierTypePercent                  // +50% Attack (additive)
	ModifierTypeMultiplier               // x2 Attack (multiplicative)
)

// StatModifier changes a stat value
type StatModifier interface {
	GetID() string
	GetType() ModifierType
	GetValue() float64
	GetDuration() time.Duration
	IsExpired() bool
	Apply(baseValue float64) float64
}

// BaseModifier provides common modifier functionality
type BaseModifier struct {
	ID        string
	Type      ModifierType
	Value     float64
	Duration  time.Duration
	StartTime time.Time
}

// NewBaseModifier creates a new base modifier
func NewBaseModifier(id string, modType ModifierType, value float64, duration time.Duration) *BaseModifier {
	return &BaseModifier{
		ID:        id,
		Type:      modType,
		Value:     value,
		Duration:  duration,
		StartTime: time.Now(),
	}
}

func (bm *BaseModifier) GetID() string {
	return bm.ID
}

func (bm *BaseModifier) GetType() ModifierType {
	return bm.Type
}

func (bm *BaseModifier) GetValue() float64 {
	return bm.Value
}

func (bm *BaseModifier) GetDuration() time.Duration {
	return bm.Duration
}

func (bm *BaseModifier) IsExpired() bool {
	if bm.Duration == 0 {
		return false // Permanent modifier
	}
	return time.Since(bm.StartTime) > bm.Duration
}

func (bm *BaseModifier) Apply(baseValue float64) float64 {
	switch bm.Type {
	case ModifierTypeFlat:
		return baseValue + bm.Value
	case ModifierTypePercent:
		return baseValue * (1.0 + bm.Value)
	case ModifierTypeMultiplier:
		return baseValue * bm.Value
	default:
		return baseValue
	}
}

// FlatModifier adds a fixed amount to a stat
type FlatModifier struct {
	*BaseModifier
}

// NewFlatModifier creates a modifier that adds a fixed value
func NewFlatModifier(id string, value float64, duration time.Duration) *FlatModifier {
	return &FlatModifier{
		BaseModifier: NewBaseModifier(id, ModifierTypeFlat, value, duration),
	}
}

// PercentModifier adds a percentage to a stat
type PercentModifier struct {
	*BaseModifier
}

// NewPercentModifier creates a modifier that adds a percentage
// value should be 0.5 for +50%, 1.0 for +100%, etc.
func NewPercentModifier(id string, value float64, duration time.Duration) *PercentModifier {
	return &PercentModifier{
		BaseModifier: NewBaseModifier(id, ModifierTypePercent, value, duration),
	}
}

// MultiplierModifier multiplies a stat
type MultiplierModifier struct {
	*BaseModifier
}

// NewMultiplierModifier creates a modifier that multiplies the stat
// value should be 2.0 for x2, 0.5 for half, etc.
func NewMultiplierModifier(id string, value float64, duration time.Duration) *MultiplierModifier {
	return &MultiplierModifier{
		BaseModifier: NewBaseModifier(id, ModifierTypeMultiplier, value, duration),
	}
}

// ConditionalModifier applies a modifier based on a condition
type ConditionalModifier struct {
	ID            string
	BaseModifier  StatModifier
	Condition     func(*StatContainer) bool
	condContainer *StatContainer
}

// NewConditionalModifier creates a conditional modifier
func NewConditionalModifier(id string, mod StatModifier, condition func(*StatContainer) bool) *ConditionalModifier {
	return &ConditionalModifier{
		ID:           id,
		BaseModifier: mod,
		Condition:    condition,
	}
}

func (cm *ConditionalModifier) GetID() string {
	return cm.ID
}

func (cm *ConditionalModifier) GetType() ModifierType {
	return cm.BaseModifier.GetType()
}

func (cm *ConditionalModifier) GetValue() float64 {
	return cm.BaseModifier.GetValue()
}

func (cm *ConditionalModifier) GetDuration() time.Duration {
	return cm.BaseModifier.GetDuration()
}

func (cm *ConditionalModifier) IsExpired() bool {
	return cm.BaseModifier.IsExpired()
}

func (cm *ConditionalModifier) Apply(baseValue float64) float64 {
	if cm.condContainer != nil && cm.Condition(cm.condContainer) {
		return cm.BaseModifier.Apply(baseValue)
	}
	return baseValue
}

// SetContainer sets the stat container reference for condition checking
func (cm *ConditionalModifier) SetContainer(container *StatContainer) {
	cm.condContainer = container
}
