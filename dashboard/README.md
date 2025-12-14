# Dashboard - Rebus Puzzle Game

A React + Vite frontend application for the Daily Rebus Puzzle game.

## Setup

1. **Install dependencies**:
```bash
npm install
```

2. **Configure environment variables**:
   - Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```
   - Edit `.env` and set `VITE_API_BASE_URL` to your backend API URL
   - For local development: `VITE_API_BASE_URL=http://localhost:8080`
   - For production: `VITE_API_BASE_URL=https://api.yourdomain.com`

3. **Run development server**:
```bash
npm run dev
```

4. **Build for production**:
```bash
npm run build
```

## Environment Variables

- `VITE_API_BASE_URL`: The base URL for the backend API. Defaults to `http://localhost:8080` if not set.

**Note**: In Vite, only environment variables prefixed with `VITE_` are exposed to the client-side code. This is a security feature to prevent accidentally exposing sensitive server-side variables.

## Development

This project uses:
- [React](https://react.dev/) for the UI
- [Vite](https://vite.dev/) for build tooling and dev server
- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react) for Fast Refresh

## Expanding the ESLint configuration

If you are developing a production application, we recommend using TypeScript with type-aware lint rules enabled. Check out the [TS template](https://github.com/vitejs/vite/tree/main/packages/create-vite/template-react-ts) for information on how to integrate TypeScript and [`typescript-eslint`](https://typescript-eslint.io) in your project.
