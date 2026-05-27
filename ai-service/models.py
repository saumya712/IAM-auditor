from typing import Literal
from pydantic import BaseModel


class GeneratePolicyRequest(BaseModel):
    prompt: str


class GeneratePolicyResponse(BaseModel):
    policy: dict  # valid AWS IAM JSON object
    explanation: str


class AuditPolicyRequest(BaseModel):
    policy: str  # raw IAM JSON string


class Finding(BaseModel):
    description: str
    risk_level: Literal["Low", "Medium", "High"]
    remediation: str


class AuditPolicyResponse(BaseModel):
    findings: list[Finding]
