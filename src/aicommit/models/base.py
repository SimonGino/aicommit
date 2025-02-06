from abc import ABC, abstractmethod
from typing import List
import re

from pydantic import BaseModel

from .config import AIConfig
from .prompts import COMMIT_TYPES


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
        prompts = {
            "en": f"""Generate a concise and descriptive git commit message for the following changes:

Branch: {commit_info.branch_name}

Files changed:
{chr(10).join(f"- {file}" for file in commit_info.files_changed)}

Changes:
{commit_info.diff_content}

Format the response as:
<title>: Brief description (max 50 chars)
<body>: Detailed explanation if needed (optional)

Focus on the purpose and impact of the changes.""",
            "zh-CN": f"""请为以下更改生成简洁且描述性的git提交信息：

分支：{commit_info.branch_name}

更改的文件：
{chr(10).join(f"- {file}" for file in commit_info.files_changed)}

更改内容：
{commit_info.diff_content}

请按以下格式回复：
<标题>：简要描述（最多50个字符）
<正文>：需要时可添加详细说明（可选）

请重点关注更改的目的和影响。""",
            "zh-TW": f"""請為以下更改生成簡潔且描述性的git提交信息：

分支：{commit_info.branch_name}

更改的文件：
{chr(10).join(f"- {file}" for file in commit_info.files_changed)}

更改內容：
{commit_info.diff_content}

請按以下格式回覆：
<標題>：簡要描述（最多50個字符）
<正文>：需要時可添加詳細說明（可選）

請重點關注更改的目的和影響。"""
        }
        return prompts[self.config.language]

    def _parse_commit_message(self, response: str) -> CommitMessage:
        # Extract the first line as title
        lines = response.strip().split('\n')
        title = lines[0].strip()
        
        # Get valid commit types for current language
        valid_types = '|'.join(COMMIT_TYPES[self.config.language].keys())
        
        # Validate commit message format
        pattern = f'^({valid_types})(\([^)]+\))?: .+'
        if not re.match(pattern, title):
            raise ValueError(f"Invalid commit message format. Title must match pattern: {pattern}")
            
        # Join remaining lines as body if they exist
        body = '\n'.join(lines[1:]).strip() if len(lines) > 1 else None
        
        return CommitMessage(title=title, body=body)