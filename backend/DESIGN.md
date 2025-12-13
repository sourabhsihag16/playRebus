# Backend Design Documentation

## Overview

The backend is a structured Go HTTP server that generates and serves rebus puzzles. It provides RESTful API endpoints for puzzle retrieval and answer verification. The system automatically generates 5 puzzles per day via a scheduled batch job and integrates with AI services for puzzle generation. Metadata is stored in PostgreSQL, while images are stored on the file system.

## Architecture

### Technology Stack

- **Go 1.21+**: Programming language
- **PostgreSQL**: Database for puzzle metadata
- **Gorilla Mux**: HTTP router and URL matcher
- **Gorilla Handlers**: CORS middleware
- **lib/pq**: PostgreSQL driver
- **Standard Library**: `net/http`, `encoding/json`, `database/sql`

### Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── models/                  # Data models
│   │   └── puzzle.go
│   ├── config/                  # Configuration management
│   │   └── config.go
│   ├── database/                # PostgreSQL database layer
│   │   └── database.go
│   ├── store/                   # Storage abstraction layer
│   │   └── store.go
│   ├── ai/                      # AI generator interface
│   │   └── generator.go
│   ├── scheduler/               # Daily batch job scheduler
│   │   └── scheduler.go
│   └── handlers/                # HTTP request handlers
│       ├── puzzle_handler.go
│       └── image_handler.go
├── storage/
│   └── images/                  # Puzzle images (file system)
├── go.mod
└── README.md
```

## Design Patterns

### 1. Interface-Based Design

The `AIGenerator` interface allows for easy swapping of AI implementations:

```go
type AIGenerator interface {
    GenerateRebusPuzzle(date string, index int, imageStore *store.Store) (*models.Puzzle, error)
}
```

**Benefits**:
- Decouples puzzle generation from specific AI service
- Easy to test with mock implementations
- Can switch AI providers without changing core logic

### 2. Layered Architecture

**Database Layer** (`internal/database/`):
- Handles all PostgreSQL operations
- Provides clean interface for data access
- Manages schema initialization

**Store Layer** (`internal/store/`):
- Abstracts storage operations
- Combines database (metadata) and file system (images)
- Provides unified API for handlers

**Handler Layer** (`internal/handlers/`):
- HTTP request/response handling
- Input validation
- Error handling

### 3. Separation of Concerns

- **Models**: Data structures only
- **Database**: Data persistence logic
- **Store**: Storage abstraction
- **AI**: Generation logic
- **Scheduler**: Batch job management
- **Handlers**: HTTP layer

## Data Models

### Puzzle Structure

```go
type Puzzle struct {
    ID        string `json:"id"`        // Unique identifier: "YYYY-MM-DD-index"
    ImageURL  string `json:"imageUrl"` // URL to puzzle image
    ImagePath string `json:"-"`        // Local file path (not exposed in API)
    Answer    string `json:"answer"`    // Correct answer (lowercase)
    Hint      string `json:"hint"`     // Hint for the puzzle
    Date      string `json:"date"`     // Date in YYYY-MM-DD format
    Index     int    `json:"index"`     // Puzzle number (0-4)
}
```

### Database Schema

```sql
CREATE TABLE puzzles (
    id VARCHAR(50) PRIMARY KEY,
    date VARCHAR(10) NOT NULL,
    index_num INTEGER NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    image_path VARCHAR(500) NOT NULL,
    answer VARCHAR(255) NOT NULL,
    hint TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(date, index_num)
);

CREATE INDEX idx_puzzles_date ON puzzles(date);
CREATE INDEX idx_puzzles_id ON puzzles(id);
```

### Request/Response Models

**VerifyRequest**:
```go
type VerifyRequest struct {
    PuzzleID string `json:"puzzleId"`
    Answer   string `json:"answer"`
}
```

**VerifyResponse**:
```go
type VerifyResponse struct {
    Correct bool `json:"correct"`
}
```

**PuzzlesResponse**:
```go
type PuzzlesResponse struct {
    Date    string   `json:"date"`
    Puzzles []Puzzle `json:"puzzles"`
}
```

## System Flow

### Daily Batch Job Flow

```
Server Startup
    ↓
Initialize Database Connection
    ↓
Initialize Scheduler
    ↓
Scheduler Starts Background Goroutine
    ↓
Wait for Scheduled Time (default: 6:00 AM)
    ↓
Check if puzzles exist for today
    ↓
If not, generate 5 puzzles:
    ├─→ For each puzzle (0-4):
    │   ├─→ Call AI Generator
    │   ├─→ Receive image data + answer + hint
    │   ├─→ Save image to file system
    │   └─→ Save metadata to PostgreSQL
    └─→ Transaction commits all puzzles
    ↓
Wait until next day
    ↓
Repeat
```

### API Request Flow

#### GET `/api/puzzles/{date}`

```
Client Request
    ↓
Handler: GetPuzzlesHandler
    ↓
Validate Date Format
    ↓
Store.GetPuzzlesForDate(date)
    ↓
Database.Query("SELECT * FROM puzzles WHERE date = $1")
    ↓
Return Puzzles Array
    ↓
JSON Response to Client
```

#### POST `/api/puzzles/verify`

```
Client Request with PuzzleID and Answer
    ↓
Handler: VerifyAnswerHandler
    ↓
Parse PuzzleID → Extract date and index
    ↓
Store.GetPuzzlesForDate(date)
    ↓
Find puzzle by ID
    ↓
Compare answers (case-insensitive)
    ↓
Return Verification Result
```

#### GET `/api/images/{filename}`

```
Client Request
    ↓
Handler: ServeImage
    ↓
Validate filename (prevent directory traversal)
    ↓
Read file from storage/images/
    ↓
Set Content-Type header
    ↓
Serve file to client
```

## API Endpoints

### GET `/api/puzzles/{date}`

**Purpose**: Retrieve puzzles for a specific date

**Path Parameters**:
- `date`: Date in `YYYY-MM-DD` format

**Response**:
- 200 OK: Returns puzzles for the date
- 400 Bad Request: Invalid date format
- 404 Not Found: No puzzles found for date

**Example Request**:
```
GET /api/puzzles/2024-01-15
```

**Example Response**:
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

**Purpose**: Verify if user's answer is correct

**Request Body**:
```json
{
  "puzzleId": "2024-01-15-0",
  "answer": "breakfast"
}
```

**Response**:
- 200 OK: Returns verification result
- 400 Bad Request: Invalid request body
- 404 Not Found: Puzzle not found

**Example Response**:
```json
{
  "correct": true
}
```

### GET `/api/images/{filename}`

**Purpose**: Serve puzzle images

**Path Parameters**:
- `filename`: Image filename (format: `{date}-{index}.png`)

**Response**:
- 200 OK: Image file
- 400 Bad Request: Invalid filename
- 404 Not Found: Image not found

### GET `/health`

**Purpose**: Health check endpoint

**Response**: `200 OK` with body `"OK"`

## Storage Architecture

### Metadata Storage (PostgreSQL)

**Why PostgreSQL?**
- ACID compliance for data integrity
- Scalable for multiple servers
- Rich querying capabilities
- Industry standard for production
- Supports transactions for batch operations

**Schema Design**:
- `id`: Primary key (composite: date + index)
- `date`: Indexed for fast date-based queries
- `index_num`: Puzzle position (0-4)
- `image_url`: URL path for serving
- `image_path`: Local file system path
- `answer`: Correct answer (lowercase)
- `hint`: Puzzle hint
- `created_at`: Timestamp for auditing

**Indexes**:
- Primary key on `id` for fast lookups
- Index on `date` for date-based queries
- Unique constraint on `(date, index_num)` prevents duplicates

### Image Storage (File System)

**Why File System?**
- Simple for MVP
- Fast local access
- Easy to migrate to cloud storage later
- No additional service dependencies

**Structure**:
```
storage/images/
├── 2024-01-15-0.png
├── 2024-01-15-1.png
├── 2024-01-15-2.png
├── 2024-01-15-3.png
└── 2024-01-15-4.png
```

**Future Migration**: Can easily move to S3, GCS, or CDN

## AI Integration Design

### Current: MockAIGenerator

```go
type MockAIGenerator struct{}
```

**Purpose**: Placeholder for development and testing

**Returns**: Hardcoded puzzle examples with placeholder images

### Real AI Integration: RealAIGenerator

**Implementation**:
```go
type RealAIGenerator struct {
    config AIGeneratorConfig
}
```

**Flow**:
1. Create request payload with prompt
2. Call AI service API
3. Receive response with:
   - `image_url` or `image_data` (base64)
   - `answer`: Correct answer
   - `hint`: Puzzle hint
4. Download/save image to file system
5. Return `Puzzle` struct

**Configuration**:
- Set `AI_API_KEY` environment variable
- Set `AI_API_URL` environment variable
- Server automatically uses `RealAIGenerator` if both are set

### AI Service Requirements

The AI service should:
- Accept a prompt/request for rebus puzzle generation
- Return an image (URL or base64 data)
- Provide the answer and hint
- Be idempotent (same input = same output for same date/index)

## Batch Job Scheduler

### Design

**Scheduler** (`internal/scheduler/`):
- Runs in background goroutine
- Calculates next run time
- Executes at configured time (default: 6:00 AM)
- Generates 5 puzzles for current date
- Uses transactions for atomicity

### Features

1. **Automatic Execution**: Runs daily without manual intervention
2. **Idempotent**: Skips if puzzles already exist for date
3. **Configurable**: Time can be configured via environment variables
4. **Resilient**: Continues running even if one puzzle generation fails
5. **Transactional**: All puzzles saved atomically

### Configuration

- `BATCH_JOB_HOUR`: Hour of day (0-23), default: 6
- `BATCH_JOB_MINUTE`: Minute of hour (0-59), default: 0

## Error Handling

### Error Types

1. **Validation Errors**: Invalid date format, malformed requests
   - Return: `400 Bad Request`

2. **Not Found Errors**: Puzzle doesn't exist
   - Return: `404 Not Found`

3. **Generation Errors**: AI service failure
   - Return: `500 Internal Server Error`
   - Logged for monitoring

4. **Database Errors**: Connection issues, query failures
   - Return: `500 Internal Server Error`
   - Logged with details

5. **Internal Errors**: Unexpected failures
   - Log error
   - Return: `500 Internal Server Error`

### Error Response Format

Currently returns plain text error messages. Can be improved to JSON:

```json
{
  "error": "Error message",
  "code": "ERROR_CODE"
}
```

## CORS Configuration

```go
handlers.AllowedOrigins([]string{"http://localhost:5173", "http://localhost:3000"})
handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"})
handlers.AllowedHeaders([]string{"Content-Type"})
```

**Purpose**: Allow frontend to make requests from different origin

**Production**: Update origins to actual frontend domain via `ALLOWED_ORIGINS` environment variable

## Server Configuration

### Environment Variables

- `PORT`: Server port (default: 8080)
- `DATABASE_URL`: PostgreSQL connection string (default: `postgres://postgres:postgres@localhost:5432/rebus_puzzles?sslmode=disable`)
- `IMAGES_PATH`: Path for storing images (default: `./storage/images`)
- `AI_API_KEY`: API key for AI service (optional)
- `AI_API_URL`: URL for AI service (optional)
- `BATCH_JOB_HOUR`: Hour for batch job (default: 6)
- `BATCH_JOB_MINUTE`: Minute for batch job (default: 0)
- `ALLOWED_ORIGINS`: Comma-separated CORS origins (optional)

### Timeouts

```go
ReadTimeout:  15 * time.Second
WriteTimeout: 15 * time.Second
IdleTimeout:  60 * time.Second
```

**Purpose**: Prevent resource exhaustion from slow/hanging connections

## Answer Verification Logic

### Comparison Algorithm

```go
strings.ToLower(strings.TrimSpace(req.Answer)) == puzzle.Answer
```

**Steps**:
1. Trim whitespace from user answer
2. Convert to lowercase
3. Compare with stored answer (already lowercase)

**Benefits**:
- Case-insensitive matching
- Handles extra spaces
- Simple and fast

**Limitations**:
- No fuzzy matching
- No partial credit
- Exact match required

## Puzzle ID Format

**Format**: `{date}-{index}`

**Example**: `2024-01-15-0`

**Benefits**:
- Unique per date and puzzle
- Contains date information
- Easy to parse
- Human-readable

## Daily Puzzle Generation

### Generation Strategy

- **Automated**: Batch job runs daily at configured time
- **Once Per Day**: Same puzzles returned for same date
- **5 Puzzles**: Fixed number per day
- **Idempotent**: Won't regenerate if puzzles exist

### Generation Timing

1. Scheduler runs at configured time (default: 6:00 AM)
2. Checks if puzzles exist for today
3. If not, generates 5 puzzles via AI service
4. Saves images to file system
5. Saves metadata to PostgreSQL in transaction
6. Subsequent requests return cached puzzles

### Storage Duration

- **Metadata**: Stored in PostgreSQL (persistent)
- **Images**: Stored on file system (persistent)
- **Data persists**: Across server restarts

## Concurrency Model

### Database Concurrency

- PostgreSQL handles concurrent access
- Transactions ensure atomicity
- Connection pooling for efficiency

### Scheduler Concurrency

- Runs in separate goroutine
- Non-blocking for HTTP requests
- Graceful shutdown on SIGTERM/SIGINT

### Request Handling

- Each HTTP request handled in separate goroutine
- Database queries are safe for concurrent access
- File system reads are safe for concurrent access

## Production Considerations

### Current Features

1. ✅ **Persistence**: PostgreSQL for metadata
2. ✅ **Image Storage**: File system (can migrate to cloud)
3. ✅ **Batch Automation**: Daily scheduled generation
4. ✅ **Structured Architecture**: Clean separation of concerns
5. ✅ **Error Handling**: Comprehensive error handling
6. ✅ **Configuration**: Environment-based configuration

### Recommended Improvements

1. **Image Storage**:
   - Migrate to S3/GCS for cloud storage
   - Use CDN for fast delivery
   - Implement image optimization

2. **Authentication**:
   - JWT tokens for user identification
   - Track user progress server-side
   - Rate limiting per user

3. **Rate Limiting**:
   - Limit requests per IP
   - Prevent abuse
   - Use middleware (e.g., `golang.org/x/time/rate`)

4. **Caching**:
   - Redis for puzzle cache
   - Reduce database load
   - Cache frequently accessed puzzles

5. **Logging**:
   - Structured logging (logrus, zap)
   - Request/response logging
   - Error tracking (Sentry, etc.)

6. **Monitoring**:
   - Health checks with database status
   - Metrics (Prometheus)
   - Distributed tracing
   - Alerting

7. **Database**:
   - Connection pooling configuration
   - Read replicas for scaling
   - Backup strategy

8. **Deployment**:
   - Docker containerization
   - Kubernetes orchestration
   - CI/CD pipeline

## Security Considerations

### Current State

- ✅ Input validation for dates
- ✅ SQL injection prevention (parameterized queries)
- ✅ Directory traversal prevention for images
- ✅ CORS protection
- ⚠️ No authentication
- ⚠️ No rate limiting

### Security Improvements Needed

1. **Input Sanitization**: Validate all inputs
2. **Authentication**: JWT tokens or OAuth
3. **Rate Limiting**: Prevent DoS attacks
4. **HTTPS**: Encrypt traffic in production
5. **API Keys**: Protect AI service credentials (use secrets management)
6. **Database Security**: Use connection pooling, limit privileges
7. **Image Validation**: Validate image files before saving

## Performance Characteristics

### Current Performance

- **Database Queries**: O(1) for date-based lookups (indexed)
- **Image Serving**: O(1) file system reads
- **Memory**: Minimal (database handles caching)
- **Scalability**: Horizontal scaling possible with shared database

### Scalability

- **Vertical**: Can handle more requests with more CPU/memory
- **Horizontal**: Can scale with multiple servers sharing PostgreSQL
- **Database**: Can add read replicas for read scaling
- **Images**: Can migrate to CDN for global distribution

## Testing Strategy

### Unit Tests

- Puzzle generation logic
- Answer verification
- Date validation
- ID parsing
- Database operations

### Integration Tests

- API endpoint testing
- Database integration
- Image storage
- Batch job execution
- Error scenarios

### Test Tools

- `testing` package (standard library)
- `httptest` for HTTP testing
- `testify` for assertions
- `testcontainers` for database testing

## Deployment

### Development

```bash
# Set up PostgreSQL
createdb rebus_puzzles

# Set environment variables
export DATABASE_URL="postgres://user:pass@localhost:5432/rebus_puzzles?sslmode=disable"

# Run server
go run main.go
```

### Production Build

```bash
go build -o server main.go
./server
```

### Docker (Recommended)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o server main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/storage ./storage
CMD ["./server"]
```

### Environment Setup

```bash
# Required
export DATABASE_URL="postgres://user:password@host:5432/dbname?sslmode=require"
export PORT=8080

# Optional
export AI_API_KEY="your-key"
export AI_API_URL="https://api.example.com/generate"
export IMAGES_PATH="/app/storage/images"
export BATCH_JOB_HOUR=6
export BATCH_JOB_MINUTE=0
```

## Database Migration

### Schema Management

Currently, schema is auto-created on startup. For production:

1. Use migration tool (e.g., `golang-migrate`, `sql-migrate`)
2. Version control schema changes
3. Run migrations as part of deployment
4. Rollback support for failed migrations

### Example Migration

```sql
-- migrations/001_initial_schema.up.sql
CREATE TABLE puzzles (
    id VARCHAR(50) PRIMARY KEY,
    date VARCHAR(10) NOT NULL,
    index_num INTEGER NOT NULL,
    image_url VARCHAR(255) NOT NULL,
    image_path VARCHAR(500) NOT NULL,
    answer VARCHAR(255) NOT NULL,
    hint TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(date, index_num)
);

CREATE INDEX idx_puzzles_date ON puzzles(date);
CREATE INDEX idx_puzzles_id ON puzzles(id);
```
