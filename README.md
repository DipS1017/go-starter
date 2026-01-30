# go-starter API

This repository contains the API for the go-starter application.

## [Folder and Component  Architecture](./docs/architecture.md)

## [Diagrams](./docs/README.md)

## Getting Started

### create .env

```bash
cp .env.example .env
```

### Run postgres server

```bash
docker compose up -d

```

### Run the application

```bash
go run main.go
```

### Live reload the application:

```bash
air
```

## Production Notes

### Health checks

- `GET /api/v1/public/healthz` returns `200` when DB/Redis are healthy and `503` when critical deps are down.

### Configuration hygiene

- Ensure `JWT_SECRET_KEY` and `API_KEY` are set in production.
- Tighten `ALLOWED_ORIGINS` in production; avoid wildcards.
