extends UITheme
class_name MMOTheme

## Traditional MMO-style UI theme
##
## Features:
## - Large health bar in top-left
## - Target frame in top-right
## - Multiple hotbars at bottom
## - More detailed information display
## - Classic MMO layout

func _setup_theme():
	theme_name = "MMO Style"
	theme_description = "Traditional MMO UI with detailed information"

	# Create health bar (top-left)
	var health_bar = HealthBar.new()
	health_bar.bar_width = 250
	health_bar.bar_height = 25
	health_bar.show_text = true
	health_bar.position = Vector2(20, 20)
	add_component(health_bar, "health_bar")

	# Create target frame (top-right)
	var target_frame = TargetFrame.new()
	target_frame.frame_width = 280
	target_frame.frame_height = 90
	target_frame.show_distance = true
	target_frame.show_health_text = true
	target_frame.auto_hide_no_target = true

	var viewport_width = get_viewport_rect().size.x
	target_frame.position = Vector2(viewport_width - target_frame.frame_width - 20, 20)

	add_component(target_frame, "target_frame")

	# Create main hotbar (bottom-center)
	var hotbar = Hotbar.new()
	hotbar.num_slots = 6
	hotbar.slot_size = 50
	hotbar.slot_spacing = 5
	hotbar.show_keybinds = true

	var hotbar_width = (hotbar.num_slots * hotbar.slot_size) + ((hotbar.num_slots - 1) * hotbar.slot_spacing)
	hotbar.position = Vector2((viewport_width - hotbar_width) / 2, get_viewport_rect().size.y - 80)

	add_component(hotbar, "hotbar_main")

	# Create secondary hotbar (bottom-center, above main)
	var hotbar2 = Hotbar.new()
	hotbar2.num_slots = 6
	hotbar2.slot_size = 45
	hotbar2.slot_spacing = 4
	hotbar2.show_keybinds = false

	var hotbar2_width = (hotbar2.num_slots * hotbar2.slot_size) + ((hotbar2.num_slots - 1) * hotbar2.slot_spacing)
	hotbar2.position = Vector2((viewport_width - hotbar2_width) / 2, get_viewport_rect().size.y - 140)

	add_component(hotbar2, "hotbar_secondary")

func _ready():
	super._ready()
	print("[MMOTheme] MMO-style UI theme loaded")
