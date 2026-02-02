# Agentic Workflow & Roles

This project was developed through an interactive session with the Gemini CLI Agent. Below is a breakdown of the "Virtual Agents" (roles) and the workflow steps that were executed to deliver the application.

## 1. The Planner Agent
*   **Objective:** Define the scope, architecture, and milestones.
*   **Output:** `docs/implementation_plan.md`
*   **Key Decisions:**
    *   Selected `bubbletea` for the TUI framework.
    *   Selected `beep` for the Audio engine (noting the `alsa-devel` requirement on Linux).
    *   Defined a 6-phase implementation roadmap to ensure steady progress.

## 2. The Infrastructure Agent
*   **Objective:** Setup the environment and dependencies.
*   **Actions:**
    *   Initialized Go module `go-subsonic` (Go 1.25).
    *   Configured `viper` for YAML configuration management.
    *   Resolved system dependencies (identified and requested `alsa-devel` for OpenSUSE).

## 3. The Backend Engineer
*   **Objective:** Implement core logic and API integration.
*   **Deliverables:**
    *   `internal/subsonic`: Custom client for Subsonic API.
        *   Authentication (Token/Salt).
        *   Browsing (switched from `getIndexes` to `getArtists` for better ID3 metadata).
        *   Stream URL generation.
    *   `internal/player`: Audio engine wrapper around `gopxl/beep`.
        *   Play, Pause, Volume controls.
        *   Implemented `ProgressStreamer` to track real-time playback position.
    *   `internal/config`: Configuration loading logic.

## 4. The Frontend Engineer (TUI)
*   **Objective:** Build and refine the interactive terminal interface.
*   **Deliverables:**
    *   `internal/app`: MVC model using the Bubble Tea framework.
    *   **Layout & Styling:**
        *   Implemented responsive multi-pane layout (Library vs. Playlist) that auto-resizes.
        *   Custom `lipgloss` styling for focused panels, bold artist names, and faint metadata.
        *   Custom `list.ItemDelegate` to render specialized row formats (e.g., "Artist **(5 Albums)**", "Track Duration").
    *   **Features:**
        *   Drill-down navigation: Artist -> Album -> Song.
        *   **Smart Navigation:** `Backspace` / `Ctrl+H` to go up a level (Artist List <-> Album List).
        *   **Playlist Management:** `Enter` to play song, `a` key to add entire album to playlist.
        *   **Playback Control:** `Space` (Pause), `+`/`-` (Volume), Real-time progress bar/timer.
        *   **Search/Filter:** Integrated list filtering, with smart `q` key handling (quits app unless typing in filter).
        *   **Metadata Display:** Added Release Year to albums and Track Numbers/Duration to songs.

## 5. The QA Agent
*   **Objective:** Verify integrity and fix bugs.
*   **Actions:**
    *   Wrote unit tests for the `subsonic` client and `app` model state.
    *   Diagnosed and fixed the `alsa` CGO build failure on OpenSUSE.
    *   Fixed TUI resizing issues by implementing strict dimension calculations.
    *   Fixed empty album counts by switching API endpoints.
    *   Refined UX issues (e.g., removing `>` prefix from selected items, fixing backspace behavior).