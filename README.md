# Kickoff Arena

A tournament management web app for eFootball players, featuring two main tools:

- **Spin Wheel** ✅ — subadmins create a tournament, share a registration link, then randomly generate a knockout bracket and track results round by round
- **Player Auction** — group tournament draft where captains bid on players *(not yet built)*

## Features

- Role-based auth (admin / subadmin) with JWT
- Admins can view, manage, and delete subadmins and all tournaments they've created
- Subadmins create Spin Wheel tournaments and share a public registration link (no login required to register)
- Configurable bracket size (2–64 players), with automatic byes for non-power-of-two counts
- Spin wheel animation to randomly seed the bracket, or a manual pick order
- Knockout bracket ("Fixtures") with automatic round progression as results are submitted
- Share registration links and completed fixtures directly to WhatsApp
- Concurrency-safe registration and bracket start (row-level locking prevents double-booking or double-starting a tournament under simultaneous requests)

## Stack

- **Frontend:** Angular (standalone components, signals)
- **Backend:** Go (Gin, layered handler → service → repository architecture)
- **Database:** PostgreSQL (primary store) + Redis (cache layer)

## Project structure

```
backend/    Go API server
  cmd/server/       entrypoint
  internal/         handlers, services, repositories, models, middleware
  migrations/        SQL schema migrations (apply in order)
frontend/   Angular app
docker-compose.yml   Local Postgres + Redis
```

## Local development

1. Copy env files and fill in real values:
   ```bash
   cp .env.example .env
   cp backend/.env.example backend/.env
   ```

2. Start Postgres + Redis:
   ```bash
   docker compose up -d
   ```

3. Apply database migrations (in order) against the running Postgres instance, e.g.:
   ```bash
   for f in backend/migrations/*.up.sql; do
     psql "$DATABASE_URL" -f "$f"
   done
   ```

4. Run the backend:
   ```bash
   cd backend && go run ./cmd/server   # API on :8080
   ```

5. Run the frontend:
   ```bash
   cd frontend && npm install && npx ng serve   # App on :4300
   ```
