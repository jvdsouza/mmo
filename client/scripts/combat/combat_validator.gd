extends Node
class_name CombatValidator

## Validates combat actions (attacks, abilities, etc.)
##
## Centralized validation logic for:
## - Range checking
## - Line of sight
## - State validation (can't attack while stunned, etc.)
## - Resource costs (mana, stamina)

enum ValidationResult {
	SUCCESS,
	NO_TARGET,
	TARGET_INVALID,
	OUT_OF_RANGE,
	NO_LINE_OF_SIGHT,
	ON_COOLDOWN,
	INSUFFICIENT_RESOURCES,
	INVALID_STATE,
	FRIENDLY_FIRE
}

## Validate an attack attempt
static func validate_attack(
	attacker: Node3D,
	target: Node,
	weapon: Weapon,
	targeting_system: TargetingSystem
) -> ValidationResult:

	# Check target exists
	if not target:
		return ValidationResult.NO_TARGET

	# Check target has ITargetable interface (duck typing)
	if not target.has_method("is_valid_target"):
		return ValidationResult.TARGET_INVALID

	if not target.has_method("is_alive"):
		return ValidationResult.TARGET_INVALID

	if not target.has_method("get_target_position"):
		return ValidationResult.TARGET_INVALID

	# Check target is valid
	if not target.is_valid_target():
		return ValidationResult.TARGET_INVALID

	# Check target is alive
	if not target.is_alive():
		return ValidationResult.TARGET_INVALID

	# Check weapon is ready (not on cooldown)
	if not weapon.is_ready():
		return ValidationResult.ON_COOLDOWN

	# Check range
	var distance = attacker.global_position.distance_to(target.get_target_position())
	if distance > weapon.attack_range:
		return ValidationResult.OUT_OF_RANGE

	# Future: Check line of sight
	# Future: Check friendly fire
	# Future: Check attacker state (stunned, silenced, etc.)
	# Future: Check resource costs

	return ValidationResult.SUCCESS

## Get human-readable message for validation result
static func get_validation_message(result: ValidationResult) -> String:
	match result:
		ValidationResult.SUCCESS:
			return "Attack validated"
		ValidationResult.NO_TARGET:
			return "No target selected"
		ValidationResult.TARGET_INVALID:
			return "Invalid target"
		ValidationResult.OUT_OF_RANGE:
			return "Target out of range"
		ValidationResult.NO_LINE_OF_SIGHT:
			return "No line of sight"
		ValidationResult.ON_COOLDOWN:
			return "Ability on cooldown"
		ValidationResult.INSUFFICIENT_RESOURCES:
			return "Not enough resources"
		ValidationResult.INVALID_STATE:
			return "Cannot attack in current state"
		ValidationResult.FRIENDLY_FIRE:
			return "Cannot attack allies"
		_:
			return "Unknown validation error"

## Check if result is successful
static func is_success(result: ValidationResult) -> bool:
	return result == ValidationResult.SUCCESS
