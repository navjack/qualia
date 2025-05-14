# Golang Mind Simulation MVP

This project simulates a simplified model of a mind, exploring concepts of internal thought generation, focus, reflection (increasing clarity), and the effort of external expression. It's inspired by discussions on consciousness, embodiment, and the challenges of actualizing internal states.

## Core Concepts

*   **MindContext**: Represents the internal world of the simulation.
    *   `Energy`: Mental energy required for actions. Replenishes over time or with `recharge`.
    *   `Thoughts`: A list of ideas or knowledge pieces generated internally.
    *   `Focus`: The currently selected thought being actively worked on.
    *   `Clarity`: A measure (0.0 to 1.0) of how well-understood or refined the focused thought is. Increased through introspection.
    *   `ExpressionThreshold`: The minimum clarity a thought needs to be successfully expressed.
*   **States**: The simulation operates in different cognitive modes:
    *   `Idle`: A base state for recharging and transitioning.
    *   `Thinking`: For generating new thoughts and choosing a thought to focus on.
    *   `Reflecting`: For introspecting on a focused thought to increase its clarity.
    *   `Acting`: For attempting to express a focused thought externally.

## How to Run

1.  Ensure you have Go installed on your machine.
2.  Save the `main.go` and `states.go` files in the same directory.
3.  Navigate to the project directory in your terminal.
4.  Run the following command to start the simulation:

    ```bash
    go run main.go states.go
    ```

5.  Follow the prompts. The simulation will display the current status and available commands for the current state.

## Commands

Global commands available from any state (unless overridden):
*   `view`: Display the current MindContext status.
*   `quit`: Exit the simulation.

### Idle State
*   `think`: Transition to the Thinking state.
*   `reflect`: Transition to the Reflecting state.
*   `act`: Transition to the Acting state.
*   `recharge`: Replenish some energy.

### Thinking State
*   `generate`: Create a new random thought (costs energy).
*   `focus <index>`: Focus on a thought from the list by its index (e.g., `focus 0`). Costs a small amount of energy.
*   `idle`: Return to the Idle state.

### Reflecting State
*   `introspect`: Increase the clarity of the currently focused thought (costs energy).
*   `unfocus`: Stop focusing on the current thought.
*   `idle`: Return to the Idle state.

### Acting State
*   `express`: Attempt to express the currently focused thought. Success depends on its clarity meeting the `ExpressionThreshold` (costs energy).
*   `idle`: Return to the Idle state.
