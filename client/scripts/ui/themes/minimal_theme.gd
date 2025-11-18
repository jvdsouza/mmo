extends UITheme
class_name MinimalTheme

## Minimal UI theme - clean and unobtrusive
##
## Features:
## - Small health bar in bottom-left
## - Compact hotbar in bottom-center
## - Target frame in top-center (only when targeting)
## - Minimal visual clutter

func _setup_theme():
	theme_name = "Minimal"
	theme_description = "Clean, minimal UI with essential information only"

	# Create health bar (bottom-left)
	var health_bar = HealthBar.new()
	health_bar.bar_width = 180
	health_bar.bar_height = 18
	health_bar.show_text = true
	health_bar.position = Vector2(20, get_viewport_rect().size.y - 40)
	add_component(health_bar, "health_bar")

	# Create hotbar (bottom-center)
	var hotbar = Hotbar.new()
	hotbar.num_slots = 6
	hotbar.slot_size = 45
	hotbar.slot_spacing = 4
	hotbar.show_keybinds = true

	var hotbar_width = (hotbar.num_slots * hotbar.slot_size) + ((hotbar.num_slots - 1) * hotbar.slot_spacing)
	var viewport_width = get_viewport_rect().size.x
	hotbar.position = Vector2((viewport_width - hotbar_width) / 2, get_viewport_rect().size.y - 70)

	add_component(hotbar, "hotbar")

	# Create target frame (top-center)
	var target_frame = TargetFrame.new()
	target_frame.frame_width = 220
	target_frame.frame_height = 75
	target_frame.show_distance = true
	target_frame.show_health_text = true
	target_frame.auto_hide_no_target = true

	var target_x = (get_viewport_rect().size.x - target_frame.frame_width) / 2
	target_frame.position = Vector2(target_x, 20)

	add_component(target_frame, "target_frame")

func _ready():
	super._ready()
	print("[MinimalTheme] Minimal UI theme loaded")
