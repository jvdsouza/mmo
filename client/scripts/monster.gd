extends CharacterBody3D

# Entity ID for combat system
@export var entity_id: String = ""
@export var monster_name: String = "Monster"

# Visual
@onready var mesh_instance = get_node_or_null("MeshInstance3D")
@onready var label = get_node_or_null("Label3D")

# Health
var current_health: int = 100
var max_health: int = 100

# Network state
var last_position: Vector3 = Vector3.ZERO

func _ready():
	# Register with combat system
	if entity_id != "" and CombatEventManager:
		CombatEventManager.register_entity(entity_id, self)
		print("[Monster] Registered with CombatEventManager: ", entity_id)

	# Update name label
	if label:
		label.text = monster_name

func _exit_tree():
	# Unregister from combat system
	if entity_id != "" and CombatEventManager:
		CombatEventManager.unregister_entity(entity_id)

func update_from_server(pos: Vector3, health: int, max_hp: int):
	"""Update monster state from server"""
	# Smoothly interpolate position
	position = position.lerp(pos, 0.3)
	last_position = pos

	# Update health
	current_health = health
	max_health = max_hp

	# Update label
	if label:
		label.text = "%s\n%d/%d" % [monster_name, current_health, max_health]

# ============================================================================
# COMBAT ANIMATION METHODS - Called by CombatEventManager
# ============================================================================

func play_attack(anim_hint: String, target_position: Vector3, event: Dictionary):
	"""Called when this monster attacks"""
	# Face the target
	if target_position != Vector3.ZERO:
		var direction = (target_position - global_position).normalized()
		if direction.length() > 0.01:
			look_at(global_position + direction, Vector3.UP)

	# Simple visual feedback (scale pulse)
	if mesh_instance:
		var tween = create_tween()
		tween.tween_property(mesh_instance, "scale", Vector3(1.2, 1.2, 1.2), 0.1)
		tween.tween_property(mesh_instance, "scale", Vector3.ONE, 0.2)

func play_hit_reaction(damage: int, is_critical: bool):
	"""Called when this monster takes damage"""
	print("[Monster] ", monster_name, " taking damage: ", damage, " critical: ", is_critical)

	# Flash red
	if mesh_instance:
		var material = mesh_instance.get_active_material(0)
		if material:
			var flash_color = Color.RED if not is_critical else Color.ORANGE_RED
			var tween = create_tween()
			tween.tween_property(material, "albedo_color", flash_color, 0.1)
			tween.tween_property(material, "albedo_color", Color.WHITE, 0.2)

	# Shake
	if mesh_instance:
		var original_pos = mesh_instance.position
		for i in range(5):
			mesh_instance.position = original_pos + Vector3(
				randf_range(-0.1, 0.1),
				randf_range(-0.1, 0.1),
				0
			)
			await get_tree().create_timer(0.02).timeout
		mesh_instance.position = original_pos

func play_death():
	"""Called when this monster dies"""
	print("[Monster] Death animation for: ", monster_name)

	# Fade out and shrink
	var tween = create_tween()
	tween.set_parallel(true)

	if mesh_instance:
		tween.tween_property(mesh_instance, "modulate:a", 0.0, 1.0)

	tween.tween_property(self, "scale", Vector3(0.5, 0.1, 0.5), 1.0)

	await tween.finished

	# Don't queue_free() - let the multiplayer_controller handle cleanup
	# Just hide it
	visible = false
