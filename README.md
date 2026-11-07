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
