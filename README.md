# ol - One-liner Command Assistant

A natural language to terminal command tool that serves as an intermediate layer between users and the terminal.

🌐 **English | [简体中文](README_CN.md)**

---

## Quick Install

### Using Install Script (Recommended)

```bash
curl -sSL https://raw.githubusercontent.com/vcvcvnvcvcvn/one-liner/main/install.sh | bash
```

### Install from Source

```bash
# Build and install
go build -o ol .
sudo mv ol /usr/local/bin/
```

### Using Makefile

```bash
# Build for current platform
make build

# Install to system
make install

# Build for all platforms
make build-all
```

## Cross-Platform Compilation

This project supports cross-compilation for multiple platforms:

| Platform | Architecture | Output Filename |
|----------|-------------|-----------------|
| Linux | amd64 | `ol-linux-amd64` |
| Linux | arm64 | `ol-linux-arm64` |
| macOS | amd64 (Intel) | `ol-darwin-amd64` |
| macOS | arm64 (Apple Silicon) | `ol-darwin-arm64` |

### Manual Cross-Compilation

```bash
# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o ol-darwin-amd64 .

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o ol-darwin-arm64 .

# Linux
GOOS=linux GOARCH=amd64 go build -o ol-linux-amd64 .
```

## Usage

```bash
# Basic usage
ol <your natural language description>

# Examples
ol list all files modified in the last 24 hours
ol find files larger than 100MB in current directory
ol show git log with graph visualization

# Other commands
ol --help      # Show help
ol --init      # Reinitialize configuration
ol --version   # Show version
```

## Interactive Controls

After generating a command, you can:

| Key | Action |
|-----|--------|
| `Enter` | Execute the generated command |
| `r` | Regenerate a different command |
| `Ctrl+C` | Cancel |

When regenerating, the system remembers previously generated commands and asks the AI to provide a different implementation.

## First Time Setup

The first time you run `ol`, it will automatically enter the initialization process where you need to choose:

1. **OpenAI Compatible API** - Supports OpenAI or compatible APIs (e.g., Azure, local models)
2. **Anthropic Compatible API** - Supports Claude API

Configuration is saved to `~/.ol_config.json`.

## Configuration Examples

### OpenAI Configuration
```json
{
  "api_type": "openai",
  "api_key": "sk-...",
  "base_url": "https://api.openai.com/v1",
  "model": "gpt-4o-mini",
  "thinking": false
}
```

### Anthropic Configuration
```json
{
  "api_type": "anthropic",
  "api_key": "sk-ant-...",
  "base_url": "https://api.anthropic.com/v1/messages",
  "model": "claude-3-5-haiku-20241022",
  "thinking": false
}
```

## Features

- 🤖 Supports OpenAI and Anthropic APIs
- ⏳ Elegant spinner animation while waiting for response
- 🛡️ Displays command and waits for user confirmation before execution
- 🔄 Press `r` to regenerate a command with different style
- 🐚 Auto-detects and uses current shell to execute commands
- 🔒 Configuration file permissions set to 0600 (user-readable only)
- 🌐 Cross-platform support: Linux, macOS

## System Prompt

The tool automatically detects the user's operating system and architecture, injecting them into the system prompt to help AI generate commands more suitable for the current platform.

---

Development Environment: Go 1.22+
