package combat

import (
	"encoding/json"
	"sync"
	"time"
)

// EventBroadcaster handles sending combat events to clients
type EventBroadcaster struct {
	// Callback to send messages to specific players
	sendToPlayer func(playerID string, message []byte)

	// Event batching
	pendingEvents map[string][]CombatEvent // playerID -> events
	batchInterval time.Duration
	mu            sync.RWMutex
}

// NewEventBroadcaster creates a new event broadcaster
func NewEventBroadcaster(sendFunc func(string, []byte)) *EventBroadcaster {
	eb := &EventBroadcaster{
		sendToPlayer:  sendFunc,
		pendingEvents: make(map[string][]CombatEvent),
		batchInterval: 50 * time.Millisecond, // Send every 50ms
	}

	// Start batch flusher
	go eb.flushLoop()

	return eb
}

// BroadcastEvent sends a combat event to relevant players
// Note: CombatEvent is defined in combat_manager.go
func (eb *EventBroadcaster) BroadcastEvent(event CombatEvent, playerIDs []string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// Add event to each player's pending queue
	for _, playerID := range playerIDs {
		if eb.pendingEvents[playerID] == nil {
			eb.pendingEvents[playerID] = []CombatEvent{}
		}
		eb.pendingEvents[playerID] = append(eb.pendingEvents[playerID], event)
	}
}

// SendEventImmediate sends an event immediately without batching
func (eb *EventBroadcaster) SendEventImmediate(event CombatEvent, playerIDs []string) {
	message := map[string]interface{}{
		"type": "combat_event",
		"data": event,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	for _, playerID := range playerIDs {
		eb.sendToPlayer(playerID, data)
	}
}

// flushLoop periodically sends batched events
func (eb *EventBroadcaster) flushLoop() {
	ticker := time.NewTicker(eb.batchInterval)
	defer ticker.Stop()

	for range ticker.C {
		eb.Flush()
	}
}

// Flush sends all pending events
func (eb *EventBroadcaster) Flush() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if len(eb.pendingEvents) == 0 {
		return
	}

	// Send batched events to each player
	for playerID, events := range eb.pendingEvents {
		if len(events) == 0 {
			continue
		}

		message := map[string]interface{}{
			"type": "combat_events_batch",
			"data": map[string]interface{}{
				"events":    events,
				"timestamp": time.Now(),
			},
		}

		data, err := json.Marshal(message)
		if err != nil {
			continue
		}

		eb.sendToPlayer(playerID, data)
	}

	// Clear pending events
	eb.pendingEvents = make(map[string][]CombatEvent)
}

// GetAttackAnimationHint returns animation hint for an ability
func GetAttackAnimationHint(abilityID string) string {
	hints := map[string]string{
		"basic_attack":  "swing_right",
		"heavy_attack":  "overhead_slam",
		"power_strike":  "thrust",
		"sweep_attack":  "sweep",
		"backstab":      "backstab",
	}

	if hint, ok := hints[abilityID]; ok {
		return hint
	}
	return "swing_right" // default
}
