package combat

import (
	"errors"
	"fmt"
)

// AbilityValidator validates whether an ability can be used
type AbilityValidator interface {
	Validate(caster Player, target Player, ability Ability) error
}

// ValidatorChain runs multiple validators in sequence
type ValidatorChain struct {
	validators []AbilityValidator
}

// NewValidatorChain creates a new validator chain
func NewValidatorChain(validators ...AbilityValidator) *ValidatorChain {
	return &ValidatorChain{
		validators: validators,
	}
}

// Validate runs all validators in the chain
func (vc *ValidatorChain) Validate(caster Player, target Player, ability Ability) error {
	for _, validator := range vc.validators {
		if err := validator.Validate(caster, target, ability); err != nil {
			return err
		}
	}
	return nil
}

// AliveValidator checks if both caster and target are alive
type AliveValidator struct{}

func (av *AliveValidator) Validate(caster Player, target Player, ability Ability) error {
	if !caster.IsAlive() {
		return errors.New("caster_is_dead")
	}
	if !target.IsAlive() {
		return errors.New("target_is_dead")
	}
	return nil
}

// CooldownValidator checks if the ability is off cooldown
type CooldownValidator struct{}

func (cv *CooldownValidator) Validate(caster Player, target Player, ability Ability) error {
	cooldownMgr := caster.GetCooldownManager()
	if cooldownMgr == nil {
		return errors.New("no_cooldown_manager")
	}

	if !cooldownMgr.IsReady(ability.GetID()) {
		remaining := cooldownMgr.GetRemaining(ability.GetID())
		return fmt.Errorf("ability_on_cooldown: %.1fs remaining", remaining.Seconds())
	}
	return nil
}

// RangeValidator checks if the target is within range
type RangeValidator struct{}

func (rv *RangeValidator) Validate(caster Player, target Player, ability Ability) error {
	casterPos := caster.GetPosition()
	targetPos := target.GetPosition()
	distance := casterPos.Distance(targetPos)

	abilityRange := ability.GetRange()
	if distance > abilityRange {
		return fmt.Errorf("target_out_of_range: %.1fm (max: %.1fm)", distance, abilityRange)
	}
	return nil
}

// ResourceValidator checks if the caster has enough resources
type ResourceValidator struct{}

func (rsv *ResourceValidator) Validate(caster Player, target Player, ability Ability) error {
	stats := caster.GetStats()
	if stats == nil {
		return errors.New("no_player_stats")
	}

	cost := ability.GetCost()

	if cost.Mana > 0 && stats.Mana < cost.Mana {
		return fmt.Errorf("not_enough_mana: need %d, have %d", cost.Mana, stats.Mana)
	}

	if cost.Stamina > 0 && stats.Stamina < cost.Stamina {
		return fmt.Errorf("not_enough_stamina: need %d, have %d", cost.Stamina, stats.Stamina)
	}

	return nil
}

// SelfTargetValidator prevents targeting yourself (optional - some abilities allow this)
type SelfTargetValidator struct {
	AllowSelfTarget bool
}

func (stv *SelfTargetValidator) Validate(caster Player, target Player, ability Ability) error {
	if !stv.AllowSelfTarget && caster.GetID() == target.GetID() {
		return errors.New("cannot_target_self")
	}
	return nil
}

// CreateDefaultValidatorChain creates a validator chain with common validators
func CreateDefaultValidatorChain() *ValidatorChain {
	return NewValidatorChain(
		&AliveValidator{},
		&CooldownValidator{},
		&RangeValidator{},
		&ResourceValidator{},
		&SelfTargetValidator{AllowSelfTarget: false},
	)
}
