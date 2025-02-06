from typing import Any, Dict, List, Optional
import asyncio
from functools import partial

from dashscope import Generation

from .base import AIProvider, CommitInfo, CommitMessage
from .prompts import SYSTEM_PROMPTS

class QwenProvider(AIProvider):
    async def generate_commit_message(self, commit_info: CommitInfo) -> CommitMessage:
        prompt = self._create_prompt(commit_info)
        response = await self._generate_response(prompt)
        return self._parse_commit_message(response)
    
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