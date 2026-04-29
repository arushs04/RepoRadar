# SupplyGraph

SupplyGraph is a backend project for scanning software assets, normalizing package inventory, and eventually exposing dependency, vulnerability, and risk analysis through structured APIs and MCP tools.

## Current Status

The project currently supports:

- parsing Syft JSON output from a saved file
- filtering and normalizing package-like artifacts
- persisting assets and scans into PostgreSQL
- defining the first database schema for assets, scans, components, and component versions

The project does not yet support:

- component and component version persistence during ingestion
- dependency graph persistence
- vulnerability enrichment with OSV or Trivy
- GraphQL APIs
- MCP tools

## Tech Stack

- Go
- PostgreSQL
- Docker Compose
- Syft

## Local Development

### Start PostgreSQL

```bash
docker compose up -d
```

### Apply the initial schema

```bash
docker exec -i supplygraph-postgres psql -U supplygraph -d supplygraph < migrations/001_init.sql
```

### Set the database connection string

If your Docker PostgreSQL instance is exposed on port `5433`:

```bash
export DATABASE_URL="postgres://supplygraph:supplygraph@localhost:5433/supplygraph?sslmode=disable"
```

If it is exposed on port `5432`, update the port in the URL accordingly.

### Run ingestion against a saved Syft JSON file

```bash
go run ./cmd/ingest /path/to/deps.json /path/to/scanned/asset
```

Example:

```bash
go run ./cmd/ingest /Users/arushsacheti/Downloads/argo-cd-master/deps.json /Users/arushsacheti/Downloads/argo-cd-master
```

## Project Layout

```text
cmd/ingest/          CLI entrypoint for first ingestion workflow
internal/db/         PostgreSQL connection and persistence helpers
internal/model/      Normalized application/domain models
internal/syft/       Syft JSON parsing and normalization logic
migrations/          Database schema SQL
docker-compose.yml   Local PostgreSQL development environment
```
