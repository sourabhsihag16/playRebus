# PlayRebus - Rebus Puzzle Game

A monorepo containing a React frontend (dashboard) and Go backend for daily rebus puzzle generation and solving.

## Structure

- `dashboard/` - React frontend application (Vite + React)
- `backend/` - Go backend API server

## Features

- **Daily Puzzle Generation**: Backend generates 5 rebus puzzles per day using AI
- **Sequential Unlocking**: Players must solve each puzzle before unlocking the next one
- **Browser Cache**: Solved puzzles are stored in browser's localStorage, so progress persists
- **Date-based Management**: Each day has its own set of puzzles
- **Modern UI**: Beautiful, responsive interface with gradient design

## How It Works

1. **Backend**: The Go server generates 5 rebus puzzles daily. When a request comes in for a date, it checks if puzzles exist for that date. If not, it generates them (currently using a mock generator - ready for AI integration).

2. **Frontend**: The React app fetches puzzles for today's date and displays them. Users can only access puzzles sequentially - they must solve puzzle 1 before puzzle 2 unlocks, and so on.

3. **Progress Tracking**: When a user solves a puzzle, it's saved to their browser's localStorage with the date as a key. If they return on the same day, their progress is restored.

4. **Answer Verification**: Users submit answers through the frontend, which are verified by the backend API.

## Getting Started

### Prerequisites

- Node.js (v18 or later)
- Go (v1.21 or later)

### Running the Application

#### 1. Start the Backend Server

```bash
cd backend
go mod download
go run main.go
```

The backend will start on `http://localhost:8080`

#### 2. Start the Frontend (in a new terminal)

```bash
cd dashboard
npm install
npm run dev
```

The frontend will start on `http://localhost:5173` (or another port if 5173 is busy)

#### 3. Open in Browser

Navigate to the frontend URL (usually `http://localhost:5173`)

## API Endpoints

### GET /api/puzzles/{date}
Get puzzles for a specific date (format: YYYY-MM-DD)

### POST /api/puzzles/verify
Verify an answer for a puzzle

See `backend/README.md` for detailed API documentation.

## AI Integration

The backend currently uses a mock AI generator. To integrate with a real AI service:

1. Implement the `AIGenerator` interface in `backend/main.go`
2. Replace the `MockAIGenerator` with your implementation
3. The AI service should generate rebus puzzle images and return the image URL and answer

See `backend/README.md` for more details on AI integration.

## Project Structure

```
playRebus/
├── dashboard/          # React frontend
│   ├── src/
│   │   ├── App.jsx     # Main puzzle game component
│   │   └── App.css     # Styling
│   └── package.json
├── backend/            # Go backend
│   ├── main.go         # Server and API handlers
│   └── go.mod
└── README.md
```

## Technologies Used

- **Frontend**: React, Vite, CSS3
- **Backend**: Go, Gorilla Mux (HTTP router), Gorilla Handlers (CORS)
- **Storage**: Browser localStorage (frontend), In-memory store (backend - replace with database in production)

