extends Node
## Manages combat event reception and routing to animation systems
##
## This singleton receives combat events from the server and routes them
## to the appropriate handlers (animations, VFX, SFX, UI updates)

signal combat_event_received(event: Dictionary)
signal damage_dealt(attacker_id: String, target_id: String, amount: int)
signal entity_died(entity_id: String, killer_id: String)

# Cache for entity lookups
var entities: Dictionary = {}  # entity_id -> Node3D

func _ready():
	# Connect to network manager
	if NetworkManager:
		NetworkManager.message_received.connect(_on_network_message)

func _on_network_message(message: Dictionary):
	match message.type:
		"combat_event":
			handle_combat_event(message.data)
		"combat_events_batch":
			handle_combat_batch(message.data)

func handle_combat_batch(data: Dictionary):
	"""Process batch of combat events"""
	var events: Array = data.events
	for event in events:
		handle_combat_event(event)

func handle_combat_event(event: Dictionary):
	"""Route combat event to appropriate handler"""

	# Determine event type (0=damage, 1=healing, 2=buff, 3=debuff, 4=death)
	var event_type: int = event.get("type", 0)

	match event_type:
		0:  # Damage
			handle_damage_event(event)
		1:  # Healing
			handle_heal_event(event)
		2:  # Buff
			handle_buff_event(event)
		3:  # Debuff
			handle_debuff_event(event)
		4:  # Death
			handle_death_event(event)
		_:
			push_warning("Unknown combat event type: %d" % event_type)

	# Emit for other systems
	combat_event_received.emit(event)

func handle_damage_event(event: Dictionary):
	"""Handle damage combat events - trigger attack and hit animations"""

	var attacker = get_entity(event.caster_id)
	var target = get_entity(event.target_id)

	if attacker:
		play_attack_animation(attacker, event)

	if target:
		play_hit_reaction(target, event)
		show_damage_number(target, event)
		update_health_bar(target, event)

	# Emit signal
	damage_dealt.emit(event.caster_id, event.target_id, event.damage)

	# Check for death
	if event.get("target_died", false):
		handle_death_event(event)

func handle_heal_event(event: Dictionary):
	"""Handle healing events"""
	var target = get_entity(event.target_id)
	if target:
		show_heal_vfx(target, event)
		show_heal_number(target, event)
		update_health_bar(target, event)

func handle_buff_event(event: Dictionary):
	"""Handle buff application"""
	var target = get_entity(event.target_id)
	if target:
		show_buff_vfx(target, event)

func handle_debuff_event(event: Dictionary):
	"""Handle debuff application"""
	var target = get_entity(event.target_id)
	if target:
		show_debuff_vfx(target, event)

func handle_death_event(event: Dictionary):
	"""Handle entity death"""
	var entity = get_entity(event.target_id)
	if entity:
		play_death_animation(entity, event)

	entity_died.emit(event.target_id, event.get("caster_id", ""))

# ============================================================================
# Animation Triggers
# ============================================================================

func play_attack_animation(attacker: Node3D, event: Dictionary):
	"""Trigger attack animation on attacker"""
	if not attacker.has_method("play_attack"):
		return

	# Get animation hint from server
	var anim_hint: String = event.get("animation_hint", "swing_right")
	var target_pos: Vector3 = get_entity_position(event.target_id)

	attacker.play_attack(anim_hint, target_pos, event)

func play_hit_reaction(target: Node3D, event: Dictionary):
	"""Trigger hit reaction animation on target"""
	if not target.has_method("play_hit_reaction"):
		return

	var damage: int = event.get("damage", 0)
	var is_critical: bool = event.get("is_critical", false)

	target.play_hit_reaction(damage, is_critical)

func play_death_animation(entity: Node3D, event: Dictionary):
	"""Trigger death animation"""
	if not entity.has_method("play_death"):
		return

	entity.play_death()

# ============================================================================
# Visual Effects
# ============================================================================

func show_damage_number(target: Node3D, event: Dictionary):
	"""Show floating damage number"""
	if not VFXManager:
		return

	var damage: int = event.get("damage", 0)
	var is_critical: bool = event.get("is_critical", false)
	var damage_type: String = event.get("damage_type", "physical")

	VFXManager.show_damage_number(
		target.global_position,
		damage,
		is_critical,
		damage_type
	)

func show_heal_number(target: Node3D, event: Dictionary):
	"""Show floating heal number"""
	if not VFXManager:
		return

	var healing: int = event.get("healing", 0)
	VFXManager.show_heal_number(target.global_position, healing)

func show_heal_vfx(target: Node3D, event: Dictionary):
	"""Show healing visual effect"""
	if not VFXManager:
		return

	VFXManager.play_heal_effect(target)

func show_buff_vfx(target: Node3D, event: Dictionary):
	"""Show buff visual effect"""
	if not VFXManager:
		return

	var buffs: Array = event.get("buffs_applied", [])
	for buff in buffs:
		VFXManager.play_buff_effect(target, buff)

func show_debuff_vfx(target: Node3D, event: Dictionary):
	"""Show debuff visual effect"""
	if not VFXManager:
		return

	var debuffs: Array = event.get("debuffs_applied", [])
	for debuff in debuffs:
		VFXManager.play_debuff_effect(target, debuff)

func update_health_bar(target: Node3D, event: Dictionary):
	"""Update entity health bar"""
	if not target.has_method("update_health"):
		return

	var current_hp: int = event.get("target_health", 0)
	var max_hp: int = event.get("target_max_health", 1)

	target.update_health(current_hp, max_hp)

# ============================================================================
# Entity Management
# ============================================================================

func register_entity(entity_id: String, entity: Node3D):
	"""Register an entity for combat event routing"""
	entities[entity_id] = entity

func unregister_entity(entity_id: String):
	"""Unregister an entity"""
	entities.erase(entity_id)

func get_entity(entity_id: String) -> Node3D:
	"""Get entity by ID"""
	return entities.get(entity_id)

func get_entity_position(entity_id: String) -> Vector3:
	"""Get entity position, or zero if not found"""
	var entity = get_entity(entity_id)
	if entity:
		return entity.global_position
	return Vector3.ZERO
