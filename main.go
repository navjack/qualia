package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"
)

const MAX_EVENT_LOG_SIZE = 7 // Number of recent events to display
var eventLog []string        // Global list to store recent events

// addEventToLog adds a new event to the global event log.
func addEventToLog(event string) {
	eventLog = append(eventLog, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), event))
	if len(eventLog) > MAX_EVENT_LOG_SIZE {
		eventLog = eventLog[len(eventLog)-MAX_EVENT_LOG_SIZE:]
	}
}

// clearScreen clears the terminal.
func clearScreen() {
	fmt.Print("\033[H\033[2J") // ANSI escape code to clear screen and move cursor to top-left
}

// renderBar creates a simple text-based progress bar.
func renderBar(current, max int, length int, colorCode string) string {
	if max == 0 {
		return strings.Repeat("-", length)
	}
	filledLength := int(float64(current) / float64(max) * float64(length))
	if filledLength < 0 {
		filledLength = 0
	}
	if filledLength > length {
		filledLength = length
	}
	bar := strings.Repeat(colorCode+"■"+"\033[0m", filledLength) + strings.Repeat("-", length-filledLength)
	return bar
}

// renderGlobalDashboard displays the state of all entities and recent events.
func renderGlobalDashboard(entities []*Entity) {
	clearScreen()
	fmt.Println("====== Qualia Simulation Dashboard (Observer Mode) ======")
	fmt.Printf("Current Time: %s | Player Autopilot: ENABLED\n", time.Now().Format("15:04:05"))
	fmt.Println(strings.Repeat("-", 60))

	for _, entity := range entities {
		entityType := "AI"
		if entity.IsPlayer {
			entityType = "Player"
		}
		fmt.Printf("| %-10s (%-6s) | State: %-12s \n", entity.ID, entityType, entity.CurrentFSMState.GetName())

		energyColor := "\033[32m" // Green
		if entity.Mind.Energy < entity.Mind.MaxEnergy/3 {
			energyColor = "\033[31m" // Red
		} else if entity.Mind.Energy < entity.Mind.MaxEnergy*2/3 {
			energyColor = "\033[33m" // Yellow
		}
		fmt.Printf("| Energy: %3d/%3d [%-20s] | Thoughts: %2d \n", entity.Mind.Energy, entity.Mind.MaxEnergy, renderBar(entity.Mind.Energy, entity.Mind.MaxEnergy, 20, energyColor), len(entity.Mind.Thoughts))

		focusedThoughtStr := "None"
		clarityBarStr := renderBar(0, 100, 20, "\033[37m") // Default empty bar (white)
		clarityValStr := "---"

		if entity.Mind.CurrentFocusIndex != -1 && entity.Mind.CurrentFocusIndex < len(entity.Mind.Thoughts) {
			focusedThought := entity.Mind.Thoughts[entity.Mind.CurrentFocusIndex]
			if len(focusedThought) > 25 {
				focusedThoughtStr = focusedThought[:22] + "..."
			} else {
				focusedThoughtStr = focusedThought
			}
			clarityPercentage := int(entity.Mind.Clarity * 100)
			clarityValStr = fmt.Sprintf("%.2f", entity.Mind.Clarity)

			clarityColor := "\033[34m" // Blue
			if entity.Mind.Clarity < 0.33 {
				clarityColor = "\033[31m" // Red
			} else if entity.Mind.Clarity < 0.66 {
				clarityColor = "\033[33m" // Yellow
			}
			clarityBarStr = renderBar(clarityPercentage, 100, 20, clarityColor)
		}
		fmt.Printf("| Focus:  %-25s | Clarity: %-4s [%-20s] \n", "'"+focusedThoughtStr+"'", clarityValStr, clarityBarStr)
		fmt.Println(strings.Repeat("-", 60))
	}

	fmt.Println("\nRecent Events:")
	if len(eventLog) == 0 {
		fmt.Println("  (No events yet)")
	}
	for i := len(eventLog) - 1; i >= 0; i-- { // Display newest first
		fmt.Printf("  %s\n", eventLog[i])
	}
	fmt.Println("===========================================================")
	// No explicit prompt in dashboard mode, it just updates.
}

// displayStatus shows the relevant information about an entity's mind.
func displayStatus(entity *Entity) {
	fmt.Printf("\n--- Status for Entity %s ---\n", entity.ID)
	fmt.Printf("Energy: %d/%d\n", entity.Mind.Energy, entity.Mind.MaxEnergy)
	fmt.Printf("Current State: %s\n", entity.CurrentFSMState.GetName())
	fmt.Println("Thoughts:")
	if len(entity.Mind.Thoughts) == 0 {
		fmt.Println("  (No thoughts yet)")
	} else {
		for i, thought := range entity.Mind.Thoughts {
			if i == entity.Mind.CurrentFocusIndex {
				fmt.Printf("  [%d] * %s (Clarity: %.2f)\n", i, thought, entity.Mind.Clarity)
			} else {
				fmt.Printf("  [%d]   %s\n", i, thought)
			}
		}
	}
	if entity.Mind.CurrentFocusIndex != -1 {
		fmt.Printf("Focused Thought Index: %d\n", entity.Mind.CurrentFocusIndex)
	} else {
		fmt.Println("Focused Thought Index: None")
	}
	fmt.Println("------------------------")
}

var autoPilotEnabled bool = false // Global flag for player's autopilot mode

// selectAIAction encapsulates the decision-making logic for an automated entity.
// It can be used for both the AI and the player in autopilot mode.
func selectAIAction(entity *Entity) []string {
	var commandParts []string

	switch entity.CurrentFSMState.(type) {
	case *IdleState:
		if entity.Mind.Energy < 30 && entity.Mind.Energy < entity.Mind.MaxEnergy {
			commandParts = []string{"recharge"}
		} else if entity.Mind.Energy > 50 && rand.Intn(2) == 0 { // 50% chance to think
			commandParts = []string{"think"}
		} else if rand.Intn(3) == 0 { // Small chance to try reflecting or acting if energy is high
			if rand.Intn(2) == 0 {
				commandParts = []string{"reflect"}
			} else {
				commandParts = []string{"act"}
			}
		}
	case *ThinkingState:
		if entity.Mind.Energy > 15 && rand.Intn(2) == 0 { // 50% chance to generate
			commandParts = []string{"generate"}
		} else if len(entity.Mind.Thoughts) > 0 && entity.Mind.CurrentFocusIndex == -1 && rand.Intn(2) == 0 {
			focusIndex := rand.Intn(len(entity.Mind.Thoughts))
			commandParts = []string{"focus", fmt.Sprintf("%d", focusIndex)}
		} else { // Default to idle or try reflecting if focused
			if entity.Mind.CurrentFocusIndex != -1 && entity.Mind.Energy > 30 && rand.Intn(2) == 0 {
				commandParts = []string{"reflect"} // Chance to go reflect if focused and has energy
			} else {
				commandParts = []string{"idle"}
			}
		}
	case *ReflectingState:
		if entity.Mind.CurrentFocusIndex != -1 && entity.Mind.Energy > 20 && entity.Mind.Clarity < 0.9 && rand.Intn(2) == 0 {
			commandParts = []string{"introspect"}
		} else if entity.Mind.CurrentFocusIndex != -1 && entity.Mind.Clarity >= entity.Mind.ExpressionThreshold && entity.Mind.Energy > 30 && rand.Intn(2) == 0 {
			commandParts = []string{"act"} // Chance to go act if clarity is good
		} else {
			commandParts = []string{"idle"}
		}
	case *ActingState:
		// Constants for AI evolution decision (mirroring states.go for now)
		const HighClarityForEvolve = 0.95
		const EnergyCostEvolve = 50
		const MinExpressionThresholdForAIDecrease = 0.20 // AI won't try to decrease if already very low
		const MaxEnergySoftCapForAI = 150                // AI prioritizes evolving MaxEnergy if below this

		// Attempt to Evolve first if conditions are met
		if entity.Mind.CurrentFocusIndex != -1 &&
			entity.Mind.Clarity >= HighClarityForEvolve &&
			entity.Mind.Energy >= EnergyCostEvolve {

			if entity.Mind.MaxEnergy < MaxEnergySoftCapForAI {
				commandParts = []string{"evolve", "max_energy", "increase"}
			} else if entity.Mind.ExpressionThreshold > MinExpressionThresholdForAIDecrease {
				commandParts = []string{"evolve", "threshold", "decrease"}
			}
		}

		// If AI didn't choose to evolve, consider expressing or idling
		if len(commandParts) == 0 {
			if entity.Mind.CurrentFocusIndex != -1 && entity.Mind.Energy > 25 && entity.Mind.Clarity >= entity.Mind.ExpressionThreshold && rand.Intn(2) == 0 {
				commandParts = []string{"express"}
			} else {
				commandParts = []string{"idle"}
			}
		}
	}
	return commandParts
}

// SerializableEntityState represents an entity's state for serialization.
type SerializableEntityState struct {
	ID                  string       `json:"id"`
	IsPlayer            bool         `json:"is_player"`
	Mind                *MindContext `json:"mind"` // MindContext is from states.go but used here
	CurrentFSMStateName string       `json:"current_fsm_state_name"`
}

// SimulationState represents the simulation state for serialization.
type SimulationState struct {
	Entities         []SerializableEntityState `json:"entities"`
	EventLog         []string                  `json:"event_log"`
	AutoPilotEnabled bool                      `json:"auto_pilot_enabled"`
	// Potentially add RNG state if deep determinism is needed, for now skipping.
}

// getStateByName converts a state name string to a State interface instance.
// This is crucial for restoring FSM states during loading.
func getStateByName(name string) State {
	switch name {
	case "Idle":
		return &IdleState{}
	case "Thinking":
		return &ThinkingState{}
	case "Reflecting":
		return &ReflectingState{}
	case "Acting":
		return &ActingState{}
	default:
		fmt.Printf("Warning: Unknown state name '%s' during load. Defaulting to IdleState.\n", name)
		return &IdleState{} // Fallback, should ideally not happen with valid save files
	}
}

// saveGame saves the current simulation state to a file.
func saveGame(filename string, entities []*Entity, currentEventLog []string, autopilot bool) error {
	simulationState := SimulationState{
		Entities:         make([]SerializableEntityState, len(entities)),
		EventLog:         currentEventLog,
		AutoPilotEnabled: autopilot,
	}

	for i, entity := range entities {
		simulationState.Entities[i] = SerializableEntityState{
			ID:                  entity.ID,
			IsPlayer:            entity.IsPlayer,
			Mind:                entity.Mind, // Revert to direct assignment
			CurrentFSMStateName: entity.CurrentFSMState.GetName(),
		}
	}

	data, err := json.MarshalIndent(simulationState, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}

// loadGame loads the simulation state from a file.
// It returns the loaded entities, event log, autopilot status, and any error encountered.
func loadGame(filename string) ([]*Entity, []string, bool, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, false, err
	}

	var simulationState SimulationState
	err = json.Unmarshal(data, &simulationState)
	if err != nil {
		return nil, nil, false, err
	}

	entities := make([]*Entity, len(simulationState.Entities))
	for i, entityState := range simulationState.Entities {
		entities[i] = &Entity{
			ID:              entityState.ID,
			IsPlayer:        entityState.IsPlayer,
			Mind:            entityState.Mind, // Revert to direct assignment
			CurrentFSMState: getStateByName(entityState.CurrentFSMStateName),
		}
	}

	return entities, simulationState.EventLog, simulationState.AutoPilotEnabled, nil
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Initialize random seed

	player := &Entity{
		ID:              "Player-1",
		IsPlayer:        true,
		Mind:            NewMindContext(),
		CurrentFSMState: &IdleState{}, // Start in Idle state
	}

	aiEntity := &Entity{
		ID:              "AI-Alpha",
		IsPlayer:        false,
		Mind:            NewMindContext(),
		CurrentFSMState: &IdleState{}, // AI also starts in Idle state
	}

	entities := []*Entity{player, aiEntity}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Mind Simulation MVP - Endless Mode with Entities")
	fmt.Println("Type 'quit' to exit.")
	fmt.Println("Type 'autopilot' to toggle player's automatic mode.")
	fmt.Println("Type 'save <filename.json>' to save the game.")
	fmt.Println("Type 'load <filename.json>' to load the game.")

	for {
		for _, currentEntity := range entities {
			// Passive energy regeneration for all entities
			if currentEntity.Mind.Energy < currentEntity.Mind.MaxEnergy {
				currentEntity.Mind.Energy++
			}

			if currentEntity.IsPlayer {
				var parts []string
				var command string

				if autoPilotEnabled {
					fmt.Printf("\n--- Player %s's turn (AUTOPILOT ACTIVE) (%s) ---\n", currentEntity.ID, currentEntity.CurrentFSMState.GetName())
					parts = selectAIAction(currentEntity)
					if len(parts) > 0 {
						command = parts[0]
						msg := fmt.Sprintf("Player %s (autopilot) attempts: %s", currentEntity.ID, strings.Join(parts, " "))
						fmt.Println(msg) // Still print attempt for clarity even in autopilot before dashboard refresh
						// addEventToLog(msg) // Event added by HandleInput wrapper later
					} else {
						fmt.Printf("Player %s (autopilot) decides to do nothing this turn.\n", currentEntity.ID)
						// No state change or input to handle, so continue entity loop
						// but need to decide if we skip the dashboard update for this entity's no-action.
						// For now, we'll let the main loop handle dashboard update.
						continue
					}
				} else { // Manual player input
					fmt.Printf("\n%s\n", currentEntity.CurrentFSMState.GetPrompt(currentEntity))
					fmt.Print("> ")
					input, _ := reader.ReadString('\n')
					input = strings.TrimSpace(input)
					parts = strings.Fields(input)

					if len(parts) == 0 {
						continue
					}
					command = parts[0]
				}

				if command == "quit" {
					fmt.Println("Exiting simulation.")
					return
				}
				if command == "view" {
					displayStatus(currentEntity)
					continue // viewing doesn't change state or end turn
				}
				if command == "autopilot" {
					autoPilotEnabled = !autoPilotEnabled
					if autoPilotEnabled {
						fmt.Println("Player autopilot ENABLED.")
					} else {
						fmt.Println("Player autopilot DISABLED.")
					}
					continue
				}

				if command == "save" {
					if len(parts) < 2 {
						fmt.Println("Usage: save <filename.json>")
						continue
					}
					filename := parts[1]
					if err := saveGame(filename, entities, eventLog, autoPilotEnabled); err != nil {
						fmt.Printf("Error saving game: %v\n", err)
					} else {
						fmt.Printf("Game saved to %s\n", filename)
						addEventToLog(fmt.Sprintf("Game state saved to %s by %s", filename, currentEntity.ID))
					}
					continue
				}

				if command == "load" {
					if len(parts) < 2 {
						fmt.Println("Usage: load <filename.json>")
						continue
					}
					filename := parts[1]
					loadedEntities, loadedEventLog, loadedAutopilot, err := loadGame(filename)
					if err != nil {
						fmt.Printf("Error loading game: %v\n", err)
					} else {
						entities = loadedEntities
						// Need to re-assign player and aiEntity pointers if they are used directly elsewhere
						// For now, assuming entities slice is the source of truth.
						for _, e := range entities {
							if e.IsPlayer {
								player = e // Update main player reference
								break
							}
						}
						// Similarly update aiEntity if needed, or just iterate entities directly.

						eventLog = loadedEventLog
						autoPilotEnabled = loadedAutopilot
						fmt.Printf("Game loaded from %s\n", filename)
						addEventToLog(fmt.Sprintf("Game state loaded from %s by %s", filename, currentEntity.ID)) // activeEntity might be stale here if player changed
						// Force a dashboard render or prompt after load
						if autoPilotEnabled {
							renderGlobalDashboard(entities)
						} else {
							fmt.Print(player.CurrentFSMState.GetPrompt(player) + " > ")
						}
					}
					continue
				}

				newState, returnedEvents := currentEntity.CurrentFSMState.HandleInput(currentEntity.ID, currentEntity.Mind, parts)
				currentEntity.CurrentFSMState = newState

				for _, event := range returnedEvents {
					addEventToLog(event)
				}

				if !autoPilotEnabled { // Only show individual status if player is manual
					displayStatus(currentEntity)
				}
			} else { // AI Entity Logic
				if !autoPilotEnabled { // If player is manual, show AI turn details for context
					fmt.Printf("\n--- AI Entity %s's turn (%s) ---\n", currentEntity.ID, currentEntity.CurrentFSMState.GetName())
				}
				var aiCommandParts []string
				aiCommandParts = selectAIAction(currentEntity)

				if len(aiCommandParts) > 0 {
					msg := fmt.Sprintf("AI %s attempts: %s", currentEntity.ID, strings.Join(aiCommandParts, " "))
					if !autoPilotEnabled {
						fmt.Println(msg)
					}
					// addEventToLog(msg) // Event added by HandleInput wrapper later
					newState, returnedEvents := currentEntity.CurrentFSMState.HandleInput(currentEntity.ID, currentEntity.Mind, aiCommandParts)
					currentEntity.CurrentFSMState = newState

					for _, event := range returnedEvents {
						addEventToLog(event)
					}

					if !autoPilotEnabled { // Only show AI status if player is manual
						displayStatus(currentEntity)
					}
				} else {
					msg := fmt.Sprintf("AI %s decides to do nothing this turn.", currentEntity.ID)
					if !autoPilotEnabled {
						fmt.Println(msg)
					}
					addEventToLog(msg) // Log this specific inaction
				}
			}

			// Pause logic remains, but dashboard update is outside this inner entity loop if autopilot is on.
			// The sleep should happen *before* the dashboard re-render for that cycle.

		} // End of for _, currentEntity := range entities

		// After all entities have had their turn in a cycle:
		if autoPilotEnabled {
			renderGlobalDashboard(entities)
			time.Sleep(1 * time.Second) // Pause for dashboard readability
		} else {
			// If player is manual, a smaller pause between player's full turn cycles might be good too, or rely on individual turn sleeps.
			// For now, let's assume the existing per-entity sleep is sufficient when player is manual.
			// The player's prompt provides a natural pause.
		}

	}
}
