extends UIComponent
class_name TargetFrame

## Target frame component showing info about current target
##
## Features:
## - Target name display
## - Target health bar
## - Level/type indicator
## - Distance indicator
## - Auto-hide when no target

## Visual configuration
@export var frame_width: float = 250.0
@export var frame_height: float = 80.0
@export var show_distance: bool = true
@export var show_health_text: bool = true
@export var auto_hide_no_target: bool = true

## Colors
@export var enemy_color: Color = Color(1.0, 0.2, 0.2)
@export var neutral_color: Color = Color(1.0, 1.0, 0.2)
@export var friendly_color: Color = Color(0.2, 1.0, 0.2)

## State
var current_target: Node = null
var target_health: int = 100
var target_max_health: int = 100
var target_name: String = ""
var target_distance: float = 0.0

## UI elements
var name_label: Label = null
var health_label: Label = null
var distance_label: Label = null
var health_bar_bg: ColorRect = null
var health_bar_fill: ColorRect = null

func _setup_component():
	component_name = "TargetFrame"
	custom_minimum_size = Vector2(frame_width, frame_height)

	# Create background panel
	var panel = Panel.new()
	panel.set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	add_child(panel)

	# Create name label
	name_label = Label.new()
	name_label.position = Vector2(10, 5)
	name_label.size = Vector2(frame_width - 20, 25)
	name_label.add_theme_font_size_override("font_size", 16)
	add_child(name_label)

	# Create health bar background
	health_bar_bg = ColorRect.new()
	health_bar_bg.position = Vector2(10, 35)
	health_bar_bg.size = Vector2(frame_width - 20, 20)
	health_bar_bg.color = Color(0.2, 0.2, 0.2, 0.8)
	add_child(health_bar_bg)

	# Create health bar fill
	health_bar_fill = ColorRect.new()
	health_bar_fill.position = Vector2(10, 35)
	health_bar_fill.size = Vector2(frame_width - 20, 20)
	health_bar_fill.color = enemy_color
	add_child(health_bar_fill)

	# Create health text label
	if show_health_text:
		health_label = Label.new()
		health_label.position = Vector2(10, 35)
		health_label.size = Vector2(frame_width - 20, 20)
		health_label.horizontal_alignment = HORIZONTAL_ALIGNMENT_CENTER
		health_label.vertical_alignment = VERTICAL_ALIGNMENT_CENTER
		health_label.add_theme_font_size_override("font_size", 12)
		add_child(health_label)

	# Create distance label
	if show_distance:
		distance_label = Label.new()
		distance_label.position = Vector2(10, 60)
		distance_label.size = Vector2(frame_width - 20, 15)
		distance_label.add_theme_font_size_override("font_size", 11)
		add_child(distance_label)

	# Start hidden if no target
	if auto_hide_no_target:
		visible = false

func _connect_signals():
	if targeting_system:
		targeting_system.target_changed.connect(_on_target_changed)
		targeting_system.target_lost.connect(_on_target_lost)

func _process(_delta):
	if not current_target or not targeting_system:
		if auto_hide_no_target and visible:
			visible = false
		return

	# Update target info
	if current_target.has_method("get_current_health"):
		target_health = current_target.get_current_health()

	if current_target.has_method("get_max_health"):
		target_max_health = current_target.get_max_health()

	# Update distance
	if player and current_target.has_method("get_target_position"):
		target_distance = player.global_position.distance_to(current_target.get_target_position())

	_update_visuals()

func _on_target_changed(old_target: Node, new_target: Node):
	current_target = new_target

	if new_target:
		# Get target info
		if new_target.has_method("get_display_name"):
			target_name = new_target.get_display_name()
		else:
			target_name = "Unknown"

		if new_target.has_method("get_current_health"):
			target_health = new_target.get_current_health()

		if new_target.has_method("get_max_health"):
			target_max_health = new_target.get_max_health()

		# Show frame
		if auto_hide_no_target:
			visible = true

		_update_visuals()
	else:
		_on_target_lost(old_target)

func _on_target_lost(_target: Node):
	current_target = null
	if auto_hide_no_target:
		visible = false

func _update_visuals():
	if not current_target:
		return

	# Update name
	if name_label:
		name_label.text = target_name

		# Color name based on faction
		if current_target.has_method("get_faction"):
			var faction = current_target.get_faction()
			match faction:
				"monster", "enemy":
					name_label.add_theme_color_override("font_color", enemy_color)
				"player", "friendly":
					name_label.add_theme_color_override("font_color", friendly_color)
				_:
					name_label.add_theme_color_override("font_color", neutral_color)

	# Update health bar
	var health_percent = float(target_health) / float(target_max_health) if target_max_health > 0 else 0.0
	if health_bar_fill:
		health_bar_fill.size.x = (frame_width - 20) * health_percent

		# Color health bar based on faction
		if current_target.has_method("get_faction"):
			var faction = current_target.get_faction()
			match faction:
				"monster", "enemy":
					health_bar_fill.color = enemy_color
				"player", "friendly":
					health_bar_fill.color = friendly_color
				_:
					health_bar_fill.color = neutral_color

	# Update health text
	if health_label:
		health_label.text = "%d / %d" % [target_health, target_max_health]

	# Update distance
	if distance_label:
		distance_label.text = "Distance: %.1fm" % target_distance
