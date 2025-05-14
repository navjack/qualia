// states.go
package main

import (
	"fmt"
	"math/rand"
	"strconv"
)

// MindContext holds the internal state of an entity's mind.
// CurrentStateName has been removed as the Entity will hold its current State object.
type MindContext struct {
	Thoughts            []string
	CurrentFocusIndex   int     // -1 if no focus
	Clarity             float64 // 0.0 to 1.0 for the focused thought
	Energy              int
	MaxEnergy           int
	ExpressionThreshold float64
}

// NewMindContext creates and initializes a new MindContext.
func NewMindContext() *MindContext {
	return &MindContext{
		Thoughts:            make([]string, 0),
		CurrentFocusIndex:   -1,
		Clarity:             0,
		Energy:              70,
		MaxEnergy:           100,
		ExpressionThreshold: 0.7,
	}
}

// Entity represents a participant in the simulation (player or AI).
type Entity struct {
	ID              string
	IsPlayer        bool
	Mind            *MindContext
	CurrentFSMState State
}

// State defines the interface for all cognitive states.
type State interface {
	HandleInput(entityID string, context *MindContext, parts []string) (State, []string)
	GetName() string
	GetPrompt(entity *Entity) string
}

// --- IdleState ---
type IdleState struct{}

func (s *IdleState) GetName() string { return "Idle" }
func (s *IdleState) GetPrompt(entity *Entity) string {
	return fmt.Sprintf("Entity %s (Idle) | Energy: %d/%d | Commands: [think | reflect | act | recharge | view | quit]", entity.ID, entity.Mind.Energy, entity.Mind.MaxEnergy)
}

func (s *IdleState) HandleInput(entityID string, ctx *MindContext, parts []string) (State, []string) {
	command := parts[0]
	var events []string

	switch command {
	case "think":
		if ctx.Energy >= 5 { // Assuming a small cost to transition
			ctx.Energy -= 5
			events = append(events, fmt.Sprintf("%s started thinking.", entityID))
			return &ThinkingState{}, events
		} else {
			events = append(events, fmt.Sprintf("%s has not enough energy to start thinking.", entityID))
			fmt.Println("Not enough energy to transition to Thinking. Energy: ", ctx.Energy)
		}
	case "reflect":
		if ctx.Energy >= 5 {
			ctx.Energy -= 5
			events = append(events, fmt.Sprintf("%s started reflecting.", entityID))
			return &ReflectingState{}, events
		} else {
			events = append(events, fmt.Sprintf("%s has not enough energy to start reflecting.", entityID))
			fmt.Println("Not enough energy to transition to Reflecting. Energy: ", ctx.Energy)
		}
	case "act":
		if ctx.Energy >= 5 {
			ctx.Energy -= 5
			events = append(events, fmt.Sprintf("%s prepared to act.", entityID))
			return &ActingState{}, events
		} else {
			events = append(events, fmt.Sprintf("%s has not enough energy to prepare to act.", entityID))
			fmt.Println("Not enough energy to transition to Acting. Energy: ", ctx.Energy)
		}
	case "recharge":
		oldEnergy := ctx.Energy
		ctx.Energy += 25
		if ctx.Energy > ctx.MaxEnergy {
			ctx.Energy = ctx.MaxEnergy
		}
		events = append(events, fmt.Sprintf("%s recharged. Energy %d -> %d.", entityID, oldEnergy, ctx.Energy))
		fmt.Printf("Energy recharged. Current energy: %d\n", ctx.Energy)
	default:
		fmt.Println("Unknown command in Idle state.")
		events = append(events, fmt.Sprintf("%s tried unknown command '%s' in Idle.", entityID, command))
	}
	return s, events
}

// --- ThinkingState ---
type ThinkingState struct{}

func (s *ThinkingState) GetName() string { return "Thinking" }
func (s *ThinkingState) GetPrompt(entity *Entity) string {
	prompt := fmt.Sprintf("Entity %s (Thinking) | Energy: %d/%d | Thoughts: %d", entity.ID, entity.Mind.Energy, entity.Mind.MaxEnergy, len(entity.Mind.Thoughts))
	if entity.Mind.CurrentFocusIndex != -1 && entity.Mind.CurrentFocusIndex < len(entity.Mind.Thoughts) {
		prompt += fmt.Sprintf(" | Focus: '%s' (Clarity: %.2f)", entity.Mind.Thoughts[entity.Mind.CurrentFocusIndex], entity.Mind.Clarity)
	}
	prompt += " | Commands: [generate | focus <index> | idle | view | quit]"
	return prompt
}

var potentialThoughts = []string{
	"the nature of reality is elusive",
	"consciousness is a complex phenomenon",
	"embodiment shapes perception",
	"meaning is constructed, not inherent",
	"the internal world is vast",
	"externalization is a lossy process",
}

func (s *ThinkingState) HandleInput(entityID string, ctx *MindContext, parts []string) (State, []string) {
	command := parts[0]
	var events []string

	if ctx.Energy < 10 && command != "idle" {
		fmt.Println("Not enough energy to think. Try 'idle' then 'recharge'.")
		events = append(events, fmt.Sprintf("%s has low energy for thinking actions.", entityID))
		return s, events
	}

	switch command {
	case "generate":
		if ctx.Energy >= 10 {
			ctx.Energy -= 10
			newThought := potentialThoughts[rand.Intn(len(potentialThoughts))]
			ctx.Thoughts = append(ctx.Thoughts, newThought)
			events = append(events, fmt.Sprintf("%s generated thought: '%s'.", entityID, newThought))
			fmt.Printf("New thought generated: '%s'. Energy: %d\n", newThought, ctx.Energy)
		} else {
			events = append(events, fmt.Sprintf("%s failed to generate thought (low energy).", entityID))
			fmt.Println("Not enough energy to generate a new thought.")
		}
	case "focus":
		if len(parts) < 2 {
			fmt.Println("Please specify the index of the thought to focus on.")
			events = append(events, fmt.Sprintf("%s tried to focus without specifying index.", entityID))
			break
		}
		index, err := strconv.Atoi(parts[1])
		if err != nil || index < 0 || index >= len(ctx.Thoughts) {
			fmt.Println("Invalid thought index.")
			events = append(events, fmt.Sprintf("%s tried to focus on invalid index '%s'.", entityID, parts[1]))
			break
		}
		if ctx.Energy >= 5 {
			ctx.Energy -= 5
			ctx.CurrentFocusIndex = index
			ctx.Clarity = 0.1 // Initial low clarity for a newly focused thought
			events = append(events, fmt.Sprintf("%s focused on thought [%d]: '%s'. Clarity reset to %.1f.", entityID, index, ctx.Thoughts[index], ctx.Clarity))
			fmt.Printf("Focused on thought: '%s'. Clarity: %.2f. Energy: %d\n", ctx.Thoughts[index], ctx.Clarity, ctx.Energy)
		} else {
			events = append(events, fmt.Sprintf("%s failed to focus (low energy).", entityID))
			fmt.Println("Not enough energy to focus.")
		}
	case "idle":
		events = append(events, fmt.Sprintf("%s transitioned to Idle from Thinking.", entityID))
		return &IdleState{}, events
	default:
		fmt.Println("Unknown command in Thinking state.")
		events = append(events, fmt.Sprintf("%s tried unknown command '%s' in Thinking.", entityID, command))
	}
	return s, events
}

// --- ReflectingState ---
type ReflectingState struct{}

func (s *ReflectingState) GetName() string { return "Reflecting" }
func (s *ReflectingState) GetPrompt(entity *Entity) string {
	prompt := fmt.Sprintf("Entity %s (Reflecting) | Energy: %d/%d", entity.ID, entity.Mind.Energy, entity.Mind.MaxEnergy)
	if entity.Mind.CurrentFocusIndex != -1 && entity.Mind.CurrentFocusIndex < len(entity.Mind.Thoughts) {
		prompt += fmt.Sprintf(" | Focus: '%s' (Clarity: %.2f)", entity.Mind.Thoughts[entity.Mind.CurrentFocusIndex], entity.Mind.Clarity)
	} else {
		prompt += " | Focus: None"
	}
	prompt += " | Commands: [introspect | unfocus | idle | view | quit]"
	return prompt
}

func (s *ReflectingState) HandleInput(entityID string, ctx *MindContext, parts []string) (State, []string) {
	command := parts[0]
	var events []string

	if ctx.Energy < 15 && command == "introspect" {
		fmt.Println("Not enough energy to introspect. Try 'idle' then 'recharge'.")
		events = append(events, fmt.Sprintf("%s has low energy for introspection.", entityID))
		return s, events
	}

	switch command {
	case "introspect":
		if ctx.CurrentFocusIndex == -1 {
			fmt.Println("No thought is currently focused. Focus on a thought first.")
			events = append(events, fmt.Sprintf("%s tried to introspect without focus.", entityID))
			break
		}
		if ctx.Energy >= 15 {
			ctx.Energy -= 15
			ctx.Clarity += 0.15 + (rand.Float64() * 0.1) // Increase clarity, with some randomness
			if ctx.Clarity > 1.0 {
				ctx.Clarity = 1.0
			}
			focusedThought := ctx.Thoughts[ctx.CurrentFocusIndex]
			events = append(events, fmt.Sprintf("%s introspected on '%s'. Clarity now %.2f.", entityID, focusedThought, ctx.Clarity))
			fmt.Printf("Introspecting... Clarity of '%s' increased to %.2f. Energy: %d\n", focusedThought, ctx.Clarity, ctx.Energy)
		} else {
			events = append(events, fmt.Sprintf("%s failed to introspect (low energy).", entityID))
			fmt.Println("Not enough energy to introspect.")
		}
	case "unfocus":
		if ctx.CurrentFocusIndex != -1 {
			focusedThought := ctx.Thoughts[ctx.CurrentFocusIndex]
			ctx.CurrentFocusIndex = -1
			ctx.Clarity = 0
			events = append(events, fmt.Sprintf("%s unfocused from '%s'.", entityID, focusedThought))
			fmt.Printf("Unfocused from thought. Clarity reset. Energy: %d\n", ctx.Energy) // No energy cost to unfocus
		} else {
			fmt.Println("No thought is currently focused.")
			events = append(events, fmt.Sprintf("%s tried to unfocus but no thought was focused.", entityID))
		}
	case "idle":
		events = append(events, fmt.Sprintf("%s transitioned to Idle from Reflecting.", entityID))
		return &IdleState{}, events
	default:
		fmt.Println("Unknown command in Reflecting state.")
		events = append(events, fmt.Sprintf("%s tried unknown command '%s' in Reflecting.", entityID, command))
	}
	return s, events
}

// --- ActingState ---
type ActingState struct{}

func (s *ActingState) GetName() string { return "Acting" }
func (s *ActingState) GetPrompt(entity *Entity) string {
	prompt := fmt.Sprintf("Entity %s (Acting) | Energy: %d/%d", entity.ID, entity.Mind.Energy, entity.Mind.MaxEnergy)
	if entity.Mind.CurrentFocusIndex != -1 && entity.Mind.CurrentFocusIndex < len(entity.Mind.Thoughts) {
		prompt += fmt.Sprintf(" | Focus: '%s' (Clarity: %.2f)", entity.Mind.Thoughts[entity.Mind.CurrentFocusIndex], entity.Mind.Clarity)
	} else {
		prompt += " | Focus: None"
	}
	prompt += " | Commands: [express | idle | view | quit]"
	return prompt
}

func (s *ActingState) HandleInput(entityID string, ctx *MindContext, parts []string) (State, []string) {
	command := parts[0]
	var events []string

	if ctx.Energy < 20 && command == "express" {
		fmt.Println("Not enough energy to express. Try 'idle' then 'recharge'.")
		events = append(events, fmt.Sprintf("%s has low energy for expressing thoughts.", entityID))
		return s, events
	}

	switch command {
	case "express":
		if ctx.CurrentFocusIndex == -1 {
			fmt.Println("No thought is currently focused. Focus and reflect first.")
			events = append(events, fmt.Sprintf("%s tried to express without focus.", entityID))
			break
		}
		focusedThought := ctx.Thoughts[ctx.CurrentFocusIndex]
		if ctx.Clarity < ctx.ExpressionThreshold {
			msg := fmt.Sprintf("FAILED TO EXPRESS: '%s'. Clarity %.2f is below threshold %.2f.", focusedThought, ctx.Clarity, ctx.ExpressionThreshold)
			events = append(events, fmt.Sprintf("%s %s", entityID, msg))
			fmt.Printf("%s Energy: %d\n", msg, ctx.Energy)
			// No energy cost for failed expression due to low clarity
			break
		}

		if ctx.Energy >= 20 {
			ctx.Energy -= 20
			msg := fmt.Sprintf("SUCCESSFULLY EXPRESSED: '%s'!", focusedThought)
			events = append(events, fmt.Sprintf("%s %s", entityID, msg))
			fmt.Printf("%s Clarity was %.2f. Energy: %d\n", msg, ctx.Clarity, ctx.Energy)

			// Reset focus and clarity after successful expression
			ctx.CurrentFocusIndex = -1
			ctx.Clarity = 0
		} else {
			events = append(events, fmt.Sprintf("%s failed to express '%s' (low energy).", entityID, focusedThought))
			fmt.Println("Not enough energy to express the thought.")
		}

	case "idle":
		events = append(events, fmt.Sprintf("%s transitioned to Idle from Acting.", entityID))
		return &IdleState{}, events
	default:
		fmt.Println("Unknown command in Acting state.")
		events = append(events, fmt.Sprintf("%s tried unknown command '%s' in Acting.", entityID, command))
	}
	return s, events
}
