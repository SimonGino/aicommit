from abc import ABC, abstractmethod
from typing import List

from pydantic import BaseModel

from .config import AIConfig


class CommitInfo(BaseModel):
    files_changed: List[str]
    diff_content: str
    branch_name: str


class CommitMessage(BaseModel):
    title: str
    body: str | None = None


class AIProvider(ABC):
    def __init__(self, config: AIConfig):
        self.config = config

    @abstractmethod
    async def generate_commit_message(self, commit_info: CommitInfo) -> CommitMessage:
        """Generate a commit message based on the changes."""
        pass

    def _create_prompt(self, commit_info: CommitInfo) -> str:
        return f"""Generate a concise and descriptive git commit message for the following changes:

Branch: {commit_info.branch_name}

Files changed:
{chr(10).join(f"- {file}" for file in commit_info.files_changed)}

Changes:
{commit_info.diff_content}

Format the response as:
<title>: Brief description (max 50 chars)
<body>: Detailed explanation if needed (optional)

Focus on the purpose and impact of the changes.""" 