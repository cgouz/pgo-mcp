# pgo-mcp
pgo-mcp
![alt text](image.png)

# checking
```bash
npx @modelcontextprotocol/inspector --cli \
  http://127.0.0.1:8001/mcp \
  --transport http \
  --method tools/list
```

# tool calling
## calling to list schemas
```bash
npx @modelcontextprotocol/inspector --cli \
  http://127.0.0.1:8001/mcp \
  --transport http \
  --method tools/call \
  --tool-name list_schemas
```

## calling to list_objects
``` bash
npx @modelcontextprotocol/inspector --cli \
  http://127.0.0.1:8001/mcp \
  --transport http \
  --method tools/call \
  --tool-name list_objects \
  --tool-arg schema_name=public \
  --tool-arg object_type=table
```

**tool args list:**
  - table
  - view
  - sequence
  - extension

unsupported object type error test:
```bash
  npx @modelcontextprotocol/inspector --cli \
  http://127.0.0.1:8001/mcp \
  --transport http \
  --method tools/call \
  --tool-name list_objects \
  --tool-arg schema_name=public \
  --tool-arg object_type=invalid
```

## calling to get_object_details
```bash
npx @modelcontextprotocol/inspector --cli \
  http://127.0.0.1:8001/mcp \
  --transport http \
  --method tools/call \
  --tool-name get_object_details \
  --tool-arg schema_name=public \
  --tool-arg object_name=procurements \
  --tool-arg object_type=table
```

# Docker

The Compose service joins an external Docker network named `pgo-mcp` by default. The existing PostgreSQL container currently uses Docker's default `bridge` network. Create a user-defined network and attach PostgreSQL once, without restarting it:

```bash
docker network create pgo-mcp
docker network connect pgo-mcp postgres
```

Copy the environment example and replace the placeholders with local secrets. Keep this file uncommitted:

```bash
cp resources/env/example.env .env.local
```

`DATABASE_URI` should use `postgres` as the host when both containers share the network. `127.0.0.1` on macOS addresses the host, while `127.0.0.1` inside a container addresses that container itself. `MCP_HOST=0.0.0.0` makes the server listen on all container interfaces; Compose still publishes it only on the host's `127.0.0.1` by default.

Build and start PGO MCP:

```bash
PGO_MCP_ENV_FILE=.env.local docker compose build
PGO_MCP_ENV_FILE=.env.local docker compose up -d
docker compose ps
docker compose logs --tail=100 pgo-mcp
curl --fail http://127.0.0.1:8001/health
```

List MCP tools with authentication:

```bash
set -a
. ./.env.local
set +a

npx @modelcontextprotocol/inspector --cli \
  http://127.0.0.1:8001/mcp \
  --transport http \
  --method tools/list \
  --header "Authorization: Bearer ${MCP_AUTH_TOKEN}" \
  --strict
```

Stop and remove only the PGO MCP application container:

```bash
PGO_MCP_ENV_FILE=.env.local docker compose down
```

Set `PGO_MCP_NETWORK` if a different user-defined network is already attached to PostgreSQL. Set `MCP_PUBLISH_PORT` to change the macOS port without changing the container's internal port.

# Continuous integration

CI runs for pushes to `main`, pull requests targeting `main`, and manual workflow dispatches. It verifies modules, formatting, vetting, package tests, compilation, the PostgreSQL-backed integration suite, and a single-platform Docker build.

The integration job starts PostgreSQL 17 with `pg_stat_statements`, creates an ephemeral read-only role and fixture database objects, starts PGO MCP on port 8002, and then runs the complete test suite. No local or production environment files are used.

Run the independent checks locally with:

```bash
go mod verify
gofmt -l .
go vet ./...
go test ./... -count=1
go build -o /tmp/pgo-mcp ./src
docker compose config -q
docker build -t pgo-mcp:ci-test .
```

# Releases

To publish a release, create and push a valid semantic-version tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow validates the tag, builds `linux/amd64` and `linux/arm64` images, publishes semantic-version tags to `ghcr.io/cgouz/pgo-mcp`, and creates a GitHub Release containing the immutable image digest. Stable releases receive `latest`; prereleases such as `v0.2.0-rc.1` do not.

After a release has been published, pull a specific version with:

```bash
docker pull ghcr.io/cgouz/pgo-mcp:0.1.0
```

For Compose, set `image: ghcr.io/cgouz/pgo-mcp:0.1.0` on the `pgo-mcp` service and remove its `build` section. Continue using the same environment file, external PostgreSQL network, port mapping, healthcheck, and security settings.
