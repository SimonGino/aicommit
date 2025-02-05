from typing import Any, Dict, List, Optional
import asyncio
from functools import partial
import re

from dashscope import Generation

from .base import AIProvider, CommitInfo, CommitMessage

COMMIT_TYPES = {
    'feat': 'New feature',
    'fix': 'Bug fix',
    'refactor': 'Code refactoring',
    'docs': 'Documentation changes',
    'style': 'Code style changes (formatting, missing semicolons, etc)',
    'test': 'Adding or modifying tests',
    'chore': 'Maintenance tasks, dependencies, build changes'
}

SYSTEM_PROMPT = """You are a helpful assistant that generates standardized git commit messages.
Follow these strict rules for commit message format:

1. Format: <type>(<scope>): <subject>

<body>

<footer>

2. Types must be one of:
- feat: New feature
- fix: Bug fix
- refactor: Code refactoring
- docs: Documentation changes
- style: Code style changes
- test: Adding or modifying tests
- chore: Maintenance tasks

3. Scope: Optional, describes the affected area (e.g., router, auth, db)
4. Subject: Short summary (50 chars or less)
5. Body: Detailed explanation (72 chars per line)
6. Footer: Optional, for breaking changes or issue references

Example:
feat(auth): implement JWT authentication

Add JWT-based authentication system with refresh tokens
- Implement token generation and validation
- Add user session management
- Set up secure cookie handling

BREAKING CHANGE: New authentication headers required
Fixes #123
"""

class QwenProvider(AIProvider):
    async def generate_commit_message(self, commit_info: CommitInfo) -> CommitMessage:
        prompt = self._create_prompt(commit_info)
        response = await self._generate_response(prompt)
        return self._parse_commit_message(response)
    
    def _parse_commit_message(self, response: str) -> CommitMessage:
        # Extract the first line as title
        lines = response.strip().split('\n')
        title = lines[0].strip()
        
        # Validate commit message format
        pattern = r'^(feat|fix|refactor|docs|style|test|chore)(\([^)]+\))?: .+'
        if not re.match(pattern, title):
            raise ValueError(f"Invalid commit message format. Title must match pattern: {pattern}")
            
        # Join remaining lines as body if they exist
        body = '\n'.join(lines[1:]).strip() if len(lines) > 1 else None
        
        return CommitMessage(title=title, body=body)
    
    async def _generate_response(self, prompt: str) -> str:
        gen = Generation()
        messages = [
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": prompt}
        ]
        loop = asyncio.get_event_loop()
        response = await loop.run_in_executor(
            None,
            partial(
                gen.call,
                model=self.config.model,
                messages=messages,
                api_key=self.config.api_key,
                temperature=self.config.temperature,
                max_tokens=self.config.max_tokens,
                result_format="message"
            )
        )
        
        if response.status_code != 200:
            raise Exception(f"Qwen API error: {response.message}")
        
        if not response.output or not response.output.choices:
            raise Exception("No response from Qwen API")
            
        return response.output.choices[0].message.content