extends Node
class_name Weapon

## Base class for all weapons
##
## Weapons define attack properties: damage, range, cooldown, type, etc.
## Different weapon types (melee, ranged, magic) extend this.

signal attack_started(target: Node)
signal attack_completed(target: Node)
signal cooldown_started(duration: float)
signal cooldown_finished()

## Weapon configuration
@export var weapon_name: String = "Basic Weapon"
@export var attack_range: float = 3.0
@export var cooldown_time: float = 1.0
@export var damage: int = 25

## Attack type for server/animation hints
@export_enum("melee", "ranged", "magic") var attack_type: String = "melee"

## Animation hint for this weapon
@export var animation_hint: String = "swing_right"

## Ability ID for server communication
@export var ability_id: String = "basic_attack"

## Current cooldown remaining
var current_cooldown: float = 0.0

## Owner of this weapon
var owner: Node

func _ready():
	set_process(true)

func _process(delta: float):
	if current_cooldown > 0:
		current_cooldown -= delta
		if current_cooldown <= 0:
			current_cooldown = 0
			cooldown_finished.emit()

## Initialize weapon with owner
func initialize(owner_node: Node) -> void:
	owner = owner_node

## Check if weapon is ready to attack (off cooldown)
func is_ready() -> bool:
	return current_cooldown <= 0

## Get remaining cooldown time
func get_cooldown_remaining() -> float:
	return current_cooldown

## Get cooldown progress (0.0 = ready, 1.0 = just used)
func get_cooldown_progress() -> float:
	if cooldown_time <= 0:
		return 0.0
	return current_cooldown / cooldown_time

## Attempt to attack a target
## Returns true if attack initiated, false if on cooldown or invalid
func attack(target: Node) -> bool:
	if not is_ready():
		return false

	if not target:
		return false

	if not target.has_method("is_valid_target"):
		return false

	if not target.is_valid_target():
		return false

	# Start cooldown
	current_cooldown = cooldown_time
	cooldown_started.emit(cooldown_time)

	# Emit attack event
	attack_started.emit(target)

	# Subclasses override this to do weapon-specific logic
	_perform_attack(target)

	return true

## Override in subclasses to implement weapon-specific attack logic
func _perform_attack(target: Node) -> void:
	# Default: just emit completed
	attack_completed.emit(target)

## Get attack data for network message
func get_attack_data(target: Node) -> Dictionary:
	return {
		"target_id": target.get_entity_id(),
		"ability_id": ability_id,
		"weapon_type": attack_type
	}

## Reset cooldown (for testing or special abilities)
func reset_cooldown() -> void:
	current_cooldown = 0.0
	cooldown_finished.emit()
