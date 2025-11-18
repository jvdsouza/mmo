extends CharacterBody3D

# Player movement
@export var speed = 5.0
@export var jump_velocity = 4.5
@export var mouse_sensitivity = 0.003
@export var rotation_speed = 10.0

# Entity ID for combat system
@export var entity_id: String = ""

# Camera
@onready var camera_pivot = $CameraPivot
@onready var camera = $CameraPivot/Camera3D
@onready var reticle = $UI/Reticle

# Animation (optional - will be null if no AnimationTree exists)
@onready var animation_tree: AnimationTree = get_node_or_null("AnimationTree")
@onready var mesh_instance = get_node_or_null("MeshInstance3D")

var animation_state = null
var has_animations = false

# Network
var is_local_player = false
var player_id = ""
var last_position = Vector3.ZERO
var last_rotation = Vector3.ZERO

# Movement state for animations
var current_speed: float = 0.0
var is_attacking: bool = false
var is_dead: bool = false

# Combat systems - NEW architecture
var targeting_system: RaycastTargeting = null
var weapon: MeleeWeapon = null

# UI system - Extensible theme architecture
var ui_theme: UITheme = null
@export_enum("Minimal", "MMO") var preferred_ui_theme: String = "Minimal"

# Get the gravity from the project settings
var gravity = ProjectSettings.get_setting("physics/3d/default_gravity")

func _ready():
	# Setup animations if available
	if animation_tree:
		animation_state = animation_tree.get("parameters/playback")
		has_animations = true
		animation_tree.active = true
		print("[Player] Animation system ready for entity: ", entity_id)
	else:
		print("[Player] No AnimationTree found - using placeholder mesh for entity: ", entity_id)

	# Register with combat system
	if entity_id != "" and CombatEventManager:
		CombatEventManager.register_entity(entity_id, self)
		print("[Player] Registered with CombatEventManager: ", entity_id)

	# Initialize combat systems for local player
	if is_local_player:
		# Setup targeting system
		targeting_system = RaycastTargeting.new()
		targeting_system.max_targeting_range = 100.0
		targeting_system.auto_update = true
		add_child(targeting_system)
		targeting_system.initialize_with_camera(self, camera)
		print("[Player] RaycastTargeting initialized")

		# Setup weapon
		weapon = MeleeWeapon.new()
		weapon.weapon_name = "Basic Sword"
		weapon.attack_range = 3.0
		weapon.cooldown_time = 1.0
		weapon.damage = 25
		weapon.animation_hint = "swing_right"
		weapon.ability_id = "basic_attack"
		add_child(weapon)
		weapon.initialize(self)
		print("[Player] MeleeWeapon initialized")

		# Connect weapon signals to animations
		weapon.attack_started.connect(_on_weapon_attack_started)
		weapon.attack_completed.connect(_on_weapon_attack_completed)

		# Setup UI theme
		_initialize_ui_theme()

	# Setup camera and UI
	if is_local_player:
		Input.mouse_mode = Input.MOUSE_MODE_CAPTURED
		camera.current = true

		# Connect reticle to targeting system
		if reticle:
			reticle.player = self
			if targeting_system:
				reticle.targeting_system = targeting_system
	else:
		# Disable camera and UI for remote players
		if camera_pivot:
			camera_pivot.queue_free()
		if reticle:
			reticle.queue_free()

func _exit_tree():
	# Unregister from combat system
	if entity_id != "" and CombatEventManager:
		CombatEventManager.unregister_entity(entity_id)

func _input(event):
	if not is_local_player or is_dead:
		return

	if event is InputEventMouseMotion:
		rotate_y(-event.relative.x * mouse_sensitivity)
		camera_pivot.rotate_x(-event.relative.y * mouse_sensitivity)
		camera_pivot.rotation.x = clamp(camera_pivot.rotation.x, -PI/2, PI/2)

	# Primary attack (left click)
	if event.is_action_pressed("primary_attack"):
		if Input.mouse_mode == Input.MOUSE_MODE_CAPTURED:
			attempt_attack()

	# Toggle mouse capture
	if event.is_action_pressed("ui_cancel"):
		if Input.mouse_mode == Input.MOUSE_MODE_CAPTURED:
			Input.mouse_mode = Input.MOUSE_MODE_VISIBLE
		else:
			Input.mouse_mode = Input.MOUSE_MODE_CAPTURED

func _physics_process(delta):
	if is_dead:
		return

	if is_local_player:
		handle_local_movement(delta)

		# Update animation based on movement
		current_speed = velocity.length()
		update_movement_animation()

		# Send position to server if changed significantly
		if position.distance_to(last_position) > 0.01 or rotation.distance_to(last_rotation) > 0.01:
			if NetworkManager:
				NetworkManager.send_position(position, rotation)
			last_position = position
			last_rotation = rotation

func handle_local_movement(delta):
	if is_attacking:
		# Reduce movement during attack
		velocity *= 0.5
		move_and_slide()
		return

	# Add gravity
	if not is_on_floor():
		velocity.y -= gravity * delta

	# Handle jump
	if Input.is_action_just_pressed("ui_accept") and is_on_floor():
		velocity.y = jump_velocity

	# Get input direction
	var input_dir = Input.get_vector("move_left", "move_right", "move_forward", "move_back")
	var direction = (transform.basis * Vector3(input_dir.x, 0, input_dir.y)).normalized()

	if direction:
		velocity.x = direction.x * speed
		velocity.z = direction.z * speed
	else:
		velocity.x = move_toward(velocity.x, 0, speed)
		velocity.z = move_toward(velocity.z, 0, speed)

	move_and_slide()

func update_movement_animation():
	if not has_animations or is_attacking:
		return

	# Transition based on speed
	if current_speed < 0.1:
		animation_state.travel("Idle")
	elif current_speed < 3.0:
		animation_state.travel("Walking")
	else:
		animation_state.travel("Running")

func set_remote_transform(pos: Vector3, rot: Vector3):
	# Smoothly interpolate remote player position
	position = position.lerp(pos, 0.3)
	rotation.y = lerp_angle(rotation.y, rot.y, 0.3)

# ============================================================================
# COMBAT ANIMATION METHODS - Called by CombatEventManager
# ============================================================================

func play_attack(anim_hint: String, target_position: Vector3, event: Dictionary):
	"""Called when this entity attacks"""
	if is_dead:
		return

	is_attacking = true

	# Face the target
	if target_position != Vector3.ZERO:
		var direction = (target_position - global_position).normalized()
		if direction.length() > 0.01:
			look_at(global_position + direction, Vector3.UP)

	# Play animation if available
	if has_animations:
		match anim_hint:
			"swing_right":
				animation_state.travel("attack_swing_right")
			"swing_left":
				animation_state.travel("attack_swing_left")
			"overhead_slam":
				animation_state.travel("attack_overhead")
			"thrust":
				animation_state.travel("attack_thrust")
			_:
				# Default attack animation
				animation_state.travel("Attack")

		# Wait for attack animation to complete
		await get_tree().create_timer(0.8).timeout
		animation_state.travel("Idle")
	else:
		# Visual feedback for placeholder mesh
		if mesh_instance:
			var tween = create_tween()
			tween.tween_property(mesh_instance, "scale", Vector3(1.2, 1.2, 1.2), 0.1)
			tween.tween_property(mesh_instance, "scale", Vector3.ONE, 0.2)

		await get_tree().create_timer(0.3).timeout

	is_attacking = false

func play_hit_reaction(damage: int, is_critical: bool):
	"""Called when this entity takes damage"""
	if is_dead:
		return

	print("[Player] Taking damage: ", damage, " critical: ", is_critical)

	# Play hit animation if available
	if has_animations:
		if is_critical:
			animation_state.travel("hit_critical")
		else:
			animation_state.travel("Hit_Reaction")

		# Return to previous state after hit
		await get_tree().create_timer(0.5).timeout
		if not is_dead and not is_attacking:
			animation_state.travel("Idle")

	# Visual feedback (works for both animated and placeholder)
	flash_damage(is_critical)

	# Screen shake for local player
	if is_local_player and camera:
		shake_camera(0.2, is_critical)

func play_death():
	"""Called when this entity dies"""
	is_dead = true

	print("[Player] Death animation for entity: ", entity_id)

	# Play death animation if available
	if has_animations:
		animation_state.travel("Death")

	# Fade out and disable
	var tween = create_tween()
	tween.set_parallel(true)

	if mesh_instance:
		tween.tween_property(mesh_instance, "modulate:a", 0.0, 2.0)

	tween.tween_property(self, "scale", Vector3(0.5, 0.1, 0.5), 2.0)

	await tween.finished

	# Could respawn here or queue_free()
	queue_free()

func flash_damage(is_critical: bool):
	"""Visual flash when taking damage"""
	if not mesh_instance:
		return

	var flash_color = Color.RED if not is_critical else Color.ORANGE
	var material = mesh_instance.get_active_material(0)

	if material:
		var tween = create_tween()
		tween.tween_property(material, "albedo_color", flash_color, 0.1)
		tween.tween_property(material, "albedo_color", Color.WHITE, 0.2)

func shake_camera(intensity: float, is_critical: bool):
	"""Camera shake effect for hits"""
	if not camera_pivot:
		return

	var shake_amount = intensity * (2.0 if is_critical else 1.0)
	var original_pos = camera_pivot.position

	for i in range(10):
		var offset = Vector3(
			randf_range(-shake_amount, shake_amount),
			randf_range(-shake_amount, shake_amount),
			0
		)
		camera_pivot.position = original_pos + offset
		await get_tree().create_timer(0.02).timeout

	camera_pivot.position = original_pos

# ============================================================================
# COMBAT / TARGETING METHODS
# ============================================================================

func attempt_attack():
	"""Try to attack using new combat architecture"""
	# Check if already attacking
	if is_attacking:
		print("[Player] Already attacking")
		return

	# Check if we have valid combat systems
	if not targeting_system or not weapon:
		print("[Player] Combat systems not initialized")
		return

	# Check if we have a valid target
	if not targeting_system.has_valid_target():
		print("[Player] No valid target")
		return

	var target = targeting_system.get_current_target()

	# Validate attack using CombatValidator
	var validation = CombatValidator.validate_attack(self, target, weapon, targeting_system)

	if not CombatValidator.is_success(validation):
		print("[Player] Attack validation failed: ", CombatValidator.get_validation_message(validation))
		return

	# Attack with weapon (weapon handles cooldown, network messages, etc.)
	var success = weapon.attack(target)

	if success:
		var distance = global_position.distance_to(target.get_target_position())
		print("[Player] Attacking: ", target.get_entity_id(), " at distance: ", distance)

# ============================================================================
# WEAPON SIGNAL HANDLERS
# ============================================================================

func _on_weapon_attack_started(target: Node):
	"""Called when weapon starts an attack"""
	if target and target.has_method("get_target_position"):
		# Play attack animation with target position
		play_attack(weapon.animation_hint, target.get_target_position(), {})

func _on_weapon_attack_completed(target: Node):
	"""Called when weapon completes an attack"""
	# Additional effects could go here (VFX, sounds, etc.)
	pass

# ============================================================================
# ITARGETABLE INTERFACE IMPLEMENTATION
# ============================================================================

func get_entity_id() -> String:
	return entity_id

func is_valid_target() -> bool:
	return not is_dead and visible

func get_target_position() -> Vector3:
	return global_position

func get_display_name() -> String:
	return player_id if player_id != "" else "Player"

func get_current_health() -> int:
	# This would be synced from server in a real implementation
	return 100

func get_max_health() -> int:
	return 100

func get_faction() -> String:
	return "player"

func get_visual_node() -> Node3D:
	return self

# ============================================================================
# UTILITY METHODS
# ============================================================================

func is_alive() -> bool:
	return not is_dead

# ============================================================================
# UI SYSTEM METHODS
# ============================================================================

func _initialize_ui_theme():
	"""Initialize UI theme based on preference"""
	# Get UI container
	var ui_container = get_node_or_null("UI")
	if not ui_container:
		print("[Player] No UI container found")
		return

	# Create theme based on preference
	match preferred_ui_theme:
		"Minimal":
			ui_theme = MinimalTheme.new()
		"MMO":
			ui_theme = MMOTheme.new()
		_:
			ui_theme = MinimalTheme.new()  # Default

	# Add theme to UI container
	ui_container.add_child(ui_theme)

	# Initialize theme with game systems
	ui_theme.initialize(self, targeting_system, weapon)

	print("[Player] UI Theme initialized: ", ui_theme.theme_name)

func switch_ui_theme(theme_name: String):
	"""Switch to a different UI theme at runtime"""
	if not is_local_player:
		return

	var ui_container = get_node_or_null("UI")
	if not ui_container:
		return

	# Cleanup old theme
	if ui_theme:
		ui_theme.cleanup()
		ui_theme.queue_free()
		ui_theme = null

	# Create new theme
	preferred_ui_theme = theme_name
	match theme_name:
		"Minimal":
			ui_theme = MinimalTheme.new()
		"MMO":
			ui_theme = MMOTheme.new()
		_:
			ui_theme = MinimalTheme.new()

	# Add and initialize new theme
	ui_container.add_child(ui_theme)
	ui_theme.initialize(self, targeting_system, weapon)

	print("[Player] Switched to UI theme: ", ui_theme.theme_name)
