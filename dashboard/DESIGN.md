# Frontend Design Documentation

## Overview

The frontend is a React application built with Vite that provides a user interface for solving daily rebus puzzles. It communicates with the Go backend API to fetch puzzles and verify answers.

## Architecture

### Technology Stack

- **React 19.2.0**: UI library for building interactive components
- **Vite 7.2.4**: Build tool and development server
- **CSS3**: Styling with modern features (gradients, flexbox, responsive design)
- **Fetch API**: For HTTP requests to the backend
- **localStorage**: Browser storage for persisting solved puzzles

### Project Structure

```
dashboard/
├── src/
│   ├── App.jsx          # Main application component
│   ├── App.css          # Application styles
│   ├── main.jsx         # Application entry point
│   └── index.css        # Global styles
├── public/              # Static assets
├── index.html           # HTML template
├── package.json         # Dependencies and scripts
└── vite.config.js       # Vite configuration
```

## Component Architecture

### Main Component: `App.jsx`

The application is built as a single-page application with one main component (`App`) that manages all the puzzle logic.

#### State Management

The component uses React hooks to manage state:

1. **`puzzles`** (Array): List of puzzles fetched from the backend
   - Initial: `[]`
   - Updated: When puzzles are fetched from API

2. **`currentPuzzleIndex`** (Number): Index of the currently displayed puzzle
   - Initial: `0`
   - Updated: When user navigates between puzzles

3. **`userAnswer`** (String): User's current answer input
   - Initial: `''`
   - Updated: As user types in the input field

4. **`isLoading`** (Boolean): Loading state while fetching puzzles
   - Initial: `true`
   - Updated: Before/after API calls

5. **`error`** (String | null): Error message if API call fails
   - Initial: `null`
   - Updated: When API calls fail

6. **`solvedPuzzles`** (Set): Set of solved puzzle indices
   - Initial: `new Set()`
   - Updated: When user solves a puzzle or loads from localStorage

## Data Flow

### 1. Initial Load

```
Component Mounts
    ↓
useEffect (load from localStorage)
    ↓
useEffect (fetch puzzles from API)
    ↓
Update puzzles state
    ↓
Render puzzles
```

### 2. Puzzle Solving Flow

```
User types answer
    ↓
onChange updates userAnswer state
    ↓
User submits form
    ↓
POST /api/puzzles/verify
    ↓
Backend verifies answer
    ↓
If correct:
    - Save to localStorage
    - Update solvedPuzzles state
    - Move to next puzzle (if available)
```

### 3. Progress Persistence

```
User solves puzzle
    ↓
saveSolvedPuzzle() function
    ↓
Convert Set to Array
    ↓
Store in localStorage with date key
    ↓
Format: rebus_solved_YYYY-MM-DD
```

## Key Functions

### `getTodayDate()`
- Returns current date in `YYYY-MM-DD` format
- Used for API requests and localStorage keys

### `isPuzzleSolved(index)`
- Checks if a puzzle at given index is solved
- Uses `solvedPuzzles` Set for O(1) lookup

### `isPuzzleUnlocked(index)`
- Determines if a puzzle can be accessed
- First puzzle (index 0) is always unlocked
- Other puzzles require previous puzzle to be solved

### `saveSolvedPuzzle(index)`
- Saves solved puzzle to localStorage
- Updates `solvedPuzzles` state
- Uses date-based key for daily puzzle management

### `handleSubmit(e)`
- Handles form submission
- Sends answer to backend for verification
- Updates UI based on response

### `goToPuzzle(index)`
- Navigates to a specific puzzle
- Only works if puzzle is unlocked
- Resets answer input

## UI Components

### 1. Header
- Displays app title and current date
- Fixed at top of page

### 2. Puzzle Selector
- Grid of puzzle buttons (1-5)
- Visual states:
  - **Active**: Currently viewing (blue background)
  - **Solved**: Completed (green background with checkmark)
  - **Locked**: Not yet unlocked (gray, disabled)
  - **Default**: Unlocked but not solved (white with blue border)

### 3. Puzzle Container
- Displays current puzzle
- Shows puzzle number and solved status
- Contains puzzle image
- Answer input form (if not solved)
- Solved message with answer reveal (if solved)

### 4. Answer Form
- Text input for user answer
- Submit button
- Only shown for unsolved puzzles

## Styling Approach

### Design System

- **Colors**:
  - Primary: `#667eea` (purple-blue)
  - Secondary: `#764ba2` (purple)
  - Success: `#10b981` (green)
  - Background: Gradient from `#667eea` to `#764ba2`

- **Typography**: System fonts for performance
- **Spacing**: Consistent padding and margins
- **Border Radius**: 8px-12px for modern look
- **Shadows**: Subtle shadows for depth

### Responsive Design

- Mobile-first approach
- Flexbox for layout
- Media queries for smaller screens
- Touch-friendly button sizes (50x50px minimum)

## API Integration

### Base URL
The API base URL is configured via the `VITE_API_BASE_URL` environment variable. 
For local development, it defaults to `http://localhost:8080`.

```javascript
// In App.jsx
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
```

To configure for deployment, create a `.env` file (or set environment variables):
```bash
VITE_API_BASE_URL=https://api.yourdomain.com
```

### Endpoints Used

1. **GET `/api/puzzles/{date}`**
   - Fetches puzzles for a specific date
   - Called on component mount
   - Returns array of puzzle objects

2. **POST `/api/puzzles/verify`**
   - Verifies user's answer
   - Called on form submission
   - Returns `{correct: boolean}`

### Error Handling

- Network errors caught and displayed to user
- Failed API calls show error message
- Image loading errors show placeholder

## Browser Storage Strategy

### localStorage Structure

**Key Format**: `rebus_solved_{YYYY-MM-DD}`

**Value Format**: JSON array of booleans
```json
[true, true, false, false, false]
```
- Index corresponds to puzzle index
- `true` = solved, `false` = not solved

### Storage Lifecycle

1. **On Solve**: Puzzle index marked as `true` in array
2. **On Load**: Array loaded and converted to Set
3. **Daily Reset**: New date = new key = fresh puzzles

### Benefits

- No server-side authentication needed
- Works offline (after initial load)
- Fast access (local storage)
- Automatic cleanup (date-based keys)

## State Synchronization

### localStorage ↔ React State

- **Read**: On component mount, load from localStorage
- **Write**: On puzzle solve, update both state and localStorage
- **Sync**: State is source of truth, localStorage is persistence layer

## Performance Considerations

1. **Lazy Loading**: Puzzles fetched only when needed
2. **Set for Lookups**: O(1) solved puzzle checks
3. **Conditional Rendering**: Only render active puzzle
4. **Image Optimization**: Error handling for failed image loads
5. **Minimal Re-renders**: State updates only when necessary

## Future Enhancements

### Potential Improvements

1. **Component Splitting**: Break App.jsx into smaller components
   - `PuzzleSelector` component
   - `PuzzleDisplay` component
   - `AnswerForm` component

2. **State Management**: Consider Context API or Redux for complex state
3. **Caching**: Cache puzzle images for offline access
4. **Animations**: Add transitions for puzzle changes
5. **Accessibility**: Add ARIA labels and keyboard navigation
6. **Error Boundaries**: Catch and handle React errors gracefully
7. **Loading States**: Skeleton screens while loading
8. **PWA Support**: Make it a Progressive Web App

## Testing Considerations

### What to Test

1. Puzzle fetching and display
2. Answer submission and verification
3. Sequential unlocking logic
4. localStorage persistence
5. Date-based puzzle management
6. Error handling
7. Responsive design

### Testing Tools

- React Testing Library
- Jest
- Cypress (for E2E)

## Browser Compatibility

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Requires localStorage support
- Requires Fetch API support
- ES6+ features used

