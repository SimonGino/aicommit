from typing import Any, Dict

from dashscope import Generation

from .base import AIProvider, CommitInfo, CommitMessage


class QwenProvider(AIProvider):
    async def generate_commit_message(self, commit_info: CommitInfo) -> CommitMessage:
        prompt = self._create_prompt(commit_info)
        
        response = await self._generate_response(prompt)
        
        # Parse the response
        message_parts = response.split("\n", 1)
        title = message_parts[0].strip()
        body = message_parts[1].strip() if len(message_parts) > 1 else None
        
        return CommitMessage(title=title, body=body)
    
    async def _generate_response(self, prompt: str) -> str:
        gen = Generation()
        response = await gen.call_async(
            model=self.config.model,
            prompt=prompt,
            api_key=self.config.api_key,
            temperature=self.config.temperature,
            max_tokens=self.config.max_tokens,
            result_format="message",
        )
        
        if response.status_code != 200:
            raise Exception(f"Qwen API error: {response.message}")
        
        return response.output.text 