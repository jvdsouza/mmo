extends Node
class_name TargetingSystem

## Base class for all targeting systems (raycast, tab-target, lock-on, etc.)
##
## Targeting systems are responsible for selecting which entity to attack.
## Different implementations provide different player experiences.

## Emitted when the current target changes
signal target_changed(old_target: Node, new_target: Node)

## Emitted when target becomes invalid (dies, out of range, etc.)
signal target_lost(target: Node)

## Emitted when attempting to target but no valid targets available
signal no_valid_targets()

## Current target (null if no target)
var current_target: Node = null

## Maximum range for target acquisition
@export var max_targeting_range: float = 100.0

## Owner of this targeting system (usually the player)
var owner: Node3D

## Initialize the targeting system
func initialize(owner_node: Node3D) -> void:
	owner = owner_node

## Update targeting (called every frame)
## Override this in subclasses to implement targeting logic
func update_targeting(delta: float) -> void:
	pass

## Get the current target
func get_current_target() -> Node:
	return current_target

## Check if we have a valid target
func has_valid_target() -> bool:
	if not current_target:
		return false

	if not is_instance_valid(current_target):
		clear_target()
		return false

	if not current_target.has_method("is_valid_target"):
		clear_target()
		return false

	if not current_target.is_valid_target():
		clear_target()
		return false

	return true

## Manually set a target (used by some targeting modes)
func set_target(target: Node) -> void:
	if target == current_target:
		return

	var old_target = current_target
	current_target = target
	target_changed.emit(old_target, current_target)

## Clear the current target
func clear_target() -> void:
	if current_target:
		var old_target = current_target
		current_target = null
		target_lost.emit(old_target)

## Check if a node is a valid targetable entity
## Uses duck typing - checks for ITargetable interface methods
func is_valid_targetable(node: Node) -> bool:
	if not node:
		return false

	# Check if node implements ITargetable interface (duck typing)
	if not node.has_method("is_valid_target"):
		return false

	if not node.has_method("is_alive"):
		return false

	if not node.is_valid_target():
		return false

	if not node.is_alive():
		return false

	return true

## Check if target is within range
func is_in_range(target: Node, range: float) -> bool:
	if not target or not owner:
		return false

	if not target.has_method("get_target_position"):
		return false

	var distance = owner.global_position.distance_to(target.get_target_position())
	return distance <= range

## Get distance to current target
func get_distance_to_target() -> float:
	if not current_target or not owner:
		return INF

	if not current_target.has_method("get_target_position"):
		return INF

	return owner.global_position.distance_to(current_target.get_target_position())
