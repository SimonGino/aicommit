from pathlib import Path
from typing import List, Tuple

import git
from git import Repo
from git.diff import Diff


def get_repo(path: Path | None = None) -> Repo:
    """Get the git repository from the current or specified path."""
    try:
        return Repo(path or Path.cwd(), search_parent_directories=True)
    except git.InvalidGitRepositoryError:
        raise Exception("Not a git repository")


def get_staged_changes(repo: Repo) -> Tuple[List[str], str]:
    """Get the staged files and their diff content."""
    if not repo.index.diff("HEAD"):
        raise Exception("No staged changes found. Use 'git add' to stage your changes.")

    # Get staged files
    staged_files = [item.a_path for item in repo.index.diff("HEAD")]
    
    # Get the diff of staged changes
    diff = repo.git.diff("--cached")
    
    return staged_files, diff


def get_current_branch(repo: Repo) -> str:
    """Get the current branch name."""
    try:
        return repo.active_branch.name
    except TypeError:
        return "HEAD-detached"


def commit_changes(repo: Repo, message: str, body: str | None = None) -> None:
    """Commit the staged changes with the given message."""
    full_message = f"{message}\n\n{body}" if body else message
    repo.index.commit(full_message) 