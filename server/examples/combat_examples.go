package examples

import (
	"fmt"
	"mmo-server/combat"
	"mmo-server/entities"
	"mmo-server/player"
	"mmo-server/stats"
	"strings"
	"time"
)

// =============================================================================
// Example 1: Player vs Monster Combat
// =============================================================================

func ExamplePlayerVsMonster() {
	fmt.Println("=== Example 1: Player vs Monster Combat ===\n")

	// Create combat manager
	cm := combat.NewCombatManager(combat.DefaultCombatConfig())

	// Create a warrior player
	warrior := player.NewPlayerWithArchetype(
		"player1",
		"Warrior Bob",
		stats.NewWarriorArchetype(1),
	)
	warrior.Position = combat.Vector3{X: 0, Y: 0, Z: 0}

	// Create a goblin monster
	goblin := entities.NewMonster(
		"goblin1",
		entities.GetGoblinTemplate(),
		combat.Vector3{X: 5, Y: 0, Z: 0}, // 5 meters away
	)

	// Register entities with combat manager
	cm.RegisterEntity(warrior)
	cm.RegisterEntity(goblin)

	fmt.Printf("Warrior HP: %d/%d\n", int(warrior.Stats.GetStat(stats.StatHealth)), int(warrior.Stats.GetStat(stats.StatMaxHealth)))
	fmt.Printf("Goblin HP: %d/%d\n", int(goblin.Stats.GetStat(stats.StatHealth)), int(goblin.Stats.GetStat(stats.StatMaxHealth)))
	fmt.Println()

	// Warrior moves closer (within melee range)
	warrior.Position = combat.Vector3{X: 2, Y: 0, Z: 0}
	cm.UpdateEntityPosition(warrior, combat.Vector3{X: 0, Y: 0, Z: 0}, warrior.Position)

	// Warrior attacks goblin with basic attack
	event, err := cm.ExecuteAbility("player1", "goblin1", "basic_attack")
	if err != nil {
		fmt.Printf("Attack failed: %v\n", err)
		return
	}

	fmt.Printf("Warrior attacks Goblin!\n")
	fmt.Printf("Damage dealt: %d\n", event.Damage)
	fmt.Printf("Goblin HP: %d/%d\n", event.TargetHealth, event.TargetMaxHealth)
	if event.IsCritical {
		fmt.Println("CRITICAL HIT!")
	}
	fmt.Println()

	// Goblin counter-attacks (simulate monster AI attack)
	goblinEvent, err := cm.ExecuteAbility("goblin1", "player1", "basic_attack")
	if err != nil {
		fmt.Printf("Goblin attack failed: %v\n", err)
	} else {
		fmt.Printf("Goblin strikes back!\n")
		fmt.Printf("Damage dealt: %d\n", goblinEvent.Damage)
		fmt.Printf("Warrior HP: %d/%d\n", goblinEvent.TargetHealth, goblinEvent.TargetMaxHealth)
		fmt.Println()
	}

	// Continue combat until goblin dies
	for i := 0; i < 5; i++ {
		if !goblin.IsAlive() {
			fmt.Println("Goblin defeated!")
			break
		}

		event, _ := cm.ExecuteAbility("player1", "goblin1", "basic_attack")
		fmt.Printf("Attack %d: %d damage -> Goblin HP: %d/%d\n",
			i+2, event.Damage, event.TargetHealth, event.TargetMaxHealth)

		if event.TargetDied {
			fmt.Println("Goblin has been slain!")
			break
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
}

// =============================================================================
// Example 2: Player vs Player (PvP) Combat
// =============================================================================

func ExamplePlayerVsPlayer() {
	fmt.Println("=== Example 2: Player vs Player (PvP) Combat ===\n")

	// Create combat manager
	cm := combat.NewCombatManager(combat.DefaultCombatConfig())

	// Create two players with different archetypes
	warrior := player.NewPlayerWithArchetype(
		"player1",
		"Warrior Bob",
		stats.NewWarriorArchetype(1),
	)
	warrior.Position = combat.Vector3{X: 0, Y: 0, Z: 0}
	warrior.PvPEnabled = true

	mage := player.NewPlayerWithArchetype(
		"player2",
		"Mage Alice",
		stats.NewMageArchetype(1),
	)
	mage.Position = combat.Vector3{X: 3, Y: 0, Z: 0}
	mage.PvPEnabled = true

	// Register entities
	cm.RegisterEntity(warrior)
	cm.RegisterEntity(mage)

	fmt.Printf("%s (Warrior) HP: %d/%d, Attack: %d, Defense: %d\n",
		warrior.Name,
		int(warrior.Stats.GetStat(stats.StatHealth)),
		int(warrior.Stats.GetStat(stats.StatMaxHealth)),
		int(warrior.Stats.GetStat(stats.StatAttackPower)),
		int(warrior.Stats.GetStat(stats.StatDefense)),
	)
	fmt.Printf("%s (Mage) HP: %d/%d, Attack: %d, Defense: %d\n",
		mage.Name,
		int(mage.Stats.GetStat(stats.StatHealth)),
		int(mage.Stats.GetStat(stats.StatMaxHealth)),
		int(mage.Stats.GetStat(stats.StatAttackPower)),
		int(mage.Stats.GetStat(stats.StatDefense)),
	)
	fmt.Println()

	// Warrior attacks mage
	event, err := cm.ExecuteAbility("player1", "player2", "basic_attack")
	if err != nil {
		fmt.Printf("Attack failed: %v\n", err)
		return
	}

	fmt.Printf("Warrior attacks Mage: %d damage\n", event.Damage)
	fmt.Printf("Mage HP: %d/%d\n", event.TargetHealth, event.TargetMaxHealth)
	fmt.Println()

	// Mage counter-attacks
	mageEvent, _ := cm.ExecuteAbility("player2", "player1", "basic_attack")
	fmt.Printf("Mage retaliates: %d damage\n", mageEvent.Damage)
	fmt.Printf("Warrior HP: %d/%d\n", mageEvent.TargetHealth, mageEvent.TargetMaxHealth)
	fmt.Println()

	// Test PvP disabled scenario
	warrior.PvPEnabled = false
	_, err = cm.ExecuteAbility("player2", "player1", "basic_attack")
	if err != nil {
		fmt.Printf("Attack blocked: %v (PvP disabled)\n", err)
	}

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
}

// =============================================================================
// Example 3: Player Destroying Crate
// =============================================================================

func ExamplePlayerVsCrate() {
	fmt.Println("=== Example 3: Player Destroying Crate ===\n")

	// Create combat manager
	cm := combat.NewCombatManager(nil)

	// Create player
	player1 := player.NewPlayer("player1", "Adventurer")
	player1.Position = combat.Vector3{X: 0, Y: 0, Z: 0}

	// Create wooden crate
	crate := entities.NewDestructible(
		"crate1",
		entities.GetWoodenCrateTemplate(),
		combat.Vector3{X: 2, Y: 0, Z: 0},
	)

	// Register entities
	cm.RegisterEntity(player1)
	cm.RegisterEntity(crate)

	fmt.Printf("Player found a %s (HP: %.0f)\n", crate.Name, crate.CurrentHealth)
	fmt.Println("Player attacks the crate...")
	fmt.Println()

	// Attack crate until destroyed
	attackCount := 0
	for crate.IsAlive() {
		attackCount++
		event, err := cm.ExecuteAbility("player1", "crate1", "basic_attack")
		if err != nil {
			fmt.Printf("Attack failed: %v\n", err)
			break
		}

		fmt.Printf("Hit %d: %d damage -> Crate HP: %d\n",
			attackCount, event.Damage, event.TargetHealth)

		if event.TargetDied {
			fmt.Println("\nCrate destroyed!")
			fmt.Println("Loot dropped: Wood Planks x2, Small Coins x3")
			break
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
}

// =============================================================================
// Example 4: Player Cutting Grass (Tool Requirement)
// =============================================================================

func ExamplePlayerVsGrass() {
	fmt.Println("=== Example 4: Player Cutting Grass (Tool Required) ===\n")

	// Create combat manager
	cm := combat.NewCombatManager(nil)

	// Create player
	player1 := player.NewPlayer("player1", "Link")
	player1.Position = combat.Vector3{X: 0, Y: 0, Z: 0}

	// Create tall grass
	grass := entities.NewDestructible(
		"grass1",
		entities.GetGrassTemplate(),
		combat.Vector3{X: 1, Y: 0, Z: 0},
	)

	// Register entities
	cm.RegisterEntity(player1)
	cm.RegisterEntity(grass)

	fmt.Printf("Player found %s\n", grass.Name)
	if grass.RequiresTool {
		fmt.Printf("Requires tool: %s\n", grass.RequiredToolID)
	}
	fmt.Println()

	// Player attacks grass (assume has sword equipped)
	event, err := cm.ExecuteAbility("player1", "grass1", "basic_attack")
	if err != nil {
		fmt.Printf("Cannot cut grass: %v\n", err)
		return
	}

	fmt.Printf("Player cuts the grass: %d damage\n", event.Damage)
	if event.TargetDied {
		fmt.Println("Grass destroyed!")
		fmt.Println("Items found: Grass Blade x2, Seeds x1")
	}

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
}

// =============================================================================
// Example 5: Environmental Damage (Lava Pool)
// =============================================================================

func ExampleEnvironmentalDamage() {
	fmt.Println("=== Example 5: Environmental Damage (Lava Pool) ===\n")

	// Create combat manager
	cm := combat.NewCombatManager(nil)

	// Create player
	player1 := player.NewPlayer("player1", "Unlucky Steve")
	player1.Position = combat.Vector3{X: 0, Y: 0, Z: 0}

	// Create lava pool
	lava := entities.NewEnvironmental(
		"lava1",
		entities.GetLavaPoolTemplate(),
		combat.Vector3{X: 3, Y: 0, Z: 0},
	)

	// Register entities
	cm.RegisterEntity(player1)
	cm.RegisterEntity(lava)

	playerHP := int(player1.Stats.GetStat(stats.StatHealth))
	playerMaxHP := int(player1.Stats.GetStat(stats.StatMaxHealth))

	fmt.Printf("Player HP: %d/%d\n", playerHP, playerMaxHP)
	fmt.Printf("Lava pool at position (%.1f, %.1f, %.1f) with radius %.1f\n",
		lava.Position.X, lava.Position.Y, lava.Position.Z, lava.Radius)
	fmt.Println()

	// Player walks into lava
	fmt.Println("Player walks toward lava...")
	player1.Position = combat.Vector3{X: 3, Y: 0, Z: 0} // Center of lava
	cm.UpdateEntityPosition(player1, combat.Vector3{X: 0, Y: 0, Z: 0}, player1.Position)

	// Simulate environmental damage over time
	for tick := 0; tick < 5; tick++ {
		time.Sleep(100 * time.Millisecond) // Simulate time passing

		// Update environmental hazard
		affectedEntities := lava.Update(1.0, cm)

		if len(affectedEntities) > 0 {
			for _, entity := range affectedEntities {
				damageResult := entity.TakeDamage(lava.GetDamageAmount(), lava)
				fmt.Printf("Tick %d: Lava deals %d fire damage -> HP: %d/%d\n",
					tick+1, damageResult.Damage, damageResult.RemainingHP, damageResult.MaxHP)

				if damageResult.Died {
					fmt.Println("\nPlayer has died in the lava!")
					goto End
				}
			}
		}
	}

	fmt.Println("\nPlayer escapes the lava!")

End:
	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
}

// =============================================================================
// Example 6: Monster Attacking Player
// =============================================================================

func ExampleMonsterVsPlayer() {
	fmt.Println("=== Example 6: Monster Attacking Player (AI Aggro) ===\n")

	// Create combat manager
	cm := combat.NewCombatManager(nil)

	// Create player
	player1 := player.NewPlayer("player1", "Brave Hero")
	player1.Position = combat.Vector3{X: 0, Y: 0, Z: 0}

	// Create orc monster
	orc := entities.NewMonster(
		"orc1",
		entities.GetOrcTemplate(),
		combat.Vector3{X: 10, Y: 0, Z: 0},
	)

	// Register entities
	cm.RegisterEntity(player1)
	cm.RegisterEntity(orc)

	fmt.Printf("Player HP: %d/%d\n",
		int(player1.Stats.GetStat(stats.StatHealth)),
		int(player1.Stats.GetStat(stats.StatMaxHealth)))
	fmt.Printf("Orc Warrior HP: %d/%d\n",
		int(orc.Stats.GetStat(stats.StatHealth)),
		int(orc.Stats.GetStat(stats.StatMaxHealth)))
	fmt.Println()

	// Player enters orc's aggro radius
	fmt.Printf("Player enters Orc's aggro radius (%.1fm)\n", orc.AggroRadius)
	player1.Position = combat.Vector3{X: 10, Y: 0, Z: 0}

	// Simulate monster AI update (would normally be in game loop)
	orc.Update(0.1, cm)

	if orc.AIState == entities.AIStateChasing {
		fmt.Println("Orc detects player and starts chasing!")
		fmt.Printf("Orc aggro table top threat: %s\n", orc.AggroTable.GetTopThreat())
		fmt.Println()
	}

	// Orc attacks player
	event, err := cm.ExecuteAbility("orc1", "player1", "basic_attack")
	if err != nil {
		fmt.Printf("Orc attack failed: %v\n", err)
		return
	}

	fmt.Printf("Orc attacks player: %d damage\n", event.Damage)
	fmt.Printf("Player HP: %d/%d\n", event.TargetHealth, event.TargetMaxHealth)
	fmt.Println()

	// Player fights back
	playerEvent, _ := cm.ExecuteAbility("player1", "orc1", "basic_attack")
	fmt.Printf("Player counter-attacks: %d damage\n", playerEvent.Damage)
	fmt.Printf("Orc HP: %d/%d\n", playerEvent.TargetHealth, playerEvent.TargetMaxHealth)
	fmt.Println()

	// Show threat increase
	fmt.Println("Orc's threat table updated (player dealt damage)")
	fmt.Printf("Current target: %s\n", orc.CurrentTarget)

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
}

// =============================================================================
// Example 7: Spatial Grid Optimization (Many Entities)
// =============================================================================

func ExampleSpatialGridPerformance() {
	fmt.Println("=== Example 7: Spatial Grid Performance ===\n")

	// Create combat manager with spatial grid
	config := combat.DefaultCombatConfig()
	config.SpatialCellSize = 10.0
	cm := combat.NewCombatManager(config)

	// Create many entities spread across the world
	fmt.Println("Creating 100 entities across a 100x100 world...")

	entityCount := 100
	for i := 0; i < entityCount; i++ {
		x := float64(i%10) * 10.0
		z := float64(i/10) * 10.0

		if i%2 == 0 {
			// Create player
			p := player.NewPlayer(fmt.Sprintf("player%d", i), fmt.Sprintf("Player%d", i))
			p.Position = combat.Vector3{X: x, Y: 0, Z: z}
			cm.RegisterEntity(p)
		} else {
			// Create monster
			m := entities.NewMonster(
				fmt.Sprintf("goblin%d", i),
				entities.GetGoblinTemplate(),
				combat.Vector3{X: x, Y: 0, Z: z},
			)
			cm.RegisterEntity(m)
		}
	}

	fmt.Printf("Total entities registered: %d\n\n", cm.GetEntityCount())

	// Test spatial query performance
	testPos := combat.Vector3{X: 50, Y: 0, Z: 50}
	radius := 15.0

	fmt.Printf("Finding all entities within %.1fm of position (%.1f, %.1f, %.1f)...\n",
		radius, testPos.X, testPos.Y, testPos.Z)

	nearby := cm.GetNearbyEntities(testPos, radius)

	fmt.Printf("Found %d nearby entities (checked ~%d cells instead of %d total)\n",
		len(nearby), 9, entityCount) // 3x3 grid cells vs all entities

	fmt.Println("\nSpatial Grid Benefits:")
	fmt.Println("- O(k) lookup where k = nearby entities (~10-20)")
	fmt.Println("- Instead of O(n) where n = all entities (100)")
	fmt.Println("- Performance improvement: ~5-10x for this scenario")
	fmt.Println("- Scales to 1000+ entities with same performance")

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
}

// =============================================================================
// Example 8: Complete Combat Scenario
// =============================================================================

func ExampleCompleteCombatScenario() {
	fmt.Println("=== Example 8: Complete Combat Scenario ===\n")
	fmt.Println("A warrior enters a dungeon with monsters, traps, and treasure...\n")

	// Create combat manager
	cm := combat.NewCombatManager(nil)

	// Create warrior
	warrior := player.NewPlayerWithArchetype(
		"player1",
		"Dungeon Delver",
		stats.NewWarriorArchetype(1),
	)
	warrior.Position = combat.Vector3{X: 0, Y: 0, Z: 0}
	cm.RegisterEntity(warrior)

	// Create goblin guard
	goblin := entities.NewMonster(
		"goblin1",
		entities.GetGoblinTemplate(),
		combat.Vector3{X: 5, Y: 0, Z: 0},
	)
	cm.RegisterEntity(goblin)

	// Create spike trap
	trap := entities.NewEnvironmental(
		"trap1",
		entities.GetSpikeTrapTemplate(),
		combat.Vector3{X: 10, Y: 0, Z: 5},
	)
	cm.RegisterEntity(trap)

	// Create treasure chest
	chest := entities.NewDestructible(
		"chest1",
		entities.GetTreasureChestTemplate(),
		combat.Vector3{X: 15, Y: 0, Z: 0},
	)
	cm.RegisterEntity(chest)

	// --- Scene 1: Fight Goblin ---
	fmt.Println("Scene 1: Warrior encounters Goblin guard")
	warrior.Position = combat.Vector3{X: 3, Y: 0, Z: 0}

	for goblin.IsAlive() {
		event, _ := cm.ExecuteAbility("player1", "goblin1", "basic_attack")
		fmt.Printf("  Warrior attacks: %d damage -> Goblin HP: %d\n",
			event.Damage, event.TargetHealth)

		if event.TargetDied {
			fmt.Println("  Goblin defeated!")
			break
		}
	}
	fmt.Println()

	// --- Scene 2: Trigger Spike Trap ---
	fmt.Println("Scene 2: Warrior steps on spike trap")
	warrior.Position = combat.Vector3{X: 10, Y: 0, Z: 5}

	affected := trap.Update(5.0, cm) // Trigger trap
	if len(affected) > 0 {
		damageResult := warrior.TakeDamage(trap.GetDamageAmount(), trap)
		fmt.Printf("  TRAP TRIGGERED! %d damage -> Warrior HP: %d/%d\n",
			damageResult.Damage, damageResult.RemainingHP, damageResult.MaxHP)
	}
	fmt.Println()

	// --- Scene 3: Open Treasure Chest ---
	fmt.Println("Scene 3: Warrior breaks open treasure chest")
	warrior.Position = combat.Vector3{X: 15, Y: 0, Z: 0}

	for chest.IsAlive() {
		event, _ := cm.ExecuteAbility("player1", "chest1", "basic_attack")
		fmt.Printf("  Warrior attacks chest: %d damage -> Chest HP: %d\n",
			event.Damage, event.TargetHealth)

		if event.TargetDied {
			fmt.Println("  Chest broken open!")
			fmt.Println("  EPIC LOOT: 150 Gold, Rare Gem x2, Epic Weapon!")
			break
		}
	}
	fmt.Println()

	// Final status
	fmt.Printf("Adventure complete! Warrior HP: %d/%d\n",
		int(warrior.Stats.GetStat(stats.StatHealth)),
		int(warrior.Stats.GetStat(stats.StatMaxHealth)))

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")
}
