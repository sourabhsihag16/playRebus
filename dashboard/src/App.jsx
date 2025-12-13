import { useState, useEffect } from 'react'
import './App.css'

// API base URL - adjust this to match your backend
const API_BASE_URL = 'http://localhost:8080'

function App() {
  const [puzzles, setPuzzles] = useState([])
  const [currentPuzzleIndex, setCurrentPuzzleIndex] = useState(0)
  const [userAnswer, setUserAnswer] = useState('')
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState(null)
  const [solvedPuzzles, setSolvedPuzzles] = useState([]) // Array of booleans: [true, false, true, ...]

  // Get today's date in YYYY-MM-DD format
  const getTodayDate = () => {
    const today = new Date()
    return today.toISOString().split('T')[0]
  }

  // Load solved puzzles from browser cache (localStorage)
  useEffect(() => {
    const today = getTodayDate()
    const savedData = localStorage.getItem(`rebus_solved_${today}`)
    if (savedData) {
      try {
        const solved = JSON.parse(savedData)
        // Ensure it's an array
        if (Array.isArray(solved)) {
          setSolvedPuzzles(solved)
          // Find the first unsolved puzzle
          const firstUnsolved = solved.findIndex((s, index) => !s)
          if (firstUnsolved !== -1) {
            setCurrentPuzzleIndex(firstUnsolved)
          } else if (solved.length > 0 && solved.every(s => s)) {
            // All puzzles solved
            setCurrentPuzzleIndex(solved.length - 1)
          }
        }
      } catch (e) {
        console.error('Error loading saved puzzles:', e)
      }
    }
  }, [])

  // Fetch puzzles from backend
  useEffect(() => {
    const fetchPuzzles = async () => {
      setIsLoading(true)
      setError(null)
      try {
        const today = getTodayDate()
        const response = await fetch(`${API_BASE_URL}/api/puzzles/${today}`)
        
        if (!response.ok) {
          throw new Error('Failed to fetch puzzles')
        }
        
        const data = await response.json()
        // Ensure image URLs are absolute (prepend API_BASE_URL if relative)
        const puzzlesWithAbsoluteUrls = (data.puzzles || []).map(puzzle => ({
          ...puzzle,
          imageUrl: puzzle.imageUrl.startsWith('http') 
            ? puzzle.imageUrl 
            : `${API_BASE_URL}${puzzle.imageUrl}`
        }))
        setPuzzles(puzzlesWithAbsoluteUrls)
      } catch (err) {
        setError(err.message)
        console.error('Error fetching puzzles:', err)
      } finally {
        setIsLoading(false)
      }
    }

    fetchPuzzles()
  }, [])

  // Check if a puzzle is solved
  const isPuzzleSolved = (index) => {
    return solvedPuzzles[index] === true
  }

  // Check if a puzzle is unlocked (previous puzzle is solved or it's the first puzzle)
  const isPuzzleUnlocked = (index) => {
    if (index === 0) return true
    return isPuzzleSolved(index - 1)
  }

  // Save solved puzzle to localStorage
  const saveSolvedPuzzle = (index) => {
    const today = getTodayDate()
    // Create a copy of the solved puzzles array
    const solvedArray = [...solvedPuzzles]
    
    // Ensure array has correct length
    while (solvedArray.length <= index) {
      solvedArray.push(false)
    }
    solvedArray[index] = true
    
    localStorage.setItem(`rebus_solved_${today}`, JSON.stringify(solvedArray))
    setSolvedPuzzles(solvedArray)
  }

  // Handle answer submission
  const handleSubmit = async (e) => {
    e.preventDefault()
    
    if (!userAnswer.trim()) {
      return
    }

    const currentPuzzle = puzzles[currentPuzzleIndex]
    if (!currentPuzzle) return

    try {
      const response = await fetch(`${API_BASE_URL}/api/puzzles/verify`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          puzzleId: currentPuzzle.id,
          answer: userAnswer.trim().toLowerCase(),
        }),
      })

      const data = await response.json()
      
      if (data.correct) {
        // Save the solved puzzle
        saveSolvedPuzzle(currentPuzzleIndex)
        
        // Clear the answer input
        setUserAnswer('')
        
        // Move to next puzzle if available
        if (currentPuzzleIndex < puzzles.length - 1) {
          setCurrentPuzzleIndex(currentPuzzleIndex + 1)
        }
      } else {
        // Clear the answer input for retry
        setUserAnswer('')
      }
    } catch (err) {
      console.error('Error verifying answer:', err)
    }
  }

  // Navigate to a puzzle
  const goToPuzzle = (index) => {
    if (isPuzzleUnlocked(index)) {
      setCurrentPuzzleIndex(index)
      setUserAnswer('')
    }
  }

  if (isLoading) {
    return (
      <div className="app">
        <div className="loading">Loading puzzles...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="app">
        <div className="error">Error: {error}</div>
      </div>
    )
  }

  const currentPuzzle = puzzles[currentPuzzleIndex]

  return (
    <div className="app">
      <header className="header">
        <button className="menu-btn" aria-label="Menu">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="3" y1="6" x2="21" y2="6"></line>
            <line x1="3" y1="12" x2="21" y2="12"></line>
            <line x1="3" y1="18" x2="21" y2="18"></line>
          </svg>
        </button>
        <h1 className="header-title">Daily Rebus Puzzles</h1>
        <div className="header-icons">
          <button className="icon-btn" aria-label="Hints">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="currentColor">
              <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 17h-2v-2h2v2zm2.07-7.75l-.9.92C13.45 12.9 13 13.5 13 15h-2v-.5c0-1.1.45-2.1 1.17-2.83l1.24-1.26c.37-.36.59-.86.59-1.41 0-1.1-.9-2-2-2s-2 .9-2 2H8c0-2.21 1.79-4 4-4s4 1.79 4 4c0 .88-.36 1.68-.93 2.25z"/>
            </svg>
          </button>
          <button className="icon-btn" aria-label="Statistics">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="20" x2="18" y2="10"></line>
              <line x1="12" y1="20" x2="12" y2="4"></line>
              <line x1="6" y1="20" x2="6" y2="14"></line>
            </svg>
          </button>
          <button className="icon-btn" aria-label="Help">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="10"></circle>
              <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path>
              <line x1="12" y1="17" x2="12.01" y2="17"></line>
            </svg>
          </button>
          <button className="icon-btn" aria-label="Settings">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="3"></circle>
              <path d="M12 1v6m0 6v6M5.64 5.64l4.24 4.24m4.24 4.24l4.24 4.24M1 12h6m6 0h6M5.64 18.36l4.24-4.24m4.24-4.24l4.24-4.24"></path>
            </svg>
          </button>
          <button className="subscribe-btn">Subscribe to Games</button>
        </div>
      </header>

      <div className="puzzle-selector">
        <h2>Puzzles</h2>
        <div className="puzzle-buttons">
          {puzzles.map((puzzle, index) => (
            <button
              key={puzzle.id}
              className={`puzzle-btn ${
                isPuzzleSolved(index) ? 'solved' : ''
              } ${!isPuzzleUnlocked(index) ? 'locked' : ''} ${
                index === currentPuzzleIndex ? 'active' : ''
              }`}
              onClick={() => goToPuzzle(index)}
              disabled={!isPuzzleUnlocked(index)}
            >
              {isPuzzleSolved(index) ? '✓' : index + 1}
            </button>
          ))}
        </div>
      </div>

      {currentPuzzle ? (
        <div className="puzzle-container">
          <div className="puzzle-info">
            <h2>Puzzle {currentPuzzleIndex + 1} of {puzzles.length}</h2>
            {isPuzzleSolved(currentPuzzleIndex) && (
              <span className="solved-badge">✓ Solved</span>
            )}
          </div>

          <div className="puzzle-image-container">
            <img
              src={currentPuzzle.imageUrl}
              alt={`Rebus puzzle ${currentPuzzleIndex + 1}`}
              className="puzzle-image"
              onError={(e) => {
                e.target.src = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="400" height="300"%3E%3Ctext x="50%25" y="50%25" text-anchor="middle" dy=".3em"%3ELoading puzzle...%3C/text%3E%3C/svg%3E'
              }}
            />
          </div>

          {!isPuzzleSolved(currentPuzzleIndex) ? (
            <form onSubmit={handleSubmit} className="answer-form">
              <input
                type="text"
                value={userAnswer}
                onChange={(e) => setUserAnswer(e.target.value)}
                placeholder="Enter your answer..."
                className="answer-input"
                autoFocus
              />
              <button type="submit" className="submit-btn">
                Submit Answer
              </button>
            </form>
          ) : (
            <div className="solved-message">
              <p>🎉 You solved this puzzle!</p>
              <p className="answer-reveal">Answer: {currentPuzzle.answer}</p>
            </div>
          )}
        </div>
      ) : (
        <div className="no-puzzles">
          <p>No puzzles available for today. Check back tomorrow!</p>
        </div>
      )}
    </div>
  )
}

export default App
