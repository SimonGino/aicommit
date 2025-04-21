# AI Commit

English | [简体中文](README.md)

An AI-powered Git commit message generator that supports multiple AI providers (Qwen, OpenAI, DeepSeek) and automatically generates commit messages compliant with the Conventional Commits specification.

## Features

- Automatically generate standardized Git commit messages
- Support for multiple AI providers:
  - Qwen (Tongyi Qianwen)
  - OpenAI (GPT-4)
  - DeepSeek
- Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification
- Multi-language support (English, Simplified Chinese, Traditional Chinese)
- Beautiful command-line interface
- Interactive commit confirmation

## Commit Message Format

Generated commit messages strictly follow this format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Supported commit types:
- feat: New feature
- fix: Bug fix
- refactor: Code refactoring
- docs: Documentation changes
- style: Code style changes
- test: Testing related changes
- chore: Other updates

## Installation

### Quick Installation (Recommended)

#### Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/install.sh | sudo bash
```

#### Windows (Run PowerShell as Administrator)
```powershell
iwr -useb https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/install.ps1 | iex
```

### Manual Installation

1. Download the latest release package:
   - Visit the [Releases](https://github.com/SimonGino/aicommit/releases) page
   - Download the version suitable for your system

2. Extract and install:
```bash
# Linux/macOS
tar xzf aicommit_*.tar.gz
sudo mv aicommit /usr/local/bin/
chmod +x /usr/local/bin/aicommit

# Windows
# Extract the zip file and add aicommit.exe to your system PATH
```

## Configuration

Configure AI provider API keys before first use:

```bash
# Configure Qwen API
aicommit config --provider qwen --api-key your-api-key-here

# Configure OpenAI API
aicommit config --provider openai --api-key your-api-key-here

# Configure DeepSeek API
aicommit config --provider deepseek --api-key your-api-key-here
```

Set output language (optional):
```bash
aicommit config --language zh-CN  # Simplified Chinese (default)
aicommit config --language en     # English
aicommit config --language zh-TW  # Traditional Chinese
```

## Usage

1. Stage the changes you want to commit:
```bash
git add .  # or specify files
```

2. Generate a commit message:
```bash
aicommit  # use default AI provider
```

Specify an AI provider:
```bash
aicommit --provider qwen     # use Qwen
aicommit --provider openai   # use OpenAI
aicommit --provider deepseek # use DeepSeek
```

Use a custom commit message:
```bash
aicommit -m "feat(auth): add user authentication"
```

Temporarily specify output language:
```bash
aicommit -l en     # Generate commit message in English
aicommit -l zh-CN  # Generate commit message in Simplified Chinese
aicommit -l zh-TW  # Generate commit message in Traditional Chinese
aicommit -l zh     # Generate commit message in Simplified Chinese (shorthand)
```

3. Generate a commit message:
```bash
aicommit report --this-week  # Generate a commit message for the current week
aicommit report --last-week  # Generate a commit message for the last week
aicommit report --since 2023-10-01 --until 2023-10-31  # Generate a commit message for a specific date range
```

## Uninstallation

```bash
# Linux/macOS
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/uninstall.sh)"

# Windows (Run PowerShell as Administrator)
iwr -useb https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/uninstall.ps1 | iex
```

## Development

1. Clone the repository:
```bash
git clone https://github.com/SimonGino/aicommit.git
cd aicommit
```

2. Install dependencies:
```bash
go mod download
```

3. Run tests:
```bash
go test ./...
```

## Contributing

Pull requests and issues are welcome!

## License

MIT 