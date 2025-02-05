from pathlib import Path
from typing import Optional
import asyncio

import typer
from rich import print as rprint
from rich.console import Console
from rich.panel import Panel

from aicommit.models.config import AIConfig, Settings
from aicommit.models.base import CommitInfo
from aicommit.models.qwen import QwenProvider
from aicommit.utils.git import commit_changes, get_current_branch, get_repo, get_staged_changes

app = typer.Typer(help="AI-powered git commit message generator")
console = Console()


def get_ai_provider(settings: Settings, provider: Optional[str] = None):
    """Get the configured AI provider."""
    provider = provider or settings.default_provider
    api_key = getattr(settings, f"{provider}_api_key")
    
    if not api_key:
        raise typer.BadParameter(f"No API key found for {provider}. Please set it using 'aicommit config --provider {provider} --api-key <your-key>'")
    
    config = AIConfig(provider=provider, api_key=api_key)
    
    if provider == "qwen":
        return QwenProvider(config)
    # Future providers will be added here
    else:
        raise typer.BadParameter(f"Provider {provider} not implemented yet")


async def async_commit(
    provider: Optional[str] = None,
    message: Optional[str] = None,
):
    """Async implementation of the commit command."""
    try:
        settings = Settings.load()
        
        repo = get_repo()
        staged_files, diff_content = get_staged_changes(repo)
        branch_name = get_current_branch(repo)
        
        if message:
            commit_changes(repo, message)
            rprint(f"[green]✓[/green] Committed changes with message: {message}")
            return
        
        with console.status("[bold blue]Generating commit message..."):
            provider_instance = get_ai_provider(settings, provider)
            commit_info = CommitInfo(
                files_changed=staged_files,
                diff_content=diff_content,
                branch_name=branch_name
            )
            commit_message = await provider_instance.generate_commit_message(commit_info)
        
        # Show the generated message
        panel = Panel(
            f"{commit_message.title}\n\n{commit_message.body or ''}",
            title="Generated Commit Message",
            border_style="blue",
        )
        console.print(panel)
        
        # Confirm and commit
        if typer.confirm("Do you want to commit with this message?", default=True):
            commit_changes(repo, commit_message.title, commit_message.body)
            rprint("[green]✓[/green] Changes committed successfully!")
        else:
            rprint("[yellow]Commit cancelled[/yellow]")
            
    except Exception as e:
        rprint(f"[red]Error:[/red] {str(e)}")
        raise typer.Exit(1)


@app.command()
def commit(
    provider: str = typer.Option(None, "--provider", "-p", help="AI provider to use"),
    message: str = typer.Option(None, "--message", "-m", help="Manual commit message (skips AI)"),
):
    """Generate and create a commit with an AI-generated message."""
    asyncio.run(async_commit(provider, message))


@app.command()
def config(
    provider: str = typer.Option(..., "--provider", "-p", help="AI provider to configure"),
    api_key: str = typer.Option(..., "--api-key", "-k", help="API key for the provider"),
):
    """Configure AI provider settings."""
    try:
        settings = Settings.load()
        settings.update_api_key(provider, api_key)
        rprint(f"[green]✓[/green] Successfully configured {provider} API key")
        rprint(f"[blue]Config file:[/blue] {settings.config_file}")
        
    except Exception as e:
        rprint(f"[red]Error:[/red] {str(e)}")
        raise typer.Exit(1)


def main():
    app()


if __name__ == "__main__":
    main() 