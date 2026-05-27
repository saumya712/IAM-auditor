# IAM AI Advisor & Auditor

An AI-powered web application for generating least-privilege AWS IAM policies from natural language and auditing existing IAM policies for security vulnerabilities.

## Architecture

The system is composed of four containerized services:

| Service | Language/Framework | Port | Responsibility |
|---|---|---|---|
| **Frontend** | React + Vite + TypeScript + Tailwind | 5173 | UI, auth state, API calls |
| **Core Backend** | Go + Gin | 8080 | Auth, JWT, routing, DB, proxy to AI service |
| **AI Service** | Python + FastAPI | 8000 | LLM orchestration, policy generation & auditing |
| **Database** | PostgreSQL 15 | 5432 | User records, policy history |

The frontend communicates only with the Go backend. The Go backend proxies AI requests to the Python service. The Python service communicates with the configured LLM provider (OpenAI, Anthropic, or Ollama).

## Features

- **Role Advisor**: Describe a job role or AWS service in natural language and receive a least-privilege IAM policy with an explanation.
- **Policy Auditor**: Upload or paste an IAM JSON policy and receive a structured security report with findings categorized by risk level (High / Medium / Low).
- **Authentication**: JWT-based registration and login with bcrypt password hashing.
- **History**: All past generations and audits are persisted and retrievable per user.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) v2+
- An API key for your chosen LLM provider (OpenAI or Anthropic)

## Getting Started

### 1. Configure environment variables

Copy the example env files and fill in your values:

```bash
cp backend/.env.example backend/.env
cp ai-service/.env.example ai-service/.env
cp frontend/.env.example frontend/.env
```

Edit `backend/.env` and set at minimum:
- `JWT_SECRET` — a long random string
- `DB_PASSWORD` — a secure database password

Edit `ai-service/.env` and set:
- `LLM_PROVIDER` — `openai` or `anthropic`
- `OPENAI_API_KEY` or `ANTHROPIC_API_KEY` — your provider API key

### 2. Start all services

```bash
docker compose up --build
```

This will start PostgreSQL, the Go backend, the Python AI service, and the React frontend. The database will be initialized automatically via GORM AutoMigrate.

### 3. Open the app

Navigate to [http://localhost:5173](http://localhost:5173) in your browser.

## Development

### Backend (Go)

```bash
cd backend
cp .env.example .env   # configure your local DB and secrets
go mod download
go run ./cmd/server
```

### AI Service (Python)

```bash
cd ai-service
cp .env.example .env   # configure your LLM provider
python -m venv .venv
source .venv/bin/activate   # or .venv\Scripts\activate on Windows
pip install -r requirements.txt
uvicorn main:app --reload --port 8000
```

### Frontend (React)

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

## Running Tests

### Backend

```bash
cd backend
go test ./...
```

### AI Service

```bash
cd ai-service
pytest
```

### Frontend

```bash
cd frontend
npm test
```

## Project Structure

```
.
├── backend/                  # Go core backend
│   ├── cmd/server/           # Entry point (main.go)
│   ├── internal/
│   │   ├── auth/             # Register & login handlers, JWT signing
│   │   ├── iam/              # Advise & audit handlers, AI proxy client
│   │   ├── history/          # History handler
│   │   ├── middleware/       # JWT & CORS middleware
│   │   ├── models/           # GORM models (User, PolicyHistory)
│   │   └── db/               # DB connection & AutoMigrate
│   ├── go.mod
│   └── .env.example
├── ai-service/               # Python FastAPI AI service
│   ├── main.py               # FastAPI app & endpoints
│   ├── models.py             # Pydantic request/response models
│   ├── llm_client.py         # LLM provider abstraction
│   ├── requirements.txt
│   └── .env.example
├── frontend/                 # React/Vite/TypeScript frontend
│   ├── src/
│   │   ├── context/          # AuthContext (JWT state)
│   │   ├── pages/            # LoginPage, SignupPage, RoleAdvisorView, PolicyAuditorView
│   │   ├── components/       # ProtectedRoute, FindingsReport, PolicyDisplay, etc.
│   │   └── layouts/          # DashboardLayout with sidebar
│   ├── package.json
│   └── .env.example
├── docker-compose.yml        # Full stack orchestration
└── README.md
```
