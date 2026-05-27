"""
IAM AI Service — FastAPI application.

Endpoints:
  POST /ai/generate-policy  — generate a least-privilege IAM policy from a prompt
  POST /ai/audit-policy     — audit an IAM JSON policy for security vulnerabilities
"""

import json
import os

from dotenv import load_dotenv  # type: ignore[import]
from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware

from llm_client import get_llm_client
from models import (
    AuditPolicyRequest,
    AuditPolicyResponse,
    Finding,
    GeneratePolicyRequest,
    GeneratePolicyResponse,
)

load_dotenv()

app = FastAPI(title="IAM AI Service", version="1.0.0")

# CORS — the AI service is only called by the Go backend, but allow all origins
# for local development convenience.
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

# ---------------------------------------------------------------------------
# System prompts
# ---------------------------------------------------------------------------

GENERATE_POLICY_SYSTEM = """You are an AWS IAM security expert. Your task is to generate
a minimal, least-privilege AWS IAM policy in valid JSON format.

Rules you MUST follow:
1. Apply the principle of least privilege — grant only the permissions explicitly required.
2. Never use wildcard actions (e.g. "s3:*") unless the user explicitly asks for full access.
3. Never use wildcard resources ("*") unless the service requires it (e.g. CloudWatch logs).
4. Use specific resource ARNs wherever possible.
5. Output ONLY a JSON object with two keys:
   - "policy": a valid AWS IAM policy JSON object (with Version, Statement, etc.)
   - "explanation": a plain-English explanation of each permission granted and why

Do NOT include any text outside the JSON object. Do NOT wrap it in markdown code fences."""

AUDIT_POLICY_SYSTEM = """You are an AWS IAM security auditor. Your task is to analyse an
AWS IAM policy JSON and identify security vulnerabilities.

For each issue found, produce a finding with:
- "description": a clear description of the security issue
- "risk_level": exactly one of "High", "Medium", or "Low"
- "remediation": concrete steps to fix the issue

Common issues to look for:
- Wildcard actions (e.g. "s3:*", "*")
- Wildcard resources ("*") where specific ARNs should be used
- Overly broad permissions that enable privilege escalation
- Missing condition keys that should restrict access
- Dangerous action combinations (e.g. iam:PassRole + ec2:RunInstances)
- Lack of MFA conditions on sensitive actions

Output ONLY a JSON object with one key:
- "findings": an array of finding objects (may be empty if the policy is secure)

Do NOT include any text outside the JSON object. Do NOT wrap it in markdown code fences."""


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _parse_llm_json(raw: str) -> dict:
    """Strip optional markdown fences and parse JSON from LLM output."""
    text = raw.strip()
    if text.startswith("```"):
        # Remove opening fence (```json or ```)
        text = text.split("\n", 1)[-1]
        # Remove closing fence
        if text.endswith("```"):
            text = text[: text.rfind("```")]
    return json.loads(text.strip())


# ---------------------------------------------------------------------------
# Endpoints
# ---------------------------------------------------------------------------

@app.post("/ai/generate-policy", response_model=GeneratePolicyResponse)
async def generate_policy(request: GeneratePolicyRequest) -> GeneratePolicyResponse:
    """Generate a least-privilege AWS IAM policy from a natural language prompt."""
    llm = get_llm_client()
    try:
        raw = await llm.complete(system=GENERATE_POLICY_SYSTEM, user=request.prompt)
    except Exception as exc:
        raise HTTPException(status_code=500, detail=f"LLM error: {exc}") from exc

    try:
        data = _parse_llm_json(raw)
        policy = data["policy"]
        explanation = data["explanation"]
    except (json.JSONDecodeError, KeyError, TypeError) as exc:
        raise HTTPException(
            status_code=500, detail=f"failed to parse LLM response: {exc}"
        ) from exc

    return GeneratePolicyResponse(policy=policy, explanation=explanation)


@app.post("/ai/audit-policy", response_model=AuditPolicyResponse)
async def audit_policy(request: AuditPolicyRequest) -> AuditPolicyResponse:
    """Audit an AWS IAM policy JSON for security vulnerabilities."""
    llm = get_llm_client()
    try:
        raw = await llm.complete(system=AUDIT_POLICY_SYSTEM, user=request.policy)
    except Exception as exc:
        raise HTTPException(status_code=500, detail=f"LLM error: {exc}") from exc

    try:
        data = _parse_llm_json(raw)
        findings_raw = data["findings"]
        findings = [Finding(**f) for f in findings_raw]
    except (json.JSONDecodeError, KeyError, TypeError, ValueError) as exc:
        raise HTTPException(
            status_code=500, detail=f"failed to parse LLM response: {exc}"
        ) from exc

    return AuditPolicyResponse(findings=findings)


@app.get("/health")
async def health() -> dict:
    return {"status": "ok"}
