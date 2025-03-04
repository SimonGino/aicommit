from pathlib import Path
from typing import List, Tuple
import os
from contextlib import contextmanager

import git
from git import Repo
from git.diff import Diff


@contextmanager
def preserve_cwd():
    """保存并在操作后恢复当前工作目录的上下文管理器。"""
    cwd = os.getcwd()
    try:
        yield
    finally:
        os.chdir(cwd)


def get_repo(path: Path | None = None) -> Tuple[Repo, str]:
    """Get the git repository from the current or specified path."""
    try:
        # 获取当前目录
        current_path = path or os.getcwd()
        # 获取 git 仓库
        repo = Repo(current_path, search_parent_directories=True)
        # 返回仓库对象和当前工作目录
        return repo, current_path
    except git.InvalidGitRepositoryError:
        raise Exception("Current directory is not a git repository. Please run 'git init' first or change to a git repository.")


def get_unstaged_changes(repo: Repo, work_dir: str) -> Tuple[List[str], str]:
    """Get the unstaged files and their diff content."""
    # 设置 GIT_WORK_TREE 和 GIT_DIR 环境变量
    env = {
        'GIT_WORK_TREE': work_dir,
        'GIT_DIR': os.path.join(repo.working_dir, '.git')
    }
    
    # 获取未暂存的文件列表
    unstaged = repo.git.diff("--name-only", env=env)
    unstaged_files = unstaged.splitlines() if unstaged else []
    
    # 获取未暂存更改的详细差异
    diff = repo.git.diff(env=env) if unstaged else ""
    
    return unstaged_files, diff


def get_staged_changes(repo: Repo, work_dir: str) -> Tuple[List[str], str]:
    """Get the staged files and their diff content."""
    # 设置 GIT_WORK_TREE 和 GIT_DIR 环境变量
    env = {
        'GIT_WORK_TREE': work_dir,
        'GIT_DIR': os.path.join(repo.working_dir, '.git')
    }
    
    # 获取暂存的文件列表
    staged = repo.git.diff("--cached", "--name-only", env=env)
    staged_files = staged.splitlines() if staged else []
    
    # 获取暂存更改的详细差异
    diff = repo.git.diff("--cached", env=env) if staged else ""
    
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