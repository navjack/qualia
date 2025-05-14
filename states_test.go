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
}

func TestIdleState_Transitions(t *testing.T) {
	ctx := NewMindContext()
	idle := &IdleState{}
	entityID := "testEntity"

	newState, _ := idle.HandleInput(entityID, ctx, strings.Fields("think"))
	assertStateType(t, &ThinkingState{}, newState)

	newState, _ = idle.HandleInput(entityID, ctx, strings.Fields("reflect"))
	assertStateType(t, &ReflectingState{}, newState)

	newState, _ = idle.HandleInput(entityID, ctx, strings.Fields("act"))
	assertStateType(t, &ActingState{}, newState)
}

func TestIdleState_Recharge(t *testing.T) {
	ctx := NewMindContext()
	idle := &IdleState{}
	entityID := "testEntity"

	// Test recharge from a lower value
	ctx.Energy = 50
	_, _ = idle.HandleInput(entityID, ctx, strings.Fields("recharge"))
	if ctx.Energy != 75 { // 50 + 25
		t.Errorf("IdleState Recharge: Expected energy 75, got %d", ctx.Energy)
	}

	// Test recharge doesn't exceed max
	ctx.Energy = 90
	_, _ = idle.HandleInput(entityID, ctx, strings.Fields("recharge"))
	if ctx.Energy != 100 { // 90 + 25 capped at 100
		t.Errorf("IdleState Recharge: Expected energy 100 (capped), got %d", ctx.Energy)
	}

	// Test recharge already at max
	ctx.Energy = 100
	_, _ = idle.HandleInput(entityID, ctx, strings.Fields("recharge"))
	if ctx.Energy != 100 {
		t.Errorf("IdleState Recharge: Expected energy 100 (already max), got %d", ctx.Energy)
	}
}

func TestIdleState_ViewAndUnknown(t *testing.T) {
	ctx := NewMindContext()
	idle := &IdleState{}
	entityID := "testEntity"

	// "view" should return the same state and not change context significantly for this test
	initialEnergy := ctx.Energy
	newState, _ := idle.HandleInput(entityID, ctx, strings.Fields("view"))
	assertStateType(t, idle, newState)
	if ctx.Energy != initialEnergy {
		t.Errorf("IdleState View: Energy changed from %d to %d", initialEnergy, ctx.Energy)
	}

	// Unknown command should return same state
	newState, _ = idle.HandleInput(entityID, ctx, strings.Fields("unknowncommand"))
	assertStateType(t, idle, newState)
}

// Next: ReflectingState tests

func TestReflectingState_Introspect(t *testing.T) {
	ctx := NewMindContext()
	reflecting := &ReflectingState{}
	entityID := "testEntity"
	ctx.Thoughts = append(ctx.Thoughts, "thought to reflect on")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.3
	initialEnergy := ctx.Energy

	newState, _ := reflecting.HandleInput(entityID, ctx, strings.Fields("introspect"))
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
	_, _ = reflecting.HandleInput(entityID, ctx, strings.Fields("introspect"))
	if ctx.Clarity > 1.0 {
		t.Errorf("ReflectingState Introspect: Clarity %f exceeded 1.0", ctx.Clarity)
	}
	// Check if it's exactly 1.0 if it was close enough (e.g. 0.95 + (0.15 to 0.25) -> likely >= 1.0)
	// The actual increment is 0.15 + (rand*0.1). So it can be up to 0.25.
	// Let's test a few introspects to ensure it reaches 1.0 and stops.
	ctx.Clarity = 0.8
	ctx.Energy = 100                                                           // Reset energy for multiple introspects
	_, _ = reflecting.HandleInput(entityID, ctx, strings.Fields("introspect")) // clarity becomes ~0.95-1.05
	if ctx.Clarity > 1.0 {
		ctx.Clarity = 1.0
	} // Simulate cap if first one overshot due to rand

	// If it's not 1.0 yet, one more should do it or cap it.
	if ctx.Clarity < 1.0 {
		_, _ = reflecting.HandleInput(entityID, ctx, strings.Fields("introspect"))
	}
	_, _ = reflecting.HandleInput(entityID, ctx, strings.Fields("introspect")) // Should stay 1.0 (or very close, then capped by logic)

	// The introspect logic caps clarity at 1.0 internally. So we just check that.
	if ctx.Clarity != 1.0 {
		t.Errorf("ReflectingState Introspect: Expected clarity to cap at 1.0, got %.2f", ctx.Clarity)
	}
}

func TestReflectingState_Introspect_NoFocus(t *testing.T) {
	ctx := NewMindContext()
	reflecting := &ReflectingState{}
	entityID := "testEntity"
	ctx.CurrentFocusIndex = -1 // No thought focused
	initialClarity := ctx.Clarity
	initialEnergy := ctx.Energy

	_, _ = reflecting.HandleInput(entityID, ctx, strings.Fields("introspect"))
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
	entityID := "testEntity"
	ctx.Thoughts = append(ctx.Thoughts, "thought to reflect on")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.3
	ctx.Energy = 5 // Not enough for introspect (costs 15)

	initialClarity := ctx.Clarity
	_, _ = reflecting.HandleInput(entityID, ctx, strings.Fields("introspect"))
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
	entityID := "testEntity"
	ctx.Thoughts = append(ctx.Thoughts, "thought to unfocus")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.5
	initialEnergy := ctx.Energy

	newState, _ := reflecting.HandleInput(entityID, ctx, strings.Fields("unfocus"))
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
	entityID := "testEntity"

	newState, _ := reflecting.HandleInput(entityID, ctx, strings.Fields("idle"))
	assertStateType(t, &IdleState{}, newState)
}

func TestReflectingState_UnknownCommand(t *testing.T) {
	ctx := NewMindContext()
	reflecting := &ReflectingState{}
	entityID := "testEntity"
	initialClarity := 0.4
	ctx.Clarity = initialClarity
	initialEnergy := ctx.Energy
	ctx.CurrentFocusIndex = 0

	newState, _ := reflecting.HandleInput(entityID, ctx, strings.Fields("unknownreflectingcommand"))
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

const (
	HighClarityForEvolveTest   = 0.95
	EnergyCostEvolveTest       = 50
	MaxEnergyEvolveAmountTest  = 10
	ThresholdEvolveAmountTest  = 0.05
	MinExpressionThresholdTest = 0.1
	MaxExpressionThresholdTest = 0.95
)

func setupContextForEvolve(t *testing.T) (*MindContext, *ActingState) {
	ctx := NewMindContext()
	ctx.Thoughts = append(ctx.Thoughts, "A profound thought for evolution")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = HighClarityForEvolveTest // Meets minimum clarity
	ctx.Energy = EnergyCostEvolveTest + 20 // Sufficient energy
	acting := &ActingState{}
	return ctx, acting
}

func TestActingState_Evolve_MaxEnergy_Success(t *testing.T) {
	ctx, acting := setupContextForEvolve(t)
	entityID := "evolveMaxEnergyEntity"
	initialMaxEnergy := ctx.MaxEnergy
	initialEnergy := ctx.Energy
	initialThoughtCount := len(ctx.Thoughts)

	newState, events := acting.HandleInput(entityID, ctx, strings.Fields("evolve max_energy increase"))
	assertStateType(t, acting, newState)

	if ctx.MaxEnergy != initialMaxEnergy+MaxEnergyEvolveAmountTest {
		t.Errorf("Evolve MaxEnergy: Expected MaxEnergy %d, got %d", initialMaxEnergy+MaxEnergyEvolveAmountTest, ctx.MaxEnergy)
	}
	if ctx.Energy != initialEnergy-EnergyCostEvolveTest {
		t.Errorf("Evolve MaxEnergy: Expected Energy %d, got %d", initialEnergy-EnergyCostEvolveTest, ctx.Energy)
	}
	if len(ctx.Thoughts) != initialThoughtCount-1 {
		t.Errorf("Evolve MaxEnergy: Expected thought to be consumed, thought count %d, expected %d", len(ctx.Thoughts), initialThoughtCount-1)
	}
	if ctx.CurrentFocusIndex != -1 {
		t.Errorf("Evolve MaxEnergy: Expected CurrentFocusIndex to be reset, got %d", ctx.CurrentFocusIndex)
	}
	if ctx.Clarity != 0 {
		t.Errorf("Evolve MaxEnergy: Expected Clarity to be reset, got %.2f", ctx.Clarity)
	}
	if len(events) == 0 || !strings.Contains(events[0], "EVOLVED: MaxEnergy increased") {
		t.Errorf("Evolve MaxEnergy: Expected evolution event, got %v", events)
	}
}

func TestActingState_Evolve_Threshold_Decrease_Success(t *testing.T) {
	ctx, acting := setupContextForEvolve(t)
	entityID := "evolveThresholdDecreaseEntity"
	initialThreshold := ctx.ExpressionThreshold
	initialEnergy := ctx.Energy

	newState, events := acting.HandleInput(entityID, ctx, strings.Fields("evolve threshold decrease"))
	assertStateType(t, acting, newState)

	expectedThreshold := initialThreshold - ThresholdEvolveAmountTest
	if ctx.ExpressionThreshold != expectedThreshold {
		t.Errorf("Evolve Threshold Decrease: Expected Threshold %.2f, got %.2f", expectedThreshold, ctx.ExpressionThreshold)
	}
	if ctx.Energy != initialEnergy-EnergyCostEvolveTest {
		t.Errorf("Evolve Threshold Decrease: Expected Energy %d, got %d", initialEnergy-EnergyCostEvolveTest, ctx.Energy)
	}
	if len(events) == 0 || !strings.Contains(events[0], "EVOLVED: ExpressionThreshold decreased") {
		t.Errorf("Evolve Threshold Decrease: Expected evolution event, got %v", events)
	}
}

func TestActingState_Evolve_Threshold_Increase_Success(t *testing.T) {
	ctx, acting := setupContextForEvolve(t)
	entityID := "evolveThresholdIncreaseEntity"
	initialThreshold := ctx.ExpressionThreshold
	initialEnergy := ctx.Energy

	newState, events := acting.HandleInput(entityID, ctx, strings.Fields("evolve threshold increase"))
	assertStateType(t, acting, newState)

	expectedThreshold := initialThreshold + ThresholdEvolveAmountTest
	if ctx.ExpressionThreshold != expectedThreshold {
		t.Errorf("Evolve Threshold Increase: Expected Threshold %.2f, got %.2f", expectedThreshold, ctx.ExpressionThreshold)
	}
	if ctx.Energy != initialEnergy-EnergyCostEvolveTest {
		t.Errorf("Evolve Threshold Increase: Expected Energy %d, got %d", initialEnergy-EnergyCostEvolveTest, ctx.Energy)
	}
	if len(events) == 0 || !strings.Contains(events[0], "EVOLVED: ExpressionThreshold increased") {
		t.Errorf("Evolve Threshold Increase: Expected evolution event, got %v", events)
	}
}

func TestActingState_Evolve_Threshold_Capping(t *testing.T) {
	entityID := "evolveThresholdCapEntity"

	// Test Min Capping
	ctxMin, actingMin := setupContextForEvolve(t)
	ctxMin.ExpressionThreshold = MinExpressionThresholdTest + 0.01 // Just above min
	actingMin.HandleInput(entityID, ctxMin, strings.Fields("evolve threshold decrease"))
	if ctxMin.ExpressionThreshold != MinExpressionThresholdTest {
		t.Errorf("Evolve Threshold Min Cap: Expected Threshold %.2f, got %.2f", MinExpressionThresholdTest, ctxMin.ExpressionThreshold)
	}
	// Try decreasing again, should stay at min
	ctxMin.Energy = EnergyCostEvolveTest + 20 // Replenish energy for test
	ctxMin.Clarity = HighClarityForEvolveTest // Reset clarity
	ctxMin.Thoughts = append(ctxMin.Thoughts, "another one")
	ctxMin.CurrentFocusIndex = len(ctxMin.Thoughts) - 1
	actingMin.HandleInput(entityID, ctxMin, strings.Fields("evolve threshold decrease"))
	if ctxMin.ExpressionThreshold != MinExpressionThresholdTest {
		t.Errorf("Evolve Threshold Min Cap (2nd attempt): Expected Threshold %.2f, got %.2f", MinExpressionThresholdTest, ctxMin.ExpressionThreshold)
	}

	// Test Max Capping
	ctxMax, actingMax := setupContextForEvolve(t)
	ctxMax.ExpressionThreshold = MaxExpressionThresholdTest - 0.01 // Just below max
	actingMax.HandleInput(entityID, ctxMax, strings.Fields("evolve threshold increase"))
	if ctxMax.ExpressionThreshold != MaxExpressionThresholdTest {
		t.Errorf("Evolve Threshold Max Cap: Expected Threshold %.2f, got %.2f", MaxExpressionThresholdTest, ctxMax.ExpressionThreshold)
	}
	// Try increasing again, should stay at max
	ctxMax.Energy = EnergyCostEvolveTest + 20 // Replenish energy
	ctxMax.Clarity = HighClarityForEvolveTest // Reset clarity
	ctxMax.Thoughts = append(ctxMax.Thoughts, "yet another")
	ctxMax.CurrentFocusIndex = len(ctxMax.Thoughts) - 1
	actingMax.HandleInput(entityID, ctxMax, strings.Fields("evolve threshold increase"))
	if ctxMax.ExpressionThreshold != MaxExpressionThresholdTest {
		t.Errorf("Evolve Threshold Max Cap (2nd attempt): Expected Threshold %.2f, got %.2f", MaxExpressionThresholdTest, ctxMax.ExpressionThreshold)
	}
}

func TestActingState_Evolve_FailureConditions(t *testing.T) {
	entityID := "evolveFailEntity"

	// Not enough arguments
	ctxArgs, actingArgs := setupContextForEvolve(t)
	_, eventsArgs := actingArgs.HandleInput(entityID, ctxArgs, strings.Fields("evolve max_energy"))
	if len(eventsArgs) == 0 || !strings.Contains(eventsArgs[0], "Usage: evolve") {
		t.Errorf("Evolve Fail Args: Expected usage message event, got %v", eventsArgs)
	}

	// No focused thought
	ctxNoFocus, actingNoFocus := setupContextForEvolve(t)
	initialMaxEnergyNF := ctxNoFocus.MaxEnergy
	ctxNoFocus.CurrentFocusIndex = -1
	_, eventsNoFocus := actingNoFocus.HandleInput(entityID, ctxNoFocus, strings.Fields("evolve max_energy increase"))
	if ctxNoFocus.MaxEnergy != initialMaxEnergyNF {
		t.Errorf("Evolve Fail NoFocus: MaxEnergy should not change, got %d", ctxNoFocus.MaxEnergy)
	}
	if len(eventsNoFocus) == 0 || !strings.Contains(eventsNoFocus[0], "without a deeply focused thought") {
		t.Errorf("Evolve Fail NoFocus: Expected no focus message event, got %v", eventsNoFocus)
	}

	// Clarity too low
	ctxLowClarity, actingLowClarity := setupContextForEvolve(t)
	initialMaxEnergyLC := ctxLowClarity.MaxEnergy
	ctxLowClarity.Clarity = HighClarityForEvolveTest - 0.1
	_, eventsLowClarity := actingLowClarity.HandleInput(entityID, ctxLowClarity, strings.Fields("evolve max_energy increase"))
	if ctxLowClarity.MaxEnergy != initialMaxEnergyLC {
		t.Errorf("Evolve Fail LowClarity: MaxEnergy should not change, got %d", ctxLowClarity.MaxEnergy)
	}
	if len(eventsLowClarity) == 0 || !strings.Contains(eventsLowClarity[0], "not high enough") {
		t.Errorf("Evolve Fail LowClarity: Expected low clarity message event, got %v", eventsLowClarity)
	}

	// Energy too low
	ctxLowEnergy, actingLowEnergy := setupContextForEvolve(t)
	initialMaxEnergyLE := ctxLowEnergy.MaxEnergy
	ctxLowEnergy.Energy = EnergyCostEvolveTest - 1
	_, eventsLowEnergy := actingLowEnergy.HandleInput(entityID, ctxLowEnergy, strings.Fields("evolve max_energy increase"))
	if ctxLowEnergy.MaxEnergy != initialMaxEnergyLE {
		t.Errorf("Evolve Fail LowEnergy: MaxEnergy should not change, got %d", ctxLowEnergy.MaxEnergy)
	}
	if len(eventsLowEnergy) == 0 || !strings.Contains(eventsLowEnergy[0], "Not enough energy") {
		t.Errorf("Evolve Fail LowEnergy: Expected low energy message event, got %v", eventsLowEnergy)
	}

	// Invalid parameter
	ctxInvalidParam, actingInvalidParam := setupContextForEvolve(t)
	initialMaxEnergyIP := ctxInvalidParam.MaxEnergy
	_, eventsInvalidParam := actingInvalidParam.HandleInput(entityID, ctxInvalidParam, strings.Fields("evolve unknown_param increase"))
	if ctxInvalidParam.MaxEnergy != initialMaxEnergyIP {
		t.Errorf("Evolve Fail InvalidParam: MaxEnergy should not change, got %d", ctxInvalidParam.MaxEnergy)
	}
	if len(eventsInvalidParam) == 0 || !strings.Contains(eventsInvalidParam[0], "Unknown parameter") {
		t.Errorf("Evolve Fail InvalidParam: Expected unknown parameter message event, got %v", eventsInvalidParam)
	}

	// Invalid direction for max_energy
	ctxInvalidDirME, actingInvalidDirME := setupContextForEvolve(t)
	initialMaxEnergyIDME := ctxInvalidDirME.MaxEnergy
	_, eventsInvalidDirME := actingInvalidDirME.HandleInput(entityID, ctxInvalidDirME, strings.Fields("evolve max_energy decrease"))
	if ctxInvalidDirME.MaxEnergy != initialMaxEnergyIDME {
		t.Errorf("Evolve Fail InvalidDir MaxEnergy: MaxEnergy should not change, got %d", ctxInvalidDirME.MaxEnergy)
	}
	if len(eventsInvalidDirME) == 0 || !strings.Contains(eventsInvalidDirME[0], "Invalid direction") {
		t.Errorf("Evolve Fail InvalidDir MaxEnergy: Expected invalid direction event, got %v", eventsInvalidDirME)
	}

	// Invalid direction for threshold
	ctxInvalidDirTH, actingInvalidDirTH := setupContextForEvolve(t)
	initialThresholdIDTH := ctxInvalidDirTH.ExpressionThreshold
	_, eventsInvalidDirTH := actingInvalidDirTH.HandleInput(entityID, ctxInvalidDirTH, strings.Fields("evolve threshold sideways"))
	if ctxInvalidDirTH.ExpressionThreshold != initialThresholdIDTH {
		t.Errorf("Evolve Fail InvalidDir Threshold: Threshold should not change, got %.2f", ctxInvalidDirTH.ExpressionThreshold)
	}
	if len(eventsInvalidDirTH) == 0 || !strings.Contains(eventsInvalidDirTH[0], "Invalid direction") {
		t.Errorf("Evolve Fail InvalidDir Threshold: Expected invalid direction event, got %v", eventsInvalidDirTH)
	}
}

func TestActingState_Express_Success(t *testing.T) {
	ctx := NewMindContext()
	acting := &ActingState{}
	entityID := "testEntity"
	ctx.Thoughts = append(ctx.Thoughts, "A brilliant idea")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.8 // Above threshold 0.7
	ctx.ExpressionThreshold = 0.7
	initialEnergy := ctx.Energy

	newState, _ := acting.HandleInput(entityID, ctx, strings.Fields("express"))
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
	entityID := "testEntity"
	ctx.Thoughts = append(ctx.Thoughts, "A muddled idea")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.5 // Below threshold 0.7
	ctx.ExpressionThreshold = 0.7
	initialEnergy := ctx.Energy

	acting.HandleInput(entityID, ctx, strings.Fields("express"))

	if ctx.Energy != initialEnergy { // No energy cost on failure due to low clarity
		t.Errorf("ActingState Express LowClarity: Energy changed, expected %d, got %d", initialEnergy, ctx.Energy)
	}
}

func TestActingState_Express_NoFocus(t *testing.T) {
	ctx := NewMindContext()
	acting := &ActingState{}
	entityID := "testEntity"
	ctx.CurrentFocusIndex = -1 // No thought focused
	initialEnergy := ctx.Energy

	acting.HandleInput(entityID, ctx, strings.Fields("express"))

	if ctx.Energy != initialEnergy { // No energy cost if no focus
		t.Errorf("ActingState Express NoFocus: Energy changed, expected %d, got %d", initialEnergy, ctx.Energy)
	}
}

func TestActingState_Express_NoEnergy(t *testing.T) {
	ctx := NewMindContext()
	acting := &ActingState{}
	entityID := "testEntity"
	ctx.Thoughts = append(ctx.Thoughts, "An energetic idea")
	ctx.CurrentFocusIndex = 0
	ctx.Clarity = 0.9 // Sufficient clarity
	ctx.ExpressionThreshold = 0.7
	ctx.Energy = 10 // Not enough for express (costs 20)

	acting.HandleInput(entityID, ctx, strings.Fields("express"))

	if ctx.Energy != 10 { // Energy should not change
		t.Errorf("ActingState Express NoEnergy: Expected energy 10, got %d", ctx.Energy)
	}
}

func TestActingState_TransitionToIdle(t *testing.T) {
	ctx := NewMindContext()
	acting := &ActingState{}
	entityID := "testEntity"

	newState, _ := acting.HandleInput(entityID, ctx, strings.Fields("idle"))
	assertStateType(t, &IdleState{}, newState)
}

func TestActingState_UnknownCommand(t *testing.T) {
	ctx := NewMindContext()
	acting := &ActingState{}
	entityID := "testEntity"
	initialEnergy := ctx.Energy
	ctx.CurrentFocusIndex = 0 // Assume some focus for consistent unknown behavior test
	ctx.Clarity = 0.8

	newState, _ := acting.HandleInput(entityID, ctx, strings.Fields("unknownactingcommand"))
	assertStateType(t, acting, newState)
	if ctx.Energy != initialEnergy {
		t.Errorf("ActingState Unknown: Energy changed")
	}
}

// All state tests added.
