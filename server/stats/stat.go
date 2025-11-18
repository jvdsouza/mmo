package stats

// StatType represents different types of statistics
type StatType string

const (
	// Resources
	StatHealth     StatType = "health"
	StatMaxHealth  StatType = "max_health"
	StatMana       StatType = "mana"
	StatMaxMana    StatType = "max_mana"
	StatStamina    StatType = "stamina"
	StatMaxStamina StatType = "max_stamina"

	// Combat Stats
	StatAttackPower    StatType = "attack_power"
	StatDefense        StatType = "defense"
	StatCritChance     StatType = "crit_chance"
	StatCritMultiplier StatType = "crit_multiplier"

	// Future stats - easy to add!
	// StatEvasion      StatType = "evasion"
	// StatArmor        StatType = "armor"
	// StatMagicResist  StatType = "magic_resist"
	// StatSpeed        StatType = "speed"
)

// Stat represents any character statistic
type Stat interface {
	GetName() StatType
	GetBaseValue() float64
	GetCurrentValue() float64
	SetBaseValue(value float64)
	SetCurrentValue(value float64)
}

// BaseStat is the default implementation of Stat
type BaseStat struct {
	Name         StatType
	BaseValue    float64
	CurrentValue float64
}

// NewBaseStat creates a new base stat
func NewBaseStat(name StatType, baseValue float64) *BaseStat {
	return &BaseStat{
		Name:         name,
		BaseValue:    baseValue,
		CurrentValue: baseValue,
	}
}

func (s *BaseStat) GetName() StatType {
	return s.Name
}

func (s *BaseStat) GetBaseValue() float64 {
	return s.BaseValue
}

func (s *BaseStat) GetCurrentValue() float64 {
	return s.CurrentValue
}

func (s *BaseStat) SetBaseValue(value float64) {
	s.BaseValue = value
}

func (s *BaseStat) SetCurrentValue(value float64) {
	s.CurrentValue = value
}
