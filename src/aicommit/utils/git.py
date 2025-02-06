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
        raise Exception("Current directory is not a git repository. Please run 'git init' first or change to a git repository.")


def get_staged_changes(repo: Repo) -> Tuple[List[str], str]:
    """Get the staged files and their diff content."""
    # 快速检查是否有暂存的更改
    staged = repo.git.diff("--cached", "--name-only")
    if not staged:
        raise Exception("No staged changes found. Use 'git add' to stage your changes.")

    # 获取暂存文件列表
    staged_files = staged.splitlines()
    
    # 获取暂存更改的详细差异
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