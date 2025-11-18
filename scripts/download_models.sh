#!/bin/bash
# Download free .glb character models for MMO

echo "🎮 Downloading free .glb character models..."

# Create assets directory
mkdir -p ../client/assets/models
mkdir -p ../client/assets/models/characters
mkdir -p ../client/assets/models/monsters

cd ../client/assets/models

# Quaternius - Ultimate Animated Characters (Free CC0)
echo ""
echo "📦 Option 1: Quaternius Ultimate Animated Characters"
echo "   Manual Download: https://quaternius.com/packs.html"
echo "   Look for: 'Ultimate Animated Character Pack'"
echo "   Download as .glb format"
echo ""

# Kenney Character Pack
echo "📦 Option 2: Kenney Character Assets"
echo "   Direct download:"
wget -nc https://kenney.nl/content/3-assets/531/platformerkit.zip -O kenney_platformer.zip 2>/dev/null || echo "   Manual: https://kenney.nl/assets/platformer-kit"
echo ""

# Mixamo via external converter
echo "📦 Option 3: Mixamo Characters (requires account)"
echo "   1. Go to: https://www.mixamo.com"
echo "   2. Sign in (free Adobe account)"
echo "   3. Select character (e.g., 'Y Bot')"
echo "   4. Download animations:"
echo "      - T-Pose"
echo "      - Idle"
echo "      - Walking"
echo "      - Running"
echo "      - Sword Slash (or Punch)"
echo "      - Hit Reaction"
echo "      - Death"
echo "   5. Export as FBX, then convert to .glb:"
echo "      - Use Blender: Import FBX → Export GLB"
echo "      - Or online: https://products.aspose.app/3d/conversion/fbx-to-glb"
echo ""

# Free .glb samples from Sketchfab
echo "📦 Option 4: Sketchfab Free Downloads"
echo "   Search with filters:"
echo "   - https://sketchfab.com/3d-models?features=downloadable&licenses=7c23a1ba438d4306920229c12afcb5f9"
echo "   - Filter: Animated, Free Download, .glb format"
echo ""

# Ready Player Me (programmatic character generation)
echo "📦 Option 5: Ready Player Me (API)"
echo "   Free customizable characters as .glb"
echo "   https://readyplayer.me/"
echo ""

echo "✨ Recommended Quick Start:"
echo ""
echo "For testing RIGHT NOW:"
echo "1. Visit: https://kenney.nl/assets/platformer-kit"
echo "2. Download and extract"
echo "3. Copy .glb files to: client/assets/models/characters/"
echo ""
echo "For production quality:"
echo "1. Use Mixamo (best animations)"
echo "2. Download character + animations as FBX"
echo "3. Convert to .glb with Blender"
echo "4. Place in client/assets/models/characters/"
echo ""

echo "📁 Expected file structure:"
echo "client/assets/models/"
echo "├── characters/"
echo "│   ├── warrior.glb          # Player model with animations"
echo "│   └── mage.glb             # Alternative class"
echo "└── monsters/"
echo "    ├── goblin.glb           # Basic enemy"
echo "    └── orc.glb              # Elite enemy"
echo ""

echo "🎯 After downloading models, run:"
echo "   cd client"
echo "   godot project.godot"
echo "   Then drag .glb files into the FileSystem panel"
echo ""
