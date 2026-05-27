# Design Document: IAM AI Advisor & Auditor

## Overview

The IAM AI Advisor & Auditor is a three-tier web application composed of four containerized services: a React/TypeScript frontend, a Go core backend (API gateway + auth), a Python AI service (LLM orchestration), and a PostgreSQL database. The two primary user-facing features are the Role Advisor (natural language → IAM policy) and the Policy Auditor (IAM JSON → security report).

Inter-service communication is HTTP/REST. The frontend talks only to the Go backend. The Go backend proxies AI requests to the Python service. The Python service talks only to the LLM provider. PostgreSQL is accessed only by the Go backend via GORM.

```
Browser (React) --JWT REST--> Core Backend (Go/Gin) --REST--> AI Service (FastAPI) --API--> LLM Provider
                                       |
                                    GORM
                                       |
                                  PostgreSQL
```

---

## Architecture

### Service Boundaries

| Service | Language | Framework | Port | Responsibility |
|---|---|---|---|---|
| Frontend | TypeScript | React + Vite + Tailwind | 5173 | UI, auth state, API calls to Go |
| Core Backend | Go | Gin | 8080 | Auth, JWT, routing, DB, proxy to Python |
| AI Service | Python | FastAPI | 8000 | LLM prompt engineering, policy gen/audit |
| Database | — | PostgreSQL 15 | 5432 | User records, policy history |

### Request Flow — Role Advisor

1. Frontend sends `POST /api/iam/advise` with JWT + prompt
2. Go backend validates JWT
3. Go backend sends `POST /ai/generate-policy {prompt}` to Python service
4. Python service calls LLM with least-privilege system prompt
5. LLM returns IAM JSON + explanation
6. Python returns `{policy, explanation}` to Go
7. Go persists PolicyHistory (type=advise) to DB
8. Go returns `{policy, explanation}` to Frontend

### Request Flow — Policy Auditor

1. Frontend sends `POST /api/iam/audit` with JWT + IAM JSON
2. Go backend validates JWT and validates JSON syntax
3. Go backend sends `POST /ai/audit-policy {policy}` to Python service
4. Python service calls LLM with security audit system prompt
5. LLM returns list of findings
6. Python returns `{findings}` to Go
7. Go persists PolicyHistory (type=audit) to DB
8. Go returns `{findings}` to Frontend

---

## Components and Interfaces

### Core Backend (Go)

#### HTTP Routes

```
POST   /api/auth/register     → authHandler.Register
POST   /api/auth/login        → authHandler.Login
GET    /api/history           → historyHandler.GetHistory   [JWT required]
POST   /api/iam/advise        → iamHandler.Advise           [JWT required]
POST   /api/iam/audit         → iamHandler.Audit            [JWT required]
```

#### Middleware

- `CORSMiddleware`: Sets `Access-Control-Allow-Origin` to the frontend origin (from env var `FRONTEND_ORIGIN`).
- `JWTMiddleware`: Validates Bearer token, extracts `user_id` claim, injects into request context.

#### Internal Packages

```
cmd/server/main.go          — entry point, DI wiring
internal/
  auth/                     — register, login handlers + bcrypt + JWT signing
  iam/                      — advise, audit handlers + proxy client
  history/                  — history handler
  middleware/               — JWT, CORS
  models/                   — GORM models (User, PolicyHistory)
  db/                       — DB connection + AutoMigrate
```

#### AI Proxy Client Interface

```go
type AIClient interface {
    GeneratePolicy(ctx context.Context, prompt string) (*GeneratePolicyResponse, error)
    AuditPolicy(ctx context.Context, policy string) (*AuditPolicyResponse, error)
}
```

### AI Service (Python)

#### HTTP Routes

```
POST /ai/generate-policy   → generate_policy(prompt: str)
POST /ai/audit-policy      → audit_policy(policy: str)
```

#### Pydantic Models

```python
class GeneratePolicyRequest(BaseModel):
    prompt: str

class GeneratePolicyResponse(BaseModel):
    policy: dict          # valid AWS IAM JSON object
    explanation: str

class AuditPolicyRequest(BaseModel):
    policy: str           # raw IAM JSON string

class Finding(BaseModel):
    description: str
    risk_level: Literal["Low", "Medium", "High"]
    remediation: str

class AuditPolicyResponse(BaseModel):
    findings: list[Finding]
```

#### LLM Abstraction

A `LLMClient` abstraction wraps the provider SDK (OpenAI/Anthropic/Ollama). The provider is selected via the `LLM_PROVIDER` environment variable. This keeps the FastAPI handlers provider-agnostic.

```python
class LLMClient(Protocol):
    async def complete(self, system: str, user: str) -> str: ...
```

### Frontend (React/TypeScript)

#### Route Structure

```
/login          → LoginPage
/signup         → SignupPage
/dashboard      → DashboardLayout (protected)
  /dashboard/advisor   → RoleAdvisorView
  /dashboard/auditor   → PolicyAuditorView
```

#### Context

```typescript
interface AuthContext {
  token: string | null;
  user: { id: string; name: string; email: string } | null;
  login(token: string, user: User): void;
  logout(): void;
}
```

#### Key Components

| Component | Responsibility |
|---|---|
| `AuthProvider` | Manages JWT in localStorage, exposes AuthContext |
| `ProtectedRoute` | Redirects to /login if no token |
| `Sidebar` | Navigation links + logout button |
| `RoleAdvisorView` | Prompt textarea, Generate button, split-pane result |
| `PolicyAuditorView` | JSON dropzone/editor, Audit button, findings report |
| `PolicyDisplay` | Syntax-highlighted JSON using `react-syntax-highlighter` |
| `FindingsReport` | Color-coded risk findings (red/yellow/green) |
| `LoadingSpinner` | Reusable spinner shown during in-flight requests |

---

## Data Models

### PostgreSQL Schema (managed via GORM AutoMigrate)

#### `users` table

| Column | Type | Constraints |
|---|---|---|
| id | SERIAL | PRIMARY KEY |
| name | VARCHAR(255) | NOT NULL |
| email | VARCHAR(255) | UNIQUE, NOT NULL |
| password | VARCHAR(255) | NOT NULL (bcrypt hash) |
| created_at | TIMESTAMP | NOT NULL |
| updated_at | TIMESTAMP | NOT NULL |

#### `policy_histories` table

| Column | Type | Constraints |
|---|---|---|
| id | SERIAL | PRIMARY KEY |
| user_id | INTEGER | FOREIGN KEY → users.id, NOT NULL |
| type | VARCHAR(10) | NOT NULL ("advise" or "audit") |
| input_prompt_or_policy | TEXT | NOT NULL |
| generated_policy | TEXT | nullable |
| analysis_report | TEXT (JSON) | nullable |
| created_at | TIMESTAMP | NOT NULL |

### GORM Models (Go)

```go
type User struct {
    gorm.Model
    Name     string `gorm:"not null"`
    Email    string `gorm:"uniqueIndex;not null"`
    Password string `gorm:"not null"`
    History  []PolicyHistory
}

type PolicyHistory struct {
    gorm.Model
    UserID               uint   `gorm:"not null"`
    Type                 string `gorm:"not null"` // "advise" | "audit"
    InputPromptOrPolicy  string `gorm:"type:text;not null"`
    GeneratedPolicy      string `gorm:"type:text"`
    AnalysisReport       string `gorm:"type:text"` // JSON string
}
```

### JWT Payload

```json
{
  "user_id": 42,
  "exp": 1700000000
}
```

---

## Correctness Properties

### Property 1: Password is never stored as plaintext

For any valid registration request with any password string, the value stored in the `users` table SHALL be a valid bcrypt hash and SHALL NOT equal the original plaintext password.

**Validates: Requirements 1.1, 1.4**

---

### Property 2: Invalid registration inputs always return 400

For any registration request where at least one required field (name, email, password) is missing or empty, the Core_Backend SHALL return a 400 status code and the user record SHALL NOT be created in the database.

**Validates: Requirements 1.3**

---

### Property 3: Login with correct credentials always returns a JWT containing user_id

For any registered user, submitting a login request with that user's correct email and password SHALL return a 200 status and a JWT token whose decoded payload contains a `user_id` claim equal to that user's database ID.

**Validates: Requirements 2.1, 2.4**

---

### Property 4: Login with wrong credentials always returns 401

For any login request where either the email does not exist or the password does not match, the Core_Backend SHALL return a 401 status and SHALL NOT issue a JWT token.

**Validates: Requirements 2.2, 2.3**

---

### Property 5: Protected endpoints reject requests without a valid JWT

For any protected endpoint (GET /api/history, POST /api/iam/advise, POST /api/iam/audit), a request sent without a valid JWT token SHALL receive a 401 response.

**Validates: Requirements 2.6, 3.3, 4.6, 5.6**

---

### Property 6: History is returned in descending creation order

For any authenticated user with N history records (N ≥ 0), a GET /api/history request SHALL return exactly N records where each record's `created_at` is greater than or equal to the next record's `created_at`.

**Validates: Requirements 3.1, 3.2**

---

### Property 7: Successful AI call persists a history record of the correct type

For any authenticated user who makes a successful POST /api/iam/advise or POST /api/iam/audit request, a subsequent GET /api/history SHALL include a record whose `type` matches the operation and whose `input_prompt_or_policy` matches the submitted input.

**Validates: Requirements 4.3, 5.3**

---

### Property 8: Invalid JSON input to /api/iam/audit returns 400 without calling AI service

For any string that is not valid JSON, submitting it to POST /api/iam/audit SHALL return a 400 status and the AI_Service SHALL NOT receive a request.

**Validates: Requirements 5.4**

---

### Property 9: AI service audit response schema is always valid

For any IAM JSON policy input, the AI_Service's /ai/audit-policy response SHALL be a JSON object containing a `findings` array where every element has a non-empty `description`, a `risk_level` of exactly "Low", "Medium", or "High", and a non-empty `remediation`.

**Validates: Requirements 5.2**

---

### Property 10: AI service generate-policy response schema is always valid

For any non-empty prompt string, the AI_Service's /ai/generate-policy response SHALL be a JSON object containing a `policy` field that is a parseable JSON object and a non-empty `explanation` string.

**Validates: Requirements 4.2**

---

### Property 11: Risk level color mapping is consistent

For any audit findings report rendered by the Frontend, every finding with `risk_level = "High"` SHALL be rendered with a red indicator, every "Medium" finding with yellow, and every "Low" finding with green.

**Validates: Requirements 8.4**

---

## Error Handling

### Core Backend (Go)

| Scenario | HTTP Status | Response Body |
|---|---|---|
| Missing/empty required field | 400 | `{"error": "field X is required"}` |
| Invalid JSON body | 400 | `{"error": "invalid JSON"}` |
| Email already registered | 409 | `{"error": "email already in use"}` |
| Wrong credentials | 401 | `{"error": "invalid credentials"}` |
| Missing/invalid JWT | 401 | `{"error": "unauthorized"}` |
| AI service unreachable | 502 | `{"error": "AI service unavailable"}` |
| AI service returns error | 502 | `{"error": "AI service error: <detail>"}` |
| Internal server error | 500 | `{"error": "internal server error"}` |

### AI Service (Python)

| Scenario | HTTP Status | Response Body |
|---|---|---|
| Empty prompt | 422 | FastAPI validation error |
| LLM provider error | 500 | `{"detail": "LLM error: <message>"}` |
| LLM returns unparseable response | 500 | `{"detail": "failed to parse LLM response"}` |

### Frontend

- Network errors and non-2xx responses display a toast or inline error message.
- Form validation (empty fields) is enforced client-side before any request is sent.
- JWT expiry is detected on any 401 response from a protected endpoint; the user is logged out and redirected to `/login`.

---

## Testing Strategy

### Core Backend (Go)

**Unit Tests** (using `testing` + `testify`):
- Auth handler: register success, duplicate email, missing fields, login success, wrong password.
- IAM handler: advise with mocked AI client, audit with invalid JSON input, 502 on AI client error.
- History handler: returns records in descending order, empty list for new user.
- JWT middleware: valid token passes, missing token returns 401, expired token returns 401.

**Property-Based Tests** (using `gopter` or `rapid`):
- Property 1: For randomly generated valid user data, stored password is always a bcrypt hash.
- Property 2: For randomly generated requests with missing/empty fields, response is always 400.
- Property 3: For any registered user with correct credentials, login returns JWT with correct user_id.
- Property 4: For any login with wrong credentials, response is always 401.
- Property 5: For any protected endpoint, request without JWT always returns 401.
- Property 6: For any user with randomly generated history records, GET /api/history returns them in descending order.
- Property 7: After a successful advise/audit call, history contains a matching record.
- Property 8: For any non-JSON string, POST /api/iam/audit returns 400.

Each property test MUST run a minimum of 100 iterations.

### AI Service (Python)

**Unit Tests** (using `pytest`):
- generate-policy: valid prompt returns policy dict and explanation string.
- audit-policy: valid IAM JSON returns list of findings.
- LLM client error propagates as 500.

**Property-Based Tests** (using `hypothesis`):
- Property 9: For any IAM JSON string, audit response always has valid schema.
- Property 10: For any non-empty prompt, generate-policy response always contains parseable policy JSON and non-empty explanation.

Each property test MUST run a minimum of 100 iterations.

### Frontend (React/TypeScript)

**Unit/Component Tests** (using `vitest` + `@testing-library/react`):
- LoginPage: renders fields, shows error on failed login, redirects on success.
- SignupPage: renders fields, shows error on failed signup.
- RoleAdvisorView: disables button while loading, shows spinner, renders split-pane on success.
- PolicyAuditorView: disables button while loading, renders findings on success.
- ProtectedRoute: redirects to /login when no token present.

**Property-Based Tests** (using `fast-check`):
- Property 11: For any array of findings with arbitrary risk_levels, FindingsReport always maps High→red, Medium→yellow, Low→green.

Each property test MUST run a minimum of 100 iterations.
