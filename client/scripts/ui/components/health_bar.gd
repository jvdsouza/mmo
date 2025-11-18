extends UIComponent
class_name HealthBar

## Health bar component showing current/max health
##
## Features:
## - Animated health changes
## - Color transitions (green -> yellow -> red)
## - Damage flash effects
## - Optional text display

## Visual configuration
@export var bar_width: float = 200.0
@export var bar_height: float = 20.0
@export var show_text: bool = true
@export var animate_changes: bool = true
@export var animation_speed: float = 0.3

## Colors
@export var color_full: Color = Color(0.2, 1.0, 0.2)  # Green
@export var color_half: Color = Color(1.0, 1.0, 0.2)  # Yellow
@export var color_low: Color = Color(1.0, 0.2, 0.2)   # Red
@export var background_color: Color = Color(0.2, 0.2, 0.2, 0.8)
@export var border_color: Color = Color(1.0, 1.0, 1.0, 0.3)

## State
var current_health: int = 100
var max_health: int = 100
var display_health: float = 100.0  # Animated value
var last_damage_time: float = 0.0

## UI elements
var label: Label = null

func _setup_component():
	component_name = "HealthBar"
	custom_minimum_size = Vector2(bar_width, bar_height)

	# Create label for health text
	if show_text:
		label = Label.new()
		label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
		label.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
		label.add_theme_font_size_override("font_size", 14)
		add_child(label)

func _connect_signals():
	# In a full implementation, player would emit health_changed signal
	# For now, we'll poll in _process()
	pass

func _process(delta):
	if not player:
		return

	# Get health from player (polling for now, signals would be better)
	var new_health = player.get_current_health() if player.has_method("get_current_health") else 100
	var new_max = player.get_max_health() if player.has_method("get_max_health") else 100

	if new_health != current_health or new_max != max_health:
		_on_health_changed(new_health, new_max)

	# Animate health bar
	if animate_changes:
		display_health = lerp(display_health, float(current_health), animation_speed)
	else:
		display_health = float(current_health)

	queue_redraw()

func _on_health_changed(new_health: int, new_max: int):
	# Detect damage
	if new_health < current_health:
		last_damage_time = Time.get_ticks_msec() / 1000.0

	current_health = new_health
	max_health = new_max

	# Update label
	if label:
		label.text = "%d / %d" % [current_health, max_health]

func _draw():
	var health_percent = display_health / max_health if max_health > 0 else 0.0
	var fill_width = bar_width * health_percent

	# Draw background
	draw_rect(Rect2(0, 0, bar_width, bar_height), background_color, true)

	# Draw health fill with color based on percentage
	var fill_color = _get_health_color(health_percent)

	# Flash red when taking damage
	var time_since_damage = (Time.get_ticks_msec() / 1000.0) - last_damage_time
	if time_since_damage < 0.2:
		fill_color = fill_color.lerp(Color.RED, 1.0 - (time_since_damage / 0.2))

	draw_rect(Rect2(0, 0, fill_width, bar_height), fill_color, true)

	# Draw border
	draw_rect(Rect2(0, 0, bar_width, bar_height), border_color, false, 2.0)

func _get_health_color(percent: float) -> Color:
	if percent > 0.5:
		# Green to yellow (100% -> 50%)
		var t = (1.0 - percent) * 2.0  # 0.0 to 1.0
		return color_full.lerp(color_half, t)
	else:
		# Yellow to red (50% -> 0%)
		var t = (0.5 - percent) * 2.0  # 0.0 to 1.0
		return color_half.lerp(color_low, t)
