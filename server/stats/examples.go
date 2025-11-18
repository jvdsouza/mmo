package stats

import "time"

// This file contains examples of how to extend the stats system
// using the Open-Closed Principle

// ============================================================================
// EXAMPLE 1: Adding a New Stat Type
// ============================================================================

// Just add a new constant - no modification to existing code!
const (
	StatEvasion     StatType = "evasion"      // NEW STAT!
	StatArmor       StatType = "armor"        // NEW STAT!
	StatMagicResist StatType = "magic_resist" // NEW STAT!
	StatSpeed       StatType = "speed"        // NEW STAT!
)

// Use it immediately:
// container.SetBaseStat(StatEvasion, 25.0)

// ============================================================================
// EXAMPLE 2: Adding a New Character Class
// ============================================================================

// ArcherArchetype - NEW CHARACTER CLASS!
type ArcherArchetype struct {
	level int
}

func NewArcherArchetype(level int) *ArcherArchetype {
	return &ArcherArchetype{level: level}
}

func (a *ArcherArchetype) GetName() string {
	return "Archer"
}

func (a *ArcherArchetype) GetBaseStats() map[StatType]float64 {
	return map[StatType]float64{
		// Balanced resources
		StatMaxHealth:  100,
		StatHealth:     100,
		StatMaxMana:    100,
		StatMana:       100,
		StatMaxStamina: 120, // High stamina for abilities

		// Moderate attack, low defense, high crit
		StatAttackPower:    15,
		StatDefense:        8,
		StatCritChance:     0.30, // 30% crit chance!
		StatCritMultiplier: 1.8,

		// NEW STATS!
		StatEvasion: 20, // High evasion
		StatSpeed:   1.2, // Faster movement
	}
}

func (a *ArcherArchetype) GetStatGrowth() map[StatType]float64 {
	return map[StatType]float64{
		StatMaxHealth:   7,
		StatMaxMana:     5,
		StatMaxStamina:  8,
		StatAttackPower: 2,
		StatCritChance:  0.01, // +1% crit per level
		StatEvasion:     1,    // +1 evasion per level
	}
}

func (a *ArcherArchetype) GetPassiveModifiers() []StatModifier {
	// "Eagle Eye": +50% crit multiplier when stamina > 90%
	eagleEye := NewConditionalModifier(
		"eagle_eye",
		NewPercentModifier("eagle_eye_effect", 0.50, 0),
		func(stats *StatContainer) bool {
			stamina := stats.GetStat(StatStamina)
			maxStamina := stats.GetStat(StatMaxStamina)
			return stamina > maxStamina*0.9
		},
	)

	return []StatModifier{eagleEye}
}

// ApplyArcherPassives applies archer-specific passives
func ApplyArcherPassives(container *StatContainer, archetype *ArcherArchetype) {
	for _, modifier := range archetype.GetPassiveModifiers() {
		container.AddModifier(StatCritMultiplier, modifier)
	}
}

// ============================================================================
// EXAMPLE 3: Adding a New Modifier Type
// ============================================================================

// DiminishingModifier - applies diminishing returns
type DiminishingModifier struct {
	ID        string
	Value     float64
	Threshold float64
	Diminish  float64 // 0.5 = 50% effective after threshold
	Duration  time.Duration
	StartTime time.Time
}

func NewDiminishingModifier(id string, value, threshold, diminish float64, duration time.Duration) *DiminishingModifier {
	return &DiminishingModifier{
		ID:        id,
		Value:     value,
		Threshold: threshold,
		Diminish:  diminish,
		Duration:  duration,
		StartTime: time.Now(),
	}
}

func (dm *DiminishingModifier) GetID() string {
	return dm.ID
}

func (dm *DiminishingModifier) GetType() ModifierType {
	return ModifierTypeFlat // Custom type
}

func (dm *DiminishingModifier) GetValue() float64 {
	return dm.Value
}

func (dm *DiminishingModifier) GetDuration() time.Duration {
	return dm.Duration
}

func (dm *DiminishingModifier) IsExpired() bool {
	if dm.Duration == 0 {
		return false
	}
	return time.Since(dm.StartTime) > dm.Duration
}

func (dm *DiminishingModifier) Apply(baseValue float64) float64 {
	if baseValue > dm.Threshold {
		excess := baseValue - dm.Threshold
		diminished := excess * dm.Diminish
		return dm.Threshold + diminished + dm.Value
	}
	return baseValue + dm.Value
}

// Usage:
// Armor has diminishing returns after 100
// container.AddModifier(StatDefense, NewDiminishingModifier("armor_1", 50, 100, 0.5, 0))

// ============================================================================
// EXAMPLE 4: Adding Equipment System
// ============================================================================

// EquipmentSlot represents where an item can be equipped
type EquipmentSlot int

const (
	SlotHead EquipmentSlot = iota
	SlotChest
	SlotLegs
	SlotFeet
	SlotMainHand
	SlotOffHand
	SlotRing1
	SlotRing2
)

// Equipment represents an equippable item
type Equipment struct {
	ID        string
	Name      string
	Slot      EquipmentSlot
	Modifiers map[StatType][]StatModifier
}

// IronSword example weapon
func NewIronSword() *Equipment {
	return &Equipment{
		ID:   "iron_sword_1",
		Name: "Iron Sword",
		Slot: SlotMainHand,
		Modifiers: map[StatType][]StatModifier{
			StatAttackPower: {
				NewFlatModifier("iron_sword_atk", 10, 0), // +10 attack
			},
		},
	}
}

// SteelHelmet example armor
func NewSteelHelmet() *Equipment {
	return &Equipment{
		ID:   "steel_helmet_1",
		Name: "Steel Helmet",
		Slot: SlotHead,
		Modifiers: map[StatType][]StatModifier{
			StatDefense: {
				NewFlatModifier("steel_helmet_def", 5, 0), // +5 defense
			},
			StatMaxHealth: {
				NewFlatModifier("steel_helmet_hp", 20, 0), // +20 max HP
			},
		},
	}
}

// Usage:
// sword := NewIronSword()
// for statType, mods := range sword.Modifiers {
//     for _, mod := range mods {
//         player.Stats.AddModifier(statType, mod)
//     }
// }

// ============================================================================
// EXAMPLE 5: Adding Buff/Debuff System
// ============================================================================

// Buff represents a temporary stat effect
type Buff struct {
	ID        string
	Name      string
	Modifiers map[StatType][]StatModifier
	Duration  time.Duration
}

// StrengthPotion example buff
func NewStrengthPotion() *Buff {
	return &Buff{
		ID:   "strength_potion_1",
		Name: "Strength Potion",
		Modifiers: map[StatType][]StatModifier{
			StatAttackPower: {
				NewPercentModifier("str_pot_atk", 0.5, 30*time.Second), // +50% attack for 30s
			},
		},
		Duration: 30 * time.Second,
	}
}

// PoisonDebuff example debuff
func NewPoisonDebuff() *Buff {
	return &Buff{
		ID:   "poison_1",
		Name: "Poison",
		Modifiers: map[StatType][]StatModifier{
			StatMaxHealth: {
				NewFlatModifier("poison_hp", -10, 10*time.Second), // -10 max HP for 10s
			},
		},
		Duration: 10 * time.Second,
	}
}

// Usage:
// potion := NewStrengthPotion()
// for statType, mods := range potion.Modifiers {
//     for _, mod := range mods {
//         player.Stats.AddModifier(statType, mod)
//     }
// }
// Modifiers automatically expire after duration!

// ============================================================================
// EXAMPLE 6: Complex Conditional Passive
// ============================================================================

// RogueArchetype with complex passives
type RogueArchetype struct {
	level int
}

func NewRogueArchetype(level int) *RogueArchetype {
	return &RogueArchetype{level: level}
}

func (r *RogueArchetype) GetName() string {
	return "Rogue"
}

func (r *RogueArchetype) GetBaseStats() map[StatType]float64 {
	return map[StatType]float64{
		StatMaxHealth:   90,
		StatHealth:      90,
		StatMaxMana:     80,
		StatMana:        80,
		StatMaxStamina:  150, // Very high stamina
		StatStamina:     150,
		StatAttackPower: 18,
		StatDefense:     6,
		StatCritChance:  0.35, // Highest crit chance
		StatEvasion:     30,   // High evasion
		StatSpeed:       1.3,  // Fastest class
	}
}

func (r *RogueArchetype) GetStatGrowth() map[StatType]float64 {
	return map[StatType]float64{
		StatMaxHealth:   6,
		StatAttackPower: 2,
		StatCritChance:  0.015, // +1.5% crit per level
		StatEvasion:     2,
	}
}

func (r *RogueArchetype) GetPassiveModifiers() []StatModifier {
	// "Assassinate": +100% crit multiplier when attacking from behind (simulated by stamina > 80%)
	assassinate := NewConditionalModifier(
		"assassinate",
		NewPercentModifier("assassinate_effect", 1.0, 0), // +100% crit multi
		func(stats *StatContainer) bool {
			stamina := stats.GetStat(StatStamina)
			maxStamina := stats.GetStat(StatMaxStamina)
			return stamina > maxStamina*0.8
		},
	)

	// "Shadow Step": +20% evasion when health < 30%
	shadowStep := NewConditionalModifier(
		"shadow_step",
		NewPercentModifier("shadow_step_effect", 0.20, 0),
		func(stats *StatContainer) bool {
			health := stats.GetStat(StatHealth)
			maxHealth := stats.GetStat(StatMaxHealth)
			return health < maxHealth*0.3
		},
	)

	return []StatModifier{assassinate, shadowStep}
}

// ApplyRoguePassives applies rogue-specific passives
func ApplyRoguePassives(container *StatContainer, archetype *RogueArchetype) {
	passives := archetype.GetPassiveModifiers()
	container.AddModifier(StatCritMultiplier, passives[0]) // Assassinate
	container.AddModifier(StatEvasion, passives[1])        // Shadow Step
}

// ============================================================================
// SUMMARY: How to Extend
// ============================================================================

/*
Adding New Features:

1. NEW STAT:
   - Add constant: const StatNewStat StatType = "new_stat"
   - Use immediately: container.SetBaseStat(StatNewStat, value)

2. NEW CHARACTER CLASS:
   - Implement CharacterArchetype interface
   - Define GetBaseStats(), GetStatGrowth(), GetPassiveModifiers()
   - Done! No changes to stat system

3. NEW MODIFIER TYPE:
   - Implement StatModifier interface
   - Define GetID(), GetType(), GetValue(), Apply(), etc.
   - Use with AddModifier()

4. NEW EQUIPMENT:
   - Create Equipment struct with modifiers
   - Apply modifiers when equipped
   - Remove modifiers when unequipped

5. NEW BUFF/DEBUFF:
   - Create Buff struct with modifiers and duration
   - Apply modifiers
   - Automatically expires after duration

Everything extends the base system without modifying it!
This is the Open-Closed Principle in action.
*/
