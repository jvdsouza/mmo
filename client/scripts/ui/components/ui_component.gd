extends Control
class_name UIComponent

## Base class for all UI components
##
## UI components are reusable pieces that can be composed into themes.
## Examples: HealthBar, TargetFrame, Hotbar, Minimap, etc.

## Emitted when component needs to be shown/hidden
signal visibility_changed(visible: bool)

## Emitted when component requests position change
signal position_requested(new_position: Vector2)

## Reference to the player (set by theme)
var player: Node3D = null

## Reference to targeting system (set by theme)
var targeting_system: Node = null

## Reference to weapon (set by theme)
var weapon: Node = null

## Component configuration
@export var component_name: String = "UIComponent"
@export var can_be_toggled: bool = true
@export var default_visible: bool = true

func _ready():
	visible = default_visible
	_setup_component()

## Override in subclasses to setup the component
func _setup_component() -> void:
	pass

## Initialize component with game systems
func initialize(player_node: Node3D, targeting_node: Node = null, weapon_node: Node = null) -> void:
	player = player_node
	targeting_system = targeting_node
	weapon = weapon_node

	_connect_signals()
	_update_component()

## Override to connect to relevant signals
func _connect_signals() -> void:
	pass

## Override to update component visuals
func _update_component() -> void:
	pass

## Toggle component visibility
func toggle() -> void:
	if can_be_toggled:
		visible = not visible
		visibility_changed.emit(visible)

## Show component
func show_component() -> void:
	visible = true
	visibility_changed.emit(true)

## Hide component
func hide_component() -> void:
	visible = false
	visibility_changed.emit(false)

## Cleanup when component is removed
func cleanup() -> void:
	# Disconnect all signals
	for signal_connection in get_incoming_connections():
		signal_connection.signal.disconnect(signal_connection.callable)
