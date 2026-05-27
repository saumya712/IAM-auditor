# Implementation Plan: IAM AI Advisor & Auditor

## Overview

Implement the three-service IAM AI Advisor application incrementally: database models and Go auth first, then the Go IAM proxy and history endpoints, then the Python AI service, then the React frontend, and finally Docker Compose wiring. Each phase ends with a checkpoint.

## Tasks

- [x] 1. Project scaffolding and shared configuration
  - Initialize Go module (`go mod init`) with Gin, GORM, PostgreSQL driver, JWT, bcrypt, and godotenv dependencies
  - Initialize Python project with `requirements.txt` including FastAPI, uvicorn, pydantic, openai, and hypothesis
  - Initialize React/Vite/TypeScript project with Tailwind CSS, react-router-dom, axios, react-syntax-highlighter, and fast-check
  - Create `.env.example` files for each service documenting required environment variables (`DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`, `JWT_SECRET`, `AI_SERVICE_URL`, `FRONTEND_ORIGIN`, `LLM_PROVIDER`, `OPENAI_API_KEY`)
  - _Requirements: 10.6_

- [x] 2. Database models and Go core backend foundation
  - [x] 2.1 Implement GORM models and database connection
    - Create `internal/models/user.go` with the `User` struct (ID, Name, Email, Password, CreatedAt, UpdatedAt, History)
    - Create `internal/models/policy_history.go` with the `PolicyHistory` struct (ID, UserID, Type, InputPromptOrPolicy, GeneratedPolicy, AnalysisReport, CreatedAt)
    - Create `internal/db/db.go` that reads DB credentials from environment variables and opens a GORM PostgreSQL connection with AutoMigrate for both models
    - _Requirements: 1.1, 3.1, 4.3, 5.3_

  - [ ]* 2.2 Write unit tests for DB connection and model migration
    - Test that AutoMigrate creates the expected tables
    - Test that the User model enforces the unique email constraint
    - _Requirements: 1.2_

- [x] 3. Go authentication endpoints
  - [x] 3.1 Implement register and login handlers
    - Create `internal/auth/handler.go` with `Register` (hash password with bcrypt, insert User, return 201) and `Login` (verify bcrypt, sign JWT with user_id claim, return 200 + token) handlers
    - Create `internal/middleware/jwt.go` with `JWTMiddleware` that validates Bearer tokens and injects `user_id` into the Gin context
    - Wire routes in `cmd/server/main.go`: `POST /api/auth/register`, `POST /api/auth/login`
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

  - [ ]* 3.2 Write property test: password is never stored as plaintext
    - **Property 1: Password is never stored as plaintext**
    - **Validates: Requirements 1.1, 1.4**
    - For randomly generated valid user inputs, verify stored password is a bcrypt hash and does not equal the plaintext
    - Tag: `// Feature: iam-ai-advisor, Property 1: password is never stored as plaintext`

  - [ ]* 3.3 Write property test: invalid registration inputs always return 400
    - **Property 2: Invalid registration inputs always return 400**
    - **Validates: Requirements 1.3**
    - Generate requests with any combination of missing/empty fields and verify 400 response and no DB record created
    - Tag: `// Feature: iam-ai-advisor, Property 2: invalid registration inputs always return 400`

  - [ ]* 3.4 Write property test: login with correct credentials returns JWT with user_id
    - **Property 3: Login with correct credentials always returns a JWT containing user_id**
    - **Validates: Requirements 2.1, 2.4**
    - For randomly generated registered users, verify login returns 200 and JWT payload contains correct user_id
    - Tag: `// Feature: iam-ai-advisor, Property 3: login with correct credentials returns JWT with user_id`

  - [ ]* 3.5 Write property test: login with wrong credentials returns 401
    - **Property 4: Login with wrong credentials always returns 401**
    - **Validates: Requirements 2.2, 2.3**
    - For random non-existent emails and random wrong passwords, verify 401 and no token issued
    - Tag: `// Feature: iam-ai-advisor, Property 4: login with wrong credentials returns 401`

  - [ ]* 3.6 Write property test: protected endpoints reject requests without valid JWT
    - **Property 5: Protected endpoints reject requests without a valid JWT**
    - **Validates: Requirements 2.6, 3.3, 4.6, 5.6**
    - For each protected endpoint, verify that requests without a token or with an invalid token return 401
    - Tag: `// Feature: iam-ai-advisor, Property 5: protected endpoints reject requests without valid JWT`

- [ ] 4. Checkpoint — auth layer
  - Ensure all tests pass before proceeding.

- [x] 5. Go history endpoint
  - [x] 5.1 Implement history handler
    - Create `internal/history/handler.go` with `GetHistory` that queries PolicyHistory records for the authenticated user ordered by `created_at DESC` and returns them as JSON
    - Wire `GET /api/history` behind `JWTMiddleware` in `cmd/server/main.go`
    - _Requirements: 3.1, 3.2, 3.3_

  - [ ]* 5.2 Write property test: history returned in descending order
    - **Property 6: History is returned in descending creation order**
    - **Validates: Requirements 3.1, 3.2**
    - Insert N randomly generated history records for a user, call GET /api/history, verify records are sorted descending by created_at and count matches N
    - Tag: `// Feature: iam-ai-advisor, Property 6: history returned in descending order`

- [x] 6. Go IAM proxy endpoints and AI client
  - [x] 6.1 Implement the AI client interface and HTTP implementation
    - Create `internal/iam/ai_client.go` defining the `AIClient` interface with `GeneratePolicy` and `AuditPolicy` methods
    - Implement `HTTPAIClient` that reads `AI_SERVICE_URL` from environment and makes HTTP POST requests to `/ai/generate-policy` and `/ai/audit-policy`
    - Return a 502 error if the AI service is unreachable or returns a non-2xx response
    - _Requirements: 4.1, 4.5, 5.1, 5.5_

  - [x] 6.2 Implement advise and audit handlers
    - Create `internal/iam/handler.go` with `Advise` (validate non-empty prompt, proxy to AI client, persist PolicyHistory of type "advise", return result) and `Audit` (validate non-empty input, validate JSON, proxy to AI client, persist PolicyHistory of type "audit", return result) handlers
    - Wire `POST /api/iam/advise` and `POST /api/iam/audit` behind `JWTMiddleware`
    - _Requirements: 4.1, 4.3, 5.1, 5.3, 5.4_

  - [ ]* 6.3 Write property test: invalid JSON to /audit returns 400 without calling AI service
    - **Property 8: Invalid JSON input to /api/iam/audit returns 400 without calling AI service**
    - **Validates: Requirements 5.4**
    - For any string that is not valid JSON, verify POST /api/iam/audit returns 400 and the mock AI client receives no call
    - Tag: `// Feature: iam-ai-advisor, Property 8: invalid JSON to audit returns 400`

  - [ ]* 6.4 Write property test: successful AI call persists history record
    - **Property 7: Successful AI call persists a history record of the correct type**
    - **Validates: Requirements 4.3, 5.3**
    - For randomly generated prompts/policies, after a successful advise or audit call, verify GET /api/history contains a record with the correct type and matching input
    - Tag: `// Feature: iam-ai-advisor, Property 7: successful AI call persists history record`

  - [ ]* 6.5 Write unit test: 502 returned when AI service is unreachable
    - Simulate AI service being down (mock returns error), verify POST /api/iam/advise and POST /api/iam/audit both return 502
    - _Requirements: 4.5, 5.5_

- [ ] 7. Checkpoint — Go backend complete
  - Ensure all tests pass before proceeding.

- [x] 8. Python AI service
  - [x] 8.1 Implement Pydantic models and LLM client abstraction
    - Create `models.py` with `GeneratePolicyRequest`, `GeneratePolicyResponse`, `AuditPolicyRequest`, `Finding`, and `AuditPolicyResponse` Pydantic models
    - Create `llm_client.py` with the `LLMClient` Protocol and concrete implementations for OpenAI and Ollama, selected via `LLM_PROVIDER` environment variable
    - _Requirements: 4.2, 5.2_

  - [x] 8.2 Implement FastAPI endpoints with prompt engineering
    - Create `main.py` with FastAPI app, `POST /ai/generate-policy` handler (system prompt enforcing least-privilege, returns `GeneratePolicyResponse`), and `POST /ai/audit-policy` handler (system prompt for security analysis, returns `AuditPolicyResponse`)
    - Ensure the generate-policy system prompt instructs the LLM to output valid AWS IAM JSON following least-privilege
    - Ensure the audit-policy system prompt instructs the LLM to return structured findings with Low/Medium/High risk levels
    - _Requirements: 4.2, 4.4, 5.2_

  - [ ]* 8.3 Write property test: audit response schema is always valid
    - **Property 9: AI service audit response schema is always valid**
    - **Validates: Requirements 5.2**
    - Using hypothesis, generate random IAM-like JSON strings, call /ai/audit-policy, verify every finding has non-empty description, risk_level in {Low, Medium, High}, and non-empty remediation
    - Tag: `# Feature: iam-ai-advisor, Property 9: audit response schema is always valid`

  - [ ]* 8.4 Write property test: generate-policy response schema is always valid
    - **Property 10: AI service generate-policy response schema is always valid**
    - **Validates: Requirements 4.2**
    - Using hypothesis, generate random non-empty prompt strings, call /ai/generate-policy, verify response contains parseable policy JSON and non-empty explanation
    - Tag: `# Feature: iam-ai-advisor, Property 10: generate-policy response schema is always valid`

  - [ ]* 8.5 Write unit tests for AI service error handling
    - Test that LLM provider errors propagate as 500 responses
    - Test that unparseable LLM responses return 500 with descriptive message
    - _Requirements: 4.2, 5.2_

- [ ] 9. Checkpoint — AI service complete
  - Ensure all tests pass before proceeding.

- [x] 10. React frontend — auth and routing
  - [x] 10.1 Implement AuthContext, ProtectedRoute, and auth pages
    - Create `src/context/AuthContext.tsx` with `AuthProvider` managing JWT in localStorage and exposing `login`, `logout`, and `token`/`user` state
    - Create `src/components/ProtectedRoute.tsx` that redirects to `/login` when no token is present
    - Create `src/pages/LoginPage.tsx` with email/password form, calls `POST /api/auth/login`, stores token on success, shows error on failure
    - Create `src/pages/SignupPage.tsx` with name/email/password form, calls `POST /api/auth/register`, redirects to login on success, shows error on failure
    - Wire routes in `src/App.tsx` using react-router-dom
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

  - [ ]* 10.2 Write component tests for auth pages
    - Test LoginPage: renders fields, shows error on 401, redirects on success
    - Test SignupPage: renders fields, shows error on 409, redirects on success
    - Test ProtectedRoute: redirects to /login when no token, renders children when token present
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6_

- [x] 11. React frontend — dashboard and Role Advisor view
  - [x] 11.1 Implement Dashboard layout and sidebar
    - Create `src/layouts/DashboardLayout.tsx` with a persistent sidebar containing navigation links to `/dashboard/advisor` and `/dashboard/auditor` and a logout button that clears the token and redirects to `/login`
    - _Requirements: 9.1, 9.2, 9.3_

  - [x] 11.2 Implement Role Advisor view
    - Create `src/pages/RoleAdvisorView.tsx` with a textarea for the prompt, a "Generate" button, a `LoadingSpinner` shown while the request is in flight (button disabled), and a split-pane result area showing syntax-highlighted JSON (using `react-syntax-highlighter`) and the explanation text
    - Validate that the prompt is non-empty before sending; show inline validation message if empty
    - Show error message on request failure
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6_

  - [ ]* 11.3 Write component tests for Role Advisor view
    - Test: button disabled and spinner shown while loading
    - Test: split-pane renders on success with policy JSON and explanation
    - Test: error message shown on failure
    - Test: empty prompt shows validation message and does not send request
    - _Requirements: 7.3, 7.4, 7.5, 7.6_

- [x] 12. React frontend — Policy Auditor view
  - [x] 12.1 Implement Policy Auditor view and FindingsReport component
    - Create `src/pages/PolicyAuditorView.tsx` with a file upload dropzone and a textarea fallback for pasting IAM JSON, an "Audit" button, a `LoadingSpinner` while in flight (button disabled), and a `FindingsReport` component
    - Create `src/components/FindingsReport.tsx` that renders findings with color-coded risk indicators: red for High, yellow for Medium, green for Low
    - Validate that the input is non-empty before sending; show inline validation message if empty
    - Show error message on request failure
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6_

  - [ ]* 12.2 Write property test: risk level color mapping is consistent
    - **Property 11: Risk level color mapping is consistent**
    - **Validates: Requirements 8.4**
    - Using fast-check, generate arrays of findings with arbitrary risk_levels, render `FindingsReport`, verify every High finding has red indicator class, Medium has yellow, Low has green
    - Tag: `// Feature: iam-ai-advisor, Property 11: risk level color mapping is consistent`

  - [ ]* 12.3 Write component tests for Policy Auditor view
    - Test: button disabled and spinner shown while loading
    - Test: findings report renders on success
    - Test: error message shown on failure
    - Test: empty input shows validation message and does not send request
    - _Requirements: 8.3, 8.5, 8.6_

- [ ] 13. Checkpoint — frontend complete
  - Ensure all tests pass before proceeding.

- [x] 14. Docker Compose and inter-service wiring
  - [x] 14.1 Write Dockerfile for Go core backend
    - Multi-stage Dockerfile: build stage compiles the Go binary, final stage uses a minimal base image
    - Expose port 8080
    - _Requirements: 10.1_

  - [x] 14.2 Write Dockerfile for Python AI service
    - Use `python:3.11-slim` base, install dependencies from requirements.txt, run with uvicorn
    - Expose port 8000
    - _Requirements: 10.1_

  - [x] 14.3 Write Dockerfile for React frontend
    - Multi-stage Dockerfile: build stage runs `vite build`, final stage serves static files with nginx
    - Expose port 80
    - _Requirements: 10.1_

  - [x] 14.4 Write docker-compose.yml
    - Define `db` service (postgres:15, named volume `pgdata`, health check using `pg_isready`, environment variables for credentials)
    - Define `ai-service` service (Python image, exposes 8000, environment variables for LLM provider and API key)
    - Define `backend` service (Go image, depends_on db with health check condition, exposes 8080, environment variables for DB connection, JWT secret, AI_SERVICE_URL, FRONTEND_ORIGIN)
    - Define `frontend` service (React image, depends_on backend, exposes 5173→80, environment variable for backend API URL)
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6_

- [x] 15. Final checkpoint — full integration
  - Ensure all tests pass and all services start cleanly with `docker compose up`.

## Notes

- Tasks marked with `*` are optional and can be skipped for a faster MVP
- Each task references specific requirements for traceability
- Property tests must run a minimum of 100 iterations each
- Go PBT uses `gopter` or `rapid`; Python PBT uses `hypothesis`; Frontend PBT uses `fast-check`
- The AI service endpoints are internal (not exposed directly to the frontend); all frontend traffic goes through the Go backend
