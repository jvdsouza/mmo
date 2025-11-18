# Hit Detection & Aiming System - Testing Guide

## What Was Implemented

### Core Features
1. **Raycast-based Hit Detection** - Camera-to-world raycast for detecting targets
2. **Attack Input Handling** - Left-click to attack
3. **Reticle/Crosshair UI** - Visual aiming indicator
4. **Network Integration** - Attacks sent to server with correct target_id
5. **Monster Spawning** - Client-side monster entities with entity_id
6. **Visual Feedback** - Reticle changes color based on target validity

### Files Created/Modified

**New Files:**
- `client/scripts/reticle.gd` - Crosshair UI with dynamic colors
- `client/scenes/reticle.tscn` - Reticle scene
- `client/scripts/monster.gd` - Monster entity script
- `client/scenes/monster.tscn` - Monster scene
- `ACTION_COMBAT_ROADMAP.md` - Full action combat system roadmap

**Modified Files:**
- `client/project.godot` - Added `primary_attack` and `secondary_attack` input actions
- `client/scripts/player_animated.gd` - Added raycast, attack logic, targeting
- `client/scripts/network_manager.gd` - Added `send_attack()`, `send_ability()`, `send_chat()`
- `client/scripts/multiplayer_controller.gd` - Added monster spawning and management
- `client/scenes/player_animated.tscn` - Added reticle UI

### System Flow

```
Player clicks (left mouse)
    ↓
player_animated.gd: attempt_attack()
    ↓
Raycast from camera → Detects entity with entity_id
    ↓
Validate: Range check (3m melee range)
    ↓
NetworkManager.send_attack(target_id)
    ↓
Server receives attack → Processes combat
    ↓
Server broadcasts combat_event
    ↓
Client receives event → Plays animations
```

## How to Test

### Prerequisites
1. Server built and running
2. Godot project open
3. Monster scene assigned in main scene

### Test 1: Basic Setup Verification

```bash
# Terminal 1 - Start server
cd /home/user/mmo
make dev
```

**Expected Output:**
```
🎮 MMO Server Starting...
Environment: development
Spawning 5 test monsters...
Monster spawned: monster_xxx (Goblin)
```

### Test 2: Client Connection & Reticle

```bash
# Open Godot
cd /home/user/mmo/client
godot project.godot
```

**Steps:**
1. Press F5 to run main scene
2. **Check**: Reticle appears in screen center (white crosshair)
3. **Check**: Console shows: `[Player] Registered with CombatEventManager: player_1`
4. **Check**: Console shows: `Monster spawned: monster_xxx (Goblin)`

**Expected**: 5 monsters spawn around the world

### Test 3: Target Detection

**Steps:**
1. Move mouse to aim at different objects
2. Aim camera at a monster (within 6m - 2x attack range)
3. **Check**: Reticle turns RED when aimed at monster in range
4. **Check**: Reticle turns ORANGE/GREEN when aimed at monster out of range
5. **Check**: Reticle stays WHITE when aimed at nothing

**Visual Reference:**
- White reticle = No target
- Red reticle = Valid target in attack range
- Orange reticle = Target detected but out of range

### Test 4: Attack Mechanics

**Steps:**
1. Aim at a monster (reticle should be RED)
2. Left-click (primary attack)
3. **Check Console for**:
   ```
   [Player] Attacking: monster_xxx at distance: 2.5
   [NetworkManager] Sent attack: monster_xxx with basic_attack
   ```
4. Try attacking when reticle is WHITE (no target)
5. **Check Console**:
   ```
   [Player] No target in range
   ```

### Test 5: Attack Cooldown

**Steps:**
1. Aim at monster
2. Spam left-click rapidly (click 5+ times fast)
3. **Check Console**:
   ```
   [Player] Attacking: monster_xxx at distance: 2.5
   [Player] Attack on cooldown: 0.85
   [Player] Attack on cooldown: 0.72
   ...
   ```
4. Wait 1 second
5. Click again - should attack successfully

**Expected**: 1 second cooldown between attacks

### Test 6: Range Validation

**Steps:**
1. Aim at a monster
2. Move CLOSE (< 3m) - reticle RED
3. Left-click - attack succeeds
4. Move FAR (> 3m) - reticle changes to ORANGE
5. Left-click - attack fails with:
   ```
   [Player] Target out of range: 4.2 > 3.0
   ```

**Expected**: Can only attack within 3m range

### Test 7: Server-Side Combat

**Steps:**
1. Attack a monster successfully
2. **Check Server Console**:
   ```
   [Combat] player_1 attacking monster_xxx
   Damage dealt: 25 (critical: false)
   ```
3. **Check Client Console** (combat event received):
   ```
   [Monster] Goblin taking damage: 25 critical: false
   ```
4. Monster should flash RED briefly

**Expected**: Full round-trip combat event

### Test 8: Monster Death

**Steps:**
1. Attack same monster repeatedly (4-5 times)
2. **Check**: Monster health decreases (visible in Label3D above monster)
3. When health reaches 0:
   - Monster plays death animation (fades out, shrinks)
   - Monster disappears from world
   - Console shows: `Removing despawned monster: monster_xxx`

**Expected**: Monster dies and is removed

### Test 9: Multiple Targets

**Steps:**
1. Spawn near multiple monsters
2. Aim at Monster A - reticle RED
3. Attack - only Monster A gets hit
4. Aim at Monster B - reticle RED
5. Attack - only Monster B gets hit

**Expected**: Raycast accurately targets specific monsters

### Test 10: Browser Test Client Integration

```bash
# Terminal 2
make test-websocket
```

**Steps:**
1. In browser, click "Attack Monster"
2. **Check Godot Console**: Combat events arrive for both browser and Godot clients
3. Monsters can be attacked from both clients
4. Combat events visible in browser message panel

**Expected**: Multi-client combat works

## Common Issues & Solutions

### Issue: No reticle visible
**Solution**: Check player_animated.tscn has UI/Reticle as child

### Issue: Reticle stays white even when aiming at monster
**Solution**:
- Check monster has CollisionShape3D
- Check monster script has `entity_id` set
- Check raycast distance (increase attack_range temporarily for testing)

### Issue: "Cannot attack - not connected to server"
**Solution**: Server not running - start with `make dev`

### Issue: "No target in range" even when close
**Solution**:
- Monster may not have collision shape
- Check player.attack_range (default: 3.0m)
- Try getting closer (< 2m)

### Issue: Monsters don't spawn
**Solution**:
- Check main scene has monster_scene assigned in MultiplayerController
- Check server has SPAWN_TEST_MONSTERS=true in .env
- Check server console for spawn messages

### Issue: Attack sends but no combat event received
**Solution**:
- Check server is processing attack (server console)
- Check CombatEventManager is registered (client console)
- Check target_id is valid (matches server monster ID)

## Performance Benchmarks

**Expected Performance:**
- Raycast: < 0.1ms per frame
- Attack latency: 20-50ms (network round-trip)
- Reticle update: 60fps
- Monster count: 100+ without lag

## Next Steps

After confirming hit detection works:

1. **Add HUD** - Health bars, damage numbers (Priority 2)
2. **Add Abilities** - Hotbar system (Priority 3)
3. **Add Monster AI** - Movement and attacks (Priority 4)
4. **Add Dodge Roll** - Defensive mechanics (Priority 5)

See `ACTION_COMBAT_ROADMAP.md` for full roadmap.

## Debug Commands

**In Godot Console** (during runtime):

```gdscript
# Check current target
print(player.current_target)

# Check attack cooldown
print(player.attack_cooldown)

# Force attack (bypass cooldown)
player.attack_cooldown = 0.0

# List all monsters
get_node("/root/Main/MultiplayerController").spawned_monsters
```

## Success Criteria

✅ Reticle visible and responsive
✅ Reticle changes color based on target
✅ Left-click sends attack to server
✅ Attacks only work within range
✅ Cooldown prevents spam
✅ Monsters receive and display damage
✅ Combat events show in console
✅ Monster death works correctly

If all criteria pass, hit detection system is **COMPLETE** ✨
