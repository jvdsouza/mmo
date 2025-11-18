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

# Combat/targeting
var current_target: Node3D = null
var attack_range: float = 3.0  # Melee range
var attack_cooldown: float = 0.0
var attack_cooldown_time: float = 1.0  # 1 second between attacks

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

	# Setup camera and UI
	if is_local_player:
		Input.mouse_mode = Input.MOUSE_MODE_CAPTURED
		camera.current = true

		# Connect reticle to player
		if reticle:
			reticle.player = self
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

	# Update attack cooldown
	if attack_cooldown > 0:
		attack_cooldown -= delta

	if is_local_player:
		handle_local_movement(delta)

		# Update animation based on movement
		current_speed = velocity.length()
		update_movement_animation()

		# Update target detection (raycast from camera)
		update_target_detection()

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

func update_target_detection():
	"""Continuously raycast from camera to detect potential targets"""
	if not camera:
		return

	var space_state = get_world_3d().direct_space_state
	var camera_pos = camera.global_position
	var camera_forward = -camera.global_transform.basis.z
	var ray_distance = attack_range * 2  # Check further than attack range for UI feedback

	# Create raycast query
	var query = PhysicsRayQueryParameters3D.create(
		camera_pos,
		camera_pos + camera_forward * ray_distance
	)
	query.exclude = [self]  # Don't hit yourself

	var result = space_state.intersect_ray(query)

	if result:
		var collider = result.collider

		# Check if it's an attackable entity (has entity_id)
		if collider.has_method("get") and collider.get("entity_id"):
			current_target = collider
		else:
			current_target = null
	else:
		current_target = null

func attempt_attack():
	"""Try to attack the current target"""
	# Check cooldown
	if attack_cooldown > 0:
		print("[Player] Attack on cooldown: ", attack_cooldown)
		return

	# Check if already attacking
	if is_attacking:
		print("[Player] Already attacking")
		return

	# Get target from raycast
	var target = get_raycast_target()

	if not target:
		print("[Player] No target in range")
		return

	var target_entity_id = target.get("entity_id")
	if not target_entity_id or target_entity_id == "":
		print("[Player] Target has no entity_id")
		return

	# Check range
	var distance = global_position.distance_to(target.global_position)
	if distance > attack_range:
		print("[Player] Target out of range: ", distance, " > ", attack_range)
		return

	# Send attack to server
	print("[Player] Attacking: ", target_entity_id, " at distance: ", distance)
	if NetworkManager:
		NetworkManager.send_attack(target_entity_id)

	# Start attack cooldown
	attack_cooldown = attack_cooldown_time

	# Play local attack animation prediction (server will confirm)
	play_attack("swing_right", target.global_position, {})

func get_raycast_target() -> Node3D:
	"""Perform raycast from camera and return the target if valid"""
	if not camera:
		return null

	var space_state = get_world_3d().direct_space_state
	var camera_pos = camera.global_position
	var camera_forward = -camera.global_transform.basis.z

	# Create raycast query
	var query = PhysicsRayQueryParameters3D.create(
		camera_pos,
		camera_pos + camera_forward * attack_range
	)
	query.exclude = [self]  # Don't hit yourself

	var result = space_state.intersect_ray(query)

	if result:
		var collider = result.collider

		# Check if it's an attackable entity
		if collider.has_method("get") and collider.get("entity_id"):
			return collider

	return null

# ============================================================================
# UTILITY METHODS
# ============================================================================

func get_current_health() -> int:
	# This would be synced from server in a real implementation
	return 100

func is_alive() -> bool:
	return not is_dead
