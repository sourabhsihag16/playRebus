# Code Explanation for Beginners

This document explains how the code works, especially for those new to JavaScript and frontend development.

## Frontend (React/JavaScript) Explanation

### What is React?

React is a JavaScript library for building user interfaces. Think of it as a way to create interactive web pages where you can update parts of the page without refreshing the whole page.

### Key Concepts in Our Code

#### 1. **useState Hook**

```javascript
const [puzzles, setPuzzles] = useState([])
```

- `useState` is a React "hook" that lets you store data that can change over time
- `puzzles` is the current value (an array of puzzles)
- `setPuzzles` is a function to update the puzzles
- `[]` is the initial value (empty array)

**Example**: When you fetch puzzles from the backend, you call `setPuzzles(data.puzzles)` to update the puzzles array, and React automatically re-renders the page to show the new puzzles.

#### 2. **useEffect Hook**

```javascript
useEffect(() => {
  // Code here runs when component loads
}, [])
```

- `useEffect` runs code at specific times (like when the page loads)
- The empty array `[]` means "run only once when the component first loads"
- We use it to fetch puzzles from the backend when the page loads

#### 3. **localStorage (Browser Cache)**

```javascript
localStorage.setItem('key', 'value')
const data = localStorage.getItem('key')
```

- `localStorage` is a way to store data in the user's browser
- Data persists even after closing the browser
- We use it to remember which puzzles the user has solved
- Format: `rebus_solved_2024-01-15` stores solved puzzles for that date

#### 4. **Fetch API (Getting Data from Backend)**

```javascript
const response = await fetch('http://localhost:8080/api/puzzles/2024-01-15')
const data = await response.json()
```

- `fetch` is a JavaScript function to make HTTP requests (like getting data from a server)
- `await` means "wait for this to finish before continuing"
- `response.json()` converts the response to a JavaScript object

#### 5. **Event Handlers**

```javascript
onChange={(e) => setUserAnswer(e.target.value)}
onSubmit={handleSubmit}
```

- `onChange` runs when the user types in an input field
- `e.target.value` is what the user typed
- `onSubmit` runs when the user submits a form (presses Enter or clicks submit)

### How the Puzzle Flow Works

1. **Page Loads**: 
   - `useEffect` runs and fetches puzzles from backend
   - Also checks localStorage for previously solved puzzles

2. **User Sees Puzzles**:
   - First puzzle is always unlocked
   - Other puzzles are locked until previous ones are solved

3. **User Solves Puzzle**:
   - Types answer and clicks "Submit"
   - Frontend sends answer to backend for verification
   - If correct, puzzle is marked as solved in localStorage
   - Next puzzle becomes unlocked

4. **User Returns Later**:
   - localStorage is checked
   - Previously solved puzzles show as solved
   - User continues from where they left off

## Backend (Go) Explanation

### What is Go?

Go (or Golang) is a programming language designed for building fast, reliable software. We use it to create our API server.

### Key Concepts

#### 1. **HTTP Handlers**

```go
func GetPuzzlesHandler(w http.ResponseWriter, r *http.Request) {
    // Handle GET request
}
```

- Functions that handle HTTP requests (GET, POST, etc.)
- `w` is for writing the response
- `r` is the incoming request

#### 2. **JSON Encoding/Decoding**

```go
json.NewEncoder(w).Encode(response)
```

- Converts Go data structures to JSON format
- JSON is a text format for exchanging data between frontend and backend

#### 3. **Mutex (Thread Safety)**

```go
store.mu.Lock()
defer store.mu.Unlock()
```

- Prevents multiple requests from modifying data at the same time
- `Lock()` means "only I can access this now"
- `Unlock()` means "others can access it now"
- `defer` ensures Unlock runs even if there's an error

#### 4. **Interface (AI Generator)**

```go
type AIGenerator interface {
    GenerateRebusPuzzle(date string, index int) (Puzzle, error)
}
```

- An interface defines what functions a type must have
- Allows us to swap different AI implementations easily
- Currently uses `MockAIGenerator`, but can be replaced with real AI

## Data Flow

```
User Types Answer
    ↓
Frontend sends POST to /api/puzzles/verify
    ↓
Backend checks if answer matches puzzle answer
    ↓
Backend returns {correct: true/false}
    ↓
Frontend updates UI and localStorage
```

## Important Files

- **dashboard/src/App.jsx**: Main React component with all the puzzle logic
- **dashboard/src/App.css**: Styling for the UI
- **backend/main.go**: Go server with API endpoints
- **localStorage**: Browser storage (not a file, but browser memory)

## Common Questions

**Q: Why do we use localStorage instead of a database?**
A: For simplicity and to work without user accounts. In production, you'd use a database with user authentication.

**Q: What happens if the user clears their browser cache?**
A: Their progress would be lost. In production, you'd store this in a database.

**Q: How do I connect a real AI service?**
A: Replace the `MockAIGenerator` in `backend/main.go` with your AI service implementation. See `backend/README.md` for details.

