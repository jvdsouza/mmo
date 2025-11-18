package player

import (
	"mmo-server/combat"
	"mmo-server/stats"
)

// Player represents a player in the game
type Player struct {
	ID       string
	Name     string
	Position combat.Vector3
	Rotation combat.Vector3

	// Stats system - OCP compliant!
	Stats     *stats.StatContainer
	Archetype stats.CharacterArchetype
	Level     int

	// Combat
	CooldownManager *combat.CooldownManager
}

// NewPlayer creates a new player with default stats
func NewPlayer(id string, name string) *Player {
	return &Player{
		ID:              id,
		Name:            name,
		Position:        combat.Vector3{X: 0, Y: 1, Z: 0},
		Rotation:        combat.Vector3{X: 0, Y: 0, Z: 0},
		Stats:           stats.NewStatContainer(),
		CooldownManager: combat.NewCooldownManager(),
		Level:           1,
	}
}

// NewPlayerWithArchetype creates a player with a specific character class
func NewPlayerWithArchetype(id string, name string, archetype stats.CharacterArchetype) *Player {
	player := NewPlayer(id, name)
	player.Archetype = archetype

	// Apply archetype's base stats
	stats.ApplyArchetype(player.Stats, archetype)

	// Apply archetype-specific passives
	switch arch := archetype.(type) {
	case *stats.WarriorArchetype:
		stats.ApplyWarriorPassives(player.Stats, arch)
	case *stats.MageArchetype:
		stats.ApplyMagePassives(player.Stats, arch)
	}

	return player
}

// Implement combat.Player interface

func (p *Player) GetID() string {
	return p.ID
}

func (p *Player) GetPosition() combat.Vector3 {
	return p.Position
}

func (p *Player) GetStats() *combat.PlayerStats {
	// Convert stats.StatContainer to combat.PlayerStats
	return &combat.PlayerStats{
		Health:         int(p.Stats.GetStat(stats.StatHealth)),
		MaxHealth:      int(p.Stats.GetStat(stats.StatMaxHealth)),
		Mana:           int(p.Stats.GetStat(stats.StatMana)),
		MaxMana:        int(p.Stats.GetStat(stats.StatMaxMana)),
		Stamina:        int(p.Stats.GetStat(stats.StatStamina)),
		MaxStamina:     int(p.Stats.GetStat(stats.StatMaxStamina)),
		AttackPower:    int(p.Stats.GetStat(stats.StatAttackPower)),
		Defense:        int(p.Stats.GetStat(stats.StatDefense)),
		CritChance:     p.Stats.GetStat(stats.StatCritChance),
		CritMultiplier: p.Stats.GetStat(stats.StatCritMultiplier),
		IsDead:         p.Stats.GetStat(stats.StatHealth) <= 0,
	}
}

func (p *Player) GetCooldownManager() *combat.CooldownManager {
	return p.CooldownManager
}

func (p *Player) TakeDamage(amount int) {
	p.Stats.ModifyCurrentValue(stats.StatHealth, float64(-amount))

	// Clamp health to 0
	if p.Stats.GetStat(stats.StatHealth) < 0 {
		p.Stats.SetBaseStat(stats.StatHealth, 0)
	}

	// Recalculate stats (for conditional modifiers like Berserker Rage)
	p.Stats.RecalculateAll()
}

func (p *Player) IsAlive() bool {
	return p.Stats.GetStat(stats.StatHealth) > 0
}

// Heal restores health
func (p *Player) Heal(amount int) {
	p.Stats.ModifyCurrentValue(stats.StatHealth, float64(amount))

	// Clamp to max health
	maxHealth := p.Stats.GetStat(stats.StatMaxHealth)
	currentHealth := p.Stats.GetStat(stats.StatHealth)
	if currentHealth > maxHealth {
		p.Stats.ModifyCurrentValue(stats.StatHealth, maxHealth-currentHealth)
	}

	// Recalculate stats
	p.Stats.RecalculateAll()
}

// RestoreMana restores mana
func (p *Player) RestoreMana(amount int) {
	p.Stats.ModifyCurrentValue(stats.StatMana, float64(amount))

	// Clamp to max mana
	maxMana := p.Stats.GetStat(stats.StatMaxMana)
	currentMana := p.Stats.GetStat(stats.StatMana)
	if currentMana > maxMana {
		p.Stats.ModifyCurrentValue(stats.StatMana, maxMana-currentMana)
	}
}

// ConsumeMana uses mana for abilities
func (p *Player) ConsumeMana(amount int) bool {
	currentMana := p.Stats.GetStat(stats.StatMana)
	if currentMana >= float64(amount) {
		p.Stats.ModifyCurrentValue(stats.StatMana, float64(-amount))
		return true
	}
	return false
}

// ConsumeStamina uses stamina for abilities
func (p *Player) ConsumeStamina(amount int) bool {
	currentStamina := p.Stats.GetStat(stats.StatStamina)
	if currentStamina >= float64(amount) {
		p.Stats.ModifyCurrentValue(stats.StatStamina, float64(-amount))
		return true
	}
	return false
}

// LevelUp increases the player's level and applies growth stats
func (p *Player) LevelUp() {
	p.Level++

	if p.Archetype != nil {
		stats.LevelUpArchetype(p.Stats, p.Archetype)

		// Restore resources on level up
		p.Stats.RestoreStat(stats.StatHealth)
		p.Stats.RestoreStat(stats.StatMana)
		p.Stats.RestoreStat(stats.StatStamina)
	}
}

// AddBuff applies a temporary stat modifier
func (p *Player) AddBuff(statType stats.StatType, modifier stats.StatModifier) {
	p.Stats.AddModifier(statType, modifier)
}

// RemoveBuff removes a buff by ID
func (p *Player) RemoveBuff(modifierID string) bool {
	return p.Stats.RemoveModifier(modifierID)
}

// Update processes per-frame updates (expired buffs, regeneration, etc.)
func (p *Player) Update(deltaTime float64) {
	// Remove expired modifiers
	p.Stats.RemoveExpiredModifiers()

	// TODO: Add resource regeneration
	// TODO: Add status effect ticking (poison, etc.)
}
