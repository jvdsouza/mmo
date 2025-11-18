# Animation Integration with Combat System

## Architecture Overview

### Server Responsibilities (Go)
- Execute combat logic
- Validate actions
- Update entity states
- **Broadcast combat events to clients**

### Client Responsibilities (Godot)
- Receive combat events
- Play animations
- Show visual effects (VFX)
- Play sounds (SFX)
- Update UI (health bars, damage numbers)

## Server-Side: Combat Event Broadcasting

### 1. Enhanced Combat Event Structure

```go
// CombatEvent represents a combat action that occurred
type CombatEvent struct {
    // Event identification
    EventID   string    `json:"event_id"`
    Timestamp time.Time `json:"timestamp"`
    EventType string    `json:"event_type"` // "melee_attack", "ability_cast", "damage_taken", "death", etc.

    // Entities involved
    CasterID  string `json:"caster_id"`
    TargetID  string `json:"target_id"`

    // Combat data
    AbilityID   string `json:"ability_id,omitempty"`
    Damage      int    `json:"damage"`
    IsCritical  bool   `json:"is_critical"`
    WasBlocked  bool   `json:"was_blocked"`
    WasDodged   bool   `json:"was_dodged"`
    DamageType  string `json:"damage_type"` // "physical", "fire", "poison", etc.

    // Results
    TargetHealth    int  `json:"target_health"`
    TargetMaxHealth int  `json:"target_max_health"`
    TargetDied      bool `json:"target_died"`

    // Animation hints (optional, helps client choose animation)
    AnimationHint string         `json:"animation_hint,omitempty"` // "swing_right", "thrust", "overhead_slam"
    Position      Vector3        `json:"position,omitempty"`
    Direction     Vector3        `json:"direction,omitempty"`
    ExtraData     map[string]any `json:"extra_data,omitempty"`
}
```

### 2. Event Types and Animation Mapping

| Event Type | Server Data | Client Animation |
|-----------|-------------|------------------|
| `melee_attack` | Attacker ID, Target ID, Damage | Attack swing, hit reaction |
| `ability_cast` | Caster ID, Ability ID | Casting animation, projectile spawn |
| `damage_taken` | Target ID, Damage, Type | Hit reaction, damage number |
| `death` | Entity ID, Killer ID | Death animation, ragdoll, despawn |
| `dodge` | Entity ID | Dodge roll animation |
| `block` | Entity ID | Block/parry animation |
| `environmental_damage` | Zone ID, Target ID, Damage | Burning/poison/shock effect |
| `heal` | Healer ID, Target ID, Amount | Healing VFX, health bar update |
| `buff_applied` | Caster ID, Target ID, Buff ID | Buff icon, aura VFX |
| `stun` | Caster ID, Target ID, Duration | Stunned animation, stars VFX |

### 3. Interest Management for Events

Combat events are only sent to clients that need to see them:

```go
func (cm *CombatManager) BroadcastCombatEvent(event *CombatEvent) {
    // Get the combat entity to determine broadcast radius
    entity := cm.GetEntity(event.CasterID)
    if entity == nil {
        return
    }

    // Calculate interest area
    interestRadius := entity.GetInterestRadius()

    // Only broadcast if entity wants combat events shown
    if !entity.ShouldBroadcastCombat() {
        return
    }

    // Find all players in interest radius
    nearbyEntities := cm.spatialGrid.GetNearbyEntities(
        entity.GetPosition(),
        interestRadius,
    )

    // Collect player IDs to notify
    playerIDs := []string{}
    for _, nearby := range nearbyEntities {
        if nearby.GetType() == EntityTypePlayer {
            playerIDs = append(playerIDs, nearby.GetID())
        }
    }

    // Add combatants themselves (always see their own combat)
    playerIDs = append(playerIDs, event.CasterID)
    if event.TargetID != "" {
        playerIDs = append(playerIDs, event.TargetID)
    }

    // Send event to relevant players
    for _, playerID := range playerIDs {
        cm.SendEventToPlayer(playerID, event)
    }
}
```

## Client-Side: Animation System (Godot)

### 1. Combat Event Receiver

```gdscript
# combat_event_manager.gd
extends Node

signal combat_event_received(event: Dictionary)

func _ready():
    NetworkManager.combat_event.connect(_on_combat_event)

func _on_combat_event(event: Dictionary):
    # Route event to appropriate handler
    match event.event_type:
        "melee_attack":
            handle_melee_attack(event)
        "ability_cast":
            handle_ability_cast(event)
        "damage_taken":
            handle_damage_taken(event)
        "death":
            handle_death(event)
        "environmental_damage":
            handle_environmental_damage(event)
        _:
            push_warning("Unknown combat event type: " + event.event_type)

    # Emit for other systems to react (UI, sound, etc.)
    combat_event_received.emit(event)

func handle_melee_attack(event: Dictionary):
    var attacker = get_entity(event.caster_id)
    var target = get_entity(event.target_id)

    if attacker:
        play_attack_animation(attacker, event)

    if target:
        play_hit_reaction(target, event)
        show_damage_number(target, event)

func handle_damage_taken(event: Dictionary):
    var target = get_entity(event.target_id)
    if target:
        play_hit_reaction(target, event)
        show_damage_number(target, event)
        update_health_bar(target, event)

func handle_death(event: Dictionary):
    var entity = get_entity(event.target_id)
    if entity:
        play_death_animation(entity, event)

func get_entity(entity_id: String) -> Node3D:
    # Find entity in scene by ID
    return get_node_or_null("/root/World/Entities/" + entity_id)
```

### 2. Entity Animation Controller

```gdscript
# entity.gd
extends CharacterBody3D

@onready var animation_tree: AnimationTree = $AnimationTree
@onready var animation_state: AnimationNodeStateMachinePlayback = animation_tree.get("parameters/playback")

enum AnimationState {
    IDLE,
    WALKING,
    RUNNING,
    ATTACKING,
    CASTING,
    HIT_REACTION,
    DODGING,
    BLOCKING,
    DYING,
    DEAD
}

var current_anim_state: AnimationState = AnimationState.IDLE

func play_attack_animation(event: Dictionary):
    # Choose attack animation based on ability/weapon
    var anim_name = "attack_melee"

    # Use animation hint if provided
    if event.has("animation_hint"):
        match event.animation_hint:
            "swing_right":
                anim_name = "attack_swing_right"
            "swing_left":
                anim_name = "attack_swing_left"
            "thrust":
                anim_name = "attack_thrust"
            "overhead_slam":
                anim_name = "attack_overhead"

    # Play animation
    animation_state.travel(anim_name)
    current_anim_state = AnimationState.ATTACKING

    # Calculate attack direction
    var target = get_entity(event.target_id)
    if target:
        look_at(target.global_position)

    # Wait for hit frame, then trigger hit VFX
    var hit_frame_time = 0.3 # 30% through animation
    await get_tree().create_timer(hit_frame_time).timeout
    trigger_hit_effect(event)

func play_hit_reaction(event: Dictionary):
    # Don't interrupt death animation
    if current_anim_state == AnimationState.DYING or current_anim_state == AnimationState.DEAD:
        return

    # Choose hit reaction based on damage type
    var hit_anim = "hit_reaction_light"

    if event.damage > 50:
        hit_anim = "hit_reaction_heavy"

    if event.is_critical:
        hit_anim = "hit_reaction_critical"

    animation_state.travel(hit_anim)
    current_anim_state = AnimationState.HIT_REACTION

    # Flash red material
    flash_damage_material()

    # Return to idle after hit reaction
    await animation_tree.animation_finished
    if current_anim_state == AnimationState.HIT_REACTION:
        animation_state.travel("idle")
        current_anim_state = AnimationState.IDLE

func play_death_animation(event: Dictionary):
    current_anim_state = AnimationState.DYING
    animation_state.travel("death")

    # Disable collision
    $CollisionShape3D.disabled = true

    # Wait for death animation to finish
    await animation_tree.animation_finished
    current_anim_state = AnimationState.DEAD

    # Optional: fade out and remove
    var tween = create_tween()
    tween.tween_property(self, "modulate:a", 0.0, 2.0)
    await tween.finished
    queue_free()

func flash_damage_material():
    # Flash red when taking damage
    var mesh = $MeshInstance3D
    var material = mesh.get_surface_override_material(0)

    if material:
        var original_color = material.albedo_color
        material.albedo_color = Color.RED

        await get_tree().create_timer(0.1).timeout
        material.albedo_color = original_color
```

### 3. Animation State Machine Setup (Godot AnimationTree)

```
                  ┌─────────┐
                  │  IDLE   │◄────────┐
                  └────┬────┘         │
                       │              │
           ┌───────────┼───────────┐  │
           │           │           │  │
           ▼           ▼           ▼  │
    ┌──────────┐ ┌─────────┐ ┌────────┴─┐
    │ WALKING  │ │ RUNNING │ │ ATTACKING│
    └──────────┘ └─────────┘ └────────┬─┘
                                      │
                  ┌───────────────────┼──────────┐
                  │                   │          │
                  ▼                   ▼          ▼
           ┌────────────┐      ┌──────────┐ ┌─────────┐
           │HIT_REACTION│      │ CASTING  │ │ DODGING │
           └────────────┘      └──────────┘ └─────────┘
                  │
                  │ (health <= 0)
                  ▼
           ┌────────────┐
           │   DYING    │
           └──────┬─────┘
                  │
                  ▼
           ┌────────────┐
           │    DEAD    │
           └────────────┘
```

### 4. Visual Effects (VFX) System

```gdscript
# vfx_manager.gd
extends Node

var hit_effect_scene = preload("res://effects/hit_effect.tscn")
var crit_effect_scene = preload("res://effects/critical_hit.tscn")
var blood_effect_scene = preload("res://effects/blood_splash.tscn")
var fire_effect_scene = preload("res://effects/fire_damage.tscn")

func trigger_hit_effect(position: Vector3, event: Dictionary):
    var effect: GPUParticles3D

    # Choose effect based on damage type
    match event.damage_type:
        "fire":
            effect = fire_effect_scene.instantiate()
        "ice":
            effect = ice_effect_scene.instantiate()
        "poison":
            effect = poison_effect_scene.instantiate()
        _:
            effect = hit_effect_scene.instantiate()

    # Critical hits get special effect
    if event.is_critical:
        var crit_effect = crit_effect_scene.instantiate()
        get_tree().root.add_child(crit_effect)
        crit_effect.global_position = position + Vector3(0, 2, 0)

    # Spawn effect
    get_tree().root.add_child(effect)
    effect.global_position = position
    effect.emitting = true

    # Auto-cleanup after particles finish
    await get_tree().create_timer(2.0).timeout
    effect.queue_free()

func show_damage_number(target: Node3D, event: Dictionary):
    var damage_label = Label3D.new()
    damage_label.text = str(event.damage)
    damage_label.font_size = 48
    damage_label.outline_size = 8

    # Color based on damage type
    if event.is_critical:
        damage_label.modulate = Color.ORANGE_RED
        damage_label.text = str(event.damage) + "!"
    elif event.damage_type == "poison":
        damage_label.modulate = Color.GREEN
    elif event.damage_type == "fire":
        damage_label.modulate = Color.ORANGE
    else:
        damage_label.modulate = Color.WHITE

    # Position above target
    get_tree().root.add_child(damage_label)
    damage_label.global_position = target.global_position + Vector3(0, 2.5, 0)

    # Animate floating up and fading
    var tween = create_tween()
    tween.set_parallel(true)
    tween.tween_property(damage_label, "position:y", damage_label.position.y + 2, 1.0)
    tween.tween_property(damage_label, "modulate:a", 0.0, 1.0)

    await tween.finished
    damage_label.queue_free()
```

### 5. Monster Animation with AI States

```gdscript
# monster.gd
extends CharacterBody3D

var ai_state = AIState.IDLE

enum AIState {
    IDLE,
    PATROLLING,
    CHASING,
    ATTACKING,
    RETURNING,
    FLEEING
}

func _process(delta):
    # Sync animations with AI state
    match ai_state:
        AIState.IDLE:
            animation_state.travel("idle")
        AIState.PATROLLING:
            animation_state.travel("walk")
        AIState.CHASING:
            animation_state.travel("run")
        AIState.ATTACKING:
            # Don't override attack animation
            pass
        AIState.RETURNING:
            animation_state.travel("walk")
        AIState.FLEEING:
            animation_state.travel("run")

func on_server_ai_state_update(new_state: String):
    # Server sends AI state updates
    match new_state:
        "idle":
            ai_state = AIState.IDLE
        "chasing":
            ai_state = AIState.CHASING
        "attacking":
            ai_state = AIState.ATTACKING
        "returning":
            ai_state = AIState.RETURNING
```

### 6. Environmental Hazard Animations

```gdscript
# environmental_hazard.gd
extends Node3D

@onready var particles: GPUParticles3D = $Particles
@onready var area: Area3D = $DamageArea

func _ready():
    # Listen for damage events
    CombatEventManager.combat_event_received.connect(_on_combat_event)

func _on_combat_event(event: Dictionary):
    if event.event_type != "environmental_damage":
        return

    # Check if this hazard caused the damage
    if event.caster_id != self.name:
        return

    # Play damage effect on affected entity
    var target = get_entity(event.target_id)
    if target:
        match self.hazard_type:
            "lava":
                apply_burning_effect(target)
            "poison":
                apply_poison_effect(target)
            "lightning":
                apply_shock_effect(target)

func apply_burning_effect(target: Node3D):
    # Spawn fire particles on target
    var fire = fire_effect.instantiate()
    target.add_child(fire)
    fire.emitting = true

    # Remove after duration
    await get_tree().create_timer(2.0).timeout
    fire.queue_free()
```

## Network Protocol

### WebSocket Message Format

```json
{
  "type": "combat_event",
  "data": {
    "event_id": "evt_12345",
    "timestamp": "2025-11-18T10:30:00Z",
    "event_type": "melee_attack",
    "caster_id": "player1",
    "target_id": "goblin1",
    "ability_id": "basic_attack",
    "damage": 25,
    "is_critical": true,
    "damage_type": "physical",
    "target_health": 75,
    "target_max_health": 100,
    "target_died": false,
    "animation_hint": "swing_right",
    "position": {"x": 10.5, "y": 0, "z": 5.2},
    "direction": {"x": 1, "y": 0, "z": 0}
  }
}
```

## Performance Optimizations

### 1. Animation Pooling

```gdscript
# Effect pooling to avoid frequent instantiation
var effect_pool = {}

func get_pooled_effect(effect_type: String) -> Node3D:
    if effect_type not in effect_pool:
        effect_pool[effect_type] = []

    # Reuse inactive effect
    for effect in effect_pool[effect_type]:
        if not effect.emitting:
            return effect

    # Create new if none available
    var new_effect = load("res://effects/" + effect_type + ".tscn").instantiate()
    effect_pool[effect_type].append(new_effect)
    return new_effect
```

### 2. LOD for Distant Entities

```gdscript
func update_animation_lod(camera_distance: float):
    if camera_distance > 50:
        # Very far: disable animations
        animation_tree.active = false
    elif camera_distance > 20:
        # Far: reduce animation update rate
        Engine.set_physics_ticks_per_second(30)
    else:
        # Close: full animation quality
        animation_tree.active = true
        Engine.set_physics_ticks_per_second(60)
```

### 3. Event Batching

Server can batch multiple combat events:

```go
type CombatEventBatch struct {
    Events    []CombatEvent `json:"events"`
    Timestamp time.Time     `json:"timestamp"`
}

// Send every 50ms instead of per-event
func (cm *CombatManager) FlushEventBatch() {
    if len(cm.pendingEvents) == 0 {
        return
    }

    batch := CombatEventBatch{
        Events:    cm.pendingEvents,
        Timestamp: time.Now(),
    }

    cm.BroadcastBatch(batch)
    cm.pendingEvents = []CombatEvent{}
}
```

## Summary

### Animation Flow

```
Server Combat               Network                Client Animation
─────────────               ───────                ────────────────
ExecuteAbility()     →                      →     play_attack_animation()
  └─ Calculate damage                              └─ AnimationTree.travel("attack")
  └─ Apply damage                                  └─ Wait for hit frame
  └─ Create CombatEvent  →  WebSocket       →     trigger_hit_effect()
  └─ Broadcast event     →  (JSON)          →     show_damage_number()
                                            →     play_hit_reaction()
                                            →     update_health_bar()
```

### Key Principles

1. **Server Authority**: Server decides what happens
2. **Client Representation**: Client shows what happened
3. **Event-Driven**: Combat events drive animations
4. **Predictive**: Can show local animations immediately, correct on server response
5. **Optimized**: Interest management, LOD, pooling
6. **Extensible**: Easy to add new event types and animations

This architecture keeps game logic secure on the server while providing responsive, visually appealing combat on the client!
