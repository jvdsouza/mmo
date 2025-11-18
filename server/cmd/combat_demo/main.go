package main

import (
	"fmt"
	"mmo-server/examples"
	"os"
	"strings"
)

func main() {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("        MMO COMBAT SYSTEM DEMONSTRATION")
	fmt.Println("     Showcasing All Combat Entity Types & Interactions")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()

	// Check if specific example requested
	if len(os.Args) > 1 {
		runSpecificExample(os.Args[1])
		return
	}

	// Run all examples
	runAllExamples()
}

func runAllExamples() {
	examples := []struct {
		name string
		fn   func()
	}{
		{"Player vs Monster", examples.ExamplePlayerVsMonster},
		{"Player vs Player (PvP)", examples.ExamplePlayerVsPlayer},
		{"Player Destroying Crate", examples.ExamplePlayerVsCrate},
		{"Player Cutting Grass", examples.ExamplePlayerVsGrass},
		{"Environmental Damage", examples.ExampleEnvironmentalDamage},
		{"Monster vs Player (AI)", examples.ExampleMonsterVsPlayer},
		{"Spatial Grid Performance", examples.ExampleSpatialGridPerformance},
		{"Complete Combat Scenario", examples.ExampleCompleteCombatScenario},
	}

	for i, ex := range examples {
		fmt.Printf("[%d/%d] Running: %s\n\n", i+1, len(examples), ex.name)
		ex.fn()
	}

	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("                  DEMONSTRATION COMPLETE")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("Key Features Demonstrated:")
	fmt.Println("  ✓ CombatEntity interface works for ALL entity types")
	fmt.Println("  ✓ Spatial grid optimization (O(k) instead of O(n²))")
	fmt.Println("  ✓ Faction system for targeting rules")
	fmt.Println("  ✓ OCP-compliant stats system")
	fmt.Println("  ✓ Monster AI with aggro management")
	fmt.Println("  ✓ Destructible objects with tool requirements")
	fmt.Println("  ✓ Environmental hazards with DoT patterns")
	fmt.Println("  ✓ PvP combat with enable/disable flags")
	fmt.Println()
}

func runSpecificExample(example string) {
	switch example {
	case "1", "monster":
		examples.ExamplePlayerVsMonster()
	case "2", "pvp":
		examples.ExamplePlayerVsPlayer()
	case "3", "crate":
		examples.ExamplePlayerVsCrate()
	case "4", "grass":
		examples.ExamplePlayerVsGrass()
	case "5", "env":
		examples.ExampleEnvironmentalDamage()
	case "6", "ai":
		examples.ExampleMonsterVsPlayer()
	case "7", "grid":
		examples.ExampleSpatialGridPerformance()
	case "8", "complete":
		examples.ExampleCompleteCombatScenario()
	default:
		fmt.Println("Usage: combat_demo [example_number]")
		fmt.Println()
		fmt.Println("Available examples:")
		fmt.Println("  1 (monster)   - Player vs Monster Combat")
		fmt.Println("  2 (pvp)       - Player vs Player (PvP)")
		fmt.Println("  3 (crate)     - Player Destroying Crate")
		fmt.Println("  4 (grass)     - Player Cutting Grass (Tool Requirement)")
		fmt.Println("  5 (env)       - Environmental Damage (Lava)")
		fmt.Println("  6 (ai)        - Monster AI Aggro System")
		fmt.Println("  7 (grid)      - Spatial Grid Performance")
		fmt.Println("  8 (complete)  - Complete Combat Scenario")
		fmt.Println()
		fmt.Println("Run without arguments to see all examples")
	}
}
