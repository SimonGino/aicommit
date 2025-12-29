# AI Commit

English | [简体中文](README.md)

An AI-powered Git commit message generator that automatically creates commit messages following the Conventional Commits specification.

## Features

- 🤖 **AI-Powered** - Automatically analyzes code changes and generates standardized commit messages
- 🎯 **Interactive** - Keyboard shortcuts for quick operation selection
- 📁 **Flexible File Selection** - Choose from staged files, select manually, or stage all
- ✏️ **Message Editing** - Edit generated messages or regenerate them
- 🔧 **Config Check** - Built-in `check` command to verify configuration and API connectivity
- 🌍 **Multi-Language** - English, Simplified Chinese, Traditional Chinese
- ☁️ **Multi-Platform** - Supports OpenAI and Azure OpenAI
- 📊 **Daily Reports** - Generate work reports from Git commit history

## Quick Start

### Installation

```bash
# Linux/macOS
curl -fsSL https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/install.sh | sudo bash

# Windows (Run PowerShell as Administrator)
iwr -useb https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/install.ps1 | iex
```

### Configuration

```bash
# Configure OpenAI API key
aicommit config --api-key your-openai-api-key

# Verify configuration
aicommit check
```

### Usage

```bash
# Interactive commit (recommended)
aicommit

# Use custom message
aicommit -m "feat: add new feature"
```

## Interactive Flow

Running `aicommit` displays an interactive interface:

```
Detected changes:

Staged:
  ✓ src/main.go

Modified (unstaged):
  • config.json

Select an action:
  [a] Use current staged content to generate commit message
  [s] Select files to stage
  [A] Stage all changes (git add .)
  [c] Cancel

Press key to select: a

Generating commit message...

✔ Generated commit message:
┌────────────────────────────────────────────────────────────┐
│ feat(main): add user authentication                        │
│                                                            │
│ - Implement JWT token validation                           │
│ - Add user login endpoint                                  │
└────────────────────────────────────────────────────────────┘

Select an action:
  [a] Accept and commit
  [e] Edit before commit
  [r] Regenerate
  [c] Cancel

Press key to select: a

✓ Changes committed
```

## Commands

| Command | Description |
|---------|-------------|
| `aicommit` | Interactive generate and commit |
| `aicommit -m "msg"` | Commit with specified message |
| `aicommit check` | Check configuration and API connectivity |
| `aicommit config` | Configure settings |
| `aicommit report` | Generate daily report |

## Configuration

### OpenAI

```bash
aicommit config --provider openai
aicommit config --api-key sk-your-api-key
aicommit config --model gpt-4o  # optional
```

### Azure OpenAI

```bash
aicommit config --provider azure
aicommit config --api-key your-azure-key
aicommit config --base-url "https://your-resource.openai.azure.com/openai/deployments/your-deployment/chat/completions"
aicommit config --azure-api-version "2024-02-15-preview"
```

### Language Settings

```bash
aicommit config --language en     # English (default)
aicommit config --language zh-CN  # Simplified Chinese
aicommit config --language zh-TW  # Traditional Chinese
```

## Daily Reports

```bash
# This week's report
aicommit report --this-week

# Last week's report
aicommit report --last-week

# Specific date range
aicommit report --since 2024-01-01 --until 2024-01-31
```

## Commit Message Format

Follows the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>
```

Supported types: `feat` | `fix` | `refactor` | `docs` | `style` | `test` | `chore`

## Development

```bash
git clone https://github.com/SimonGino/aicommit.git
cd aicommit
go mod download
go test ./...
go build -o aicommit ./cmd/aicommit
```

## Uninstallation

```bash
# Linux/macOS
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/uninstall.sh)"

# Windows (Run PowerShell as Administrator)
iwr -useb https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/uninstall.ps1 | iex
```

## License

MIT