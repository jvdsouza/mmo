extends Control
class_name UITheme

## Base class for all UI themes
##
## Themes are different layouts/styles that compose UIComponents.
## Examples: MinimalTheme, MMOTheme, ModernTheme, etc.
##
## Themes can be swapped at runtime for different player preferences.

## Emitted when theme is fully loaded
signal theme_loaded()

## Emitted when theme is being unloaded
signal theme_unloading()

## Theme configuration
@export var theme_name: String = "Default Theme"
@export var theme_description: String = ""
@export var supports_scaling: bool = true

## References to game systems
var player: Node3D = null
var targeting_system: Node = null
var weapon: Node = null

## Component storage
var components: Dictionary = {}  # name -> UIComponent

func _ready():
	# Themes should cover the full screen
	set_anchors_and_offsets_preset(Control.PRESET_FULL_RECT)
	mouse_filter = Control.MOUSE_FILTER_IGNORE  # Don't block game input

## Initialize theme with game systems
func initialize(player_node: Node3D, targeting_node: Node = null, weapon_node: Node = null) -> void:
	player = player_node
	targeting_system = targeting_node
	weapon = weapon_node

	_setup_theme()
	_initialize_components()

	theme_loaded.emit()

## Override in subclasses to create and position components
func _setup_theme() -> void:
	push_error("UITheme._setup_theme() must be overridden")

## Initialize all components with game systems
func _initialize_components() -> void:
	for component in components.values():
		if component is UIComponent:
			component.initialize(player, targeting_system, weapon)

## Add a component to the theme
func add_component(component: UIComponent, component_name: String) -> void:
	components[component_name] = component
	add_child(component)

## Get a component by name
func get_component(component_name: String) -> UIComponent:
	return components.get(component_name, null)

## Remove a component
func remove_component(component_name: String) -> void:
	var component = components.get(component_name)
	if component:
		component.cleanup()
		component.queue_free()
		components.erase(component_name)

## Toggle component visibility
func toggle_component(component_name: String) -> void:
	var component = get_component(component_name)
	if component:
		component.toggle()

## Cleanup when theme is switched
func cleanup() -> void:
	theme_unloading.emit()

	# Cleanup all components
	for component in components.values():
		if component is UIComponent:
			component.cleanup()

	components.clear()

## Get all component names
func get_component_names() -> Array:
	return components.keys()
