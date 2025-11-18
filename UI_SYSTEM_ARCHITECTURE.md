# UI System Architecture

## Overview

The UI system is a **fully extensible, theme-based architecture** that allows players to swap between different UI layouts and styles at runtime. It follows the same design principles as the combat system: composition, signals, and the Strategy pattern.

## Architecture Pattern

```
┌─────────────┐
│   Player    │
│             │
│ ui_theme ───┼──> UITheme (Strategy)
│             │         │
│             │         ├─> MinimalTheme
│             │         ├─> MMOTheme
│             │         └─> CustomTheme
│             │
│             │    Each theme composes:
│             │         │
│             │         ├─> HealthBar (UIComponent)
│             │         ├─> TargetFrame (UIComponent)
│             │         ├─> Hotbar (UIComponent)
│             │         └─> ... more components
└─────────────┘
```

## Core Classes

### 1. UIComponent (Base Class)

**Location:** `client/scripts/ui/components/ui_component.gd`

**Purpose:** Base class for all reusable UI pieces (health bars, hotbars, minimaps, etc.)

**Key Features:**
- Automatically receives references to player, targeting_system, weapon
- Signal-based updates
- Can be toggled on/off
- Cleanup method for proper disposal

**Example:**
```gdscript
extends UIComponent
class_name HealthBar

func _connect_signals():
    # Subscribe to player health changes
    if player:
        player.health_changed.connect(_on_health_changed)

func _on_health_changed(current: int, max: int):
    # Update health bar visuals
    update_bar(current, max)
```

### 2. UITheme (Base Class)

**Location:** `client/scripts/ui/ui_theme.gd`

**Purpose:** Base class for complete UI layouts that compose UIComponents

**Key Features:**
- Composes multiple UIComponents into a layout
- Manages component lifecycle
- Can be swapped at runtime
- Each theme creates its own unique layout

**Example:**
```gdscript
extends UITheme
class_name MinimalTheme

func _setup_theme():
    theme_name = "Minimal"

    # Create health bar in bottom-left
    var health_bar = HealthBar.new()
    health_bar.position = Vector2(20, viewport.y - 40)
    add_component(health_bar, "health_bar")

    # Create hotbar in bottom-center
    var hotbar = Hotbar.new()
    hotbar.position = Vector2(center.x, viewport.y - 70)
    add_component(hotbar, "hotbar")
```

## Built-in Components

### HealthBar

**Location:** `client/scripts/ui/components/health_bar.gd`

**Features:**
- Animated health changes
- Color transitions (green → yellow → red)
- Damage flash effects
- Optional health text display

**Configuration:**
```gdscript
var health_bar = HealthBar.new()
health_bar.bar_width = 200
health_bar.bar_height = 20
health_bar.show_text = true
health_bar.animate_changes = true
```

### TargetFrame

**Location:** `client/scripts/ui/components/target_frame.gd`

**Features:**
- Shows target name, health, distance
- Auto-hides when no target
- Color-coded by faction (enemy/neutral/friendly)
- Subscribes to targeting_system.target_changed signal

**Configuration:**
```gdscript
var target_frame = TargetFrame.new()
target_frame.show_distance = true
target_frame.auto_hide_no_target = true
```

### Hotbar

**Location:** `client/scripts/ui/components/hotbar.gd`

**Features:**
- Configurable number of slots
- Cooldown overlay on slot 1 (weapon)
- Keybind display (1-6, etc.)
- Subscribes to weapon cooldown signals

**Configuration:**
```gdscript
var hotbar = Hotbar.new()
hotbar.num_slots = 6
hotbar.slot_size = 50
hotbar.show_keybinds = true
```

## Built-in Themes

### MinimalTheme

**Location:** `client/scripts/ui/themes/minimal_theme.gd`

**Layout:**
- Small health bar in bottom-left
- Compact hotbar in bottom-center
- Target frame in top-center (only when targeting)

**Best for:** Clean, FPS-style gameplay with minimal screen clutter

### MMOTheme

**Location:** `client/scripts/ui/themes/mmo_theme.gd`

**Layout:**
- Large health bar in top-left
- Target frame in top-right
- Main hotbar at bottom-center
- Secondary hotbar above main hotbar

**Best for:** Traditional MMO gameplay with detailed information

## Usage

### In Player Script

The player automatically initializes the UI theme in `_ready()`:

```gdscript
# Variables
var ui_theme: UITheme = null
@export_enum("Minimal", "MMO") var preferred_ui_theme: String = "Minimal"

# In _ready()
func _initialize_ui_theme():
    match preferred_ui_theme:
        "Minimal":
            ui_theme = MinimalTheme.new()
        "MMO":
            ui_theme = MMOTheme.new()

    ui_container.add_child(ui_theme)
    ui_theme.initialize(self, targeting_system, weapon)
```

### Switching Themes at Runtime

```gdscript
# Player has a method to switch themes
player.switch_ui_theme("MMO")  # Switch to MMO theme
player.switch_ui_theme("Minimal")  # Switch back to Minimal
```

## Creating Custom Components

### Step 1: Extend UIComponent

```gdscript
extends UIComponent
class_name MyCustomComponent

@export var custom_setting: float = 100.0

func _setup_component():
    component_name = "MyCustom"
    custom_minimum_size = Vector2(200, 50)

    # Create your UI elements here
    var label = Label.new()
    add_child(label)

func _connect_signals():
    # Subscribe to relevant signals
    if player:
        player.custom_event.connect(_on_custom_event)

    if weapon:
        weapon.attack_started.connect(_on_attack)

func _update_component():
    # Update visuals based on game state
    queue_redraw()

func _draw():
    # Custom drawing code
    draw_rect(Rect2(0, 0, 200, 50), Color.BLUE, true)
```

### Step 2: Use in a Theme

```gdscript
extends UITheme
class_name MyCustomTheme

func _setup_theme():
    theme_name = "My Custom Theme"

    var my_component = MyCustomComponent.new()
    my_component.position = Vector2(100, 100)
    my_component.custom_setting = 150.0
    add_component(my_component, "my_component")
```

## Creating Custom Themes

### Step 1: Extend UITheme

```gdscript
extends UITheme
class_name CompactTheme

func _setup_theme():
    theme_name = "Compact"
    theme_description = "Ultra-minimal UI for maximum immersion"

    # Create your layout
    var health = HealthBar.new()
    health.bar_width = 150
    health.bar_height = 15
    health.show_text = false
    health.position = Vector2(10, viewport.y - 25)
    add_component(health, "health")

    var hotbar = Hotbar.new()
    hotbar.num_slots = 4
    hotbar.slot_size = 40
    hotbar.show_keybinds = false
    hotbar.position = Vector2(center.x, viewport.y - 50)
    add_component(hotbar, "hotbar")
```

### Step 2: Register in Player

```gdscript
# In player_animated.gd
@export_enum("Minimal", "MMO", "Compact") var preferred_ui_theme: String = "Minimal"

func _initialize_ui_theme():
    match preferred_ui_theme:
        "Minimal":
            ui_theme = MinimalTheme.new()
        "MMO":
            ui_theme = MMOTheme.new()
        "Compact":
            ui_theme = CompactTheme.new()
```

## Benefits of This Architecture

### 1. Complete Theme Swapping
**Before:** UI hardcoded in scene, can't change layout
**After:** `player.switch_ui_theme("MMO")` instantly changes entire UI

### 2. Component Reusability
Same HealthBar component can be used in:
- MinimalTheme (bottom-left, small)
- MMOTheme (top-left, large)
- CustomTheme (wherever you want)

### 3. Easy Customization
Players/modders can create custom themes without touching:
- Player code
- Combat systems
- Core game logic

### 4. Signal-Based Updates
Components automatically update when game state changes:
- TargetFrame updates when targeting_system.target_changed fires
- Hotbar updates when weapon.cooldown_started fires
- HealthBar updates when player health changes

### 5. No Blast Radius
Want to add a minimap component?
1. Create `Minimap` extending `UIComponent`
2. Add to themes: `add_component(Minimap.new(), "minimap")`
3. Done! No player code changes needed.

## Architecture Comparison

Similar to the combat system refactor, the UI follows these patterns:

| Pattern | Combat System | UI System |
|---------|--------------|-----------|
| **Strategy** | TargetingSystem | UITheme |
| **Composition** | Player has targeting + weapon | Theme has components |
| **Signal-based** | target_changed, cooldown_started | All components subscribe |
| **Duck typing** | ITargetable | UIComponent |
| **Swappable** | RaycastTargeting ↔ TabTargeting | MinimalTheme ↔ MMOTheme |

## File Structure

```
client/scripts/ui/
├── ui_theme.gd                    # Base theme class
├── components/
│   ├── ui_component.gd            # Base component class
│   ├── health_bar.gd              # Health bar component
│   ├── target_frame.gd            # Target info component
│   └── hotbar.gd                  # Ability hotbar component
└── themes/
    ├── minimal_theme.gd           # Minimal UI layout
    └── mmo_theme.gd               # Traditional MMO layout
```

## Testing Theme Switching

In-game, you can test theme switching:

```gdscript
# In console or debug script
player.switch_ui_theme("Minimal")
await get_tree().create_timer(2.0).timeout
player.switch_ui_theme("MMO")
```

## Future Enhancements

Easy to add:

1. **More Components:**
   - Minimap
   - Quest tracker
   - Inventory window
   - Chat window
   - Damage numbers
   - Buff/debuff icons

2. **More Themes:**
   - CompactTheme (ultra-minimal)
   - RPGTheme (classic JRPG style)
   - FPSTheme (Halo/Doom style)
   - StreamerTheme (viewer-friendly layout)

3. **Theme Settings:**
   - Scale adjustment
   - Color customization
   - Component visibility toggles
   - Save/load preferences

4. **Dynamic Layout:**
   - Responsive to screen size
   - Drag-and-drop positioning
   - Per-component customization

## Summary

The UI system provides:
- ✅ **Extensible** - Easy to add new components and themes
- ✅ **Composable** - Themes mix and match components
- ✅ **Swappable** - Change entire UI at runtime
- ✅ **Signal-based** - Automatic updates via events
- ✅ **Clean architecture** - No coupling to player/combat systems
- ✅ **Zero blast radius** - Add features without breaking existing code

This architecture makes it trivial to support different player preferences, create seasonal themes, or allow community-created UI mods!
