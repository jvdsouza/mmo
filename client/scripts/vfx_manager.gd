extends Node
## VFX Manager - Handles all visual effects for combat
##
## Manages particle effects, damage numbers, and other combat visuals

# Preloaded effect scenes
var hit_effect_scene = preload("res://effects/hit_effect.tscn") if ResourceLoader.exists("res://effects/hit_effect.tscn") else null
var crit_effect_scene = preload("res://effects/critical_hit.tscn") if ResourceLoader.exists("res://effects/critical_hit.tscn") else null
var fire_effect_scene = preload("res://effects/fire_damage.tscn") if ResourceLoader.exists("res://effects/fire_damage.tscn") else null
var poison_effect_scene = preload("res://effects/poison_damage.tscn") if ResourceLoader.exists("res://effects/poison_damage.tscn") else null
var heal_effect_scene = preload("res://effects/heal_effect.tscn") if ResourceLoader.exists("res://effects/heal_effect.tscn") else null

# Damage number label (created procedurally)
const DAMAGE_NUMBER_FONT_SIZE = 48
const DAMAGE_NUMBER_OUTLINE = 8

# Effect pooling for performance
var effect_pools: Dictionary = {}
var max_pool_size: int = 20

# ============================================================================
# Hit Effects
# ============================================================================

func play_hit_effect(position: Vector3, event: Dictionary):
	"""Play hit visual effect based on damage type"""

	var damage_type: String = event.get("damage_type", "physical")
	var is_critical: bool = event.get("is_critical", false)

	# Get appropriate effect scene
	var effect: GPUParticles3D

	match damage_type:
		"fire":
			effect = get_pooled_effect("fire")
		"ice":
			effect = get_pooled_effect("ice")
		"poison":
			effect = get_pooled_effect("poison")
		"lightning":
			effect = get_pooled_effect("lightning")
		_:  # physical or default
			effect = get_pooled_effect("hit")

	if effect:
		# Position and trigger
		effect.global_position = position
		effect.emitting = true
		effect.restart()

	# Add critical hit effect
	if is_critical and crit_effect_scene:
		var crit_effect = crit_effect_scene.instantiate()
		get_tree().root.add_child(crit_effect)
		crit_effect.global_position = position + Vector3(0, 2, 0)
		crit_effect.emitting = true

		# Auto-cleanup
		await get_tree().create_timer(2.0).timeout
		crit_effect.queue_free()

func play_heal_effect(target: Node3D):
	"""Play healing visual effect"""

	if not heal_effect_scene:
		return

	var heal_effect = heal_effect_scene.instantiate()
	target.add_child(heal_effect)
	heal_effect.position = Vector3(0, 1, 0)

	if heal_effect is GPUParticles3D:
		heal_effect.emitting = true

	# Auto-cleanup
	await get_tree().create_timer(3.0).timeout
	heal_effect.queue_free()

func play_buff_effect(target: Node3D, buff_id: String):
	"""Play buff visual effect"""

	# Different effects based on buff type
	var buff_color: Color

	match buff_id:
		"strength_buff":
			buff_color = Color.RED
		"defense_buff":
			buff_color = Color.STEEL_BLUE
		"speed_buff":
			buff_color = Color.YELLOW
		"regeneration":
			buff_color = Color.GREEN
		_:
			buff_color = Color.WHITE

	# Create aura effect
	create_aura_effect(target, buff_color)

func play_debuff_effect(target: Node3D, debuff_id: String):
	"""Play debuff visual effect"""

	var debuff_color: Color

	match debuff_id:
		"poison":
			debuff_color = Color.GREEN
		"burn":
			debuff_color = Color.ORANGE
		"slow":
			debuff_color = Color.DEEP_SKY_BLUE
		"stun":
			debuff_color = Color.YELLOW
		_:
			debuff_color = Color.PURPLE

	create_aura_effect(target, debuff_color, true)

# ============================================================================
# Damage/Heal Numbers
# ============================================================================

func show_damage_number(position: Vector3, damage: int, is_critical: bool, damage_type: String):
	"""Show floating damage number"""

	var label = Label3D.new()
	label.text = str(damage)
	label.font_size = DAMAGE_NUMBER_FONT_SIZE
	label.outline_size = DAMAGE_NUMBER_OUTLINE
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED

	# Color based on damage type and critical
	if is_critical:
		label.modulate = Color.ORANGE_RED
		label.text = str(damage) + "!"
		label.font_size = 64  # Bigger for crits
	else:
		match damage_type:
			"fire":
				label.modulate = Color.ORANGE
			"ice":
				label.modulate = Color.CYAN
			"poison":
				label.modulate = Color.GREEN
			"lightning":
				label.modulate = Color.YELLOW
			_:
				label.modulate = Color.WHITE

	# Add to scene
	get_tree().root.add_child(label)
	label.global_position = position + Vector3(0, 2, 0)

	# Animate floating up and fading
	var tween = create_tween()
	tween.set_parallel(true)
	tween.tween_property(label, "position:y", label.position.y + 2.5, 1.5).set_ease(Tween.EASE_OUT)
	tween.tween_property(label, "modulate:a", 0.0, 1.5).set_delay(0.5)

	# Cleanup
	await tween.finished
	label.queue_free()

func show_heal_number(position: Vector3, healing: int):
	"""Show floating heal number"""

	var label = Label3D.new()
	label.text = "+" + str(healing)
	label.font_size = DAMAGE_NUMBER_FONT_SIZE
	label.outline_size = DAMAGE_NUMBER_OUTLINE
	label.modulate = Color.GREEN_YELLOW
	label.billboard = BaseMaterial3D.BILLBOARD_ENABLED

	# Add to scene
	get_tree().root.add_child(label)
	label.global_position = position + Vector3(0, 2, 0)

	# Animate
	var tween = create_tween()
	tween.set_parallel(true)
	tween.tween_property(label, "position:y", label.position.y + 2, 1.2).set_ease(Tween.EASE_OUT)
	tween.tween_property(label, "modulate:a", 0.0, 1.2).set_delay(0.3)

	await tween.finished
	label.queue_free()

# ============================================================================
# Helper Functions
# ============================================================================

func create_aura_effect(target: Node3D, color: Color, is_debuff: bool = false):
	"""Create aura effect around entity"""

	# Create simple particle aura
	var particles = GPUParticles3D.new()
	target.add_child(particles)
	particles.position = Vector3(0, 1, 0)

	# Configure particles
	particles.amount = 20
	particles.lifetime = 0.8
	particles.emitting = true

	# Material
	var material = ParticleProcessMaterial.new()
	material.emission_shape = ParticleProcessMaterial.EMISSION_SHAPE_SPHERE
	material.emission_sphere_radius = 0.5
	material.direction = Vector3(0, 1, 0)
	material.spread = 180.0
	material.gravity = Vector3(0, -1, 0) if is_debuff else Vector3(0, 1, 0)
	material.initial_velocity_min = 0.5
	material.initial_velocity_max = 1.0
	material.color = color

	particles.process_material = material

	# Auto-cleanup after 3 seconds
	await get_tree().create_timer(3.0).timeout
	particles.emitting = false
	await get_tree().create_timer(1.0).timeout
	particles.queue_free()

func get_pooled_effect(effect_type: String) -> GPUParticles3D:
	"""Get effect from pool or create new one"""

	# Initialize pool for this type
	if effect_type not in effect_pools:
		effect_pools[effect_type] = []

	# Find inactive effect in pool
	for effect in effect_pools[effect_type]:
		if not effect.emitting:
			return effect

	# Create new effect if pool not full
	if effect_pools[effect_type].size() < max_pool_size:
		var new_effect = create_effect(effect_type)
		if new_effect:
			effect_pools[effect_type].append(new_effect)
			get_tree().root.add_child(new_effect)
			return new_effect

	# Pool full, reuse oldest
	return effect_pools[effect_type][0] if effect_pools[effect_type].size() > 0 else null

func create_effect(effect_type: String) -> GPUParticles3D:
	"""Create effect based on type"""

	# Try to load from scene first
	var scene_path = "res://effects/" + effect_type + "_effect.tscn"
	if ResourceLoader.exists(scene_path):
		var scene = load(scene_path)
		return scene.instantiate()

	# Create procedural effect as fallback
	return create_procedural_effect(effect_type)

func create_procedural_effect(effect_type: String) -> GPUParticles3D:
	"""Create simple procedural particle effect"""

	var particles = GPUParticles3D.new()
	particles.amount = 30
	particles.lifetime = 0.5
	particles.one_shot = true

	var material = ParticleProcessMaterial.new()
	material.emission_shape = ParticleProcessMaterial.EMISSION_SHAPE_SPHERE
	material.emission_sphere_radius = 0.3

	# Configure based on type
	match effect_type:
		"fire":
			material.color = Color.ORANGE
			material.direction = Vector3(0, 1, 0)
		"ice":
			material.color = Color.CYAN
			material.direction = Vector3(0, -1, 0)
		"poison":
			material.color = Color.GREEN
			material.direction = Vector3(0, 0.5, 0)
		"lightning":
			material.color = Color.YELLOW
			material.direction = Vector3(0, 1, 0)
		_:  # hit
			material.color = Color.WHITE
			material.direction = Vector3(0, 0, 0)
			material.spread = 180.0

	material.initial_velocity_min = 1.0
	material.initial_velocity_max = 2.0

	particles.process_material = material

	return particles
