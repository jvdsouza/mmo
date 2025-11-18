extends Node
class_name ITargetable

## Interface for entities that can be targeted in combat
##
## Entities that want to be targetable should extend this class and implement its methods.
## This provides a contract for what targeting systems can expect.

## Get the unique entity ID for network communication
func get_entity_id() -> String:
	push_error("ITargetable.get_entity_id() must be overridden")
	return ""

## Check if this entity is currently a valid target
## Returns false if dead, invulnerable, friendly, etc.
func is_valid_target() -> bool:
	push_error("ITargetable.is_valid_target() must be overridden")
	return false

## Get the world position for distance calculations and aiming
func get_target_position() -> Vector3:
	push_error("ITargetable.get_target_position() must be overridden")
	return Vector3.ZERO

## Get display name for UI
func get_display_name() -> String:
	push_error("ITargetable.get_display_name() must be overridden")
	return "Unknown"

## Get current health for UI display
func get_current_health() -> int:
	push_error("ITargetable.get_current_health() must be overridden")
	return 0

## Get maximum health for UI display
func get_max_health() -> int:
	push_error("ITargetable.get_max_health() must be overridden")
	return 100

## Check if entity is alive
func is_alive() -> bool:
	push_error("ITargetable.is_alive() must be overridden")
	return true

## Get faction/team for friend/foe detection
## Returns: "player", "monster", "neutral", etc.
func get_faction() -> String:
	push_error("ITargetable.get_faction() must be overridden")
	return "neutral"

## Optional: Get visual node for effects (health bars, etc.)
func get_visual_node() -> Node3D:
	if self is Node3D:
		return self as Node3D
	return null
