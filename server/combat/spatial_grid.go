package combat

import (
	"math"
	"sync"
)

// GridCell represents a cell coordinate in the spatial grid
type GridCell struct {
	X int
	Z int
}

// SpatialGrid divides the world into cells for fast spatial queries
type SpatialGrid struct {
	cellSize float64
	cells    map[GridCell][]CombatEntity
	mu       sync.RWMutex
}

// NewSpatialGrid creates a new spatial grid
func NewSpatialGrid(cellSize float64) *SpatialGrid {
	return &SpatialGrid{
		cellSize: cellSize,
		cells:    make(map[GridCell][]CombatEntity),
	}
}

// GetCell converts world position to grid cell
func (sg *SpatialGrid) GetCell(pos Vector3) GridCell {
	return GridCell{
		X: int(math.Floor(pos.X / sg.cellSize)),
		Z: int(math.Floor(pos.Z / sg.cellSize)),
	}
}

// AddEntity adds an entity to the grid
func (sg *SpatialGrid) AddEntity(entity CombatEntity) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	cell := sg.GetCell(entity.GetPosition())
	sg.cells[cell] = append(sg.cells[cell], entity)
}

// RemoveEntity removes an entity from the grid
func (sg *SpatialGrid) RemoveEntity(entity CombatEntity) {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	cell := sg.GetCell(entity.GetPosition())
	entities := sg.cells[cell]

	// Find and remove entity
	for i, e := range entities {
		if e.GetID() == entity.GetID() {
			sg.cells[cell] = append(entities[:i], entities[i+1:]...)
			break
		}
	}

	// Clean up empty cells
	if len(sg.cells[cell]) == 0 {
		delete(sg.cells, cell)
	}
}

// UpdateEntity updates an entity's position in the grid
func (sg *SpatialGrid) UpdateEntity(entity CombatEntity, oldPos, newPos Vector3) {
	oldCell := sg.GetCell(oldPos)
	newCell := sg.GetCell(newPos)

	// If cell changed, move entity
	if oldCell != newCell {
		sg.mu.Lock()
		defer sg.mu.Unlock()

		// Remove from old cell
		if entities, ok := sg.cells[oldCell]; ok {
			for i, e := range entities {
				if e.GetID() == entity.GetID() {
					sg.cells[oldCell] = append(entities[:i], entities[i+1:]...)
					break
				}
			}
			if len(sg.cells[oldCell]) == 0 {
				delete(sg.cells, oldCell)
			}
		}

		// Add to new cell
		sg.cells[newCell] = append(sg.cells[newCell], entity)
	}
}

// GetNearbyEntities returns all entities within radius of position
func (sg *SpatialGrid) GetNearbyEntities(pos Vector3, radius float64) []CombatEntity {
	sg.mu.RLock()
	defer sg.mu.RUnlock()

	// Calculate cell range to check
	minCell := sg.GetCell(Vector3{X: pos.X - radius, Y: 0, Z: pos.Z - radius})
	maxCell := sg.GetCell(Vector3{X: pos.X + radius, Y: 0, Z: pos.Z + radius})

	nearby := []CombatEntity{}
	seen := make(map[string]bool) // Prevent duplicates

	// Check all cells in range
	for x := minCell.X; x <= maxCell.X; x++ {
		for z := minCell.Z; z <= maxCell.Z; z++ {
			cell := GridCell{X: x, Z: z}
			if entities, ok := sg.cells[cell]; ok {
				for _, entity := range entities {
					// Skip duplicates
					if seen[entity.GetID()] {
						continue
					}

					// Check actual distance
					distance := pos.Distance(entity.GetPosition())
					if distance <= radius {
						nearby = append(nearby, entity)
						seen[entity.GetID()] = true
					}
				}
			}
		}
	}

	return nearby
}

// GetEntitiesInCell returns all entities in a specific cell
func (sg *SpatialGrid) GetEntitiesInCell(cell GridCell) []CombatEntity {
	sg.mu.RLock()
	defer sg.mu.RUnlock()

	if entities, ok := sg.cells[cell]; ok {
		// Return copy to avoid race conditions
		result := make([]CombatEntity, len(entities))
		copy(result, entities)
		return result
	}

	return []CombatEntity{}
}

// HasLineOfSight checks if there's a clear line between two positions
func (sg *SpatialGrid) HasLineOfSight(from, to Vector3) bool {
	sg.mu.RLock()
	defer sg.mu.RUnlock()

	// Get all cells along the ray
	cells := sg.raycastCells(from, to)

	// Check each cell for obstacles
	for _, cell := range cells {
		if entities, ok := sg.cells[cell]; ok {
			for _, entity := range entities {
				// Only check solid obstacles
				if entity.GetType() == EntityTypeDestructible {
					// Check if obstacle blocks the ray
					// Simple AABB check for now
					entPos := entity.GetPosition()
					if sg.pointNearLine(entPos, from, to, 1.0) {
						return false // Blocked!
					}
				}
			}
		}
	}

	return true // Clear shot
}

// raycastCells returns all cells along a line (Bresenham's algorithm)
func (sg *SpatialGrid) raycastCells(from, to Vector3) []GridCell {
	cells := []GridCell{}

	x0 := int(math.Floor(from.X / sg.cellSize))
	z0 := int(math.Floor(from.Z / sg.cellSize))
	x1 := int(math.Floor(to.X / sg.cellSize))
	z1 := int(math.Floor(to.Z / sg.cellSize))

	dx := abs(x1 - x0)
	dz := abs(z1 - z0)
	sx := sign(x1 - x0)
	sz := sign(z1 - z0)
	err := dx - dz

	x, z := x0, z0

	for {
		cells = append(cells, GridCell{X: x, Z: z})

		if x == x1 && z == z1 {
			break
		}

		e2 := 2 * err
		if e2 > -dz {
			err -= dz
			x += sx
		}
		if e2 < dx {
			err += dx
			z += sz
		}
	}

	return cells
}

// pointNearLine checks if a point is near a line segment
func (sg *SpatialGrid) pointNearLine(point, lineStart, lineEnd Vector3, threshold float64) bool {
	// Vector from line start to point
	dx := point.X - lineStart.X
	dz := point.Z - lineStart.Z

	// Vector from line start to end
	lx := lineEnd.X - lineStart.X
	lz := lineEnd.Z - lineStart.Z

	// Line length squared
	lineLen2 := lx*lx + lz*lz
	if lineLen2 == 0 {
		return false
	}

	// Project point onto line
	t := (dx*lx + dz*lz) / lineLen2
	t = math.Max(0, math.Min(1, t)) // Clamp to line segment

	// Nearest point on line
	nearX := lineStart.X + t*lx
	nearZ := lineStart.Z + t*lz

	// Distance from point to nearest point on line
	distX := point.X - nearX
	distZ := point.Z - nearZ
	dist := math.Sqrt(distX*distX + distZ*distZ)

	return dist <= threshold
}

// GetEntityCount returns total number of entities in grid
func (sg *SpatialGrid) GetEntityCount() int {
	sg.mu.RLock()
	defer sg.mu.RUnlock()

	count := 0
	for _, entities := range sg.cells {
		count += len(entities)
	}
	return count
}

// Clear removes all entities from the grid
func (sg *SpatialGrid) Clear() {
	sg.mu.Lock()
	defer sg.mu.Unlock()

	sg.cells = make(map[GridCell][]CombatEntity)
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func sign(x int) int {
	if x > 0 {
		return 1
	} else if x < 0 {
		return -1
	}
	return 0
}
