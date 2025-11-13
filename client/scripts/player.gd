extends CharacterBody3D

# Player movement
@export var speed = 5.0
@export var jump_velocity = 4.5
@export var mouse_sensitivity = 0.003
@export var rotation_speed = 10.0

# Camera
@onready var camera_pivot = $CameraPivot
@onready var camera = $CameraPivot/Camera3D

# Network
var is_local_player = false
var player_id = ""
var last_position = Vector3.ZERO
var last_rotation = Vector3.ZERO

# Get the gravity from the project settings
var gravity = ProjectSettings.get_setting("physics/3d/default_gravity")

func _ready():
	if is_local_player:
		# Capture mouse for local player
		Input.mouse_mode = Input.MOUSE_MODE_CAPTURED
		camera.current = true
	else:
		# Disable camera for remote players
		camera_pivot.queue_free()

func _input(event):
	if not is_local_player:
		return

	if event is InputEventMouseMotion:
		rotate_y(-event.relative.x * mouse_sensitivity)
		camera_pivot.rotate_x(-event.relative.y * mouse_sensitivity)
		camera_pivot.rotation.x = clamp(camera_pivot.rotation.x, -PI/2, PI/2)

	# Toggle mouse capture
	if event.is_action_pressed("ui_cancel"):
		if Input.mouse_mode == Input.MOUSE_MODE_CAPTURED:
			Input.mouse_mode = Input.MOUSE_MODE_VISIBLE
		else:
			Input.mouse_mode = Input.MOUSE_MODE_CAPTURED

func _physics_process(delta):
	if is_local_player:
		handle_local_movement(delta)

		# Send position to server if changed significantly
		if position.distance_to(last_position) > 0.01 or rotation.distance_to(last_rotation) > 0.01:
			NetworkManager.send_position(position, rotation)
			last_position = position
			last_rotation = rotation

func handle_local_movement(delta):
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

func set_remote_transform(pos: Vector3, rot: Vector3):
	# Smoothly interpolate remote player position
	position = position.lerp(pos, 0.3)
	rotation.y = lerp_angle(rotation.y, rot.y, 0.3)
