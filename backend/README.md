# Backend - Rebus Puzzle Generator

A structured Go backend server for generating and serving rebus puzzles with daily batch job automation.

## Project Structure

```
backend/
├── main.go                  # Application entry point
├── internal/
│   ├── models/              # Data models
│   │   └── puzzle.go
│   ├── config/              # Configuration management
│   │   └── config.go
│   ├── store/               # Storage layer (file system)
│   │   └── store.go
│   ├── ai/                  # AI generator interface and implementations
│   │   └── generator.go
│   ├── scheduler/           # Daily batch job scheduler
│   │   └── scheduler.go
│   └── handlers/            # HTTP request handlers
│       ├── puzzle_handler.go
│       └── image_handler.go
├── storage/                 # Storage directory (created at runtime)
│   ├── images/             # Puzzle images
│   └── puzzles.json        # Puzzle metadata
├── go.mod
├── go.sum
└── README.md
```

## Features

- **Daily Batch Job**: Automatically generates 5 rebus puzzles at the start of each day (configurable time)
- **Image Storage**: Stores puzzle images on the file system
- **Metadata Persistence**: Saves puzzle data (answers, hints) to JSON file
- **RESTful API**: Endpoints for retrieving puzzles and verifying answers
- **Structured Architecture**: Clean separation of concerns with internal packages
- **AI Integration Ready**: Interface-based design for easy AI service integration

## Setup

1. **Install Go** (version 1.21 or later)

2. **Install dependencies**:
```bash
go mod download
```

3. **Run the server**:
```bash
go run main.go
```

Or build and run:
```bash
go build -o server main.go
./server
```

The server will start on port 8080 by default (or the port specified in the `PORT` environment variable).

## Development Tools

### Hot Reloading with Air

Air automatically rebuilds and restarts your Go application when you make changes to your code. This makes development much faster!

**Installation:**

```bash
# Using Homebrew (macOS)
brew install air

# Or using Go install (correct package path)
go install github.com/air-verse/air@latest
```

**Usage:**

From the `backend` directory, simply run:

```bash
air
```

Air will:
- Watch for changes in `.go` files
- Automatically rebuild your application
- Restart the server when changes are detected
- Show colored output for build status

The configuration is in `.air.toml`. You can customize it to watch additional file types or exclude specific directories.

### Debugging with VS Code

The project includes a VS Code launch configuration for debugging with Delve (the Go debugger).

**Prerequisites:**

1. Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) for VS Code
2. Delve will be automatically installed when you install the Go extension

**Usage:**

1. Open the project in VS Code
2. Set breakpoints in your code by clicking in the gutter (left of line numbers)
3. Press `F5` or go to Run → Start Debugging
4. Select "Launch Server (Debug)" from the dropdown
5. The debugger will start and stop at your breakpoints

**Debug Features:**
- Set breakpoints anywhere in your code
- Step through code line by line
- Inspect variables and their values
- View call stack
- Evaluate expressions in the debug console

**Note:** When debugging, the server runs in debug mode. For hot reloading during development, use Air instead.

## Configuration

### Environment Variables

- `PORT`: Server port (default: 8080)
- `STORAGE_PATH`: Path for storing puzzle metadata (default: `./storage`)
- `IMAGES_PATH`: Path for storing puzzle images (default: `./storage/images`)
- `AI_API_KEY`: API key for AI image generation service (optional, uses mock if not set)
- `AI_API_URL`: URL endpoint for AI image generation service (optional)
- `ALLOWED_ORIGINS`: Comma-separated list of allowed CORS origins (optional)

### Batch Job Configuration

The batch job runs daily at 6:00 AM by default. To change this, modify the `BatchJobHour` and `BatchJobMinute` values in `internal/config/config.go` or add environment variables.

## API Endpoints

### GET `/api/puzzles/{date}`

Get puzzles for a specific date (format: YYYY-MM-DD).

**Response:**
```json
{
  "date": "2024-01-15",
  "puzzles": [
    {
      "id": "2024-01-15-0",
      "imageUrl": "/api/images/2024-01-15-0.png",
      "answer": "breakfast",
      "hint": "break + fast",
      "date": "2024-01-15",
      "index": 0
    },
    ...
  ]
}
```

### POST `/api/puzzles/verify`

Verify an answer for a puzzle.

**Request:**
```json
{
  "puzzleId": "2024-01-15-0",
  "answer": "breakfast"
}
```

**Response:**
```json
{
  "correct": true
}
```

### GET `/api/images/{filename}`

Serve puzzle images. The filename format is `{date}-{index}.png`.

### GET `/health`

Health check endpoint. Returns `200 OK` with body `"OK"`.

## AI Integration

The backend uses an interface-based design for AI generation, making it easy to integrate with any AI service.

### Current Implementation

- **MockAIGenerator**: Returns placeholder puzzles for development/testing
- **RealAIGenerator**: Ready for integration with actual AI services

### Integrating Your AI Service

1. **Update `internal/ai/generator.go`**: Modify the `RealAIGenerator.GenerateRebusPuzzle` method to match your AI service's API format.

2. **Set Environment Variables**:
```bash
export AI_API_KEY="your-api-key"
export AI_API_URL="https://your-ai-service.com/api/generate"
```

3. **AI Service Requirements**:
   - Accept a prompt/request for rebus puzzle generation
   - Return an image (URL or base64 data)
   - Provide the answer and hint for the puzzle
   - The response should include:
     - `image_url` or `image_data` (base64)
     - `answer`: The correct answer
     - `hint`: A hint for the puzzle

### Example AI Service Response Format

```json
{
  "image_url": "https://...",
  "image_data": "base64-encoded-image-data",
  "answer": "breakfast",
  "hint": "break + fast"
}
```

## Daily Batch Job

The scheduler automatically:
1. Runs at the configured time each day (default: 6:00 AM)
2. Generates 5 rebus puzzles for the current date
3. Stores images in the `storage/images/` directory
4. Saves metadata to `storage/puzzles.json`
5. Skips generation if puzzles already exist for that date

The batch job runs in a background goroutine and continues running as long as the server is active.

## Storage

- **Images**: Stored as PNG files in `storage/images/` with format `{date}-{index}.png`
- **Metadata**: Stored in `storage/puzzles.json` as JSON
- **Persistence**: Data persists across server restarts

## Development

### Building

```bash
go build -o server main.go
```

### Testing

```bash
go test ./...
```

## Production Considerations

- Replace file system storage with a database (PostgreSQL, MySQL) for scalability
- Use cloud storage (S3, GCS) for images instead of local file system
- Add authentication and rate limiting
- Implement proper logging and monitoring
- Use environment-based configuration files
- Set up proper backup strategies for puzzle data
