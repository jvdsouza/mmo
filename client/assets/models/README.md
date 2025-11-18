# 3D Models Directory

This directory is structured for organizing game assets.

## Directory Structure

```
models/
├── characters/     # Player models (.glb files)
├── monsters/       # Enemy models (.glb files)
└── environment/    # Props, destructibles (.glb files)
```

## Quick Start - Adding Your First Model

1. **Download a free character model** (see `/docs/ADDING_3D_MODELS.md` for sources)
2. **Copy the .glb file** into `characters/` folder
3. **Open Godot** - The file will auto-import
4. **Verify import**: Click the .glb file in FileSystem panel to see preview
5. **Use in scene**: Drag the .glb into your player scene

## Recommended Sources

- **Quaternius**: https://quaternius.com/packs.html (Best - already .glb format)
- **Kenney**: https://kenney.nl/assets (Simple, good for testing)
- **Mixamo**: https://www.mixamo.com (Professional, needs .fbx to .glb conversion)

## Expected Animation Names

Your imported models should have these animations:
- `Idle` or `idle` - Standing still
- `Walking` or `walk` - Movement
- `Running` or `run` - Fast movement
- `Attack` or `attack_1` - Basic attack
- `Hit_Reaction` or `hit` - Taking damage
- `Death` or `die` - Death animation

## Current Status

🔴 **No models imported yet**

The animation system code is ready and waiting for assets. Once you add a .glb file here:
1. Godot will auto-import it with animations
2. The `CombatEventManager` will route server events to it
3. Animations will play automatically based on combat events

See `/docs/ADDING_3D_MODELS.md` for the complete guide.
