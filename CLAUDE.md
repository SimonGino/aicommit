# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Build
go build -o aicommit ./cmd/aicommit

# Run tests
go test ./...

# Run a single test
go test -run TestName ./internal/ai/

# Format code (required before commit)
gofmt -s -w .

# Lint (CI uses golangci-lint)
golangci-lint run --timeout=5m
```

## Architecture

This is an AI-powered Git commit message generator written in Go.

### Module Structure

```
cmd/aicommit/main.go     # CLI entry point using urfave/cli/v2
internal/
├── ai/provider.go       # AI provider interface + OpenAI/Azure implementations
├── config/config.go     # JSON config in ~/.config/aicommit/config.json
├── git/git.go           # Git operations wrapper (exec.Command based)
└── interactive/prompt.go # Terminal UI with single-key selection
```

### Key Patterns

- **Provider Pattern**: `ai.Provider` interface abstracts AI backends (OpenAI, Azure OpenAI)
- **Repository Pattern**: `git.Repository` encapsulates all git operations
- Config stored as JSON at `~/.config/aicommit/config.json`

### CLI Commands

- Default: Interactive commit message generation
- `config`: Set API keys, provider, model, language
- `check`: Verify configuration and API connectivity
- `report`: Generate work reports from commit history

### Version Injection

Version info is injected via GoReleaser ldflags at build time. See `.goreleaser.yaml`.
