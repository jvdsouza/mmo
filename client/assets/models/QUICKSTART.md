# 5-Minute Model Integration Guide

## You Have a .glb File - Now What?

### Step 1: Copy the File (10 seconds)

```bash
# From the project root
cp /path/to/your/character.glb client/assets/models/characters/
```

### Step 2: Open Godot (30 seconds)

```bash
cd client
godot project.godot
```

Godot will automatically import the .glb file. You'll see:
- `character.glb`
- `character.glb.import` (auto-generated)

### Step 3: Preview the Model (20 seconds)

1. In Godot FileSystem panel (bottom-left)
2. Navigate to `res://assets/models/characters/`
3. Click on `character.glb`
4. Preview appears in 3D viewport
5. Check Inspector panel - you should see:
   - Scene structure
   - AnimationPlayer node
   - List of animations

### Step 4: Add to Player Scene (2 minutes)

#### Option A: Replace Capsule (Easiest)

1. Open `res://scenes/player_animated.tscn`
2. Delete the `MeshInstance3D` node (the capsule)
3. Drag `character.glb` from FileSystem onto the `Player` root node
4. The .glb becomes a child of Player
5. **Important**: Check if imported model has `AnimationPlayer`
   - If yes: Continue to Step 5
   - If no: You're done! Run the scene (F6)

#### Option B: Create New Scene (More Control)

1. Scene → New Scene
2. Add `CharacterBody3D` as root, name it "Player"
3. Drag `character.glb` as child
4. Add `CollisionShape3D` as child
   - Add CapsuleShape3D
   - Adjust size to match model
5. Add `Node3D` as child, name it "CameraPivot"
   - Position at head height (Y: 1.5)
   - Add `Camera3D` as child of CameraPivot
   - Position camera behind player (Z: 5)
6. Attach `res://scripts/player_animated.gd` script to root
7. Set `entity_id` in Inspector (e.g., "player_1")

### Step 5: Setup AnimationTree (2 minutes)

**Only if your model has animations**

1. With Player node selected
2. Add Child Node → `AnimationTree`
3. In Inspector for AnimationTree:
   - **Tree Root**: Click → Create "AnimationNodeStateMachine"
   - **Anim Player**: Select the AnimationPlayer from imported model
     - Path will be like: `../character/AnimationPlayer`
   - **Active**: ✓ Check this box

4. Click "AnimationTree" at top → Opens state machine editor
5. Add animations:
   - Right-click → Add Animation → Select "Idle"
   - Right-click → Add Animation → Select "Walking"
   - Right-click → Add Animation → Select "Attack"
   - Right-click → Add Animation → Select "Hit_Reaction"
   - Right-click → Add Animation → Select "Death"

6. Connect transitions:
   - Click Idle node
   - Click "Connect node..." button
   - Click Walking node
   - This creates arrow: Idle → Walking
   - Repeat for common transitions:
     - Idle ↔ Walking
     - Walking ↔ Running (if you have run animation)
     - Any → Attack
     - Any → Hit_Reaction
     - Any → Death

7. **Important**: Make sure Start node connects to Idle

### Step 6: Test It (30 seconds)

1. Press **F6** to run current scene
2. You should see your 3D model!
3. Move around (WASD) - model should move
4. If animations exist and AnimationTree is set up:
   - Standing still → Idle animation plays
   - Moving → Walking animation plays

### Step 7: Test with Server (1 minute)

1. Start server:
   ```bash
   # In terminal, from project root
   make dev
   ```

2. In Godot, press F5 (run main scene)

3. In browser, open test client:
   ```bash
   make test-websocket
   # Or open: client/test/websocket_test.html
   ```

4. Click "Attack Monster" in test client
5. Your character should play attack animation!

## Common Issues

### Model is Too Big/Small

1. Click .glb file in FileSystem
2. Import tab (top of screen)
3. Find "Scale" setting
4. Set to 1.0 (or 0.1 for large models)
5. Click "Reimport"

### Model is Facing Wrong Direction

1. Click .glb file
2. Import tab → Root Rotation
3. Set Y to 180
4. Reimport

### Animations Not Playing

- Check AnimationTree is **Active** (checkbox in Inspector)
- Check **Anim Player** is set to correct AnimationPlayer path
- Check animation names match what the script expects:
  - Must have: `Idle`, `Walking`, `Attack`, `Hit_Reaction`, `Death`
  - Or rename them in AnimationPlayer

### Can't Find AnimationPlayer

After importing .glb:
1. Expand the .glb node in scene tree
2. Look for `AnimationPlayer` child
3. If not there, your model doesn't have animations
4. Download a different model with animations

## What If My Model Doesn't Have Animations?

**That's OK!** The `player_animated.gd` script gracefully falls back:
- It checks if AnimationTree exists
- If not, uses simple scaling effects on the mesh
- Your character still works, just no animations

You can add animations later from a different source:
- Download animations separately from Mixamo
- Import them and merge with your model

## Next Steps

Once your model is working:

1. **Add more characters**: Repeat for different player types
2. **Add monsters**: Put monster models in `assets/models/monsters/`
3. **Add props**: Crates, barrels in `assets/models/environment/`
4. **Tune animations**: Adjust speeds in AnimationTree
5. **Add VFX**: Particle effects for hits (VFXManager is ready)

## Need Help?

- Full guide: `/docs/ADDING_3D_MODELS.md`
- Model sources: Quaternius.com (easiest), Mixamo.com (most options)
- Godot docs: https://docs.godotengine.org/en/stable/tutorials/assets_pipeline/importing_3d_scenes.html
