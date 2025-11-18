extends TargetingSystem
class_name RaycastTargeting

## Raycast-based targeting for action combat
##
## Targets whatever the camera/cursor is aimed at via physics raycast.
## Perfect for FPS-style or action combat.

## Camera to raycast from
var camera: Camera3D

## Physics layers to raycast against (default: all)
@export var collision_mask: int = 0xFFFFFFFF

## Automatically update targeting every frame
@export var auto_update: bool = true

func _ready():
	set_physics_process(auto_update)

func _physics_process(delta: float):
	if auto_update:
		update_targeting(delta)

## Initialize with camera reference
func initialize_with_camera(owner_node: Node3D, camera_node: Camera3D) -> void:
	super.initialize(owner_node)
	camera = camera_node

## Update targeting via raycast from camera
func update_targeting(delta: float) -> void:
	if not camera or not owner:
		return

	var new_target = perform_raycast()

	# Target changed
	if new_target != current_target:
		var old_target = current_target
		current_target = new_target
		target_changed.emit(old_target, new_target)

## Perform raycast from camera and return target
func perform_raycast() -> Node:
	var space_state = owner.get_world_3d().direct_space_state
	var camera_pos = camera.global_position
	var camera_forward = -camera.global_transform.basis.z

	# Raycast to max range
	var query = PhysicsRayQueryParameters3D.create(
		camera_pos,
		camera_pos + camera_forward * max_targeting_range
	)
	query.exclude = [owner]  # Don't hit self
	query.collision_mask = collision_mask

	var result = space_state.intersect_ray(query)

	if result and result.has("collider"):
		var collider = result.collider

		# Check if it's a valid targetable entity
		if is_valid_targetable(collider):
			return collider

	return null

## Get raycast hit point in world (useful for projectiles, effects)
func get_raycast_hit_point() -> Vector3:
	if not camera or not owner:
		return Vector3.ZERO

	var space_state = owner.get_world_3d().direct_space_state
	var camera_pos = camera.global_position
	var camera_forward = -camera.global_transform.basis.z

	var query = PhysicsRayQueryParameters3D.create(
		camera_pos,
		camera_pos + camera_forward * max_targeting_range
	)
	query.exclude = [owner]
	query.collision_mask = collision_mask

	var result = space_state.intersect_ray(query)

	if result and result.has("position"):
		return result.position

	# No hit - return point at max range
	return camera_pos + camera_forward * max_targeting_range

## Check if raycast would hit given target
func is_aiming_at(target: Node) -> bool:
	var raycast_target = perform_raycast()
	return raycast_target == target
