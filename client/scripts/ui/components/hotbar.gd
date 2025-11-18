extends UIComponent
class_name Hotbar

## Hotbar component for abilities and items
##
## Features:
## - Configurable number of slots
## - Cooldown overlay
## - Keybind display
## - Visual feedback on use

## Configuration
@export var num_slots: int = 6
@export var slot_size: float = 50.0
@export var slot_spacing: float = 5.0
@export var show_keybinds: bool = true

## Colors
@export var slot_color: Color = Color(0.2, 0.2, 0.2, 0.8)
@export var slot_border_color: Color = Color(0.6, 0.6, 0.6, 1.0)
@export var cooldown_color: Color = Color(0.0, 0.0, 0.0, 0.7)
@export var keybind_color: Color = Color(1.0, 1.0, 1.0, 0.9)

## Slots
class HotbarSlot:
	var index: int
	var rect: Rect2
	var keybind: String
	var cooldown_percent: float = 0.0
	var is_active: bool = false

var slots: Array[HotbarSlot] = []

func _setup_component():
	component_name = "Hotbar"

	var total_width = (num_slots * slot_size) + ((num_slots - 1) * slot_spacing)
	custom_minimum_size = Vector2(total_width, slot_size)

	# Create slots
	for i in range(num_slots):
		var slot = HotbarSlot.new()
		slot.index = i
		slot.rect = Rect2(
			i * (slot_size + slot_spacing),
			0,
			slot_size,
			slot_size
		)
		slot.keybind = str(i + 1) if i < 9 else "0"
		slots.append(slot)

func _connect_signals():
	# Connect to weapon cooldown signals
	if weapon:
		weapon.cooldown_started.connect(_on_cooldown_started)
		weapon.cooldown_finished.connect(_on_cooldown_finished)
		weapon.attack_started.connect(_on_attack_started)

func _process(_delta):
	# Update weapon cooldown on slot 1
	if weapon and slots.size() > 0:
		slots[0].cooldown_percent = weapon.get_cooldown_progress()

	queue_redraw()

func _on_cooldown_started(_duration: float):
	if slots.size() > 0:
		slots[0].is_active = true

func _on_cooldown_finished():
	if slots.size() > 0:
		slots[0].is_active = false

func _on_attack_started(_target: Node):
	# Visual feedback on slot 1
	pass

func _draw():
	for slot in slots:
		_draw_slot(slot)

func _draw_slot(slot: HotbarSlot):
	var rect = slot.rect

	# Draw slot background
	draw_rect(rect, slot_color, true)

	# Draw slot border
	var border_width = 2.0 if slot.is_active else 1.0
	draw_rect(rect, slot_border_color, false, border_width)

	# Draw cooldown overlay
	if slot.cooldown_percent > 0.0:
		var cooldown_height = rect.size.y * slot.cooldown_percent
		var cooldown_rect = Rect2(
			rect.position.x,
			rect.position.y + (rect.size.y - cooldown_height),
			rect.size.x,
			cooldown_height
		)
		draw_rect(cooldown_rect, cooldown_color, true)

		# Draw cooldown text
		var cd_text = "%.1f" % (slot.cooldown_percent * (weapon.cooldown_time if weapon else 1.0))
		var font = get_theme_default_font()
		var font_size = 16
		var text_size = font.get_string_size(cd_text, HORIZONTAL_ALIGNMENT_CENTER, -1, font_size)
		var text_pos = rect.position + (rect.size / 2.0) - (text_size / 2.0)
		draw_string(font, text_pos, cd_text, HORIZONTAL_ALIGNMENT_CENTER, -1, font_size, Color.WHITE)

	# Draw keybind
	if show_keybinds and slot.cooldown_percent == 0.0:
		var font = get_theme_default_font()
		var font_size = 14
		var text_pos = rect.position + Vector2(5, rect.size.y - 5)
		draw_string(font, text_pos, slot.keybind, HORIZONTAL_ALIGNMENT_LEFT, -1, font_size, keybind_color)

	# Draw weapon icon placeholder for slot 1
	if slot.index == 0:
		var icon_text = "⚔"  # Sword emoji
		var font = get_theme_default_font()
		var font_size = 24
		var text_size = font.get_string_size(icon_text, HORIZONTAL_ALIGNMENT_CENTER, -1, font_size)
		var text_pos = rect.position + (rect.size / 2.0) - (text_size / 2.0)

		if slot.cooldown_percent == 0.0:
			draw_string(font, text_pos, icon_text, HORIZONTAL_ALIGNMENT_CENTER, -1, font_size, Color.WHITE)
