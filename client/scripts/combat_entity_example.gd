extends CharacterBody3D
## Example combat entity with animation integration
##
## This shows how to implement the animation methods called by CombatEventManager

@export var entity_id: String = ""
@export var health: float = 100.0
@export var max_health: float = 100.0

@onready var animation_tree: AnimationTree = $AnimationTree
@onready var animation_state: AnimationNodeStateMachinePlayback
@onready var mesh: MeshInstance3D = $MeshInstance3D
@onready var health_bar: ProgressBar = $HealthBar3D/SubViewport/HealthBar

enum AnimState {
	IDLE,
	WALKING,
	RUNNING,
	ATTACKING,
	HIT_REACTION,
	DYING,
	DEAD
}

var current_anim_state: AnimState = AnimState.IDLE
var is_dead: bool = false

func _ready():
	# Get animation state machine
	if animation_tree:
		animation_state = animation_tree.get("parameters/playback")

	# Register with combat event manager
	if CombatEventManager:
		CombatEventManager.register_entity(entity_id, self)

func _exit_tree():
	# Unregister when removed
	if CombatEventManager:
		CombatEventManager.unregister_entity(entity_id)

# ============================================================================
# Combat Animation Methods (called by CombatEventManager)
# ============================================================================

func play_attack(anim_hint: String, target_position: Vector3, event: Dictionary):
	"""Play attack animation based on server hint"""

	if is_dead:
		return

	# Face target
	if target_position != Vector3.ZERO:
		look_at(target_position, Vector3.UP)

	# Choose animation based on hint
	var anim_name: String = "attack_basic"

	match anim_hint:
		"swing_right":
			anim_name = "attack_swing_right"
		"swing_left":
			anim_name = "attack_swing_left"
		"overhead_slam":
			anim_name = "attack_overhead"
		"thrust":
			anim_name = "attack_thrust"
		"sweep":
			anim_name = "attack_sweep"
		_:
			anim_name = "attack_basic"

	# Play animation
	if animation_state:
		animation_state.travel(anim_name)
		current_anim_state = AnimState.ATTACKING

	# Trigger hit VFX after hit frame (30% through animation)
	var hit_timing: float = 0.3
	await get_tree().create_timer(hit_timing).timeout

	# Show hit effect at target position
	if VFXManager:
		VFXManager.play_hit_effect(target_position, event)

	# Return to idle
	await get_tree().create_timer(0.5).timeout
	if current_anim_state == AnimState.ATTACKING:
		animation_state.travel("idle")
		current_anim_state = AnimState.IDLE

func play_hit_reaction(damage: int, is_critical: bool):
	"""Play hit reaction animation when taking damage"""

	if is_dead:
		return

	# Choose hit reaction based on damage amount
	var hit_anim: String = "hit_light"

	if is_critical:
		hit_anim = "hit_critical"
	elif damage > 50:
		hit_anim = "hit_heavy"

	# Play animation
	if animation_state:
		animation_state.travel(hit_anim)
		current_anim_state = AnimState.HIT_REACTION

	# Flash damage color
	flash_damage_material()

	# Return to idle after hit reaction
	await get_tree().create_timer(0.4).timeout
	if current_anim_state == AnimState.HIT_REACTION:
		animation_state.travel("idle")
		current_anim_state = AnimState.IDLE

func play_death():
	"""Play death animation"""

	is_dead = true
	current_anim_state = AnimState.DYING

	# Play death animation
	if animation_state:
		animation_state.travel("death")

	# Disable collision
	$CollisionShape3D.disabled = true

	# Wait for animation to finish
	await get_tree().create_timer(2.0).timeout
	current_anim_state = AnimState.DEAD

	# Fade out and remove
	var tween = create_tween()
	tween.tween_property(self, "modulate:a", 0.0, 2.0)
	await tween.finished

	queue_free()

func update_health(current_hp: int, max_hp: int):
	"""Update health bar UI"""

	health = float(current_hp)
	max_health = float(max_hp)

	if health_bar:
		health_bar.max_value = max_health
		health_bar.value = health

	# Update color based on health percentage
	var health_percent: float = health / max_health

	if health_percent > 0.6:
		health_bar.modulate = Color.GREEN
	elif health_percent > 0.3:
		health_bar.modulate = Color.YELLOW
	else:
		health_bar.modulate = Color.RED

# ============================================================================
# Visual Effects
# ============================================================================

func flash_damage_material():
	"""Flash red when taking damage"""

	if not mesh:
		return

	var material = mesh.get_surface_override_material(0)
	if not material:
		material = mesh.get_active_material(0)

	if material:
		# Store original color
		var original_color = material.albedo_color

		# Flash red
		material.albedo_color = Color.RED

		# Restore after delay
		await get_tree().create_timer(0.15).timeout
		material.albedo_color = original_color

# ============================================================================
# Movement Animations (for reference)
# ============================================================================

func _physics_process(delta):
	# Update movement animations based on velocity
	if is_dead:
		return

	# Only update movement animations if not in combat
	if current_anim_state in [AnimState.ATTACKING, AnimState.HIT_REACTION, AnimState.DYING]:
		return

	# Determine animation based on velocity
	var speed = velocity.length()

	if speed < 0.1:
		if current_anim_state != AnimState.IDLE:
			animation_state.travel("idle")
			current_anim_state = AnimState.IDLE
	elif speed < 3.0:
		if current_anim_state != AnimState.WALKING:
			animation_state.travel("walk")
			current_anim_state = AnimState.WALKING
	else:
		if current_anim_state != AnimState.RUNNING:
			animation_state.travel("run")
			current_anim_state = AnimState.RUNNING
