# Combat System Refactor - In Progress

## ✅ Completed

### New Architecture Created

**Folder Structure:**
```
client/scripts/combat/
├── targeting/
│   ├── targeting_system.gd       # Base class with signals
│   └── raycast_targeting.gd      # Camera raycast implementation
├── interfaces/
│   └── targetable.gd              # ITargetable interface definition
├── weapons/
│   ├── weapon.gd                  # Base weapon class
│   └── melee_weapon.gd            # Melee weapon implementation
└── combat_validator.gd            # Attack validation logic
```

### Core Systems Implemented

**1. ITargetable Interface** (`combat/interfaces/targetable.gd`)
- Defines contract for targetable entities
- Methods: get_entity_id(), is_valid_target(), get_target_position(), etc.
- Uses duck typing (GDScript doesn't support true interfaces)

**2. TargetingSystem** (`combat/targeting/targeting_system.gd`)
- Base class for all targeting modes
- Signals: target_changed, target_lost, no_valid_targets
- Methods: get_current_target(), has_valid_target(), is_in_range()
- Fully decoupled from player

**3. RaycastTargeting** (`combat/targeting/raycast_targeting.gd`)
- Camera-based raycast targeting
- Auto-updates every frame
- perform_raycast() returns targetable entity
- get_raycast_hit_point() for projectiles/VFX

**4. Weapon System** (`combat/weapons/weapon.gd`)
- Base weapon class with cooldowns
- Signals: attack_started, attack_completed, cooldown_started/finished
- Properties: attack_range, cooldown_time, damage, animation_hint
- is_ready() checks cooldown
- attack(target) initiates attack

**5. MeleeWeapon** (`combat/weapons/melee_weapon.gd`)
- Extends Weapon
- _perform_attack() sends network message
- Instant-hit melee attacks

**6. CombatValidator** (`combat/combat_validator.gd`)
- Centralized validation logic
- validate_attack() checks: target, range, cooldown, state
- Returns ValidationResult enum
- Easy to extend (LOS, friendly fire, resources)

**7. Monster ITargetable Implementation** (`scripts/monster.gd`)
- Implements all ITargetable methods
- get_entity_id(), is_valid_target(), is_alive(), etc.
- get_faction() returns "monster"
- Ready for targeting system

---

## ✅ Refactor Complete!

### Player Refactor - COMPLETED

**File:** `client/scripts/player_animated.gd`

**Changes Made:**
1. ✅ Added ITargetable interface implementation
2. ✅ Created RaycastTargeting instance in _ready()
3. ✅ Created MeleeWeapon instance in _ready()
4. ✅ Refactored attempt_attack() to use CombatValidator
5. ✅ Removed old targeting code (update_target_detection, get_raycast_target)
6. ✅ Connected weapon signals to animations
7. ✅ Removed hardcoded attack_range and cooldown properties

### Reticle Refactor - COMPLETED

**File:** `client/scripts/reticle.gd`

**Changes Made:**
1. ✅ Updated to use targeting_system.has_valid_target()
2. ✅ Uses targeting_system.get_current_target() instead of player.current_target
3. ✅ Uses weapon.attack_range instead of player.attack_range
4. ✅ Fully decoupled from player internals

### Type System - COMPLETED

**Files:** `weapon.gd`, `combat_validator.gd`

**Changes Made:**
1. ✅ Changed all ITargetable type hints to Node (duck typing)
2. ✅ Added has_method() checks in weapon.attack() and CombatValidator
3. ✅ Consistent duck typing throughout all combat systems

---

## ⏳ Remaining Tasks

### 1. ✅ Complete Player Refactor - DONE

- [x] Add ITargetable interface methods to player
- [x] Initialize targeting_system (RaycastTargeting)
- [x] Initialize weapon (MeleeWeapon)
- [x] Refactor attempt_attack() to use CombatValidator
- [x] Remove old targeting code (update_target_detection, get_raycast_target)
- [x] Connect weapon signals to animations

### 2. ✅ Update Reticle - DONE

- [x] Remove direct player.current_target access
- [x] Use targeting_system.get_current_target() instead
- [x] Use targeting_system.has_valid_target()
- [x] Use weapon.attack_range from player.weapon

### 3. Testing - NEXT
**Estimated Time:** 30 minutes

- [ ] Test in Godot with server running
- [ ] Verify targeting still works
- [ ] Verify attacks still work
- [ ] Check reticle color changes
- [ ] Test cooldowns
- [ ] Test range validation
- [ ] Check monster damage/death

### 4. Commit & Push
**Estimated Time:** 10 minutes

- [ ] Review all changes
- [ ] Create comprehensive commit message
- [ ] Push to branch: `claude/raylib-comparison-research-011CV5gnH9bFEGaAKxWZVsQ4`

---

## 🎯 Benefits of New Architecture

### Extensibility
**Before:** Change targeting = rewrite player script
**After:** Add new TargetingSystem subclass

Example - Adding tab targeting:
```gdscript
# tab_targeting.gd - NO player changes needed!
extends TargetingSystem
class_name TabTargeting

func update_targeting(delta):
    if Input.is_action_just_pressed("tab_target"):
        cycle_to_next_target()
```

### Testability
**Before:** Can't test targeting without player
**After:** Test targeting independently

### Weapon Swapping
**Before:** Hardcoded in player
**After:**
```gdscript
player.weapon = RangedWeapon.new()  # Now shoots projectiles!
player.weapon = MagicWeapon.new()   # Now casts spells!
```

### No Blast Radius
**Change targeting mode:** Only swap targeting_system instance
**Add weapon type:** Create new Weapon subclass
**Change validation:** Edit CombatValidator only

---

## 📁 New File Structure

```
client/scripts/
├── combat/                         # NEW
│   ├── targeting/
│   │   ├── targeting_system.gd
│   │   └── raycast_targeting.gd
│   ├── interfaces/
│   │   └── targetable.gd
│   ├── weapons/
│   │   ├── weapon.gd
│   │   └── melee_weapon.gd
│   └── combat_validator.gd
├── player_animated.gd              # TO REFACTOR
├── monster.gd                      # ✅ REFACTORED
└── reticle.gd                      # TO UPDATE
```

---

## 🚀 Next Session TODO

1. **Refactor player_animated.gd** (30 min)
   - Add ITargetable interface
   - Create targeting_system and weapon
   - Use CombatValidator
   - Remove old code

2. **Update reticle.gd** (15 min)
   - Use targeting signals
   - Decouple from player

3. **Test Everything** (30 min)
   - Run in Godot
   - Verify combat works
   - Check all features

4. **Commit & Document** (20 min)
   - Comprehensive commit message
   - Update documentation
   - Push to branch

**Total Time:** ~2 hours

---

## 🔍 Testing Checklist (After Refactor)

- [ ] Server starts and spawns monsters
- [ ] Player connects and spawns
- [ ] Reticle appears (white)
- [ ] Aiming at monster turns reticle RED
- [ ] Left-click attacks monster
- [ ] Cooldown prevents spam (1s)
- [ ] Range validation works (3m)
- [ ] Monster takes damage
- [ ] Monster dies after ~4 hits
- [ ] Combat events in console
- [ ] No errors or warnings

---

## 💡 Future Enhancements (Easy Now!)

**With this architecture, we can easily add:**

1. **Tab Targeting** - New TargetingSystem subclass
2. **Lock-on Targeting** - Another subclass
3. **Ranged Weapons** - New Weapon subclass
4. **Magic Weapons** - Another Weapon subclass
5. **Weapon Stats** - Already in Weapon base class
6. **Line of Sight** - Add to CombatValidator
7. **Friendly Fire** - Add to CombatValidator
8. **Resource Costs** - Add to Weapon and CombatValidator

All without touching player code!

---

## Current Branch

`claude/raylib-comparison-research-011CV5gnH9bFEGaAKxWZVsQ4`

**Status:** ✅ Refactor 100% complete! Ready for testing and commit.

**Files Modified:**
- `client/scripts/combat/weapons/weapon.gd` - Fixed type hints for duck typing
- `client/scripts/combat/combat_validator.gd` - Fixed type hints for duck typing
- `client/scripts/player_animated.gd` - Complete refactor to use new architecture
- `client/scripts/reticle.gd` - Updated to use targeting system

**Files Created:**
- `client/scripts/combat/interfaces/targetable.gd`
- `client/scripts/combat/targeting/targeting_system.gd`
- `client/scripts/combat/targeting/raycast_targeting.gd`
- `client/scripts/combat/weapons/weapon.gd`
- `client/scripts/combat/weapons/melee_weapon.gd`
- `client/scripts/combat/combat_validator.gd`

**Files Previously Modified:**
- `client/scripts/monster.gd` - Added ITargetable implementation
