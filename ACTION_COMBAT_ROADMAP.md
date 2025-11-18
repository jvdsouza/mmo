# Action Combat System Roadmap

## Priority Tasks for GitHub Project

### 🎯 Priority 1: Hit Detection & Aiming System
**Status**: In Progress
**Labels**: `combat`, `gameplay`, `P0`

**Description**:
Implement raycast-based hit detection for action combat. Players currently cannot attack in-game (only via browser test client with manual target IDs).

**Tasks**:
- [ ] Add raycast from camera center to detect entities
- [ ] Implement mouse click attack input
- [ ] Add `send_attack()` method to NetworkManager
- [ ] Create reticle/crosshair UI
- [ ] Dynamic reticle (color change when aiming at enemy)
- [ ] Hit marker visual feedback
- [ ] Range validation (melee vs ranged)

**Acceptance Criteria**:
- Player can click to attack enemies they're aiming at
- Reticle appears in screen center
- Reticle changes color when over valid target
- Server receives attack with correct `target_id`
- Hit feedback displays on successful hit

**Estimated Effort**: 2-3 hours

---

### 🎨 Priority 2: Game HUD/UI System
**Status**: Todo
**Labels**: `ui`, `gameplay`, `P1`

**Description**:
Essential visual feedback for combat. Players need to see health, damage, and status.

**Tasks**:
- [ ] Health bars (player + enemies)
- [ ] Floating damage numbers (integrate with VFXManager)
- [ ] Player name labels over characters
- [ ] Combat log/event feed
- [ ] Status indicators (buffs/debuffs)
- [ ] Boss health bars (large enemies)
- [ ] Hit direction indicators

**Acceptance Criteria**:
- Health bars visible above all entities
- Damage numbers float up on hit
- Combat log shows recent events
- UI scales properly at different resolutions
- Performance: 60fps with 10+ entities visible

**Estimated Effort**: 4-5 hours

---

### ⚔️ Priority 3: Ability/Hotbar System
**Status**: Todo
**Labels**: `combat`, `abilities`, `ui`, `P1`

**Description**:
Multiple combat abilities with directional casting. Server supports "ability" messages but client only has basic attack concept.

**Tasks**:
- [ ] Hotbar UI (slots 1-9)
- [ ] Ability data structure (name, cooldown, cost, range)
- [ ] Cooldown timer display
- [ ] Resource management (mana/stamina)
- [ ] Directional ability casting (cast where you're facing)
- [ ] Ability tooltips
- [ ] Visual effects per ability type
- [ ] Sound hooks (when implemented)

**Abilities to Implement**:
- Basic Attack (instant, melee)
- Fireball (projectile, ranged)
- Heal (self-target)
- Charge/Dash (movement ability)
- Area of Effect attack

**Acceptance Criteria**:
- 5+ abilities usable from hotbar
- Cooldowns prevent spam
- Visual feedback for each ability
- Server validates ability usage
- Abilities respect facing direction

**Estimated Effort**: 6-8 hours

---

### 🤖 Priority 4: Monster AI & Behavior
**Status**: Todo
**Labels**: `ai`, `monsters`, `gameplay`, `P2`

**Description**:
Dynamic monster behavior. Currently monsters spawn but don't move or act.

**Tasks**:
- [ ] Aggro detection (player proximity)
- [ ] Chase behavior (pathfinding)
- [ ] Attack behavior (animations + damage)
- [ ] Return to spawn point (leash)
- [ ] Patrol routes (idle behavior)
- [ ] Death and respawn cycle
- [ ] Difficulty scaling (easy/normal/hard monsters)
- [ ] Boss AI (special behaviors)

**AI States**:
- Idle: Patrol or stand
- Aggro: Detect player in range
- Chase: Move toward player
- Attack: In range, perform attack
- Flee: Low health retreat (optional)
- Return: Go back to spawn
- Dead: Death animation, respawn timer

**Acceptance Criteria**:
- Monsters detect and chase players
- Monsters attack when in range
- Monsters return to spawn when player too far
- Death/respawn cycle works
- Performance: 100+ monsters without lag

**Estimated Effort**: 8-10 hours

---

### 🏃 Priority 5: Dodge Roll / Dash System
**Status**: Todo
**Labels**: `combat`, `movement`, `gameplay`, `P2`

**Description**:
Core action combat defensive mechanic. Allows skill expression and counter-play.

**Tasks**:
- [ ] Dodge roll input (Spacebar/Shift)
- [ ] Dodge animation (even with capsule)
- [ ] I-frames (invincibility frames)
- [ ] Stamina cost system
- [ ] Cooldown timer
- [ ] Visual trail effect during dodge
- [ ] Sound hook for dodge
- [ ] Cancel animations during dodge

**Mechanics**:
- Duration: 0.4s
- I-frames: 0.3s (first 75% of dodge)
- Cooldown: 1.0s
- Stamina cost: 25
- Distance: 3-5 units
- Direction: Based on movement input (or backward if stationary)

**Acceptance Criteria**:
- Dodge roll works in all directions
- I-frames prevent damage during roll
- Cannot spam (cooldown + stamina cost)
- Works with network latency (client prediction)
- Visual feedback clear

**Estimated Effort**: 4-5 hours

---

### 📷 Priority 6: Camera Improvements
**Status**: Todo
**Labels**: `camera`, `ux`, `P3`

**Description**:
Better camera control for action combat feel.

**Tasks**:
- [ ] Adjustable camera distance (mouse wheel)
- [ ] Camera collision detection (no wall clipping)
- [ ] Combat camera shake (on hit)
- [ ] Soft lock-on system (optional, faces enemy)
- [ ] FOV options
- [ ] Camera smoothing settings
- [ ] Third-person offset adjustment

**Acceptance Criteria**:
- Camera doesn't clip through walls
- Zoom in/out smooth and responsive
- Camera shake enhances impact feel
- Settings persist across sessions

**Estimated Effort**: 3-4 hours

---

### 🧪 Priority 7: Comprehensive Testing & Polish
**Status**: Todo
**Labels**: `testing`, `qa`, `P3`

**Description**:
End-to-end testing and polish pass.

**Tasks**:
- [ ] Test full combat loop (hit detection → server → animation)
- [ ] Network latency testing (simulate 100ms+ lag)
- [ ] Load testing (100+ concurrent players)
- [ ] Edge case testing (negative health, collision bugs)
- [ ] Performance profiling (client + server)
- [ ] Memory leak detection
- [ ] Documentation updates
- [ ] Video demo creation

**Acceptance Criteria**:
- All systems work together seamlessly
- No critical bugs
- Documented known issues
- Performance targets met (60fps client, 20Hz server)

**Estimated Effort**: 4-6 hours

---

## Total Estimated Effort
**30-45 hours** for complete action combat system

## Current Progress
- ✅ Server-side combat system
- ✅ Event broadcasting
- ✅ Client animation integration
- ✅ VFXManager foundation
- 🚧 Hit detection (in progress)
- ⏳ Remaining priorities

## Dependencies
- **Priority 1** must complete before 2 & 3 (need to attack to test UI/abilities)
- **Priority 2 & 3** can be done in parallel
- **Priority 4** depends on 1 (monsters need to detect when hit)
- **Priority 5 & 6** can be done anytime
- **Priority 7** happens after 1-6

## Notes for GitHub Project Setup
1. Create project called "Action Combat System"
2. Add columns: Backlog, Todo, In Progress, Done
3. Copy each priority as a separate issue
4. Link issues to project
5. Add appropriate labels
6. Set milestones if desired
