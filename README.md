# AI Commit

An AI-powered git commit message generator that helps you write better commit messages. Currently supports Qwen API, with planned support for OpenAI, Claude, and DeepSeek.

## Features

- Generate meaningful commit messages based on your staged changes
- Support for multiple AI providers (currently Qwen, with more coming soon)
- Easy configuration and API key management
- Beautiful CLI interface with rich formatting
- Cross-platform support (Windows, macOS, Linux)

## Installation

### Option 1: Using Installation Scripts (Recommended)

#### On macOS/Linux:
```bash
# Download the installation script
curl -O https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/install.sh

# Make it executable
chmod +x install.sh

# Run the installer (requires sudo)
sudo ./install.sh
```

#### On Windows:
1. Open PowerShell as Administrator
2. Run the following commands:
```powershell
# Download and run the installation script
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/install.ps1" -OutFile "install.ps1"
.\install.ps1
```

### Option 2: Manual Installation
1. Download the latest binary for your platform from the [releases page](https://github.com/SimonGino/aicommit/releases)
2. Add the binary to your system PATH:
   - Windows: Copy to `C:\Program Files\aicommit\`
   - macOS/Linux: Copy to `/usr/local/bin/`

### Option 3: Using pip (Python package)
```bash
pip install aicommit
```

## Configuration

Before using the tool, you need to configure your AI provider API key. Currently, Qwen API is supported:

```bash
aicommit config --provider qwen --api-key your-api-key-here
```

## Usage

1. Stage your changes using git:
```bash
git add .  # or specific files
```

2. Generate and commit with AI-generated message:
```bash
aicommit
```

You can also specify a different provider:
```bash
aicommit --provider qwen
```

Or use it with a manual message (skips AI):
```bash
aicommit -m "your message"
```

```
## Uninstallation

### On macOS/Linux:
```bash
# Download the uninstallation script
curl -O https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/uninstall.sh

# Make it executable
chmod +x uninstall.sh

# Run the uninstaller (requires sudo)
sudo ./uninstall.sh
```

### On Windows:
1. Open PowerShell as Administrator
2. Run the following commands:
```powershell
# Download and run the uninstallation script
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/SimonGino/aicommit/main/scripts/uninstall.ps1" -OutFile "uninstall.ps1"
.\uninstall.ps1
```

## Project Structure

```
aicommit/
├── src/
│   └── aicommit/
│       ├── __init__.py
│       ├── cli/          # CLI interface
│       │   ├── __init__.py
│       │   └── main.py
│       ├── models/       # Data models and AI providers
│       │   ├── __init__.py
│       │   ├── base.py
│       │   ├── config.py
│       │   └── qwen.py
│       └── utils/        # Utility functions
│           ├── __init__.py
│           └── git.py
├── scripts/             # Installation scripts
│   ├── install.sh      # Unix installation script
│   ├── uninstall.sh    # Unix uninstallation script
│   ├── install.ps1     # Windows installation script
│   └── uninstall.ps1   # Windows uninstallation script
├── tests/              # Test files
├── .github/            # GitHub configuration
│   └── workflows/      # GitHub Actions workflows
├── pyproject.toml     # Project configuration
└── README.md         # This file
```

## Development

1. Clone the repository
2. Install PDM if you haven't already:
```bash
pip install pdm
```

3. Install dependencies:
```bash
pdm install
```

4. Run the CLI:
```bash
pdm run aicommit
```

## Building

### Building Python Package
To build the package for Python distribution:

```bash
pdm build
```

### Building Binary Executable
To build a standalone binary executable:

```bash
pdm run build-binary
```

This will create a single executable file in the `dist` directory that can be run without Python installation.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT