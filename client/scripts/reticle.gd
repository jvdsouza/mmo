extends Control

# Reticle appearance
@export var reticle_size: float = 8.0
@export var reticle_thickness: float = 2.0
@export var reticle_gap: float = 4.0

# Colors
@export var default_color: Color = Color(1, 1, 1, 0.8)  # White
@export var target_color: Color = Color(1, 0.2, 0.2, 1.0)  # Red
@export var friendly_color: Color = Color(0.2, 1, 0.2, 1.0)  # Green

var current_color: Color = default_color
var has_target: bool = false

# References to combat systems (set externally)
var player: Node3D = null
var targeting_system: Node = null  # RaycastTargeting instance

func _ready():
	# Make sure we're centered and cover full screen
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	mouse_filter = Control.MOUSE_FILTER_IGNORE  # Don't block mouse events

func _process(_delta):
	# Update color based on targeting system
	if targeting_system and targeting_system.has_valid_target():
		var target = targeting_system.get_current_target()

		if target and player:
			# Get weapon from player to check attack range
			var weapon = player.get("weapon")
			var attack_range = weapon.attack_range if weapon else 3.0

			# Check if in range
			var distance = player.global_position.distance_to(target.get_target_position())

			if distance <= attack_range:
				current_color = target_color  # In range, can attack
				has_target = true
			else:
				current_color = friendly_color.lerp(default_color, 0.5)  # Out of range
				has_target = true
		else:
			current_color = default_color
			has_target = false
	else:
		current_color = default_color
		has_target = false

	queue_redraw()  # Trigger _draw()

func _draw():
	var center = size / 2.0

	# Draw crosshair
	draw_crosshair(center, current_color)

	# Optional: Draw hit marker effect
	if has_target:
		draw_circle_outline(center, reticle_size + reticle_gap + 4.0, current_color, 1.0)

func draw_crosshair(center: Vector2, color: Color):
	"""Draw a simple crosshair"""
	var gap = reticle_gap
	var length = reticle_size

	# Top line
	draw_line(
		Vector2(center.x, center.y - gap - length),
		Vector2(center.x, center.y - gap),
		color,
		reticle_thickness
	)

	# Bottom line
	draw_line(
		Vector2(center.x, center.y + gap),
		Vector2(center.x, center.y + gap + length),
		color,
		reticle_thickness
	)

	# Left line
	draw_line(
		Vector2(center.x - gap - length, center.y),
		Vector2(center.x - gap, center.y),
		color,
		reticle_thickness
	)

	# Right line
	draw_line(
		Vector2(center.x + gap, center.y),
		Vector2(center.x + gap + length, center.y),
		color,
		reticle_thickness
	)

	# Center dot
	draw_circle(center, 1.5, color)

func draw_circle_outline(center: Vector2, radius: float, color: Color, thickness: float):
	"""Draw a circle outline"""
	var points = 32
	var angle_step = TAU / points

	for i in range(points):
		var angle1 = i * angle_step
		var angle2 = (i + 1) * angle_step

		var point1 = center + Vector2(cos(angle1), sin(angle1)) * radius
		var point2 = center + Vector2(cos(angle2), sin(angle2)) * radius

		draw_line(point1, point2, color, thickness)

# Alternative: Simple dot reticle
func draw_dot_reticle(center: Vector2, color: Color):
	"""Alternative reticle style - just a dot"""
	draw_circle(center, 3.0, color)
	draw_circle(center, 1.5, Color(0, 0, 0, 0.8))  # Black center

# Alternative: Circle reticle
func draw_circle_reticle(center: Vector2, color: Color):
	"""Alternative reticle style - circle"""
	draw_circle_outline(center, 10.0, color, 2.0)
	draw_circle(center, 1.5, color)
