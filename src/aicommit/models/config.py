from pathlib import Path
from typing import Literal, Optional
import sys

if sys.version_info >= (3, 11):
    import tomllib
else:
    import tomli as tomllib

from pydantic import BaseModel, Field


class AIConfig(BaseModel):
    provider: Literal["qwen", "openai", "claude", "deepseek"] = "qwen"
    api_key: str
    model: str = Field(default="qwen-max")
    temperature: float = Field(default=0.7, ge=0.0, le=1.0)
    max_tokens: int = Field(default=500, gt=0)
    language: Literal["en", "zh-CN", "zh-TW"] = "en"


class Settings(BaseModel):
    qwen_api_key: Optional[str] = None
    openai_api_key: Optional[str] = None
    claude_api_key: Optional[str] = None
    deepseek_api_key: Optional[str] = None
    default_provider: Literal["qwen", "openai", "claude", "deepseek"] = "qwen"
    language: Literal["en", "zh-CN", "zh-TW"] = "en"

    @property
    def config_dir(self) -> Path:
        return Path.home() / ".config" / "aicommit"

    @property
    def config_file(self) -> Path:
        return self.config_dir / "config.toml"

    def ensure_config_dir(self) -> None:
        self.config_dir.mkdir(parents=True, exist_ok=True)

    @classmethod
    def load(cls) -> "Settings":
        """Load settings from config file."""
        config_file = cls().config_file
        if not config_file.exists():
            return cls()

        try:
            with open(config_file, "rb") as f:
                data = tomllib.load(f)
            return cls(**data)
        except Exception:
            return cls()

    def save(self) -> None:
        """Save settings to config file."""
        self.ensure_config_dir()
        import tomlkit
        with open(self.config_file, "w", encoding="utf-8") as f:
            tomlkit.dump(self.model_dump(exclude_none=True), f)

    def update_api_key(self, provider: str, api_key: str) -> None:
        """Update API key for a provider."""
        setattr(self, f"{provider}_api_key", api_key)
        self.save()

    def update_language(self, language: str) -> None:
        """Update language setting."""
        if language not in ["en", "zh-CN", "zh-TW"]:
            raise ValueError("Language must be one of: en, zh-CN, zh-TW")
        self.language = language
        self.save()