# AI Delivery Operating System (AI-OS)

AI-OS is a state-of-the-art Operational Execution System designed to autonomously track tasks, ingest external events, and maintain project memory using Large Language Models (LLMs). It acts as an intelligent layer on top of your existing tools.

## Architecture

The system consists of two main components:
1. **Frontend**: A Next.js (React) application that provides the "Live Dashboard" and the AI chat interface.
2. **Backend**: A Go-based API server that handles AI routing, database persistence (SQLite), and external integrations.

### Key Features
- **4-Tier AI Routing**: Automatically routes tasks between Extraction (identifying tasks in text), Conversation (chatting with the user), Reasoning (deep thinking), and Fallback modes.
- **Self-Improving Evaluator**: You can provide feedback on AI responses, and the AI will permanently adjust its core instructions based on your feedback.
- **Multi-Integration Hub**: Concurrently pulls data from Google Sheets, Jira, Slack, and GitLab to maintain an accurate, up-to-date project context.
- **Persistent Memory**: All chat history and operational tasks are saved to a local SQLite database (`aios.db`).
- **Automated Executive Reporting**: Automatically generates weekly project status reports.

## Getting Started

### Prerequisites
- Node.js (for the frontend)
- Go (for the backend)
- API Keys for your preferred LLMs (Gemini, DeepSeek, Groq, or OpenRouter)

### 1. Backend Setup
Navigate to the backend directory:
```bash
cd backend
```

Copy the example environment file and fill in your API keys:
```bash
cp .env.example .env
```

Start the backend server (runs on port 8080):
```bash
go run cmd/api/main.go
```

### 2. Frontend Setup
Navigate to the frontend directory:
```bash
cd frontend
```

Install dependencies:
```bash
npm install
```

Start the development server (runs on port 3000):
```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser to view the Live Dashboard.

## Setting Up Integrations
To connect the AI-OS to your Google Sheets, Slack, Jira, or GitLab, please read the [INTEGRATIONS.md](./INTEGRATIONS.md) guide.

## Modifying Integrations
If you have multiple Slack channels, Jira projects, or Google Sheets, you can configure them in the `backend/integrations.json` file. The backend will concurrently sync all listed targets.

## Troubleshooting
- **Hydration Mismatch**: If you see a React Hydration Mismatch error mentioning `pronounceRootElement`, it is caused by a text-to-speech browser extension. It is harmless in development, but you can disable the extension for `localhost` to remove the error.
