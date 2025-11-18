# Extensible Stats System - Open-Closed Principle Design

## Open-Closed Principle (OCP)

**"Software entities should be open for extension but closed for modification"**

This means:
- ✅ **Open for extension**: Can add new stats, modifiers, character types
- ❌ **Closed for modification**: Don't change existing stat system code

---

## Design Goals

### What We Need to Support (Now & Future)

**Current:**
- Base stats: Health, Mana, Stamina, Attack, Defense
- Simple combat calculations

**Future Extensions (WITHOUT modifying core):**
- New stats: CritChance, CritMultiplier, Evasion, Armor, MagicResist
- Character classes: Warrior, Mage, Archer, Rogue
- Gear modifiers: Sword +10 Attack, Helmet +20 Defense
- Temporary buffs: Strength potion +50% Attack for 30s
- Debuffs: Poison -5 Health per second
- Passive abilities: "Berserker Rage" +10% Attack when <50% health
- Stat scaling: Int increases Mana, Str increases Attack
- Level-based growth: Stats increase per level

---

## Architecture

### 1. Stat Interface - What IS a Stat?

```go
// Stat represents any character statistic
type Stat interface {
    GetName() string
    GetBaseValue() float64
    GetCurrentValue() float64
    SetBaseValue(value float64)
}

// StatType enum for common stats
type StatType string

const (
    StatHealth       StatType = "health"
    StatMaxHealth    StatType = "max_health"
    StatMana         StatType = "mana"
    StatMaxMana      StatType = "max_mana"
    StatStamina      StatType = "stamina"
    StatMaxStamina   StatType = "max_stamina"
    StatAttackPower  StatType = "attack_power"
    StatDefense      StatType = "defense"
    StatCritChance   StatType = "crit_chance"
    StatCritMultiplier StatType = "crit_multiplier"
    // Easy to add: StatEvasion, StatArmor, StatSpeed, etc.
)
```

### 2. Stat Modifiers - How Stats Change

```go
// StatModifier changes a stat value
type StatModifier interface {
    GetType() ModifierType
    GetValue() float64
    GetDuration() time.Duration
    Apply(baseValue float64) float64
}

// ModifierType determines how modifier is applied
type ModifierType int

const (
    ModifierTypeFlat ModifierType = iota  // +10 Attack
    ModifierTypePercent                    // +50% Attack
    ModifierTypeMultiplier                 // x2 Attack
)

// Examples:
// Flat: 100 base + 10 flat = 110
// Percent: 100 base + 50% = 150
// Multiplier: 100 base * 2 = 200
```

### 3. Stat Container - Holds All Stats

```go
// StatContainer manages a collection of stats
type StatContainer struct {
    stats     map[StatType]*BaseStat
    modifiers map[StatType][]StatModifier
    mu        sync.RWMutex
}

// Methods:
// - GetStat(statType) - Get final calculated value
// - GetBaseStat(statType) - Get base value only
// - AddModifier(statType, modifier) - Add temporary buff
// - RemoveModifier(modifierID) - Remove buff
// - RecalculateAll() - Update all calculated values
```

### 4. Character Archetype - Different Builds

```go
// CharacterArchetype defines initial stats for character types
type CharacterArchetype interface {
    GetName() string
    GetBaseStats() map[StatType]float64
    GetStatGrowth() map[StatType]float64  // Per-level growth
    GetPassiveModifiers() []StatModifier
}

// Examples:
type WarriorArchetype struct{}
type MageArchetype struct{}
type ArcherArchetype struct{}
```

---

## Implementation Pattern

### Base Implementation (Closed for Modification)

```go
// BaseStat - Core stat implementation
type BaseStat struct {
    Name         StatType
    BaseValue    float64
    CurrentValue float64  // Cached calculated value
}

func (s *BaseStat) GetName() string {
    return string(s.Name)
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
```

### Extensions (Open for Extension)

#### Extension 1: Add New Stat Type
```go
// Just add to the const - no modification to BaseStat!
const (
    StatEvasion StatType = "evasion"  // NEW!
)

// Use it:
container.AddStat(StatEvasion, 10.0)
```

#### Extension 2: Add Modifier Type
```go
// FlatModifier - adds a fixed amount
type FlatModifier struct {
    Value    float64
    Duration time.Duration
}

func (fm *FlatModifier) Apply(baseValue float64) float64 {
    return baseValue + fm.Value
}

// PercentModifier - adds a percentage
type PercentModifier struct {
    Percent  float64  // 0.5 = 50%
    Duration time.Duration
}

func (pm *PercentModifier) Apply(baseValue float64) float64 {
    return baseValue * (1.0 + pm.Percent)
}

// ConditionalModifier - applies based on condition
type ConditionalModifier struct {
    BaseModifier StatModifier
    Condition    func(*StatContainer) bool
}

func (cm *ConditionalModifier) Apply(baseValue float64) float64 {
    if cm.Condition(container) {
        return cm.BaseModifier.Apply(baseValue)
    }
    return baseValue
}
```

#### Extension 3: Add Character Class
```go
// No modification to base stat system!
type WarriorArchetype struct {
    level int
}

func (w *WarriorArchetype) GetBaseStats() map[StatType]float64 {
    return map[StatType]float64{
        StatMaxHealth:   150,  // High HP
        StatMaxMana:     50,   // Low mana
        StatAttackPower: 20,   // High attack
        StatDefense:     15,   // High defense
    }
}

func (w *WarriorArchetype) GetStatGrowth() map[StatType]float64 {
    return map[StatType]float64{
        StatMaxHealth:   10,  // +10 HP per level
        StatAttackPower: 2,   // +2 attack per level
    }
}

func (w *WarriorArchetype) GetPassiveModifiers() []StatModifier {
    return []StatModifier{
        // "Berserker Rage": +20% attack when health < 50%
        &ConditionalModifier{
            BaseModifier: &PercentModifier{Percent: 0.20},
            Condition: func(stats *StatContainer) bool {
                health := stats.GetStat(StatHealth)
                maxHealth := stats.GetStat(StatMaxHealth)
                return health < maxHealth * 0.5
            },
        },
    }
}
```

#### Extension 4: Add Gear System
```go
// Equipment modifies stats - no change to stat system!
type Equipment struct {
    Name      string
    Slot      EquipmentSlot
    Modifiers []StatModifier
}

type Sword struct {
    Equipment
}

func NewIronSword() *Sword {
    return &Sword{
        Equipment: Equipment{
            Name: "Iron Sword",
            Slot: SlotMainHand,
            Modifiers: []StatModifier{
                &FlatModifier{Value: 10},  // +10 attack
            },
        },
    }
}

// When equipped:
player.Stats.AddModifiers(StatAttackPower, sword.Modifiers)

// When unequipped:
player.Stats.RemoveModifiers(sword.ModifierIDs)
```

#### Extension 5: Add Buff System
```go
// Buffs are just temporary modifiers!
type Buff struct {
    ID        string
    Name      string
    Modifiers map[StatType][]StatModifier
    Duration  time.Duration
    StartTime time.Time
}

func NewStrengthPotion() *Buff {
    return &Buff{
        ID:   "strength_potion_123",
        Name: "Strength Potion",
        Modifiers: map[StatType][]StatModifier{
            StatAttackPower: {
                &PercentModifier{Percent: 0.5, Duration: 30*time.Second},
            },
        },
        Duration: 30 * time.Second,
    }
}

// Apply buff:
for statType, mods := range buff.Modifiers {
    for _, mod := range mods {
        player.Stats.AddModifier(statType, mod)
    }
}

// Buff expires automatically after duration
```

---

## Calculation Flow

### Without Modifiers:
```
GetStat(StatAttackPower)
  ↓
Return BaseValue (100)
```

### With Modifiers:
```
GetStat(StatAttackPower)
  ↓
Start with BaseValue: 100
  ↓
Apply FlatModifier (+10): 110
  ↓
Apply PercentModifier (+50%): 165
  ↓
Apply ConditionalModifier (if health < 50%): 198
  ↓
Return CurrentValue: 198
```

### Modifier Application Order:
1. **Flat modifiers** first (addition)
2. **Percent modifiers** second (multiplication)
3. **Multiplier modifiers** last (final scaling)
4. **Conditional modifiers** can apply at any stage

---

## Usage Examples

### Example 1: Create a Warrior
```go
// Create warrior with base stats
warrior := NewPlayer("player_123")
archetype := &WarriorArchetype{}

// Initialize stats from archetype
for statType, value := range archetype.GetBaseStats() {
    warrior.Stats.SetBaseStat(statType, value)
}

// Apply passive modifiers
for _, mod := range archetype.GetPassiveModifiers() {
    warrior.Stats.AddPassiveModifier(mod)
}

// Result:
// MaxHealth: 150
// AttackPower: 20 (or 24 if health < 50% due to Berserker Rage)
```

### Example 2: Equip Gear
```go
sword := NewIronSword()
warrior.EquipItem(SlotMainHand, sword)

// Internally:
warrior.Stats.AddModifier(StatAttackPower, &FlatModifier{Value: 10})

// Result:
// AttackPower: 30 (20 base + 10 from sword)
```

### Example 3: Drink Potion
```go
potion := NewStrengthPotion()
warrior.ApplyBuff(potion)

// Result (for 30 seconds):
// AttackPower: 45 (30 base+gear * 1.5 from potion)

// After 30 seconds:
// AttackPower: 30 (buff expires automatically)
```

### Example 4: Take Damage
```go
warrior.TakeDamage(80)

// Health: 150 -> 70 (below 50% threshold)
// Berserker Rage activates!
// AttackPower: 45 * 1.20 = 54

// Warrior is now stronger when low on health!
```

### Example 5: Level Up
```go
warrior.LevelUp()

// Growth rates from archetype:
for statType, growth := range archetype.GetStatGrowth() {
    currentBase := warrior.Stats.GetBaseStat(statType)
    warrior.Stats.SetBaseStat(statType, currentBase + growth)
}

// Result:
// MaxHealth: 150 -> 160 (+10 per level)
// AttackPower: 20 -> 22 (+2 per level)
```

---

## Adding New Features (Without Modifying Core)

### Add New Stat: Evasion
```go
// 1. Add constant
const StatEvasion StatType = "evasion"

// 2. Use it
player.Stats.SetBaseStat(StatEvasion, 10.0)

// 3. Add to archetype
func (r *RogueArchetype) GetBaseStats() map[StatType]float64 {
    return map[StatType]float64{
        StatEvasion: 25,  // Rogues have high evasion!
    }
}

// That's it! No modification to stat system core
```

### Add New Modifier Type: Diminishing Returns
```go
// New modifier type for diminishing returns
type DiminishingModifier struct {
    Value      float64
    Threshold  float64
    Diminish   float64
}

func (dm *DiminishingModifier) Apply(baseValue float64) float64 {
    if baseValue > dm.Threshold {
        excess := baseValue - dm.Threshold
        diminished := excess * dm.Diminish
        return dm.Threshold + diminished + dm.Value
    }
    return baseValue + dm.Value
}

// Example: Defense has diminishing returns after 100
// 100 defense = 50% damage reduction
// 200 defense = 60% damage reduction (not 100%)
```

### Add New Character: Mage
```go
type MageArchetype struct{}

func (m *MageArchetype) GetBaseStats() map[StatType]float64 {
    return map[StatType]float64{
        StatMaxHealth:   80,   // Low HP
        StatMaxMana:     200,  // High mana
        StatAttackPower: 5,    // Low physical
        StatDefense:     5,    // Low defense
    }
}

func (m *MageArchetype) GetPassiveModifiers() []StatModifier {
    return []StatModifier{
        // "Arcane Mastery": +30% mana regen when above 80% mana
        &ConditionalModifier{
            BaseModifier: &PercentModifier{Percent: 0.30},
            Condition: func(stats *StatContainer) bool {
                mana := stats.GetStat(StatMana)
                maxMana := stats.GetStat(StatMaxMana)
                return mana > maxMana * 0.8
            },
        },
    }
}

// No changes to stat system! Just implement the interface
```

---

## Benefits of This Design

### ✅ Open for Extension
- Add new stats: Just add constant
- Add new modifiers: Implement `StatModifier` interface
- Add new classes: Implement `CharacterArchetype` interface
- Add new features: Composition over modification

### ✅ Closed for Modification
- Core stat system never changes
- BaseStat implementation stable
- StatContainer logic untouched
- Safe to extend without breaking existing features

### ✅ Testable
```go
func TestWarriorBerserkerRage(t *testing.T) {
    warrior := NewWarriorPlayer()
    warrior.Stats.SetBaseStat(StatHealth, 40)  // Below 50%
    warrior.Stats.SetBaseStat(StatMaxHealth, 100)

    attack := warrior.Stats.GetStat(StatAttackPower)
    expected := 20 * 1.20  // Base * Berserker Rage
    assert.Equal(t, expected, attack)
}
```

### ✅ Maintainable
- Each archetype is independent
- Modifiers are composable
- Easy to understand and debug
- Clear separation of concerns

### ✅ Performant
- Cached calculated values
- Only recalculate when modifiers change
- Efficient modifier lookup
- Thread-safe with mutex

---

## File Structure

```
server/
├── stats/
│   ├── stat.go              # Stat interface & BaseStat
│   ├── stat_container.go    # StatContainer with modifiers
│   ├── modifier.go          # Modifier interface & types
│   ├── archetype.go         # Character archetype interface
│   ├── warrior.go           # Warrior implementation
│   ├── mage.go              # Mage implementation
│   ├── archer.go            # Archer implementation (future)
│   └── equipment.go         # Equipment modifiers (future)
├── player/
│   └── player.go            # Player has StatContainer
└── combat/
    └── damage_calculator.go  # Uses stats for combat
```

---

## Summary

### Open-Closed Principle Implementation

**Closed (Never Modified):**
- `Stat` interface
- `BaseStat` implementation
- `StatContainer` core logic
- `StatModifier` interface

**Open (Easy to Extend):**
- New stat types (add constant)
- New modifiers (implement interface)
- New character classes (implement interface)
- New features (composition)

**How to Action This:**

1. **Adding a stat**: Add constant, use in archetype
2. **Adding a class**: Implement `CharacterArchetype`
3. **Adding a buff**: Create modifier, apply to stats
4. **Adding gear**: Create equipment with modifiers

**No modification to core system required!**

---

## Next: Implementation

Ready to implement this design in code?

Files to create:
1. `server/stats/stat.go` - Core interfaces
2. `server/stats/stat_container.go` - Container with modifiers
3. `server/stats/modifier.go` - Modifier types
4. `server/stats/archetype.go` - Archetype interface
5. `server/stats/warrior.go` - Example archetype
6. Update `server/player/player.go` - Integrate stats

Shall I implement this now?
