# Ishni boshlash

Ushbu qo‘llanma PGO MCP’ni tayyor Docker image, Docker Compose yoki source kod orqali ishga tushirishni tushuntiradi.

## Talablar

Docker usuli uchun:

- Docker Engine yoki Docker Desktop;
- PostgreSQL serveri;
- minimal privilege’ga ega PostgreSQL login roli;
- kamida 32 belgili tasodifiy `MCP_AUTH_TOKEN`.

Source kod usuli uchun qo‘shimcha ravishda Go `1.25.7` talab qilinadi. Workload va top Query vositalari uchun PostgreSQL’da `pg_stat_statements` extension’i yoqilgan bo‘lishi kerak.

## Docker image bilan ishga tushirish

Bu usul repository’ni clone qilishni talab qilmaydi.

```bash
docker pull ghcr.io/cgouz/pgo-mcp:0.1.0
export MCP_AUTH_TOKEN="$(openssl rand -hex 32)"
```

PostgreSQL macOS host’ida ishlayotgan bo‘lsa:

```bash
docker run --rm \
  --name pgo-mcp \
  --read-only \
  --cap-drop ALL \
  --security-opt no-new-privileges:true \
  -p 127.0.0.1:8001:8001 \
  -e DATABASE_URI='postgresql://pgo_mcp_reader:CHANGE_ME@host.docker.internal:5432/procurement?sslmode=disable' \
  -e MCP_HOST=0.0.0.0 \
  -e MCP_PORT=8001 \
  -e MCP_AUTH_TOKEN \
  ghcr.io/cgouz/pgo-mcp:0.1.0
```

`127.0.0.1` container ichida aynan o‘sha containerni bildiradi. U host’dagi yoki boshqa container’dagi PostgreSQL’ni bildirmaydi. Docker Desktop’da host uchun `host.docker.internal`, boshqa container uchun esa umumiy user-defined network’dagi container DNS nomi ishlatiladi.

## Docker Compose bilan ishga tushirish

Bo‘sh katalogda quyidagi `docker-compose.yml` faylini yarating:

```yaml
services:
  pgo-mcp:
    image: ghcr.io/cgouz/pgo-mcp:0.1.0
    restart: unless-stopped
    env_file:
      - .docker.env
    environment:
      MCP_HOST: 0.0.0.0
      MCP_PORT: 8001
    ports:
      - "127.0.0.1:8001:8001"
    networks:
      - pgo-mcp
    read_only: true
    cap_drop:
      - ALL
    security_opt:
      - no-new-privileges:true

networks:
  pgo-mcp:
    external: true
```

Secret’larni commit qilinmaydigan `.docker.env` fayliga yozing:

```dotenv
DATABASE_URI=postgresql://pgo_mcp_reader:CHANGE_ME@postgres:5432/procurement?sslmode=disable
DB_MAX_CONNS=5
DB_MIN_CONNS=0
MCP_AUTH_TOKEN=REPLACE_WITH_A_RANDOM_64_CHARACTER_HEX_TOKEN
QUERY_TIMEOUT=10s
QUERY_MAX_ROWS=100
QUERY_MAX_BYTES=1048576
```

Mavjud `postgres` containerini umumiy network’ga ulash:

```bash
docker network create pgo-mcp
docker network connect pgo-mcp postgres
docker compose up -d
docker compose ps
docker compose logs --tail=100 pgo-mcp
```

`docker network connect` PostgreSQL containerini restart qilmaydi. Agar network allaqachon mavjud yoki container avval ulangan bo‘lsa, tegishli yaratish/ulash buyrug‘ini takrorlamang.

## Source koddan ishga tushirish

```bash
git clone https://github.com/cgouz/pgo-mcp.git
cd pgo-mcp
cp resources/env/example.env .env.local
```

`.env.local` ichidagi qiymatlarni almashtiring, so‘ng shell’ga yuklang:

```bash
set -a
. ./.env.local
set +a
go run ./src
```

Server `SIGINT` yoki `SIGTERM` olganda HTTP serverni 10 soniyagacha kutib, graceful shutdown bajaradi va Connection Pool’ni yopadi.

## Health endpoint

```bash
curl --fail http://127.0.0.1:8001/health
```

Muvaffaqiyatli javob `HTTP 200` va `ok` matnidir. Bu endpoint authentication talab qilmaydi va faqat process HTTP so‘rovlariga javob berayotganini bildiradi; u PostgreSQL uchun chuqur health audit emas.

## MCP Inspector bilan tekshirish

```bash
npx @modelcontextprotocol/inspector --cli \
  http://127.0.0.1:8001/mcp \
  --transport http \
  --method tools/list \
  --header "Authorization: Bearer ${MCP_AUTH_TOKEN}" \
  --strict
```

Natijada to‘qqizta vosita ko‘rinishi kerak.

## Integration test

Testlar `pgo_mcp_test` nomli alohida baza, `pg_stat_statements`, `public.procurements` fixture’i va `pgo_mcp_reader` rolini kutadi. CI tayyor fixture sifatida `tests/integration/fixture.sql` faylidan foydalanadi. Production bazasida bu fixture’ni ishlatmang.

Test serveri port `8002` da ishlayotganda:

```bash
export MCP_TEST_ENDPOINT=http://127.0.0.1:8002/mcp
export MCP_AUTH_TOKEN=REPLACE_WITH_TEST_TOKEN
go test ./... -count=1
```

## Ko‘p uchraydigan xatolar

### `DATABASE_URI is required`

`DATABASE_URI` process yoki container muhitiga uzatilmagan. Environment fayli avtomatik o‘qilmaydi; uni shell orqali yuklash yoki Compose `env_file` bilan ko‘rsatish kerak.

### `connect: connection refused`

PostgreSQL host nomi, porti va Docker network’ni tekshiring. Container ichida `127.0.0.1` boshqa containerni anglatmaydi.

### `MCP_AUTH_TOKEN must be at least 32 characters`

Kamida 32 belgili token bering. Tavsiya etilgan generatsiya: `openssl rand -hex 32`.

### `401 Unauthorized`

Faqat bitta `Authorization` header yuborilganini va qiymat `Bearer <token>` formatida ekanini tekshiring.

### `pg_stat_statements extension is not installed`

PostgreSQL startup konfiguratsiyasida `shared_preload_libraries='pg_stat_statements'` ni yoqing, serverni restart qiling va kerakli bazada `CREATE EXTENSION pg_stat_statements;` bajaring.

### `object not found or access denied`

Obyekt nomi bilan birga `USAGE`, `SELECT` yoki sequence huquqlarini tekshiring. Server faqat ulangan rol ko‘ra oladigan obyektlarni qaytaradi.
