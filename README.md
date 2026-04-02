# go-subsonic

A lightweight Terminal UI (TUI) Subsonic client written in Go.

> **Note:** This project has been exclusively tested with **[Navidrome](https://www.navidrome.org/)**. While it uses the standard Subsonic API, compatibility with other servers (e.g., Airsonic, Gonic) is not guaranteed.

## Features

- **Minimal and focused** on playing music in an **album-oriented way**.
- Browse music by Artist and Album.
- Search and filter your library.
- Simple playlist management.
- Audio playback with volume control and progress tracking.
- Low resource footprint.

## Configuration

The application looks for a `config.yaml` file in the current directory or at `~/.config/go-subsonic/config.yaml`.

### Sample `config.yaml`

```yaml
server:
  url: "https://your-navidrome-instance.com"
  username: "your_username"
  password: "your_password"

player:
  buffer_size: 10
```

## Linux Build Dependencies (ALSA)

Audio playback uses `gopxl/beep`, which requires ALSA headers on Linux.

Install dependencies (names vary by distro):
- **Debian/Ubuntu:** `libasound2-dev pkg-config`
- **Fedora:** `alsa-lib-devel pkgconf-pkg-config`
- **openSUSE:** `alsa-devel pkgconf-pkg-config`

## Installation

### Homebrew (macOS/Linux)

```bash
brew install rbagd/tap/go-subsonic
```

### Go Install

```bash
go install github.com/rytis/go-subsonic/cmd/go-subsonic@latest
```

Or download the latest release from the [Releases](https://github.com/rytis/go-subsonic/releases) page.

## Build/Test Without Audio

To build or test without audio support (e.g., in CI environments):

```bash
go test -tags noaudio ./...
go build -tags noaudio ./cmd/go-subsonic
```

---

*This project was essentially built by **Gemini** through an interactive agent session.*

