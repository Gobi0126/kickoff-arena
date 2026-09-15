# kickoff-arena

A web app for eFootball players featuring two main tools:

- **Spin Wheel** — register players, spin to randomly generate opponent matchups
- **Player Auction** — group tournament draft where captains bid on players

## Stack

- **Frontend:** Angular
- **Backend:** Go (Gin, layered controller/service/repository architecture)
- **Database:** PostgreSQL (primary store) + Redis (real-time/cache layer)

## Project structure

```
backend/    Go API server
frontend/   Angular app
docker-compose.yml   Local Postgres + Redis
```

## Local development

```bash
docker compose up -d              # Postgres + Redis
cd backend && go run cmd/server/main.go   # API on :8080
cd frontend && npx ng serve               # App on :4200
```

Copy `.env.example` → `.env` (root) and `backend/.env.example` → `backend/.env`, filling in real values before running.
