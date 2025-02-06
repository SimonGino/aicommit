from typing import Any, Dict, List, Optional
from openai import OpenAI

from .base import AIProvider, CommitInfo, CommitMessage
from .config import AIConfig
from .prompts import SYSTEM_PROMPTS

class OpenAIProvider(AIProvider):
    def __init__(self, config: AIConfig):
        super().__init__(config)
        self.client = OpenAI(api_key=config.api_key)

    async def generate_commit_message(self, commit_info: CommitInfo) -> CommitMessage:
        prompt = self._create_prompt(commit_info)
        response = await self._generate_response(prompt)
        return self._parse_commit_message(response)
    
    async def _generate_response(self, prompt: str) -> str:
        response = await self.client.chat.completions.create(
            model=self.config.model or "gpt-4o-mini",
            messages=[
                {"role": "system", "content": SYSTEM_PROMPTS[self.config.language]},
                {"role": "user", "content": prompt}
            ],
            max_tokens=self.config.max_tokens,
            temperature=self.config.temperature,
            stream=False
        )
        
        if not response.choices:
            raise Exception("No response from OpenAI API")
            
        return response.choices[0].message.content 