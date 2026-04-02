# Agentic Workflow & Roles

This project was developed through an interactive session with the Gemini CLI Agent. Below is a breakdown of the "Virtual Agents" (roles) and the workflow steps that were executed to deliver the application.

## 1. The Planner Agent
*   **Objective:** Define the scope, architecture, and milestones.
*   **Key Decisions:**
    *   Selected `bubbletea` for the TUI framework.
    *   Selected `beep` for the Audio engine (noting the `alsa-devel` requirement on Linux).
    *   Defined a 6-phase implementation roadmap to ensure steady progress.

## 2. The Infrastructure Agent
*   **Objective:** Setup the environment and dependencies.
*   **Actions:**
    *   Initialized Go module `go-subsonic` (Go 1.25).
    *   Configured `viper` for YAML configuration management.
    *   Set up GitHub Actions for automated releases.

## 3. The Backend Engineer
*   **Objective:** Implement core logic and API integration.
*   **Deliverables:**
    *   `internal/subsonic`: Custom client for Subsonic API (Token/Salt Auth).
    *   `internal/player`: Audio engine wrapper around `gopxl/beep` with volume and progress tracking.
    *   `internal/config`: Configuration loading using `viper`.

## 4. The Frontend Engineer (TUI)
*   **Objective:** Build and refine the interactive terminal interface.
*   **Deliverables:**
    *   `internal/app`: MVC model using the Bubble Tea framework.
    *   **Layout & Styling:** Responsive multi-pane layout using `lipgloss`.
    *   **Features:** Navigation (Artist -> Album -> Song), Search/Filter, Playlist Management, Playback Control.

## 5. The QA Agent
*   **Objective:** Verify integrity and fix bugs.
*   **Actions:**
    *   Wrote unit tests for core components.
    *   Diagnosed and fixed cross-platform build issues.
    *   Refined UX based on manual testing.
