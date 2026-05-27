# Requirements Document

## Introduction

This document defines the requirements for an AI-powered AWS IAM Advisor & Auditor web application. The system consists of three decoupled services: a React/TypeScript frontend, a Go (Gin) core backend that handles authentication and acts as an API gateway, and a Python (FastAPI) AI service that communicates with an LLM to generate and audit IAM policies. PostgreSQL stores user credentials and history. The application provides two primary features: a Role Advisor that generates least-privilege IAM policies from natural language prompts, and a Policy Auditor that analyzes uploaded IAM JSON policies for security risks.

## Glossary

- **Core_Backend**: The Go (Gin) service responsible for authentication, routing, and proxying requests to the AI service.
- **AI_Service**: The Python (FastAPI) service responsible for LLM orchestration and IAM policy generation/auditing.
- **Frontend**: The React/TypeScript/Tailwind CSS single-page application served via Vite.
- **Database**: The PostgreSQL instance storing user credentials and policy history.
- **JWT**: JSON Web Token used for stateless authentication between the Frontend and Core_Backend.
- **User**: A registered account with email, hashed password, and associated history.
- **PolicyHistory**: A persisted record of a user's past IAM generation or audit request and its result.
- **Role_Advisor**: The feature that generates a least-privilege IAM policy from a natural language prompt.
- **Policy_Auditor**: The feature that analyzes a user-supplied IAM JSON policy for security vulnerabilities.
- **LLM**: Large Language Model (OpenAI, Anthropic, or Ollama) used by the AI_Service to generate and audit policies.

---

## Requirements

### Requirement 1: User Registration

**User Story:** As a new user, I want to create an account, so that I can securely access the IAM advisor and auditor features.

#### Acceptance Criteria

1. WHEN a POST request is sent to `/api/auth/register` with a valid name, email, and password, THE Core_Backend SHALL create a new User record with the password hashed using bcrypt and return a 201 status.
2. WHEN a registration request is received with an email that already exists in the Database, THE Core_Backend SHALL return a 409 Conflict error with a descriptive message.
3. WHEN a registration request is received with a missing or empty required field (name, email, or password), THE Core_Backend SHALL return a 400 Bad Request error.
4. THE Core_Backend SHALL store passwords exclusively as bcrypt hashes and SHALL NOT store plaintext passwords.

---

### Requirement 2: User Authentication

**User Story:** As a registered user, I want to log in with my credentials, so that I can receive a JWT token to access protected features.

#### Acceptance Criteria

1. WHEN a POST request is sent to `/api/auth/login` with a valid email and correct password, THE Core_Backend SHALL return a signed JWT token and a 200 status.
2. WHEN a login request is received with an email that does not exist in the Database, THE Core_Backend SHALL return a 401 Unauthorized error.
3. WHEN a login request is received with an incorrect password, THE Core_Backend SHALL return a 401 Unauthorized error.
4. THE JWT token issued by THE Core_Backend SHALL contain the user's ID as a claim and SHALL be signed with a secret read from an environment variable.
5. WHILE a valid JWT token is present in the Authorization header, THE Core_Backend SHALL grant access to protected endpoints.
6. IF a request to a protected endpoint is received without a valid JWT token, THEN THE Core_Backend SHALL return a 401 Unauthorized error.

---

### Requirement 3: Policy History Retrieval

**User Story:** As an authenticated user, I want to retrieve my past IAM generations and audits, so that I can review and reuse previous results.

#### Acceptance Criteria

1. WHEN a GET request is sent to `/api/history` with a valid JWT token, THE Core_Backend SHALL return all PolicyHistory records associated with the authenticated user's ID, ordered by creation date descending.
2. WHEN a GET request is sent to `/api/history` with a valid JWT token and no history exists for the user, THE Core_Backend SHALL return an empty list and a 200 status.
3. IF a GET request is sent to `/api/history` without a valid JWT token, THEN THE Core_Backend SHALL return a 401 Unauthorized error.

---

### Requirement 4: Role Advisor — IAM Policy Generation

**User Story:** As an authenticated user, I want to describe a job role or service in natural language, so that I can receive a least-privilege AWS IAM policy.

#### Acceptance Criteria

1. WHEN a POST request is sent to `/api/iam/advise` with a valid JWT token and a non-empty prompt string, THE Core_Backend SHALL forward the prompt to the AI_Service's `/ai/generate-policy` endpoint.
2. WHEN THE AI_Service receives a prompt at `/ai/generate-policy`, THE AI_Service SHALL return a response containing a valid AWS IAM JSON policy object and a textual explanation of the policy choices.
3. WHEN THE AI_Service returns a successful response, THE Core_Backend SHALL persist a PolicyHistory record of type "advise" containing the input prompt, the generated policy, and the explanation, then return the result to the Frontend.
4. THE AI_Service SHALL instruct the LLM to follow the principle of least privilege when generating IAM policies.
5. IF THE AI_Service is unreachable or returns an error, THEN THE Core_Backend SHALL return a 502 Bad Gateway error to the Frontend.
6. IF a POST request is sent to `/api/iam/advise` without a valid JWT token, THEN THE Core_Backend SHALL return a 401 Unauthorized error.

---

### Requirement 5: Policy Auditor — IAM Policy Analysis

**User Story:** As an authenticated user, I want to upload a custom IAM JSON policy, so that I can receive a security analysis identifying vulnerabilities and remediation steps.

#### Acceptance Criteria

1. WHEN a POST request is sent to `/api/iam/audit` with a valid JWT token and a non-empty IAM JSON string, THE Core_Backend SHALL forward the policy to the AI_Service's `/ai/audit-policy` endpoint.
2. WHEN THE AI_Service receives an IAM JSON string at `/ai/audit-policy`, THE AI_Service SHALL return a structured list of security findings, where each finding includes a description, a risk level of "Low", "Medium", or "High", and remediation steps.
3. WHEN THE AI_Service returns a successful audit response, THE Core_Backend SHALL persist a PolicyHistory record of type "audit" containing the input policy and the analysis report, then return the result to the Frontend.
4. IF the submitted IAM JSON string is not valid JSON, THEN THE Core_Backend SHALL return a 400 Bad Request error without forwarding to the AI_Service.
5. IF THE AI_Service is unreachable or returns an error, THEN THE Core_Backend SHALL return a 502 Bad Gateway error to the Frontend.
6. IF a POST request is sent to `/api/iam/audit` without a valid JWT token, THEN THE Core_Backend SHALL return a 401 Unauthorized error.

---

### Requirement 6: Frontend Authentication UI

**User Story:** As a user, I want login and signup pages, so that I can authenticate and access the application.

#### Acceptance Criteria

1. THE Frontend SHALL provide a Login page with email and password input fields and a submit button.
2. THE Frontend SHALL provide a Signup page with name, email, and password input fields and a submit button.
3. WHEN a user successfully logs in, THE Frontend SHALL store the JWT token in local storage and redirect the user to the Dashboard.
4. WHEN a user successfully registers, THE Frontend SHALL redirect the user to the Login page.
5. IF a login or registration request fails, THEN THE Frontend SHALL display a descriptive error message to the user.
6. WHILE a valid JWT token is present in local storage, THE Frontend SHALL protect the Dashboard route from unauthenticated access.

---

### Requirement 7: Frontend Role Advisor View

**User Story:** As an authenticated user, I want a Role Advisor interface, so that I can generate IAM policies from natural language descriptions.

#### Acceptance Criteria

1. THE Frontend SHALL provide a Role Advisor view with a textarea for entering a natural language prompt and a "Generate" button.
2. WHEN the "Generate" button is clicked with a non-empty prompt, THE Frontend SHALL send a POST request to `/api/iam/advise` with the JWT token in the Authorization header.
3. WHILE the request is in flight, THE Frontend SHALL display a loading spinner and disable the "Generate" button.
4. WHEN a successful response is received, THE Frontend SHALL display the generated IAM JSON policy with syntax highlighting and the textual explanation in a split-pane layout.
5. IF the request fails, THEN THE Frontend SHALL display a descriptive error message.
6. WHEN the "Generate" button is clicked with an empty prompt, THE Frontend SHALL prevent the request and display a validation message.

---

### Requirement 8: Frontend Policy Auditor View

**User Story:** As an authenticated user, I want a Policy Auditor interface, so that I can analyze custom IAM policies for security vulnerabilities.

#### Acceptance Criteria

1. THE Frontend SHALL provide a Policy Auditor view with a file upload dropzone or text editor for inputting an IAM JSON policy and an "Audit" button.
2. WHEN the "Audit" button is clicked with a non-empty policy input, THE Frontend SHALL send a POST request to `/api/iam/audit` with the JWT token in the Authorization header.
3. WHILE the request is in flight, THE Frontend SHALL display a loading spinner and disable the "Audit" button.
4. WHEN a successful audit response is received, THE Frontend SHALL display an interactive report with findings categorized by risk level: High findings displayed in red, Medium in yellow, and Low in green.
5. IF the request fails, THEN THE Frontend SHALL display a descriptive error message.
6. WHEN the "Audit" button is clicked with an empty input, THE Frontend SHALL prevent the request and display a validation message.

---

### Requirement 9: Dashboard Navigation

**User Story:** As an authenticated user, I want a dashboard with a sidebar, so that I can switch between the Role Advisor and Policy Auditor views.

#### Acceptance Criteria

1. THE Frontend SHALL provide a Dashboard layout with a persistent sidebar containing navigation links to the Role Advisor and Policy Auditor views.
2. WHEN a sidebar navigation link is clicked, THE Frontend SHALL render the corresponding view without a full page reload.
3. THE Frontend SHALL provide a logout action in the sidebar or header that clears the JWT token from local storage and redirects the user to the Login page.

---

### Requirement 10: Containerized Deployment

**User Story:** As a developer, I want a unified Docker Compose configuration, so that I can spin up all services with a single command.

#### Acceptance Criteria

1. THE System SHALL provide a `docker-compose.yml` file that defines services for the Core_Backend, AI_Service, Frontend, and Database.
2. THE Database service SHALL be configured with a persistent named volume to preserve data across container restarts.
3. THE Database service SHALL include a health check so that dependent services wait for it to be ready before starting.
4. THE Core_Backend SHALL read the AI_Service URL from an environment variable to enable inter-service communication.
5. THE Core_Backend SHALL be configured with CORS to allow requests from the Frontend's origin.
6. WHERE environment-specific configuration is needed, THE System SHALL use environment variables defined in the `docker-compose.yml` or a `.env` file.
