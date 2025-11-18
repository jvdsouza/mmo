extends Weapon
class_name MeleeWeapon

## Melee weapon implementation
##
## Instant-hit melee attacks (swords, axes, fists, etc.)
## Uses raycast or collision detection for hit validation.

## Override to send network attack and play animation
func _perform_attack(target: ITargetable) -> void:
	# Send attack to server
	if NetworkManager:
		NetworkManager.send_attack(target.get_entity_id(), ability_id)

	# Emit for animation/VFX systems
	attack_completed.emit(target)
