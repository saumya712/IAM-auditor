"""
LLM client abstraction.

The provider is selected via the LLM_PROVIDER environment variable:
  - "openai"    → OpenAI ChatCompletion (default)
  - "anthropic" → Anthropic Messages API
  - "ollama"    → Ollama local HTTP API

All implementations expose the same async `complete(system, user) -> str` interface.
"""

import json
import os
from typing import Protocol, runtime_checkable

import httpx


@runtime_checkable
class LLMClient(Protocol):
    async def complete(self, system: str, user: str) -> str:
        """Send a system + user prompt and return the assistant's text response."""
        ...


class OpenAIClient:
    """LLM client backed by the OpenAI ChatCompletion API."""

    def __init__(self) -> None:
        try:
            from openai import AsyncOpenAI  # type: ignore
        except ImportError as exc:
            raise RuntimeError("openai package is required for OpenAI provider") from exc

        self._client = AsyncOpenAI(api_key=os.environ["OPENAI_API_KEY"])
        self._model = os.getenv("OPENAI_MODEL", "gpt-4o-mini")

    async def complete(self, system: str, user: str) -> str:
        response = await self._client.chat.completions.create(
            model=self._model,
            messages=[
                {"role": "system", "content": system},
                {"role": "user", "content": user},
            ],
            temperature=0,
        )
        return response.choices[0].message.content or ""


class AnthropicClient:
    """LLM client backed by the Anthropic Messages API."""

    def __init__(self) -> None:
        try:
            import anthropic as _anthropic  # type: ignore
        except ImportError as exc:
            raise RuntimeError("anthropic package is required for Anthropic provider") from exc

        self._client = _anthropic.AsyncAnthropic(api_key=os.environ["ANTHROPIC_API_KEY"])
        self._model = os.getenv("ANTHROPIC_MODEL", "claude-3-haiku-20240307")

    async def complete(self, system: str, user: str) -> str:
        response = await self._client.messages.create(
            model=self._model,
            max_tokens=4096,
            system=system,
            messages=[{"role": "user", "content": user}],
        )
        return response.content[0].text if response.content else ""


class OllamaClient:
    """LLM client backed by a local Ollama HTTP API."""

    def __init__(self) -> None:
        self._base_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
        self._model = os.getenv("OLLAMA_MODEL", "llama3")

    async def complete(self, system: str, user: str) -> str:
        payload = {
            "model": self._model,
            "messages": [
                {"role": "system", "content": system},
                {"role": "user", "content": user},
            ],
            "stream": False,
        }
        async with httpx.AsyncClient(timeout=60) as client:
            resp = await client.post(f"{self._base_url}/api/chat", json=payload)
            resp.raise_for_status()
            data = resp.json()
            return data.get("message", {}).get("content", "")


def get_llm_client() -> LLMClient:
    """Factory that returns the appropriate LLMClient based on LLM_PROVIDER env var."""
    provider = os.getenv("LLM_PROVIDER", "openai").lower()
    if provider == "openai":
        return OpenAIClient()
    if provider == "anthropic":
        return AnthropicClient()
    if provider == "ollama":
        return OllamaClient()
    raise ValueError(f"Unknown LLM_PROVIDER: {provider!r}. Choose 'openai', 'anthropic', or 'ollama'.")
