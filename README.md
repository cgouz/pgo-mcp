# PGO MCP

[![CI](https://github.com/cgouz/pgo-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/cgouz/pgo-mcp/actions/workflows/ci.yml)
[![GitHub release](https://img.shields.io/github/v/release/cgouz/pgo-mcp)](https://github.com/cgouz/pgo-mcp/releases)
[![Go](https://img.shields.io/badge/Go-1.25.7-00ADD8?logo=go)](https://go.dev/)

PGO MCP — Go dasturlash tilida yozilgan PostgreSQL uchun Model Context Protocol (MCP) serveri. U ma’lumotlar bazasi tuzilmasini ko‘rish, SQL so‘rovlarini faqat o‘qish tranzaksiyasida bajarish, bajarilish rejasini va indekslardan foydalanishni tahlil qilish hamda ma’lumotlar bazasi holatini kuzatish imkonini beradi.

Server MCP Streamable HTTP transportidan foydalanadi. `/mcp` endpoint’i Bearer authentication bilan himoyalangan, `/health` endpoint’i esa ochiq readiness tekshiruvi uchun mo‘ljallangan.

## Imkoniyatlar

- PostgreSQL schema, table, view, sequence va extension obyektlarini ko‘rish.
- SQL natijalarini satr va javob hajmi cheklovlari bilan olish.
- `EXPLAIN (FORMAT JSON)` va ixtiyoriy `ANALYZE` rejimi.
- Alohida Query va umumiy workload bo‘yicha Index tahlili.
- Connection, transaction, cache, deadlock, temporary file va vacuum statistikasi.
- `pg_stat_statements` asosida eng muhim so‘rovlarni ko‘rish.
- Read-only transaction, statement timeout va PostgreSQL privilege tekshiruvlari.
- `linux/amd64` va `linux/arm64` uchun non-root Docker image.

## MCP vositalari

| Vosita | Vazifasi |
| --- | --- |
| `list_schemas` | Mavjud schemalarni qaytaradi. |
| `list_objects` | Schema ichidagi table, view, sequence yoki extension ro‘yxatini beradi. |
| `get_object_details` | Obyekt ustunlari, constraint, index, definition yoki metadata ma’lumotlarini beradi. |
| `execute_sql` | SQL so‘rovini read-only transaction ichida bajaradi. |
| `explain_query` | Query bajarilish rejasini JSON ko‘rinishida qaytaradi. |
| `analyze_query_indexes` | Bitta Query rejasi va unga aloqador indekslarni tahlil qiladi. |
| `analyze_workload_indexes` | `pg_stat_statements` va Index statistikasi asosida workload’ni tahlil qiladi. |
| `analyze_db_health` | Ma’lumotlar bazasi holati va ehtimoliy muammolarni ko‘rsatadi. |
| `get_top_queries` | So‘rovlarni vaqt, chaqiruv yoki qaytargan satrlar bo‘yicha saralaydi. |

To‘liq input va output tavsifi: [MCP vositalari ma’lumotnomasi](docs/tools.md).

## Docker orqali tezkor ishga tushirish

Image’ni yuklab oling:

```bash
docker pull ghcr.io/cgouz/pgo-mcp:0.1.0
```

Kamida 32 belgili token yarating va serverni ishga tushiring:

```bash
export MCP_AUTH_TOKEN="$(openssl rand -hex 32)"

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

macOS va Docker Desktop’da `host.docker.internal` container ichidan host’dagi PostgreSQL’ga murojaat qiladi. PostgreSQL boshqa container’da ishlasa, ikkala containerni user-defined Docker network’ga ulang va URI’da PostgreSQL container DNS nomidan foydalaning.

```bash
curl --fail http://127.0.0.1:8001/health
```

Batafsil yo‘riqnoma: [Ishni boshlash](docs/getting-started.md).

## Konfiguratsiya

Asosiy o‘zgaruvchilar:

```dotenv
DATABASE_URI=postgresql://pgo_mcp_reader:CHANGE_ME@postgres:5432/procurement?sslmode=disable
DB_MAX_CONNS=5
DB_MIN_CONNS=0
MCP_HOST=0.0.0.0
MCP_PORT=8001
MCP_AUTH_TOKEN=REPLACE_WITH_A_RANDOM_64_CHARACTER_HEX_TOKEN
QUERY_TIMEOUT=10s
QUERY_MAX_ROWS=100
QUERY_MAX_BYTES=1048576
```

Aniq default qiymatlar va validatsiya chegaralari: [Konfiguratsiya ma’lumotnomasi](docs/configuration.md).

## MCP client ulanishi

Server manzili: `http://127.0.0.1:8001/mcp`. Har bir MCP so‘roviga quyidagi header yuborilishi kerak:

```http
Authorization: Bearer REPLACE_WITH_YOUR_TOKEN
```

Client’ga oid yo‘riqnomalar:

- [Claude Code](docs/clients/claude.md)
- [Cursor](docs/clients/cursor.md)
- [Visual Studio Code](docs/clients/vscode.md)

## Xavfsizlik

PGO MCP’ni faqat minimal privilege’ga ega PostgreSQL roli bilan ishlating. Bearer token transportni shifrlamaydi; localhost’dan tashqarida reverse proxy orqali TLS/HTTPS ishlating. `SELECT` huquqining o‘zi barcha function yoki tashqi side effect xavflarini yo‘q qilmaydi, shuning uchun roldan xavfli function’larni bajarish huquqini ham olib tashlash zarur.

Batafsil tavsiyalar va read-only rol namunasi: [Xavfsizlik](docs/security.md).

## Dasturlash va test

Talablar: Go `1.25.7`, PostgreSQL 17 va ixtiyoriy ravishda Docker.

```bash
git clone https://github.com/cgouz/pgo-mcp.git
cd pgo-mcp
go mod verify
go vet ./...
go test ./src/... -count=1
go build -o /tmp/pgo-mcp ./src
```

Integration test alohida `pgo_mcp_test` bazasi, fixture va ishlayotgan PGO MCP serverini talab qiladi. CI bu muhitni PostgreSQL 17 service container’ida avtomatik yaratadi. Lokal tartib [Ishni boshlash](docs/getting-started.md#source-koddan-ishga-tushirish) bo‘limida berilgan.

## CI/CD va release

CI `main` branch’ga push, `main` uchun pull request va qo‘lda ishga tushirishda format, module, vet, test, build, PostgreSQL-backed integration hamda Docker build tekshiruvlarini bajaradi.

`v*` SemVer tag push qilinganda release workflow `linux/amd64` va `linux/arm64` image’larini GHCR’ga chiqaradi va GitHub Release yaratadi:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Bu buyruqlar faqat maintainer release yaratmoqchi bo‘lganda bajariladi.

## Hujjatlar

- [Ishni boshlash](docs/getting-started.md)
- [Konfiguratsiya](docs/configuration.md)
- [MCP vositalari](docs/tools.md)
- [Xavfsizlik](docs/security.md)
- [Claude Code](docs/clients/claude.md)
- [Cursor](docs/clients/cursor.md)
- [Visual Studio Code](docs/clients/vscode.md)

## Litsenziya

Repository’da hozircha `LICENSE` fayli mavjud emas. Ochiq repository mavjudligi kodga avtomatik ravishda foydalanish, o‘zgartirish yoki tarqatish huquqini bermaydi. Ochiq manbali release uchun maintainer alohida litsenziya tanlab, `LICENSE` faylini qo‘shishi kerak.
