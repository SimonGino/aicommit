# AI Commit

English | [简体中文](README.md)

An AI-powered Git commit message generator that automatically generates commit messages following the Conventional Commits specification.

## Features

- Automatically generate standardized Git commit messages
- Support for custom API URL and model
- Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification
- Support for multiple languages (English, Simplified Chinese, Traditional Chinese)
- Beautiful command line interface
- Interactive commit confirmation
- Support for daily report generation

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
- docs: Documentation updates
- style: Code style changes
- test: Test related
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

1. Download the latest release:
   - Visit the [Releases](https://github.com/SimonGino/aicommit/releases) page
   - Choose the version suitable for your system

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

Configure the API key before first use:

```bash
# Configure API key
aicommit config --api-key your-api-key-here

# Configure custom API base URL (optional)
aicommit config --base-url https://your-custom-api-url.com/v1

# Configure custom model (optional, default is gpt-4o)
aicommit config --model gpt-4-turbo
```

Set the output language (optional):
```bash
aicommit config --language zh-CN  # Simplified Chinese
aicommit config --language en     # English (default)
aicommit config --language zh-TW  # Traditional Chinese
```

## Usage

1. Stage the changes you want to commit:
```bash
git add .  # or specify files
```

2. Generate a commit message:
```bash
aicommit
```

Use a custom commit message:
```bash
aicommit -m "feat(auth): add user authentication"
```

Temporarily specify the output language:
```bash
aicommit -l en     # generate commit message in English
aicommit -l zh-CN  # generate commit message in Simplified Chinese
aicommit -l zh-TW  # generate commit message in Traditional Chinese
aicommit -l zh     # generate commit message in Simplified Chinese (shorthand)
```

3. Use `aicommit report` to generate daily reports

   Generate work reports based on your Git commit history.

   ```bash
   # Generate report for this week (default author is current Git config)
   aicommit report --this-week

   # Generate report for last week
   aicommit report --last-week

   # Generate report for a specific date range
   aicommit report --since 2023-10-01 --until 2023-10-31

   # Generate report for a specific author
   aicommit report --this-week --author "user@example.com"
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