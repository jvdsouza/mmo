# Stats System Usage Guide

## Quick Reference

This guide shows you how to USE the OCP-compliant stats system in practice.

---

## Table of Contents

1. [Creating a Player with Stats](#creating-a-player-with-stats)
2. [Using Character Archetypes](#using-character-archetypes)
3. [Adding Equipment](#adding-equipment)
4. [Applying Buffs/Debuffs](#applying-buffsdebuffs)
5. [Leveling Up](#leveling-up)
6. [Extending with New Content](#extending-with-new-content)

---

## Creating a Player with Stats

### Basic Player (No Class)

```go
import (
    "mmo-server/player"
    "mmo-server/stats"
)

// Create basic player
p := player.NewPlayer("player_123", "Alice")

// Manually set stats
p.Stats.SetBaseStat(stats.StatMaxHealth, 100)
p.Stats.SetBaseStat(stats.StatHealth, 100)
p.Stats.SetBaseStat(stats.StatAttackPower, 10)
p.Stats.SetBaseStat(stats.StatDefense, 5)

// Get current stats
health := p.Stats.GetStat(stats.StatHealth)
attack := p.Stats.GetStat(stats.StatAttackPower)

fmt.Printf("Health: %.0f, Attack: %.0f\n", health, attack)
```

### Player with Class Archetype

```go
// Create a Warrior
warriorArch := stats.NewWarriorArchetype(1)
warrior := player.NewPlayerWithArchetype("player_123", "Conan", warriorArch)

// Stats are automatically set from archetype!
// Warrior gets:
// - 150 HP
// - 20 Attack
// - 15 Defense
// - Berserker Rage passive

fmt.Printf("%s the %s\n", warrior.Name, warrior.Archetype.GetName())
fmt.Printf("HP: %.0f\n", warrior.Stats.GetStat(stats.StatMaxHealth))
fmt.Printf("Attack: %.0f\n", warrior.Stats.GetStat(stats.StatAttackPower))
```

---

## Using Character Archetypes

### Create Different Classes

```go
// Warrior - High HP, High Attack
warrior := player.NewPlayerWithArchetype(
    "player_1",
    "Conan",
    stats.NewWarriorArchetype(1),
)

// Mage - Low HP, High Mana
mage := player.NewPlayerWithArchetype(
    "player_2",
    "Gandalf",
    stats.NewMageArchetype(1),
)

// Archer - Balanced, High Crit
archer := player.NewPlayerWithArchetype(
    "player_3",
    "Legolas",
    stats.NewArcherArchetype(1),
)

// Compare stats
fmt.Printf("Warrior HP: %.0f\n", warrior.Stats.GetStat(stats.StatMaxHealth))  // 150
fmt.Printf("Mage HP: %.0f\n", mage.Stats.GetStat(stats.StatMaxHealth))        // 80
fmt.Printf("Archer HP: %.0f\n", archer.Stats.GetStat(stats.StatMaxHealth))    // 100
```

### Passive Abilities Work Automatically

```go
warrior := player.NewPlayerWithArchetype("w1", "Warrior", stats.NewWarriorArchetype(1))

// Full health attack
attack1 := warrior.Stats.GetStat(stats.StatAttackPower)
fmt.Printf("Attack at full health: %.0f\n", attack1)  // 20

// Take damage below 50%
warrior.TakeDamage(80)  // Health: 150 -> 70 (below 50%)

// Berserker Rage activates!
attack2 := warrior.Stats.GetStat(stats.StatAttackPower)
fmt.Printf("Attack at low health: %.0f\n", attack2)   // 24 (20 * 1.20)
```

---

## Adding Equipment

### Create Equipment

```go
// Define a weapon
sword := &stats.Equipment{
    ID:   "iron_sword_1",
    Name: "Iron Sword",
    Slot: stats.SlotMainHand,
    Modifiers: map[stats.StatType][]stats.StatModifier{
        stats.StatAttackPower: {
            stats.NewFlatModifier("sword_atk", 10, 0),  // +10 attack
        },
    },
}

// Define armor
helmet := &stats.Equipment{
    ID:   "steel_helmet_1",
    Name: "Steel Helmet",
    Slot: stats.SlotHead,
    Modifiers: map[stats.StatType][]stats.StatModifier{
        stats.StatDefense: {
            stats.NewFlatModifier("helmet_def", 5, 0),  // +5 defense
        },
        stats.StatMaxHealth: {
            stats.NewFlatModifier("helmet_hp", 20, 0),  // +20 max HP
        },
    },
}
```

### Equip Items

```go
func EquipItem(p *player.Player, equipment *stats.Equipment) {
    // Apply all modifiers from the equipment
    for statType, modifiers := range equipment.Modifiers {
        for _, mod := range modifiers {
            p.Stats.AddModifier(statType, mod)
        }
    }
}

// Use it
warrior := player.NewPlayerWithArchetype("w1", "Warrior", stats.NewWarriorArchetype(1))

fmt.Printf("Base attack: %.0f\n", warrior.Stats.GetStat(stats.StatAttackPower))  // 20

EquipItem(warrior, sword)

fmt.Printf("With sword: %.0f\n", warrior.Stats.GetStat(stats.StatAttackPower))   // 30
```

### Unequip Items

```go
func UnequipItem(p *player.Player, equipment *stats.Equipment) {
    // Remove all modifiers from the equipment
    for _, modifiers := range equipment.Modifiers {
        for _, mod := range modifiers {
            p.Stats.RemoveModifier(mod.GetID())
        }
    }
}

// Use it
UnequipItem(warrior, sword)

fmt.Printf("After unequip: %.0f\n", warrior.Stats.GetStat(stats.StatAttackPower))  // 20
```

---

## Applying Buffs/Debuffs

### Create a Buff

```go
// Strength Potion: +50% attack for 30 seconds
strengthBuff := &stats.Buff{
    ID:   "strength_potion_1",
    Name: "Strength Potion",
    Modifiers: map[stats.StatType][]stats.StatModifier{
        stats.StatAttackPower: {
            stats.NewPercentModifier("str_pot", 0.5, 30*time.Second),
        },
    },
    Duration: 30 * time.Second,
}
```

### Apply Buff

```go
func ApplyBuff(p *player.Player, buff *stats.Buff) {
    for statType, modifiers := range buff.Modifiers {
        for _, mod := range modifiers {
            p.Stats.AddModifier(statType, mod)
        }
    }
}

// Use it
warrior := player.NewPlayerWithArchetype("w1", "Warrior", stats.NewWarriorArchetype(1))

fmt.Printf("Base attack: %.0f\n", warrior.Stats.GetStat(stats.StatAttackPower))  // 20

ApplyBuff(warrior, strengthBuff)

fmt.Printf("With buff: %.0f\n", warrior.Stats.GetStat(stats.StatAttackPower))    // 30 (20 * 1.5)

// Wait 30 seconds or manually remove expired modifiers
time.Sleep(31 * time.Second)
warrior.Stats.RemoveExpiredModifiers()

fmt.Printf("After expiry: %.0f\n", warrior.Stats.GetStat(stats.StatAttackPower)) // 20
```

### Create a Debuff

```go
// Poison: -10 max HP for 10 seconds
poison := &stats.Buff{
    ID:   "poison_1",
    Name: "Poison",
    Modifiers: map[stats.StatType][]stats.StatModifier{
        stats.StatMaxHealth: {
            stats.NewFlatModifier("poison_hp", -10, 10*time.Second),
        },
    },
    Duration: 10 * time.Second,
}

ApplyBuff(warrior, poison)  // Max HP: 150 -> 140 for 10 seconds
```

---

## Leveling Up

### Basic Level Up

```go
warrior := player.NewPlayerWithArchetype("w1", "Warrior", stats.NewWarriorArchetype(1))

fmt.Printf("Level %d\n", warrior.Level)
fmt.Printf("HP: %.0f\n", warrior.Stats.GetStat(stats.StatMaxHealth))       // 150
fmt.Printf("Attack: %.0f\n", warrior.Stats.GetStat(stats.StatAttackPower)) // 20

// Level up!
warrior.LevelUp()

fmt.Printf("Level %d\n", warrior.Level)
fmt.Printf("HP: %.0f\n", warrior.Stats.GetStat(stats.StatMaxHealth))       // 160 (+10)
fmt.Printf("Attack: %.0f\n", warrior.Stats.GetStat(stats.StatAttackPower)) // 22 (+2)
```

### Multiple Levels

```go
// Level up 10 times
for i := 0; i < 10; i++ {
    warrior.LevelUp()
}

// Level 11 warrior stats:
// HP: 150 + (10 * 10) = 250
// Attack: 20 + (10 * 2) = 40
```

---

## Extending with New Content

### Add a New Stat

```go
// 1. Add constant (in stats/stat.go or stats/examples.go)
const StatSpeed stats.StatType = "speed"

// 2. Use immediately!
player.Stats.SetBaseStat(StatSpeed, 1.0)

// 3. Add to archetype
type RogueArchetype struct{ level int }

func (r *RogueArchetype) GetBaseStats() map[stats.StatType]float64 {
    return map[stats.StatType]float64{
        // ... other stats
        StatSpeed: 1.3,  // Rogues are fast!
    }
}

// That's it! No modification to core stats system
```

### Add a New Character Class

```go
// 1. Implement CharacterArchetype interface
type PaladinArchetype struct {
    level int
}

func (p *PaladinArchetype) GetName() string {
    return "Paladin"
}

func (p *PaladinArchetype) GetBaseStats() map[stats.StatType]float64 {
    return map[stats.StatType]float64{
        stats.StatMaxHealth:    140,  // High HP
        stats.StatHealth:       140,
        stats.StatMaxMana:      120,  // Moderate mana
        stats.StatMana:         120,
        stats.StatAttackPower:  18,   // Good attack
        stats.StatDefense:      20,   // High defense
        stats.StatCritChance:   0.08,
        stats.StatCritMultiplier: 1.6,
    }
}

func (p *PaladinArchetype) GetStatGrowth() map[stats.StatType]float64 {
    return map[stats.StatType]float64{
        stats.StatMaxHealth:   8,
        stats.StatMaxMana:     6,
        stats.StatAttackPower: 2,
        stats.StatDefense:     2,
    }
}

func (p *PaladinArchetype) GetPassiveModifiers() []stats.StatModifier {
    // "Divine Shield": +30% defense when health > 80%
    divineShield := stats.NewConditionalModifier(
        "divine_shield",
        stats.NewPercentModifier("shield_effect", 0.30, 0),
        func(stats *stats.StatContainer) bool {
            health := stats.GetStat(stats.StatHealth)
            maxHealth := stats.GetStat(stats.StatMaxHealth)
            return health > maxHealth*0.8
        },
    )

    return []stats.StatModifier{divineShield}
}

// 2. Create apply function
func ApplyPaladinPassives(container *stats.StatContainer, arch *PaladinArchetype) {
    for _, mod := range arch.GetPassiveModifiers() {
        container.AddModifier(stats.StatDefense, mod)
    }
}

// 3. Update NewPlayerWithArchetype to handle paladin
// (or just create a custom function for paladin)

// 4. Use it!
paladin := player.NewPlayerWithArchetype(
    "p1",
    "Uther",
    &PaladinArchetype{level: 1},
)

// Apply passives
if paladinArch, ok := paladin.Archetype.(*PaladinArchetype); ok {
    ApplyPaladinPassives(paladin.Stats, paladinArch)
}
```

### Add a New Modifier Type

```go
// Example: Stacking modifier that gets stronger each application

type StackingModifier struct {
    ID          string
    ValuePerStack float64
    CurrentStacks int
    MaxStacks     int
    Duration      time.Duration
    StartTime     time.Time
}

func (sm *StackingModifier) GetID() string { return sm.ID }
func (sm *StackingModifier) GetType() stats.ModifierType { return stats.ModifierTypeFlat }
func (sm *StackingModifier) GetValue() float64 { return sm.ValuePerStack * float64(sm.CurrentStacks) }
func (sm *StackingModifier) GetDuration() time.Duration { return sm.Duration }
func (sm *StackingModifier) IsExpired() bool {
    if sm.Duration == 0 { return false }
    return time.Since(sm.StartTime) > sm.Duration
}
func (sm *StackingModifier) Apply(baseValue float64) float64 {
    return baseValue + sm.GetValue()
}

func (sm *StackingModifier) AddStack() {
    if sm.CurrentStacks < sm.MaxStacks {
        sm.CurrentStacks++
        sm.StartTime = time.Now()  // Refresh duration
    }
}

// Usage: "Fury" buff - +2 attack per stack, max 5 stacks
fury := &StackingModifier{
    ID:           "fury_stacks",
    ValuePerStack: 2,
    CurrentStacks: 0,
    MaxStacks:    5,
    Duration:     10 * time.Second,
}

// Each hit adds a stack
player.Stats.AddModifier(stats.StatAttackPower, fury)
fury.AddStack()  // +2 attack
fury.AddStack()  // +4 attack
fury.AddStack()  // +6 attack
// ... up to +10 attack at 5 stacks
```

---

## Complete Example: Creating a Fully Equipped Player

```go
package main

import (
    "fmt"
    "time"
    "mmo-server/player"
    "mmo-server/stats"
)

func main() {
    // 1. Create warrior
    warrior := player.NewPlayerWithArchetype(
        "warrior_1",
        "Conan",
        stats.NewWarriorArchetype(1),
    )

    fmt.Printf("=== %s the %s ===\n", warrior.Name, warrior.Archetype.GetName())
    PrintStats(warrior)

    // 2. Equip gear
    sword := stats.NewIronSword()
    helmet := stats.NewSteelHelmet()

    EquipItem(warrior, sword)
    EquipItem(warrior, helmet)

    fmt.Println("\n=== After Equipping Gear ===")
    PrintStats(warrior)

    // 3. Drink potion
    potion := stats.NewStrengthPotion()
    ApplyBuff(warrior, potion)

    fmt.Println("\n=== After Strength Potion ===")
    PrintStats(warrior)

    // 4. Take damage (activate Berserker Rage)
    warrior.TakeDamage(100)

    fmt.Println("\n=== After Taking Damage (Berserker Rage!) ===")
    PrintStats(warrior)

    // 5. Level up
    warrior.LevelUp()

    fmt.Println("\n=== After Level Up ===")
    fmt.Printf("Level: %d\n", warrior.Level)
    PrintStats(warrior)
}

func PrintStats(p *player.Player) {
    fmt.Printf("HP: %.0f / %.0f\n",
        p.Stats.GetStat(stats.StatHealth),
        p.Stats.GetStat(stats.StatMaxHealth))
    fmt.Printf("Attack: %.0f\n", p.Stats.GetStat(stats.StatAttackPower))
    fmt.Printf("Defense: %.0f\n", p.Stats.GetStat(stats.StatDefense))
    fmt.Printf("Crit: %.1f%%\n", p.Stats.GetStat(stats.StatCritChance)*100)
}

func EquipItem(p *player.Player, equipment *stats.Equipment) {
    for statType, modifiers := range equipment.Modifiers {
        for _, mod := range modifiers {
            p.Stats.AddModifier(statType, mod)
        }
    }
}

func ApplyBuff(p *player.Player, buff *stats.Buff) {
    for statType, modifiers := range buff.Modifiers {
        for _, mod := range modifiers {
            p.Stats.AddModifier(statType, mod)
        }
    }
}

// Output:
// === Conan the Warrior ===
// HP: 150 / 150
// Attack: 20
// Defense: 15
// Crit: 10.0%
//
// === After Equipping Gear ===
// HP: 150 / 170     (+20 from helmet)
// Attack: 30        (+10 from sword)
// Defense: 20       (+5 from helmet)
// Crit: 10.0%
//
// === After Strength Potion ===
// HP: 150 / 170
// Attack: 45        (+50% from potion: 30 * 1.5)
// Defense: 20
// Crit: 10.0%
//
// === After Taking Damage (Berserker Rage!) ===
// HP: 50 / 170      (below 50%, Berserker active!)
// Attack: 54        (+20% from Berserker: 45 * 1.2)
// Defense: 20
// Crit: 10.0%
//
// === After Level Up ===
// Level: 2
// HP: 180 / 180     (+10 base HP, healed to full)
// Attack: 56        (Base 22, +10 sword, +50% potion, +20% Berserker)
// Defense: 21       (+1 from level, +5 from helmet)
// Crit: 10.0%
```

---

## Best Practices

### 1. Always Use Archetypes for Character Classes
```go
// ✅ Good: Use archetype
warrior := player.NewPlayerWithArchetype("id", "name", stats.NewWarriorArchetype(1))

// ❌ Bad: Manually set every stat
warrior := player.NewPlayer("id", "name")
warrior.Stats.SetBaseStat(stats.StatMaxHealth, 150)
warrior.Stats.SetBaseStat(stats.StatAttackPower, 20)
// ... tedious and error-prone
```

### 2. Use Modifiers for Temporary Changes
```go
// ✅ Good: Use modifiers (automatically expire)
buff := stats.NewPercentModifier("buff_id", 0.5, 30*time.Second)
player.Stats.AddModifier(stats.StatAttackPower, buff)

// ❌ Bad: Directly modify base stats
oldAttack := player.Stats.GetBaseStat(stats.StatAttackPower)
player.Stats.SetBaseStat(stats.StatAttackPower, oldAttack * 1.5)
// How do you undo this? Hard to track!
```

### 3. Clean Up Expired Modifiers
```go
// In game loop or update function
func (p *Player) Update(deltaTime float64) {
    p.Stats.RemoveExpiredModifiers()
}
```

### 4. Extend, Don't Modify
```go
// ✅ Good: Add new archetype
type NecromancerArchetype struct { /* ... */ }

// ❌ Bad: Modify existing archetype
// Don't change WarriorArchetype's stats - create a new class instead!
```

---

## Summary

The OCP stats system allows you to:

✅ **Create players** with different classes easily
✅ **Equip gear** that modifies stats
✅ **Apply buffs/debuffs** with automatic expiration
✅ **Level up** with automatic stat growth
✅ **Extend** with new stats, classes, and modifiers

All without modifying the core stats system!

See `docs/STATS_OCP_DESIGN.md` for architectural details.
See `server/stats/examples.go` for more extension examples.
