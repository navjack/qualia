# Golang Mind Simulation - Qualia

This project simulates a simplified model of a mind, exploring concepts of internal thought generation, focus, reflection (increasing clarity), and the effort of external expression. It has evolved to support multiple interacting entities and an observer dashboard.

## Core Concepts (per Entity)

Each entity in the simulation possesses its own `MindContext`:

*   **Entity ID**: A unique identifier (e.g., "Player-1", "AI-Alpha").
*   **Type**: Player or AI.
*   **Energy**: Mental energy required for actions. Replenishes over time or with `recharge`.
*   **Thoughts**: A list of ideas or knowledge pieces generated internally.
*   **Focus**: The currently selected thought being actively worked on.
*   **Clarity**: A measure (0.0 to 1.0) of how well-understood or refined the focused thought is. Increased through introspection.
*   **ExpressionThreshold**: The minimum clarity a thought needs to be successfully expressed.
*   **States**: Entities operate in different cognitive modes:
    *   `Idle`: A base state for recharging and transitioning.
    *   `Thinking`: For generating new thoughts and choosing a thought to focus on.
    *   `Reflecting`: For introspecting on a focused thought to increase its clarity.
    *   `Acting`: For attempting to express a focused thought externally.

## Key Features

*   **Multi-Entity Simulation**: The simulation now runs with multiple entities (currently one Player and one AI), each with their own independent mind and state.
*   **Player Autopilot Mode**: The Player entity can be toggled into an "autopilot" mode. In this mode, the simulation makes decisions for the player, allowing for a passive observation experience.
*   **Global Dashboard**: When autopilot is enabled for the player, a text-based dashboard is displayed in the terminal. This dashboard provides a real-time overview of:
    *   The current state, energy levels, thought count, focused thought, and clarity for all entities.
    *   A log of recent significant events (e.g., thought generation, state changes, actions taken).
*   **Enhanced Event Logging**: The system logs more detailed events, offering better insight into the internal workings and interactions of the entities.

## Autopilot Dashboard Preview

When `autopilot` mode is enabled for the player, the terminal displays a global dashboard that updates in real-time. Here's an example of what it looks like:

```text
====== Qualia Simulation Dashboard (Observer Mode) ======
Current Time: 20:39:13 | Player Autopilot: ENABLED
------------------------------------------------------------
| Player-1   (Player) | State: Idle         
| Energy:  70/100 [■■■■■■■■■■■■■■------] | Thoughts:  0 
| Focus:  'None'                    | Clarity: ---  [--------------------] 
------------------------------------------------------------
| AI-Alpha   (AI    ) | State: Thinking     
| Energy:  55/100 [■■■■■■■■■■■---------] | Thoughts:  1 
| Focus:  'None'                    | Clarity: ---  [--------------------] 
------------------------------------------------------------

Recent Events:
  [20:39:13] AI-Alpha generated thought: 'consciousness is a complex phenomenon'.
  [20:39:13] Player-1 transitioned to Idle from Acting.
  [20:39:12] AI-Alpha started thinking.
  [20:39:12] Player-1 prepared to act.
  [20:39:11] AI AI-Alpha decides to do nothing this turn.
  [20:39:10] AI-Alpha transitioned to Idle from Acting.
  [20:39:09] AI-Alpha prepared to act.
===========================================================
```

## How to Run

1.  Ensure you have Go installed on your machine.
2.  Save the `main.go` and `states.go` files in the same directory.
3.  Navigate to the project directory in your terminal.
4.  Run the following command to start the simulation:

    ```bash
    go run main.go states.go
    ```

5.  Follow the prompts. If player autopilot is off, you will interact directly with your entity. If on, the dashboard will appear.

## Commands

### Global Commands
These commands are generally available to the player when autopilot is OFF:

*   `autopilot`: Toggles the player's autopilot mode on or off.
    *   When **ON**: The simulation takes over player decisions, and the Global Dashboard is displayed, updating in real-time.
    *   When **OFF**: You control the player entity directly, and the dashboard is not shown.
*   `view`: Display the current status (Energy, Thoughts, Focus, Clarity) of your player entity. (Only available/relevant when autopilot is OFF).
*   `quit`: Exit the simulation.

### State-Specific Commands (for Player when Autopilot is OFF, and for AI logic)
The following commands are used by entities to navigate their cognitive processes. When playing manually, these are the inputs you'll use. The AI and the player on autopilot also use this underlying command structure.

#### Idle State
*   `think`: Transition to the Thinking state.
*   `reflect`: Transition to the Reflecting state.
*   `act`: Transition to the Acting state.
*   `recharge`: Replenish some energy.

#### Thinking State
*   `generate`: Create a new random thought (costs energy).
*   `focus <index>`: Focus on a thought from the list by its index (e.g., `focus 0`). Costs a small amount of energy.
*   `idle`: Return to the Idle state.

#### Reflecting State
*   `introspect`: Increase the clarity of the currently focused thought (costs energy).
*   `unfocus`: Stop focusing on the current thought and clear clarity. (Note: `unfocus` is now primarily handled by focusing on another thought or changing state if no thought is focused; an explicit `unfocus` command within Reflecting might not be active or necessary in the current `states.go` implementation but the concept remains).
*   `idle`: Return to the Idle state.

#### Acting State
*   `express`: Attempt to express the currently focused thought. Success depends on its clarity meeting the `ExpressionThreshold` (costs energy).
*   `idle`: Return to the Idle state.
