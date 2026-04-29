# SupplyGraph

SupplyGraph is a backend project for scanning software assets, normalizing package inventory, enriching package versions with vulnerability data, and eventually exposing dependency, vulnerability, and risk analysis through structured APIs and MCP tools.

## Current Status

The project currently supports:

- parsing Syft JSON output from a saved file
- filtering and normalizing package-like artifacts
- persisting assets and scans into PostgreSQL
- persisting normalized components, component versions, and scan membership
- enriching npm package versions with OSV vulnerability data
- persisting vulnerabilities and findings into PostgreSQL

The project does not yet support:

- dependency graph persistence
- Trivy enrichment
- GraphQL APIs
- MCP tools
- scheduled or background scan processing
- test coverage

## Tech Stack

- Go
- PostgreSQL
- Docker Compose
- Syft
- OSV

## Local Development

### Start PostgreSQL

```bash
docker compose up -d
```

### Apply database migrations

```bash
docker exec -i supplygraph-postgres psql -U supplygraph -d supplygraph < migrations/001_init.sql
docker exec -i supplygraph-postgres psql -U supplygraph -d supplygraph < migrations/002_vulnerabilities.sql
```

### Set the database connection string

If your Docker PostgreSQL instance is exposed on port `5432`:

```bash
export DATABASE_URL="postgres://supplygraph:supplygraph@localhost:5432/supplygraph?sslmode=disable"
```

If you remap the container to a different host port, update the URL accordingly.

### Generate a Syft JSON scan

```bash
syft /path/to/asset -o json > deps.json
```

### Run ingestion against a saved Syft JSON file

```bash
go run ./cmd/ingest /path/to/deps.json /path/to/scanned/asset
```

Example:

```bash
go run ./cmd/ingest /Users/arushsacheti/Downloads/argo-cd-master/deps.json /Users/arushsacheti/Downloads/argo-cd-master
```

### Inspect stored data

```bash
docker exec -it supplygraph-postgres psql -U supplygraph -d supplygraph
```

Example queries:

```sql
SELECT * FROM assets;
SELECT COUNT(*) FROM scans;
SELECT COUNT(*) FROM components;
SELECT COUNT(*) FROM component_versions;
SELECT COUNT(*) FROM vulnerabilities;
SELECT COUNT(*) FROM findings;
```

## Current Data Model

Implemented tables:

- `assets`
- `scans`
- `components`
- `component_versions`
- `scan_component_versions`
- `vulnerabilities`
- `findings`

## Project Layout

```text
cmd/ingest/          CLI entrypoint for ingestion workflow
internal/ingest/     Inventory persistence and OSV enrichment workflows
internal/db/         PostgreSQL connection and persistence helpers
internal/model/      Normalized application/domain models
internal/osv/        OSV client and response types
internal/syft/       Syft JSON parsing and normalization logic
migrations/          Database schema SQL
docker-compose.yml   Local PostgreSQL development environment
```
