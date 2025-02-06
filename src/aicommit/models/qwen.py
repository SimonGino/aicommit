from typing import Any, Dict, List, Optional
import asyncio
from functools import partial
import re

from dashscope import Generation

from .base import AIProvider, CommitInfo, CommitMessage

COMMIT_TYPES = {
    'en': {
        'feat': 'New feature',
        'fix': 'Bug fix',
        'refactor': 'Code refactoring',
        'docs': 'Documentation changes',
        'style': 'Code style changes (formatting, missing semicolons, etc)',
        'test': 'Adding or modifying tests',
        'chore': 'Maintenance tasks, dependencies, build changes'
    },
    'zh-CN': {
        'feat': '新功能',
        'fix': '修复缺陷',
        'refactor': '代码重构',
        'docs': '文档更新',
        'style': '代码格式',
        'test': '测试相关',
        'chore': '其他更新'
    },
    'zh-TW': {
        'feat': '新功能',
        'fix': '修復缺陷',
        'refactor': '代碼重構',
        'docs': '文檔更新',
        'style': '代碼格式',
        'test': '測試相關',
        'chore': '其他更新'
    }
}

SYSTEM_PROMPTS = {
    'en': """You are a helpful assistant that generates standardized git commit messages.
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
""",
    'zh-CN': """您是一个帮助生成标准化git提交信息的助手。
请严格遵循以下提交信息格式规则：

1. 格式：<类型>(<范围>): <主题>

<正文>

<脚注>

2. 类型必须是以下之一：
- feat: 新功能
- fix: 修复缺陷
- refactor: 代码重构
- docs: 文档更新
- style: 代码格式
- test: 测试相关
- chore: 其他更新

3. 范围：可选，描述影响的区域（如：router、auth、db）
4. 主题：简短摘要（不超过50个字符）
5. 正文：详细说明（每行不超过72个字符）
6. 脚注：可选，用于说明重大变更或引用问题编号

示例：
feat(认证): 实现JWT认证系统

添加基于JWT的认证系统，支持刷新令牌
- 实现令牌生成和验证
- 添加用户会话管理
- 设置安全Cookie处理

重大变更：需要新的认证头
修复 #123
""",
    'zh-TW': """您是一個幫助生成標準化git提交信息的助手。
請嚴格遵循以下提交信息格式規則：

1. 格式：<類型>(<範圍>): <主題>

<正文>

<腳註>

2. 類型必須是以下之一：
- feat: 新功能
- fix: 修復缺陷
- refactor: 代碼重構
- docs: 文檔更新
- style: 代碼格式
- test: 測試相關
- chore: 其他更新

3. 範圍：可選，描述影響的區域（如：router、auth、db）
4. 主題：簡短摘要（不超過50個字符）
5. 正文：詳細說明（每行不超過72個字符）
6. 腳註：可選，用於說明重大變更或引用問題編號

示例：
feat(認證): 實現JWT認證系統

添加基於JWT的認證系統，支持刷新令牌
- 實現令牌生成和驗證
- 添加用戶會話管理
- 設置安全Cookie處理

重大變更：需要新的認證頭
修復 #123
"""
}

class QwenProvider(AIProvider):
    async def generate_commit_message(self, commit_info: CommitInfo) -> CommitMessage:
        prompt = self._create_prompt(commit_info)
        response = await self._generate_response(prompt)
        return self._parse_commit_message(response)
    
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
    
    async def _generate_response(self, prompt: str) -> str:
        gen = Generation()
        messages = [
            {"role": "system", "content": SYSTEM_PROMPTS[self.config.language]},
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