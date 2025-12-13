# Migration Guide - Project Restructuring

This document explains the restructuring of the backend project from a single-file architecture to a structured, maintainable codebase.

## What Changed

### Before
- Single `main.go` file with all code
- In-memory storage only (data lost on restart)
- No batch job automation
- No image storage
- No hint field in puzzles

### After
- Structured package layout following Go best practices
- File system persistence for puzzles and images
- Daily batch job scheduler (runs at 6:00 AM by default)
- Image storage with serving endpoint
- Hint field added to puzzle model
- Clean separation of concerns

## New Structure

### Package Organization

1. **`cmd/server/`**: Application entry point
   - `main.go`: Server initialization, routing, and startup

2. **`internal/models/`**: Data models
   - `puzzle.go`: Puzzle, VerifyRequest, VerifyResponse, PuzzlesResponse

3. **`internal/config/`**: Configuration management
   - `config.go`: Loads configuration from environment variables

4. **`internal/store/`**: Storage layer
   - `store.go`: File system storage for puzzles and images
   - Persists data to JSON file
   - Manages image file storage

5. **`internal/ai/`**: AI generation interface
   - `generator.go`: AIGenerator interface and implementations
   - `MockAIGenerator`: For development/testing
   - `RealAIGenerator`: Ready for AI service integration

6. **`internal/scheduler/`**: Batch job scheduler
   - `scheduler.go`: Daily puzzle generation at configured time
   - Handles automatic generation and manual triggers

7. **`internal/handlers/`**: HTTP handlers
   - `puzzle_handler.go`: Puzzle retrieval and verification
   - `image_handler.go`: Image serving

## Key Features Added

### 1. Daily Batch Job
- Automatically generates 5 puzzles each day at 6:00 AM
- Runs in background goroutine
- Skips if puzzles already exist for the date
- Can be manually triggered if needed

### 2. Image Storage
- Images stored in `storage/images/` directory
- Format: `{date}-{index}.png`
- Served via `/api/images/{filename}` endpoint
- Images persist across server restarts

### 3. Data Persistence
- Puzzle metadata stored in `storage/puzzles.json`
- Loads existing puzzles on startup
- No data loss on server restart

### 4. Hint Field
- Puzzles now include a `hint` field
- Helps users solve rebus puzzles
- Stored with puzzle metadata

## Running the New Structure

### Development
```bash
go run cmd/server/main.go
```

### Production Build
```bash
go build -o server cmd/server/main.go
./server
```

## Environment Variables

Set these for production:
```bash
export PORT=8080
export STORAGE_PATH=./storage
export IMAGES_PATH=./storage/images
export AI_API_KEY=your-api-key-here
export AI_API_URL=https://your-ai-service.com/api/generate
```

## AI Integration

To integrate with your AI service:

1. Update `internal/ai/generator.go` - modify `RealAIGenerator.GenerateRebusPuzzle()` to match your API
2. Set `AI_API_KEY` and `AI_API_URL` environment variables
3. The generator will automatically be used if both are set

## API Changes

### New Endpoint
- `GET /api/images/{filename}` - Serve puzzle images

### Updated Response
Puzzle objects now include:
```json
{
  "id": "2024-01-15-0",
  "imageUrl": "/api/images/2024-01-15-0.png",
  "answer": "breakfast",
  "hint": "break + fast",  // NEW
  "date": "2024-01-15",
  "index": 0
}
```

## Benefits

1. **Maintainability**: Clear separation of concerns
2. **Testability**: Each package can be tested independently
3. **Scalability**: Easy to add features (database, caching, etc.)
4. **Reliability**: Data persists across restarts
5. **Automation**: Daily batch job eliminates manual generation
6. **Extensibility**: Easy to swap AI providers or storage backends
