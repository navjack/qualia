// states_test.go
package main

import (
	"strings"
	"testing"
)

// Helper to assert the type of the returned state
func assertStateType(t *testing.T, expected State, actual State) {
	t.Helper()
	if actual.GetName() != expected.GetName() {
		t.Errorf("Expected state to be %s, but got %s", expected.GetName(), actual.GetName())
	}
}

func TestNewMindContext(t *testing.T) {
	ctx := NewMindContext()

	if ctx.Energy != 70 {
		t.Errorf("NewMindContext: Expected initial Energy 70, got %d", ctx.Energy)
	}
	if ctx.MaxEnergy != 100 {
		t.Errorf("NewMindContext: Expected initial MaxEnergy 100, got %d", ctx.MaxEnergy)
	}
	if ctx.CurrentFocusIndex != -1 {
		t.Errorf("NewMindContext: Expected initial CurrentFocusIndex -1, got %d", ctx.CurrentFocusIndex)
	}
	if len(ctx.Thoughts) != 0 {
		t.Errorf("NewMindContext: Expected initial Thoughts to be empty, got %d items", len(ctx.Thoughts))
	}
	if ctx.ExpressionThreshold != 0.7 {
		t.Errorf("NewMindContext: Expected initial ExpressionThreshold 0.7, got %f", ctx.ExpressionThreshold)
	}
	if ctx.CurrentStateName != "Idle" { // As set by NewMindContext
		t.Errorf("NewMindContext: Expected initial CurrentStateName 'Idle', got %s", ctx.CurrentStateName)
	}
}

func TestIdleState_Transitions(t *testing.T) {
	ctx := NewMindContext()
	idle := &IdleState{}

	newState := idle.HandleInput(ctx, strings.Fields("think"))
	assertStateType(t, &ThinkingState{}, newState)

	newState = idle.HandleInput(ctx, strings.Fields("reflect"))
	assertStateType(t, &ReflectingState{}, newState)

	newState = idle.HandleInput(ctx, strings.Fields("act"))
	assertStateType(t, &ActingState{}, newState)
}

func TestIdleState_Recharge(t *testing.T) {
	ctx := NewMindContext()
	idle := &IdleState{}

	// Test recharge from a lower value
	ctx.Energy = 50
	idle.HandleInput(ctx, strings.Fields("recharge"))
	if ctx.Energy != 70 { // 50 + 20
		t.Errorf("IdleState Recharge: Expected energy 70, got %d", ctx.Energy)
	}

	// Test recharge doesn't exceed max
	ctx.Energy = 90
	idle.HandleInput(ctx, strings.Fields("recharge"))
	if ctx.Energy != 100 { // 90 + 20 capped at 100
		t.Errorf("IdleState Recharge: Expected energy 100 (capped), got %d", ctx.Energy)
	}

	// Test recharge already at max
	ctx.Energy = 100
	idle.HandleInput(ctx, strings.Fields("recharge"))
	if ctx.Energy != 100 {
		t.Errorf("IdleState Recharge: Expected energy 100 (already max), got %d", ctx.Energy)
	}
}

func TestIdleState_ViewAndUnknown(t *testing.T) {
	ctx := NewMindContext()
	idle := &IdleState{}

	// "view" should return the same state and not change context significantly for this test
	initialEnergy := ctx.Energy
	newState := idle.HandleInput(ctx, strings.Fields("view"))
	assertStateType(t, idle, newState)
	if ctx.Energy != initialEnergy {
		t.Errorf("IdleState View: Energy changed from %d to %d", initialEnergy, ctx.Energy)
	}

	// Unknown command should return same state
	newState = idle.HandleInput(ctx, strings.Fields("unknowncommand"))
	assertStateType(t, idle, newState)
}

// Next: ReflectingState tests

func TestReflectingState_Introspect(t *testing.T) {
	ctx := NewMindContext()
	reflecting := &ReflectingState{}
	ctx.Thoughts = append(ctx.Thoughts, "thought to reflect on")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.3
	initialEnergy := ctx.Energy

	newState := reflecting.HandleInput(ctx, strings.Fields("introspect"))
	assertStateType(t, reflecting, newState) // Should remain in ReflectingState

	if ctx.Clarity <= 0.3 {
		t.Errorf("ReflectingState Introspect: Expected clarity to increase from 0.3, got %.2f", ctx.Clarity)
	}
	expectedEnergy := initialEnergy - 15
	if ctx.Energy != expectedEnergy {
		t.Errorf("ReflectingState Introspect: Expected energy %d, got %d", expectedEnergy, ctx.Energy)
	}

	// Test clarity capping at 1.0
	ctx.Clarity = 0.95
	ctx.Energy = 50 // Ensure enough energy
	reflecting.HandleInput(ctx, strings.Fields("introspect"))
	if ctx.Clarity > 1.0 {
		t.Errorf("ReflectingState Introspect: Clarity %f exceeded 1.0", ctx.Clarity)
	}
	// Check if it's exactly 1.0 if it was close enough (e.g. 0.95 + 0.1 = 1.05, capped to 1.0)
	// The actual increment is 0.1. So 0.95 + 0.1 = 1.05, which should be capped to 1.0.
	// If original clarity was 0.9, 0.9 + 0.1 = 1.0.
	// Let's test a few introspects to ensure it reaches 1.0 and stops.
	ctx.Clarity = 0.8
	reflecting.HandleInput(ctx, strings.Fields("introspect")) // 0.8 + 0.1 = 0.9
	reflecting.HandleInput(ctx, strings.Fields("introspect")) // 0.9 + 0.1 = 1.0
	reflecting.HandleInput(ctx, strings.Fields("introspect")) // Should stay 1.0
	if ctx.Clarity != 1.0 {
		t.Errorf("ReflectingState Introspect: Expected clarity to cap at 1.0, got %.2f", ctx.Clarity)
	}
}

func TestReflectingState_Introspect_NoFocus(t *testing.T) {
	ctx := NewMindContext()
	reflecting := &ReflectingState{}
	ctx.CurrentFocusIndex = -1 // No thought focused
	initialClarity := ctx.Clarity
	initialEnergy := ctx.Energy

	reflecting.HandleInput(ctx, strings.Fields("introspect"))
	if ctx.Clarity != initialClarity {
		t.Errorf("ReflectingState Introspect NoFocus: Clarity changed, expected %.2f, got %.2f", initialClarity, ctx.Clarity)
	}
	if ctx.Energy != initialEnergy {
		t.Errorf("ReflectingState Introspect NoFocus: Energy changed, expected %d, got %d", initialEnergy, ctx.Energy)
	}
}

func TestReflectingState_Introspect_NoEnergy(t *testing.T) {
	ctx := NewMindContext()
	reflecting := &ReflectingState{}
	ctx.Thoughts = append(ctx.Thoughts, "thought to reflect on")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.3
	ctx.Energy = 5 // Not enough for introspect (costs 15)

	initialClarity := ctx.Clarity
	reflecting.HandleInput(ctx, strings.Fields("introspect"))
	if ctx.Clarity != initialClarity {
		t.Errorf("ReflectingState Introspect NoEnergy: Clarity changed, expected %.2f, got %.2f", initialClarity, ctx.Clarity)
	}
	if ctx.Energy != 5 { // Energy should not change
		t.Errorf("ReflectingState Introspect NoEnergy: Expected energy 5, got %d", ctx.Energy)
	}
}

func TestReflectingState_Unfocus(t *testing.T) {
	ctx := NewMindContext()
	reflecting := &ReflectingState{}
	ctx.Thoughts = append(ctx.Thoughts, "thought to unfocus")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.5
	initialEnergy := ctx.Energy

	newState := reflecting.HandleInput(ctx, strings.Fields("unfocus"))
	assertStateType(t, reflecting, newState)

	if ctx.CurrentFocusIndex != -1 {
		t.Errorf("ReflectingState Unfocus: Expected CurrentFocusIndex -1, got %d", ctx.CurrentFocusIndex)
	}
	// Clarity being reset to 0 on unfocus is a design choice reflected here.
	// If it were to persist with the thought, this test would change.
	if ctx.Clarity != 0.0 {
		t.Errorf("ReflectingState Unfocus: Expected Clarity to reset to 0.0, got %.2f", ctx.Clarity)
	}
	if ctx.Energy != initialEnergy { // Unfocus costs no energy
		t.Errorf("ReflectingState Unfocus: Energy changed from %d to %d", initialEnergy, ctx.Energy)
	}
}

func TestReflectingState_TransitionToIdle(t *testing.T) {
	ctx := NewMindContext()
	reflecting := &ReflectingState{}

	newState := reflecting.HandleInput(ctx, strings.Fields("idle"))
	assertStateType(t, &IdleState{}, newState)
}

func TestReflectingState_UnknownCommand(t *testing.T) {
	ctx := NewMindContext()
	reflecting := &ReflectingState{}
	initialClarity := 0.4
	ctx.Clarity = initialClarity
	initialEnergy := ctx.Energy
	ctx.CurrentFocusIndex = 0

	newState := reflecting.HandleInput(ctx, strings.Fields("unknownreflectingcommand"))
	assertStateType(t, reflecting, newState)
	if ctx.Clarity != initialClarity {
		t.Errorf("ReflectingState Unknown: Clarity changed")
	}
	if ctx.Energy != initialEnergy {
		t.Errorf("ReflectingState Unknown: Energy changed")
	}
	if ctx.CurrentFocusIndex != 0 {
		t.Errorf("ReflectingState Unknown: Focus changed")
	}
}

// Next: ActingState tests

func TestActingState_Express_Success(t *testing.T) {
	ctx := NewMindContext()
	acting := &ActingState{}
	ctx.Thoughts = append(ctx.Thoughts, "A brilliant idea")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.8 // Above threshold 0.7
	ctx.ExpressionThreshold = 0.7
	initialEnergy := ctx.Energy

	newState := acting.HandleInput(ctx, strings.Fields("express"))
	assertStateType(t, acting, newState) // Should remain in ActingState

	expectedEnergy := initialEnergy - 20
	if ctx.Energy != expectedEnergy {
		t.Errorf("ActingState Express Success: Expected energy %d, got %d", expectedEnergy, ctx.Energy)
	}
	// Optionally, check if the thought was "consumed" or marked, if that's a feature.
	// For now, it just costs energy and prints a success message (not testable here).
}

func TestActingState_Express_LowClarity(t *testing.T) {
	ctx := NewMindContext()
	acting := &ActingState{}
	ctx.Thoughts = append(ctx.Thoughts, "A muddled idea")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.5 // Below threshold 0.7
	ctx.ExpressionThreshold = 0.7
	initialEnergy := ctx.Energy

	acting.HandleInput(ctx, strings.Fields("express"))

	if ctx.Energy != initialEnergy { // No energy cost on failure due to low clarity
		t.Errorf("ActingState Express LowClarity: Energy changed, expected %d, got %d", initialEnergy, ctx.Energy)
	}
}

func TestActingState_Express_NoFocus(t *testing.T) {
	ctx := NewMindContext()
	acting := &ActingState{}
	ctx.CurrentFocusIndex = -1 // No thought focused
	initialEnergy := ctx.Energy

	acting.HandleInput(ctx, strings.Fields("express"))

	if ctx.Energy != initialEnergy { // No energy cost if no focus
		t.Errorf("ActingState Express NoFocus: Energy changed, expected %d, got %d", initialEnergy, ctx.Energy)
	}
}

func TestActingState_Express_NoEnergy(t *testing.T) {
	ctx := NewMindContext()
	acting := &ActingState{}
	ctx.Thoughts = append(ctx.Thoughts, "An energetic idea")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.9 // Sufficient clarity
	ctx.ExpressionThreshold = 0.7
	ctx.Energy = 10 // Not enough for express (costs 20)

	acting.HandleInput(ctx, strings.Fields("express"))

	if ctx.Energy != 10 { // Energy should not change
		t.Errorf("ActingState Express NoEnergy: Expected energy 10, got %d", ctx.Energy)
	}
}

func TestActingState_TransitionToIdle(t *testing.T) {
	ctx := NewMindContext()
	acting := &ActingState{}

	newState := acting.HandleInput(ctx, strings.Fields("idle"))
	assertStateType(t, &IdleState{}, newState)
}

func TestActingState_UnknownCommand(t *testing.T) {
	ctx := NewMindContext()
	acting := &ActingState{}
	initialEnergy := ctx.Energy
	ctx.CurrentFocusIndex = 0 // Assume some focus for consistent unknown behavior test
	ctx.Clarity = 0.8

	newState := acting.HandleInput(ctx, strings.Fields("unknownactingcommand"))
	assertStateType(t, acting, newState)
	if ctx.Energy != initialEnergy {
		t.Errorf("ActingState Unknown: Energy changed")
	}
}

// All state tests added.
